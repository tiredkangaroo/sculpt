package sculpt

import (
	"fmt"
	"reflect"
)

// FIXME: do not require Condition to be joined
// with and. allow something like or as well.

type Condition = string

type Query struct {
	// Columns specifies which columns to return.
	Columns []string

	// Distinct specifies if the columns should be unique.
	Distinct bool

	// Conditions specifies the conditions for the rows being returned.
	Conditions []Condition
}

func EqualTo(name string, value any) Condition {
	v, err := anyToSQLString(value)
	if err != nil {
		panic(err.Error())
	}
	return fmt.Sprintf(`"%s" = %s`, name, v)
}

func GreaterThan(name string, value any) Condition {
	v, err := anyToSQLString(value)
	if err != nil {
		panic(err.Error())
	}
	return fmt.Sprintf(`"%s" > %s`, name, v)
}

func LessThan(name string, value any) Condition {
	v, err := anyToSQLString(value)
	if err != nil {
		panic(err.Error())
	}
	return fmt.Sprintf(`"%s" < %s`, name, v)
}

func GreaterEqualOrEqualTo(name string, value any) Condition {
	v, err := anyToSQLString(value)
	if err != nil {
		panic(err.Error())
	}
	return fmt.Sprintf(`"%s" >= %s`, name, v)
}

func LessThanOrEqualTo(name string, value any) Condition {
	v, err := anyToSQLString(value)
	if err != nil {
		panic(err.Error())
	}
	return fmt.Sprintf(`"%s" <= %s`, name, v)
}

func NotEqualTo(name string, value any) Condition {
	v, err := anyToSQLString(value)
	if err != nil {
		panic(err.Error())
	}
	return fmt.Sprintf(`"%s" <> %s`, name, v)
}

func Between(name string, range1 any, range2 any) Condition {
	v, err := anyToSQLString(range1)
	if err != nil {
		panic(err.Error())
	}
	v2, err := anyToSQLString(range2)
	if err != nil {
		panic(err.Error())
	}
	return fmt.Sprintf(`"%s" BETWEEN %s AND %s`, name, v, v2)
}

func Like(name string, value string) Condition {
	return fmt.Sprintf(`"%s" LIKE %s`, name, value)
}

func In(name string, value ...any) Condition {
	statement := fmt.Sprintf(`"%s" IN (`, name)
	for i, val := range value {
		v, err := anyToSQLString(val)
		if err != nil {
			panic(err.Error())
		}
		statement += v
		if i+1 < len(value) {
			statement += `, `
		}
	}
	statement += `)`
	return statement
}

// RunQuery runs the query specified on the model.
func RunQuery[I any](m *Model, query Query) ([]I, error) {
	s := m.raw
	sv := reflect.ValueOf(s)
	st := reflect.TypeOf(s).Elem()

	m, ok := ModelRegistry[st.Name()]
	if !ok {
		panic(`cannot run query on an unregistered model`)
	}

	rawt := reflect.TypeOf(m.raw)
	if rawt != sv.Type() {
		return []I{}, ModelTypeMismatch(rawt.Name(), sv.Type().Name())
	}

	sv = sv.Elem()

	statement := `SELECT `
	if query.Distinct {
		statement += `DISTINCT `
	}
	if len(query.Columns) == 0 {
		statement += `*`
	}
	for i, c := range query.Columns {
		statement += c
		if i+1 < len(query.Columns) {
			statement += `, `
		}
	}
	statement += ` FROM ` + `"` + m.Name + `"`
	if len(query.Conditions) != 0 {
		statement += ` WHERE `
		for i, c := range query.Conditions {
			statement += c
			if i+1 < len(query.Conditions) {
				statement += ` AND `
			}
		}
	}
	statement += `;`
	rows, err := ActiveDB.Query(statement)
	if err != nil {
		return []I{}, err
	}
	defer rows.Close()
	schemas := []I{}
	for rows.Next() { // if there is no next it returns false and loop closes
		newSchema := reflect.New(sv.Type())
		nse := newSchema.Elem()
		ptrs := []interface{}{}
		for i := range nse.NumField() {
			field := nse.Field(i)
			fieldt := st.Field(i)
			if len(query.Columns) == 0 { //SELECT *
				ptrs = append(ptrs, field.Addr().Interface())
				continue
			}
			for _, c := range query.Columns {
				if fieldt.Name == c {
					ptrs = append(ptrs, field.Addr().Interface())
				}
			}
		}
		err := rows.Scan(ptrs...)
		if err != nil {
			LogError(`an error occured during scanning rows to schema: %s`, err.Error())
			continue
		}
		nsi, ok := newSchema.Interface().(I)
		if !ok {
			LogError(`something went wrong`)
			continue
		}
		schemas = append(schemas, nsi)
	}
	return schemas, nil
}
