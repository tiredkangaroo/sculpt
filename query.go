package sculpt

import (
	"fmt"
	"reflect"
	"sculpt/internals/sql"
	"slices"
	"strings"
)

type Query[T any] struct {
	distinct bool
	model    *Model[T]

	ascdesc sql.ASCDESC
	orderby string

	fields     []string
	conditions []Condition
}

// Distinct makes the query return ONLY distinct results.
func (q *Query[T]) Distinct() *Query[T] {
	q.distinct = true
	return q
}

// AscendingOrder sets the order of the results to ascending.
func (q *Query[T]) AscendingOrder() *Query[T] {
	q.ascdesc = sql.ASC
	return q
}

// DescendingOrder sets the order of the results to descending.
func (q *Query[T]) DescendingOrder() *Query[T] {
	q.ascdesc = sql.DESC
	return q
}

// IncludeFields allows manual specification of which fields to populate in the result. If
// not called, or left empty, all fields will be given.
//
// It panics if a field provided does not exist on the model.
func (q *Query[T]) IncludeFields(f ...string) *Query[T] {
	columnNames := make([]string, len(q.model.columns))
	for i, c := range q.model.columns {
		columnNames[i] = c.name
	}
	for _, field := range f {
		if !slices.Contains(columnNames, field) {
			panic(fmt.Sprintf("field %s is not on the model", field))
		}
		q.fields = append(q.fields, field)
	}
	return q
}

// Conditions adds conditions to the query. These conditions are combined, in order to
// specifically get the results that meet the criteria.
func (q *Query[T]) Conditions(c ...Condition) *Query[T] {
	for _, cd := range c {
		q.conditions = append(q.conditions, cd)
	}
	return q
}

// compile makes a SQL statement for pgx from the query.
func (q *Query[T]) compile() (string, []any, error) {
	// start of query
	statement := "SELECT "
	if len(q.fields) == 0 {
		for _, c := range q.model.columns {
			q.fields = append(q.fields, c.name)
		}
	}

	// distinct
	if q.distinct {
		statement += "DISTINCT "
	}

	// continuing start of query
	if len(q.fields) > 0 {
		statement += strings.Join(q.fields, ", ")
	}
	statement += fmt.Sprintf(" FROM %s ", q.model.name)

	// placeholders for pgx
	a := []any{} // pgx query arguments
	j := 0       // uses a counter to replace placeholders for pgx

	// WHERE
	if len(q.conditions) > 0 {
		statement += "WHERE "
	}
	for i, c := range q.conditions {
		statement += replaceAllFunc(c.s, "<_sculpt>", func() string {
			j++
			return fmt.Sprintf("%d", j)
		})
		if i != len(q.conditions)-1 {
			statement += " AND "
		}
		a = append(a, c.a...)
	}

	// ORDER BY
	if q.orderby != "" {
		j++
		statement += fmt.Sprintf(" ORDER BY $%d", j)
		a = append(a, q.orderby)
	}
	if q.ascdesc != sql.NoASCDESC && q.orderby == "" {
		return "", nil, fmt.Errorf("cannot use ascending or descending order without an order field")
	}

	//     - ASC/DESC
	switch q.ascdesc {
	case sql.ASC:
		statement += " ASC"
	case sql.DESC:
		statement += " DESC"
	}

	statement += ";"
	return statement, a, nil
}

// Do compiles and executes the query and returns the results.
func (q *Query[T]) Do() ([]T, error) {
	statement, a, err := q.compile()
	if err != nil {
		return nil, err
	}
	rows, err := sql.Query(statement, a...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	results := []T{}

	for rows.Next() {
		values := make([]any, len(q.model.columns))
		for i, column := range q.model.columns {
			if column.nullable {
				values[i] = reflect.New(optionalType(column.t)).Interface()
			} else {
				values[i] = reflect.New(column.t).Interface()
			}
		}
		err = rows.Scan(values...)
		if err != nil {
			return nil, err
		}
		result := reflect.New(reflect.TypeFor[T]()).Elem()
		for i, c := range q.model.columns {
			var value reflect.Value
			if c.nullable {
				value = reflect.New(c.t)
				value.MethodByName("Set").Call([]reflect.Value{reflect.ValueOf(values[i]).Elem()})
			} else {
				value = reflect.ValueOf(values[i])
			}
			result.FieldByName(c.name).Set(value.Elem())
		}
		r := result.Interface().(T) // literally impossible to fail
		results = append(results, r)
	}
	return results, nil
}
