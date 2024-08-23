package sculpt

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"unsafe"
)

// FIXME: do not require Condition to be joined
// with and. allow something like or as well.

type Condition struct {
	S string
	A []any
}

type Query struct {
	// Columns specifies which columns to return.
	Columns []string

	// Distinct specifies if the columns should be unique.
	Distinct bool

	// Conditions specifies the conditions for the rows being returned.
	Conditions []Condition
}

func EqualTo(name string, value any) Condition {
	return Condition{S: fmt.Sprintf(`"%s" = *!`, name), A: []any{value}}
}

func GreaterThan(name string, value any) Condition {
	return Condition{S: fmt.Sprintf(`"%s" > *!`, name), A: []any{value}}
}

func LessThan(name string, value any) Condition {
	return Condition{S: fmt.Sprintf(`"%s" < *!`, name), A: []any{value}}
}

func GreaterEqualOrEqualTo(name string, value any) Condition {
	return Condition{S: fmt.Sprintf(`"%s" >= *!`, name), A: []any{value}}
}

func LessThanOrEqualTo(name string, value any) Condition {
	return Condition{S: fmt.Sprintf(`"%s" <= *!`, name), A: []any{value}}
}

func NotEqualTo(name string, value any) Condition {
	return Condition{S: fmt.Sprintf(`"%s" <> *!`, name), A: []any{value}}
}

func Between(name string, range1 any, range2 any) Condition {
	return Condition{S: fmt.Sprintf(`"%s" BETWEEN *! AND *!`, name), A: []any{range1, range2}}
}

func Like(name string, value string) Condition {
	return Condition{S: fmt.Sprintf(`"%s" LIKE *!`, name), A: []any{value}}
}

func In(name string, value ...any) Condition {
	placeholders := make([]string, len(value))
	for i := range value {
		placeholders[i] = fmt.Sprintf("*!", i+1)
	}
	statement := fmt.Sprintf(`"%s" IN (%s)`, name, strings.Join(placeholders, ", "))
	return Condition{S: statement, A: value}
}

func valuesToSchema(columns []*Column, values []any, schema reflect.Type) (*reflect.Value, error) {
	if schema.Kind() == reflect.Pointer {
		schema = schema.Elem()
	}
	newSchema := reflect.New(schema)
	j := 0
	for j < len(columns) {
		v := values[j]
		column := columns[j]
		cname := columns[j].Name
		field := newSchema.Elem().FieldByName(cname)
		fmt.Println("on", newSchema.Elem().Type().Name(), "looking up", cname)
		if column.Kind.String() == "ReferenceField" {
			ck := column.Kind.(ReferenceField)
			ckrcl := len(ck.References.Columns) + 1
			ns, err := valuesToSchema(columns[j+1:j+ckrcl], values[j+1:j+ckrcl], reflect.TypeOf(ck.References.raw))
			if err != nil {
				return nil, err
			}
			j += ckrcl
			field.Set(*ns)
			continue
		}
		vD := *v.(*any)
		vDT := reflect.TypeOf(vD)
		switch vDA := vD.(type) {
		case int, int8, int16, int32, int64:
			switch column.Kind.String() {
			case "IntegerField", "ReferenceField":
			default:
				return nil, FieldTypeMismatchGot(cname, field.Type().Kind().String(), vDT.Kind().String())
			}
			up := unsafe.Pointer(&v)
			i64 := (*int64)(up)
			field.SetInt(*i64)
		case bool:
			if column.Kind.String() != "BooleanField" {
				return nil, FieldTypeMismatchGot(cname, field.Type().Kind().String(), vDT.Kind().String())
			}
			field.SetBool(vDA)
		case string:
			if column.Kind.String() != "TextField" {
				return nil, FieldTypeMismatchGot(cname, field.Type().Kind().String(), vDT.Kind().String())
			}
			field.SetString(vDA)
		default:
			ok := reflect.TypeOf(vD).AssignableTo(field.Type())
			if !ok {
				return nil, FieldTypeMismatchGot(cname, field.Type().Kind().String(), vDT.Kind().String())
			}
			field.Set(reflect.ValueOf(vD))
		}
		j += 1
	}
	return &newSchema, nil
}
func rowToSchema(schema reflect.Type, m *Model, columns []*Column, rows *sql.Rows) (*reflect.Value, error) {
	if schema.Kind() == reflect.Pointer {
		schema = schema.Elem()
	}
	values := make([]any, len(columns))
	for i, c := range columns {
		var fieldType reflect.Type
		if c.model.Name == schema.Name() {
			field, _ := schema.FieldByName(c.Name)
			fieldType = field.Type
		} else {
			field, _ := reflect.TypeOf(c.model.raw).Elem().FieldByName(c.Name)
			fieldType = field.Type
		}
		fmt.Println("at index", i, "schema", schema.Name(), "column", c.Name)
		v := reflect.Zero(fieldType).Interface()
		values[i] = &v
	}
	err := rows.Scan(values...)
	if err != nil {
		return nil, err
	}
	return valuesToSchema(columns, values, schema)
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
	arguments := []any{}
	if query.Distinct {
		statement += `DISTINCT `
	}
	if len(query.Columns) == 0 {
		for _, c := range m.Columns {
			query.Columns = append(query.Columns, c.Name)
		}
	}
	innerJoins := ""
	qcl := len(query.Columns)
	queriedColumns := []*Column{}
	for i, c := range query.Columns {
		statement += fmt.Sprintf(`"%s"`, c)
		var column Column
		queriedColumns = append(queriedColumns, &column)
		for _, col := range m.Columns {
			if col.Name == c {
				column = *col
				fmt.Println("set column", col.Name)
				if col.Kind.String() == "ReferenceField" {
					ck := col.Kind.(ReferenceField)
					o := fmt.Sprintf(`"%s"."%s"="%s"."%s"`, m.Name, c, ck.References.Name, ck.References.PrimaryKeyColumn.Name)
					ijs := fmt.Sprintf(` INNER JOIN "%s" ON %s`, ck.References.Name, o)
					innerJoins += ijs
					for j, ckrcc := range ck.References.Columns {
						if j == 0 {
							fmt.Println(ckrcc.Name, "gets a comma before it")
							statement += ", "
						}
						statement += fmt.Sprintf(`"%s"`, ckrcc.Name)
						if j+1 < len(ck.References.Columns) {
							statement += `, `
						}
						queriedColumns = append(queriedColumns, ckrcc)
					}
				}
				break
			}
		}
		// if column == nil { //not found
		// 	return nil, QueryOnColumnThatDoesNotExist(m.Name, c)
		// }
		if i+1 < qcl {
			statement += `, `
		}
	}
	fmt.Println(queriedColumns)
	statement += ` FROM ` + `"` + m.Name + `"`
	statement += innerJoins
	sa, a := buildWhere(query)
	statement += sa
	arguments = append(arguments, a...)

	statement += `;`
	statement = replaceStatementPlaceholders(statement) // replaces *! in the WHERE clauses with $1, $2, etc..
	rows, err := ActiveDB.Query(statement, arguments...)
	if err != nil {
		return []I{}, err
	}
	defer rows.Close()

	var schemas []I
	for rows.Next() {
		newSchema, err := rowToSchema(st, m, queriedColumns, rows)
		if err != nil {
			return []I{}, err
		}
		nsi := newSchema.Interface()
		schemas = append(schemas, nsi.(I))
	}
	return schemas, nil
}
