package sculpt

import (
	"fmt"
)

type PopulatedFields = []Field

// Model
type Model struct {
	Name   string
	Fields []Field
}

type Query struct {
	Columns  []string
	DISTINCT bool
	Where    map[string]interface{}
	Order_By string
}

type Result struct {
	Statement       string
	PopulatedModels []Model
}

type Field interface {
	IS_PRIMARY_KEY() bool
	IS_UNIQUE() bool
	Name() string
	Kind() string
	// Populate(any) error
}

type IDField struct {
	PRIMARY_KEY    bool
	UNIQUE         bool
	ColumnName     string
	Auto           bool
	PopulatedValue string
}

type TextField struct {
	PRIMARY_KEY    bool
	UNIQUE         bool
	ColumnName     string
	Minimum_Length int
	Maximum_Length int
	PopulatedValue string
}

type IntegerField struct {
	PRIMARY_KEY    bool
	UNIQUE         bool
	ColumnName     string
	PopulatedValue int
}

// NewModel provides a function to create a new Model object.
func NewModel(name string, values ...Field) (m *Model) {
	return &Model{
		Name:   name,
		Fields: values,
	}
}

// IS_PRIMARY_KEY provides an interface to check if it is
// the primary key on the model that it is on.
func (idf IDField) IS_PRIMARY_KEY() bool {
	return idf.PRIMARY_KEY
}

// IS_UNIQUE provides an interface to check if it is
// unique on the model that it is on.
func (idf IDField) IS_UNIQUE() bool {
	return idf.UNIQUE
}

// Name provides an interface to return the name
// the
func (idf IDField) Name() string {
	return idf.ColumnName
}

func (idf IDField) Value() string {
	return idf.PopulatedValue
}
func (idf IDField) Kind() string {
	return "idfield"
}
func (idf IDField) Populate(value any) error {
	switch value.(type) {
	case string:
		break
	default:
		return FieldTypeMismatch(idf.Name(), "string")
	}
	return nil
}

func (tf TextField) IS_PRIMARY_KEY() bool {
	return tf.PRIMARY_KEY
}
func (tf TextField) IS_UNIQUE() bool {
	return tf.UNIQUE
}
func (tf TextField) Name() string {
	return tf.ColumnName
}
func (tf TextField) Value() string {
	return tf.PopulatedValue
}
func (tf TextField) Kind() string {
	return "textfield"
}

func (inf IntegerField) IS_PRIMARY_KEY() bool {
	return inf.PRIMARY_KEY
}
func (inf IntegerField) IS_UNIQUE() bool {
	return inf.UNIQUE
}
func (inf IntegerField) Name() string {
	return inf.ColumnName
}
func (inf IntegerField) Value() int {
	return inf.PopulatedValue
}
func (inf IntegerField) Kind() string {
	return "integerfield"
}

// used in models.go/New
func getKind(v interface{}) string {
	switch v.(type) {
	case string:
		return "string"
	case int:
		return "integer"
	default:
		return "not assigned"
	}
}

// Creates the table in the postgres database.
// if ifNotExists is true, the table will only be created if it does not already exist.
func (entity *Model) Create(ifNotExists bool) error {
	if !Connected() {
		return OperationRequiresDatabaseConnection("create table")
	}
	statement := "CREATE TABLE "
	if ifNotExists {
		statement += "IF NOT EXISTS "
	}
	statement += entity.Name + " ("
	fields := entity.Fields
	for j := 0; j < len(fields); j++ {
		switch field := fields[j].(type) {
		case IDField:
			statement += fmt.Sprintf("%s VARCHAR(32)", field.Name())
			if field.PRIMARY_KEY {
				statement += " PRIMARY KEY"
			}
			if field.UNIQUE {
				statement += " UNIQUE"
			}
		case IntegerField:
			statement += fmt.Sprintf("%s INT", field.Name())
			if field.PRIMARY_KEY {
				statement += " PRIMARY KEY"
			}
			if field.UNIQUE {
				statement += " UNIQUE"
			}
		case TextField:
			maxl := "MAX"
			if field.Maximum_Length != 0 {
				maxl = fmt.Sprint(field.Maximum_Length)
			}
			statement += fmt.Sprintf("%s VARCHAR(%s)", field.Name(), maxl)
		}
		if j+1 != len(fields) {
			statement += ", "
		}
	}
	statement += ");"
	LogStatement(statement)
	_, err := DB.Exec(statement)
	return err
}

// new model ready for population
func (m *Model) New(pv map[string]interface{}) (Model, error) {
	var PopulatedFields PopulatedFields
	for _, f := range m.Fields {
		valueSetForField, found := pv[f.Name()]

		if found == false {
			if f.Kind() == "idfield" {
				if f.(IDField).Auto {
					valueSetForField = GenerateUID()
				} else {
					return Model{}, MissingFieldForPopulation(f.Name())
				}
			}
		}

		vSFFKind := getKind(valueSetForField)
		h := f
		switch f.Kind() {
		case "textfield":
			if vSFFKind != "string" {
				return Model{}, FieldTypeMismatch(f.Name(), "string")
			} else {
				h := h.(TextField)
				h.PopulatedValue = valueSetForField.(string)
				PopulatedFields = append(PopulatedFields, h)
			}
		case "integerfield":
			if vSFFKind != "integer" {
				return Model{}, FieldTypeMismatch(f.Name(), "integer")
			} else {
				h := h.(IntegerField)
				h.PopulatedValue = valueSetForField.(int)
				PopulatedFields = append(PopulatedFields, &h)
			}
		case "idfield":
			if vSFFKind != "idfield" {
				return Model{}, FieldTypeMismatch(f.Name(), "idfield")
			} else {
				h := h.(IDField)
				h.PopulatedValue = valueSetForField.(string)
				PopulatedFields = append(PopulatedFields, &h)
			}
		}
		PopulatedFields = append(PopulatedFields, f)
	}
	return Model{Name: m.Name, Fields: PopulatedFields}, nil
}

// function used within Models.go to construct "WHERE" in sql
func where(w map[string]interface{}) string {
	i := 0
	statement := ""
	if len(w) > 0 {
		statement += " WHERE "
	}
	for n, v := range w {
		statement += n
		statement += "="
		switch val := v.(type) {
		case string:
			statement += "'" + val + "' "

		case int:
			statement += fmt.Sprintf("%d ", val)
		}
		i++
		if i != len(w) {
			statement += "AND "
		}
	}
	return statement
}

// get from query
func (m *Model) Get(q *Query) (Result, error) {
	statement := "SELECT "
	if q.DISTINCT == true {
		statement += "DISTINCT "
	}
	if len(q.Columns) == 0 {
		statement += "*"
	} else {
		for j := 0; j < len(q.Columns); j++ {
			statement += q.Columns[j]
			if j+1 != len(q.Columns) {
				statement += ", "
			}
		}
	}
	statement += " FROM "
	statement += m.Name + " "

	statement += where(q.Where)
	if len(q.Order_By) > 0 {
		statement += "ORDER BY "
		statement += q.Order_By
	}
	statement += ";"
	LogStatement(statement)
	result, err := DB.Query(statement)
	if err != nil {
		return Result{}, err
	}
	defer result.Close()
	Columns, err := result.Columns()
	if err != nil {
		return Result{}, err
	}
	var rows []interface{}
	row := make([]interface{}, len(Columns))
	rowp := make([]interface{}, len(Columns)) //rows as pointer
	for j := 0; j < len(row); j++ {
		rowp[j] = &row[j] //make into pointers
	}
	for result.Next() {
		err := result.Scan(rowp...) //scan into the pointer (will scan into the real thing)
		if err != nil {
			return Result{}, err
		}
		rows = append(rows, row...) //add to rows
	}
	var rc []map[string]interface{}
	for k := 0; k < len(rows); k += len(Columns) {
		rcm := make(map[string]interface{}, len(Columns))
		for l := 0; l < len(Columns); l++ {
			rcm[Columns[l]] = rows[k+l]
		}
		rc = append(rc, rcm)
	}
	var h []Model
	for i := range rc {
		var fields []Field
		for _, f := range m.Fields {
			rci := rc[i]
			v := rci[f.Name()]
			field := f
			switch field.Kind() {
			case "idfield":
				vas, ok := v.(string)
				if !ok {
					continue
				}
				fidf := field.(IDField)
				fidf.PopulatedValue = vas
				field = fidf
			case "textfield":
				vas, ok := v.(string)
				if !ok {
					continue
				}
				ftf := field.(TextField)
				ftf.PopulatedValue = vas
				field = ftf
			case "integerfield":
				vai, ok := v.(int)
				if !ok {
					continue
				}
				finf := field.(IntegerField)
				finf.PopulatedValue = vai
				field = finf
			}
			fields = append(fields, field)
		}
		newModel := Model{
			Name:   m.Name,
			Fields: fields,
		}
		h = append(h, newModel)
	}
	return Result{
		Statement:       statement,
		PopulatedModels: h,
	}, nil
}

func (m *Model) Delete(w map[string]interface{}) error {
	statement := "DELETE FROM "
	statement += m.Name
	statement += " " + where(w)
	statement += ";"
	LogStatement(statement)
	_, err := DB.Exec(statement)
	return err
}

// Save saves the current instance of Model into postgres with values.
func (et *Model) Save() error {
	statement := "INSERT INTO " // statement: INSERT INTO
	statement += et.Name        // statement: INSERT INTO MY_TABLE
	statement += " VALUES ("    // statement: INSERT INTO MY_TABLE VALUES(
	fields := et.Fields
	for j := 0; j < len(fields); j++ {
		switch field := fields[j].(type) {
		case IDField:
			statement += "'" + field.Value() + "'"
		case IntegerField:
			statement += fmt.Sprint(field.Value())
		case TextField:
			statement += "'" + field.Value() + "'"
		}
		if j+1 != len(fields) {
			statement += ","
		}
	}
	statement += ");"
	LogStatement(statement)
	_, err := DB.Exec(statement)
	return err
}
