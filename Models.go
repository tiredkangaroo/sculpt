package sculpt

import (
	"fmt"
	"reflect"
)

type Model struct {
	raw     any
	Name    string
	Columns []*Column
}

type Field = interface{}
type IDField struct{}
type IntegerField struct{}
type TextField struct{}

type Column struct {
	PRIMARY_KEY bool
	UNIQUE      bool
	Name        string
	Kind        Field
	Validations []string
}

type Row struct {
	Model  *Model
	Values map[string]any
}

type Condition = string

// NewModel creates a new pointer to a Model from the
// passed in struct. See the reference on how to create
// the struct.
func NewModel(schema any) *Model {
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

		// column.Kind
		kind := field.Tag.Get("kind")
		switch kind {

		case "IDField":
			switch svf.Interface().(type) {
			case string:
				column.Kind = IDField{}
			default:
				panic("type for IDField must be string")
			}
		case "IntegerField":
			switch svf.Interface().(type) {
			case int, int8, int16, int32, int64:
				column.Kind = IntegerField{}
			default:
				panic("type for IntegerField must be int, int8, int16, int32, int64")
			}
		case "TextField":
			switch svf.Interface().(type) {
			case string:
				column.Kind = TextField{}
			default:
				panic("type for TextField must be string")
			}
		case "":
			panic(fmt.Sprintf("field %s does not specfiy a kind in its struct tag", column.Name))
		default:
			panic(fmt.Sprintf("field %s has a kind that is not IDField, IntegerField, or TextField.", column.Name))
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

		column.Name = name
		column.PRIMARY_KEY = primary_key
		column.UNIQUE = unique
		column.Validations = validators

		model.Columns = append(model.Columns, column)
	}
	return model
}

func (m *Model) NewRow(s any) (*Row, error) {
	sv := reflect.ValueOf(s)

	rawt := reflect.TypeOf(m.raw)
	if sv.Type() != rawt {
		LogInfo("%s %s", sv.Type(), rawt)
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
			ok, err := validator.Func(vn)
			if !ok {
				return nil, ValidationFailed(vn, column.Name, field.Interface(), err)
			}
			LogInfo("validation %s on column %s passed", vn, column.Name)
		}
	}
	return newRow, nil
}

func RunQuery[I any](m *Model, s I, query Query) ([]I, error) {
	sv := reflect.ValueOf(s)

	rawt := reflect.TypeOf(m.raw)
	if rawt != sv.Type() {
		return []I{}, ModelTypeMismatch(rawt.Name(), sv.Type().Name())
	}

	sv = sv.Elem()
	st := reflect.TypeOf(s).Elem()

	statement := "SELECT "
	if query.Distinct {
		statement += "DISTINCT "
	}
	if len(query.Columns) == 0 {
		statement += "*"
	}
	for i, c := range query.Columns {
		statement += c
		if i+1 < len(query.Columns) {
			statement += ", "
		}
	}
	statement += " FROM " + m.Name
	if len(query.Conditions) != 0 {
		statement += " WHERE "
		for i, c := range query.Conditions {
			statement += c
			if i+1 < len(query.Conditions) {
				statement += " AND "
			}
		}
	}
	statement += ";"
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
			LogError("an error occured during scanning rows to schema: %s", err.Error())
			continue
		}
		nsi, ok := newSchema.Interface().(I)
		if !ok {
			LogError("something went wrong")
			continue
		}
		schemas = append(schemas, nsi)
	}
	return schemas, nil
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
		if i+1 < len(m.Columns) {
			statement += ","
		}
	}
	statement += ");"
	_, err := ActiveDB.Execute(statement)
	if err != nil {
		return err
	}
	return nil
}

func (r *Row) Save() error {
	statement := fmt.Sprintf("INSERT INTO %s (", r.Model.Name)
	sp2 := "VALUES ("

	for i, c := range r.Model.Columns {
		statement += c.Name
		switch c.Kind.(type) {
		case TextField, IDField:
			sp2 += fmt.Sprintf(`'%s'`, r.Values[c.Name])
		case IntegerField:
			sp2 += fmt.Sprintf("%d", r.Values[c.Name])
		}
		if i+1 < len(r.Model.Columns) {
			statement += ","
			sp2 += ","
		}
	}
	statement += ") "
	sp2 += ") "
	statement += sp2
	statement += ";"
	_, err := ActiveDB.Execute(statement)
	return err
}
