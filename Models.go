package sculpt

import (
	"fmt"
	"reflect"
)

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

type Field = interface{}
type IDField struct{}
type IntegerField struct{}
type TextField struct{}

type Model struct {
	raw     any
	Name    string
	Columns []*Column
}

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
		return nil, ModelTypeMismatch(sv.Type().Name(), rawt.Name())
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
