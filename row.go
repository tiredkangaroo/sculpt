package sculpt

import (
	"fmt"
	"reflect"
)

type Row struct {
	Model  *Model
	Values map[string]any
}

func (r *Row) runValidations() error {
	for _, column := range r.Model.Columns {
		valueForColumn := r.Values[column.Name]
		for _, vn := range column.Validations {
			validator, ok := registeredValidators[vn]
			if !ok {
				return ValidatorDoesNotExist(vn, column.Name)
			}
			if validator.Func == nil {
				return ValidatorHasNoFunc(vn, column.Name)
			}
			if validator.Kind.String() != column.Kind.String() {
				return ValidatorCannotBeUsedForKind(vn, validator.Kind, column.Name, column.Kind)
			}
			err := validator.Func(valueForColumn)
			if err != nil {
				return ValidationFailed(vn, column.Name, valueForColumn, err)
			}
		}
	}
	return nil
}

func (r *Row) save() error {
	statement := fmt.Sprintf(`INSERT INTO "%s" (`, r.Model.Name)
	sp2 := "VALUES ("

	for i, c := range r.Model.Columns {
		statement += `"` + c.Name + `"`
		switch c.Kind.String() {
		case "TextField":
			sp2 += fmt.Sprintf(`'%s'`, r.Values[c.Name])
		case "IntegerField":
			sp2 += fmt.Sprintf("%d", r.Values[c.Name])
		case "ReferenceField":
			var vf reflect.Value
			if c.NULLABLE {
				value := r.Values[c.Name].(interface{})
				vf = reflect.ValueOf(value)
			} else {
				value := r.Values[c.Name].(*interface{})
				vf = reflect.ValueOf(value).Elem()
			}
			field := vf.FieldByName(c.Kind.(ReferenceField).References.PrimaryKeyColumn.Name)
			valueSQL, err := anyToSQLString(field.Interface())
			if err != nil {
				return err
			}
			sp2 += valueSQL
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

func (r *Row) Save() error {
	err := r.runValidations()
	if err != nil {
		return err
	}
	return r.save()
}

func (r *Row) SaveWithForce() error {
	return r.save()
}
