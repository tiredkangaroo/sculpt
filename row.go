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
	statementArguments := []any{}
	sp2 := "VALUES ("

	for i, c := range r.Model.Columns {
		statement += `"` + c.Name + `"`
		sp2 += fmt.Sprintf("$%d", i+1) // $1, $2, $3, etc. (added 1 to i because $0 probably isn't allowed)
		switch c.Kind.String() {
		case "TextField", "IntegerField":
			statementArguments = append(statementArguments, r.Values[c.Name])
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
			statementArguments = append(statementArguments, field.Interface())
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
	fmt.Println(statement, statementArguments)
	_, err := ActiveDB.Execute(statement, statementArguments...)
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
