package sculpt

import (
	"fmt"
	"reflect"
)

var ModelRegistry = make(map[string]*Model)

type Model struct {
	raw     any
	Name    string
	Columns []*Column
}

// New creates a new pointer to a Model from the
// passed in struct. See the reference on how to create
// the struct.
func Register(schema any) *Model {
	st := reflect.TypeOf(schema)
	sv := reflect.ValueOf(schema)

	if st.Kind() != reflect.Pointer {
		panic("schema must be pointer to STRUCT.")
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

		//column.NULLABLE
		nullable := false
		if field.Type.Kind() == reflect.Pointer {
			nullable = true
		}

		svf := sv.Field(i)

		column := new(Column)

		//column.Name
		name := st.Field(i).Name

		//column.PRIMARY_KEY
		if i == 0 {
			column.PRIMARY_KEY = true
		}

		//column.UNIQUE
		unique := boolFromTag("unique", field.Tag, false)

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
			if validator.Kind != column.Kind {
				panic(fmt.Sprintf("validator %s handles %s not %s", v_name, kindToString(validator.Kind), kindToString(column.Kind)))
			}
		}

		column.model = model
		column.Name = name
		column.UNIQUE = unique
		column.NULLABLE = nullable
		column.Validations = validators

		model.Columns = append(model.Columns, column)
	}
	mn := model.Name
	ModelRegistry[mn] = model
	return model
}

// Migrate migrates the model to a new schema.
func (m *Model) Migrate() error {
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
                    a.attnum;`, m.Name, "public")
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
	newModel := Register(m.raw)
	additions, alterations, deletions := compareColumns(oldColumns, newModel.Columns)
	if len(additions) == 0 && len(alterations) == 0 && len(deletions) == 0 {
		LogError("Migrations: %s", "No migrations.")
		return nil
	}
	LogInfo("Migrations:", "Migrations for %s (%s+%d%s, %s!%d%s, %s-%d%s):", m.Name, greenbgBlackF, len(additions), normal, yellowbgBlackF, len(alterations), normal, redbgWhiteF, len(deletions), normal)
	statements := []string{}
	for _, deletion := range deletions {
		LogInfo("Migrations:", "Migration change detected: Deleted field: %s (%T)", deletion.Name, deletion.Kind)
		statement := fmt.Sprintf(`ALTER TABLE "%s" DROP COLUMN "%s";`, m.Name, deletion.Name)
		statements = append(statements, statement)
	}
	for _, alteration := range alterations {
		LogInfo("Migrations:", "Migration change detected: Alteration.")
		statements = append(statements, fmt.Sprintf(`ALTER TABLE "%s" "%s";`, m.Name, alteration))
	}
	for _, addition := range additions {
		LogInfo("Migrations:", "Migration change detected: Added new field: %s (%T)", addition.Name, addition.Kind)
		kts, _ := kindToSQL(addition.Kind)
		statement = fmt.Sprintf(`ALTER TABLE "%s" ADD "%s" %s`, m.Name, addition.Name, kts)
		statement += extraColumnProperties(addition)
		statement += ";"
		statements = append(statements, statement)
	}
	for _, statement := range statements {
		_, err := ActiveDB.Execute(statement)
		if err != nil {
			LogError("Migrations failed to apply: %s", err.Error())
			continue
		}
	}
	fmt.Println(statements)
	return nil
}

// Delete deletes the model inside the PostgreSQL database.
// The pointer to Model that delete was called with is set to nil.
func (m *Model) Delete() error {
	statement := fmt.Sprintf(`DROP TABLE "%s";`, m.Name)
	_, err := ActiveDB.Execute(statement)
	if err != nil {
		return err
	}
	m = nil
	return nil
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
			LogDebug("Validator:", "validation %s on column %s passed (input: %s)", vn, column.Name, field.Interface())
		}
	}
	return newRow, nil
}

// NewNE is equivelent to New, however, it does not return an error.
// In place of returning an error, it panics. This is only to be used
// with prior knowledge that New would not have returned an error in the
// first place.
func (m *Model) NewNE(s any) *Row {
	r, err := m.New(s)
	if err != nil {
		panic(err)
	}
	return r
}

func (m *Model) Save() error {
	//FIXME: only does if not exists
	statement := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS "%s" (`, m.Name)
	for i, column := range m.Columns {
		_typ, err := kindToSQL(column.Kind)
		if err != nil {
			return err
		}
		statement += fmt.Sprintf(`"%s" %s`, column.Name, _typ)
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
