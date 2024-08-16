package sculpt

import (
	"database/sql"
	"fmt"
	"reflect"
	"unsafe"
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

func rowToSchema(schema reflect.Type, m *Model, columns []string, rows *sql.Rows) (*reflect.Value, error) {
	if schema.Kind() == reflect.Pointer {
		schema = schema.Elem()
	}
	newSchema := reflect.New(schema)
	var values []any
	for _, c := range columns {
		field, _ := schema.FieldByName(c)
		v := reflect.Zero(field.Type).Interface()
		values = append(values, &v)
	}
	err := rows.Scan(values...)
	if err != nil {
		return nil, err
	}
	for j, v := range values {
		cname := columns[j]
		column := getColumnByName(m, cname)
		field := newSchema.Elem().FieldByName(cname)
		vD := *v.(*any)
		if column.Kind.String() == "ReferenceField" {
			ck := column.Kind.(ReferenceField)
			referencedRows, err := ActiveDB.Query(
				fmt.Sprintf(
					`SELECT * FROM "%s" WHERE "%s" = '%v';`,
					ck.References.Name,
					ck.References.PrimaryKeyColumn.Name,
					vD),
			)
			if err != nil {
				return nil, err
			}
			nc := []string{}
			for _, c := range ck.References.Columns {
				nc = append(nc, c.Name)
			}
			for referencedRows.Next() {
				s, err := rowToSchema(field.Type(), ck.References, nc, referencedRows)
				if err != nil {
					return nil, err
				}
				field.Set(*s)
			}
		} else {
			vDT := reflect.TypeOf(vD)
			switch vDA := vD.(type) {
			case int, int8, int16, int32, int64:
				if column.Kind.String() != "IntegerField" {
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
		}
	}
	return &newSchema, nil
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
		for _, c := range m.Columns {
			query.Columns = append(query.Columns, c.Name)
		}
	}
	for i, c := range query.Columns {
		statement += fmt.Sprintf(`"%s"`, c)
		if i+1 < len(query.Columns) {
			statement += `, `
		}
	}
	statement += ` FROM ` + `"` + m.Name + `"`
	statement += buildWhere(query)
	statement += `;`
	rows, err := ActiveDB.Query(statement)
	if err != nil {
		return []I{}, err
	}
	defer rows.Close()

	var schemas []I
	for rows.Next() {
		newSchema, err := rowToSchema(st, m, query.Columns, rows)
		if err != nil {
			return []I{}, err
		}
		nsi := newSchema.Interface()
		schemas = append(schemas, nsi.(I))
	}
	return schemas, nil
}
