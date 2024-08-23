package sculpt

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
)

var ModelRegistry = make(map[string]*Model)

type Model struct {
	raw              any
	PrimaryKeyColumn *Column
	Name             string
	Columns          []*Column
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
		name := field.Name

		//column.PRIMARY_KEY
		primary_key := false
		if strings.HasPrefix(name, "PrimaryKey") && len(name) > 10 {
			primary_key = true
			column.PRIMARY_KEY = boolFromTag("primary_key", field.Tag, primary_key) // struct tag override
			// FIXME: should probably remove PrimaryKey from name is primary_key is still true
		} else {
			primary_key = boolFromTag("primary_key", field.Tag, primary_key) // only specfied in struct tag
		}
		if primary_key == true {
			model.PrimaryKeyColumn = column
		}

		//column.UNIQUE
		unique := boolFromTag("unique", field.Tag, false)

		svfi := svf.Interface()
		// column.kind
		switch svfi.(type) {
		case string, *string:
			column.Kind = TextField{
				MaximumLength: uintFromTag("maximum_length", field.Tag, 4096),
			}
		case int, int8, int16, int32, int64, *int, *int8, *int16, *int32, *int64:
			column.Kind = IntegerField{}
		case bool, *bool:
			column.Kind = BooleanField{}
		case interface{}, *interface{}:
			var found *Model
			for name, m := range ModelRegistry {
				if nullable {
					if field.Type.Elem().Name() == name {
						found = m
						break
					}
				} else {
					if field.Type.Name() == name {
						found = m
						break
					}
				}
			}
			if found == nil {
				panic("cannot reference non-existent model (reference from field " + field.Name + ")")
			}
			if found.PrimaryKeyColumn == nil {
				panic(fmt.Sprintf("cannot reference model %s. it does not have a primary key", found.Name))
			}
			onDelete := field.Tag.Get("on_delete")
			switch onDelete {
			case "cascade", "no action", "restrict":
			case "set null":
				if nullable == false {
					panic(`on_delete cannot be set to "set null" on a non-nullable field`)
				}
			default:
				panic("unsupported on_delete action: " + onDelete)
			}
			column.Kind = ReferenceField{
				References: found,
				OnDelete:   onDelete,
			}
		}

		//column.Validators
		validators := arrayFromTag("validators", field.Tag)
		for _, v_name := range validators {
			validator, ok := registeredValidators[v_name]
			if !ok {
				panic(fmt.Sprintf("validator %s on field %s is not registered", v_name, name))
			}
			if validator.Kind.String() != column.Kind.String() {
				panic(fmt.Sprintf("validator %s handles %s not %s", v_name, validator.Kind.String(), column.Kind.String()))
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
                    t.typname AS data_type,
                    CASE
                        WHEN t.typname = 'varchar' THEN a.atttypmod - 4
                        ELSE NULL
                    END AS character_maximum_length
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
                    a.attnum`, m.Name, "public")
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
		var maxLength sql.NullInt64

		err := rows.Scan(&columnName, &columnNullable, &columnPrimaryKey, &columnUnique, &columnKind, &maxLength)
		if err != nil {
			return err
		}
		column.Name = columnName
		column.PRIMARY_KEY = columnPrimaryKey
		column.UNIQUE = columnUnique
		column.NULLABLE = columnNullable
		switch columnKind {
		case "int4", "int8", "serial", "bigserial":
			column.Kind = IntegerField{}
		case "text", "varchar":
			if maxLength.Valid {
				column.Kind = TextField{MaximumLength: uint(maxLength.Int64)}
			} else {
				column.Kind = TextField{}
			}
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
		statements = append(statements, fmt.Sprintf(`ALTER TABLE "%s" %s;`, m.Name, alteration))
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
	return nil
}

func (m *Model) Delete(query Query) (err error) {
	w, a := buildWhere(query)
	statement := fmt.Sprintf(`DELETE FROM "%s"%s`, m.Name, w)
	statement = replaceStatementPlaceholders(statement)
	_, err = ActiveDB.Execute(statement, a...)
	return
}

// Delete deletes the model inside the PostgreSQL database.
// The pointer to Model that delete was called with is set to nil.
func (m *Model) DeleteModel() error {
	statement := fmt.Sprintf(`DROP TABLE "%s" CASCADE;`, m.Name)
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
		if field.Type().Kind() == reflect.Pointer {
			newRow.Values[column.Name] = field.Elem().Interface()
		} else {
			newRow.Values[column.Name] = field.Interface()
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
