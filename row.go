package sculpt

import (
	"fmt"
	"reflect"
)

type Row struct {
	Model  *Model
	Values map[string]any
}

func (r *Row) Save() error {
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
