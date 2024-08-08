package sculpt

import "fmt"

type Row struct {
	Model  *Model
	Values map[string]any
}

func (r *Row) Save() error {
	statement := fmt.Sprintf(`INSERT INTO "%s" (`, r.Model.Name)
	sp2 := "VALUES ("

	for i, c := range r.Model.Columns {
		statement += `"` + c.Name + `"`
		switch c.Kind {
		case TextField:
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
