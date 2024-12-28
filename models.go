package sculpt

import (
	"fmt"
	"reflect"
	"sculpt/internals/sql"
)

// Model represents a table in the database.
type Model[T any] struct {
	name    string
	t       reflect.Type
	columns []Column
}

// Query creates a new Query to get stored data with the model.
func (m *Model[T]) Query() *Query[T] {
	return &Query[T]{model: m}
}

// Save uses the Postgres connection to save the struct into to the database.
func (m *Model[T]) Save(v T) error {
	statement := fmt.Sprintf(`INSERT INTO %s (`, m.name)
	values_statement := `VALUES (`
	values := []any{}

	rv := reflect.ValueOf(v)
	i := 0 // i maintains a counter of placeholder usage for pgx
	for j, column := range m.columns {
		statement += column.name
		if column.autoincrement {
			values_statement += `DEFAULT`
			if j < len(m.columns)-1 {
				statement += `, `
				values_statement += `, `
			}
			continue
		}
		i++
		values_statement += fmt.Sprintf(`$%d`, i)
		if j < len(m.columns)-1 {
			statement += `, `
			values_statement += `, `
		}
		field := rv.FieldByName(column.name)
		if column.nullable {
			// call the Optional.Nil method
			nilcheck := field.MethodByName("Nil").Call([]reflect.Value{})
			if nilcheck[0].Bool() { // if the optional is nil
				values = append(values, nil)
				continue
			}

			v := field.MethodByName("Value").Call([]reflect.Value{}) // call the Optional.Value method
			values = append(values, v[0].Interface())
		} else {
			values = append(values, field.Interface())
		}
	}
	statement += `) `
	values_statement += `);`
	statement += values_statement
	_, err := sql.Execute(statement, values...)
	return err
}

// Create uses the Postgres connection to create the table in the database, if it does not
// already exist.
func (m *Model[T]) Create() error {
	statement := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (`, m.name)
	for i, column := range m.columns {
		statement += fmt.Sprintf(`%s %s`, column.name, column.sqltype.String())
		if !column.nullable {
			statement += " NOT NULL"
		}
		if column.primarykey {
			statement += " PRIMARY KEY"
		}
		if column.unique {
			statement += " UNIQUE"
		}
		if i < len(m.columns)-1 {
			statement += `, `
		}
	}
	statement += `);`
	_, err := sql.Execute(statement)
	return err
}

// New makes a new Model from a struct.
func New[T any]() (*Model[T], error) {
	rt := reflect.TypeFor[T]()
	if rt.Kind() != reflect.Struct {
		return nil, fmt.Errorf("T must be a struct")
	}
	m := new(Model[T])

	m.name = rt.Name()
	m.t = rt

	for i := range rt.NumField() {
		field := rt.Field(i)
		column, err := handleColumn(field)
		if err != nil {
			return nil, fmt.Errorf("column %s error: %v", field.Name, err)
		}
		m.columns = append(m.columns, column)
	}
	return m, nil
}
