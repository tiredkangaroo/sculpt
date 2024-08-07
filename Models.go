package sculpt

import (
	"fmt"
	"reflect"
	"strings"
)

type Model struct {
	raw     any
	Name    string
	Columns []*Column
}

// New creates a new pointer to a Model from the
// passed in struct. See the reference on how to create
// the struct.
func New(schema any) *Model {
	st := reflect.TypeOf(schema)
	sv := reflect.ValueOf(schema)

	if st.Kind() != reflect.Pointer {
		panic("schema must be POINTER to struct.")
	}

	st = st.Elem()
	sv = sv.Elem()

	if st.Kind() != reflect.Struct {
		panic("schema must be pointer to STRUCT.")
	}

	model := new(Model)
	model.Name = st.Name()
	model.raw = schema
	for i := range st.NumField() {
		field := st.Field(i)
		svf := sv.Field(i)

		column := new(Column)

		//column.Name
		name := st.Field(i).Name

		//column.PRIMARY_KEY
		primary_key := boolFromTag("primary_key", field.Tag, false)

		//column.UNIQUE
		unique := boolFromTag("unique", field.Tag, false)

		//column.NULLABLE
		nullable := boolFromTag("nullable", field.Tag, true)

		// column.Kind
		kind := field.Tag.Get("kind")
		switch kind {
		case "IntegerField":
			switch svf.Interface().(type) {
			case int, int8, int16, int32, int64:
				column.Kind = IntegerField
			default:
				panic("type for IntegerField must be int, int8, int16, int32, int64")
			}
		case "TextField":
			switch svf.Interface().(type) {
			case string:
				column.Kind = TextField
			default:
				panic("type for TextField must be string")
			}
		case "":
			switch svf.Interface().(type) {
			case string:
				column.Kind = TextField
			case int, int8, int16, int32, int64:
				column.Kind = IntegerField
			default:
				panic(fmt.Sprintf("field %s does not specfiy a kind in its struct tag", column.Name))
			}
		default:
			panic(fmt.Sprintf("field %s has a kind that is not IntegerField, or TextField.", column.Name))
		}

		//column.Validators
		validators := arrayFromTag("validators", field.Tag)
		for _, v_name := range validators {
			validator, ok := registeredValidators[v_name]
			if !ok {
				panic(fmt.Sprintf("validator %s on field %s is not registered", v_name, name))
			}
			ckt := reflect.TypeOf(column.Kind)
			vkt := reflect.TypeOf(validator.Kind)
			if ckt != vkt {
				panic(fmt.Sprintf("validator %s handles %s not %s", v_name, vkt.String(), ckt.String()))
			}
		}

		column.model = model
		column.Name = name
		column.PRIMARY_KEY = primary_key
		column.UNIQUE = unique
		column.NULLABLE = nullable
		column.Validations = validators

		model.Columns = append(model.Columns, column)
	}
	return model
}

// Migrate migrates the model to a new schema.
func (m *Model) Migrate() error {
	newModel := New(m.raw)
	// Detect changes
	statement := fmt.Sprintf(`SELECT
                    a.attname AS column_name,
                    NOT (a.attnotnull OR (t.typname = 'bool' AND a.atttypmod = -1)) AS nullable,
                    (SELECT count(*) = 1 FROM pg_constraint c WHERE c.conrelid = a.attrelid AND c.conkey[1] = a.attnum AND c.contype = 'p') AS primary_key,
                    (SELECT count(*) = 1 FROM pg_constraint c WHERE c.conrelid = a.attrelid AND c.conkey[1] = a.attnum AND c.contype = 'u') AS unique,
                    t.typname AS data_type
                FROM
                    pg_attribute a
                JOIN
                    pg_type t ON a.atttypid = t.oid
                JOIN
                    pg_class c ON a.attrelid = c.oid
                JOIN
                    pg_namespace n ON c.relnamespace = n.oid
                WHERE
                    a.attnum > 0 AND NOT a.attisdropped AND c.relname = '%s' AND n.nspname = '%s'
                ORDER BY
                    a.attnum;`, strings.ToLower(m.Name), "public")
	var oldColumns []*Column
	rows, err := ActiveDB.Query(statement)
	if err != nil {
		return err
	}
	defer rows.Close()
	if rows == nil {
		return RowsAreEmpty()
	}
	for rows.Next() {
		column := new(Column)
		column.model = m
		var columnName, columnKind string
		var columnPrimaryKey, columnUnique, columnNullable bool
		err := rows.Scan(&columnName, &columnNullable, &columnPrimaryKey, &columnUnique, &columnKind)
		if err != nil {
			return err
		}
		column.Name = columnName
		column.PRIMARY_KEY = columnPrimaryKey
		column.UNIQUE = columnUnique
		column.NULLABLE = columnNullable
		switch columnKind {
		case "int4", "int8", "serial", "bigserial":
			column.Kind = IntegerField
		case "text", "varchar":
			column.Kind = TextField
		default:
			// Default case can be expanded to handle other types
			return UnknownTypeFromDatabase(columnKind)
		}
		oldColumns = append(oldColumns, column)
	}
	additions, alterations, deletions := compareColumns(oldColumns, newModel.Columns)
	if len(additions) == 0 && len(alterations) == 0 && len(deletions) == 0 {
		LogInfo("Migrations:", "No migrations.")
		return nil
	}
	LogInfo("Migrations:", "Migrations for %s (%s+%d%s, %s!%d%s, %s-%d%s):", m.Name, greenbgBlackF, len(additions), normal, yellowbgBlackF, len(alterations), normal, redbgWhiteF, len(deletions), normal)
	statements := []string{}
	for _, deletion := range deletions {
		LogInfo("Migrations:", "Migration change detected: Deleted field: %s (%T)", deletion.Name, deletion.Kind)
		statement := fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s;", m.Name, deletion.Name)
		statements = append(statements, statement)
	}
	for _, alteration := range alterations {
		LogInfo("Migrations:", "Migration change detected: Alteration.")
		statements = append(statements, fmt.Sprintf("ALTER TABLE %s %s", m.Name, alteration))
	}
	for _, addition := range additions {
		LogInfo("Migrations:", "Migration change detected: Added new field: %s (%T)", addition.Name, addition.Kind)
		kts, _ := kindToSQL(addition.Kind)
		statement = fmt.Sprintf("ALTER TABLE %s ADD %s %s", m.Name, addition.Name, kts)
		statement += extraColumnProperties(addition)
		statement += ";"
		statements = append(statements, statement)
	}
	// for _, statement := range statements {
	// 	_, err := ActiveDB.Execute(statement)
	// 	if err != nil {
	// 		LogError("Migrations failed to apply: %s", err.Error())
	// 		return err
	// 	}
	// }
	fmt.Println(statements)
	return nil
}

// Delete deletes the model inside the PostgreSQL database.
// The pointer to Model that delete was called with is set to nil.
func (m *Model) Delete() error {
	statement := fmt.Sprintf("DROP TABLE %s;", m.Name)
	_, err := ActiveDB.Execute(statement)
	if err != nil {
		return err
	}
	m = nil
	return nil
}

// Truncate removes all rows inside the table associated
// with the Model.
func (m *Model) Truncate() error {
	statement := fmt.Sprintf("TRUNCATE TABLE %s;", m.Name)
	_, err := ActiveDB.Execute(statement)
	return err
}

// New creates a new row associated with the model.
func (m *Model) New(s any) (*Row, error) {
	sv := reflect.ValueOf(s)

	rawt := reflect.TypeOf(m.raw)
	if sv.Type() != rawt {
		LogInfo("", "%s %s", sv.Type(), rawt)
		return nil, ModelTypeMismatch(rawt.Name(), sv.Type().Name())
	}

	sv = sv.Elem()

	newRow := new(Row)
	newRow.Model = m
	newRow.Values = make(map[string]any)
	for _, column := range m.Columns {
		field := sv.FieldByName(column.Name)
		newRow.Values[column.Name] = field.Interface()
		for _, vn := range column.Validations {
			validator := registeredValidators[vn]
			if validator.Func == nil {
				return nil, ValidatorHasNoFunc(vn, column.Name)
			}
			if validator.Kind != column.Kind {
				return nil, ValidatorCannotBeUsedForKind(vn, validator.Kind, column.Name, column.Kind)
			}
			ok, err := validator.Func(field.Interface())
			if !ok {
				return nil, ValidationFailed(vn, column.Name, field.Interface(), err)
			}
			LogInfo("validation %s on column %s passed", vn, column.Name)
		}
	}
	return newRow, nil
}

func (m *Model) Save() error {
	//FIXME: only does if not exists
	statement := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (", m.Name)
	for i, column := range m.Columns {
		_typ, err := kindToSQL(column.Kind)
		if err != nil {
			return err
		}
		statement += fmt.Sprintf("%s %s", column.Name, _typ)
		statement += extraColumnProperties(*column)
		if i+1 < len(m.Columns) {
			statement += ", "
		}
	}
	statement += ");"
	_, err := ActiveDB.Execute(statement)
	if err != nil {
		return err
	}
	return nil
}
