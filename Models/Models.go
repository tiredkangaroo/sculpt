package models

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"sculpt/Log"
	"sculpt/Manager"
)

type IDField struct {
	PRIMARY_KEY bool
	UNIQUE      bool
	Name        string //Column Name
	Auto        bool   //Automatically populate with a random id?
	value       string
}

func GenerateUID() string {
	id := ""
	for range 10 {
		r := rand.IntN(25) + 1
		id += fmt.Sprintf("%c", ('A' - 1 + r))
	}
	id += fmt.Sprintf("%02d", rand.IntN(26))
	return id
}

func (idf *IDField) populate(value string) error {
	if len(value) > 32 {
		return fmt.Errorf("(populating %s error) The value provided does not meet the ID length constraint (<12).", idf.Name)
	}
	idf.value = value
	return nil
}

type TextField struct {
	PRIMARY_KEY    bool
	UNIQUE         bool
	Name           string
	Minimum_Length int
	Maximum_Length int
	value          string
}

func (tf *TextField) populate(value string) error {
	valuelen := len(value)
	if valuelen < tf.Minimum_Length {
		return fmt.Errorf("(populating %s error) The value provided does not meet the minimum length.", tf.Name)
	}
	if tf.Maximum_Length != 0 { //0 is the default value for the integer type
		if valuelen > tf.Maximum_Length {
			return fmt.Errorf("(populating %s error) The value provided does not meet the maximum length.", tf.Name)
		}
	}
	tf.value = value
	return nil
}

type IntegerField struct {
	PRIMARY_KEY bool
	UNIQUE      bool
	Name        string
	value       int
}

func (inf *IntegerField) populate(value int) {
	inf.value = value
}

type Model struct {
	Name   string
	Fields []interface{}
}

type Query struct {
	Columns  []string
	DISTINCT bool
	Where    map[string]interface{}
	Order_By string
}
type Result struct {
	*Model
	*Query
	Statement string
	Objects   []map[string]interface{}
}

func (m *Model) New(values ...interface{}) (*PopulatedModel, error) {
	mi := m //model instance
	if len(values) != len(mi.Fields) {
		return &PopulatedModel{}, errors.New("(population failed for model) Insuffcient paramaters.")
	}
	for j := 0; j < len(mi.Fields); j++ {
		switch field := mi.Fields[j].(type) {
		case IDField:
			var vj string
			if field.Auto {
				v := GenerateUID()
				vj = v
			} else {
				v, ok := values[j].(string)
				if !ok {
					return &PopulatedModel{}, errors.New("(population failed for IDField) The value provided does not meet the string type requirement for the ID Field.")
				}
				vj = v
			}
			err := field.populate(vj)
			if err != nil {
				return &PopulatedModel{}, err
			}
			mi.Fields[j] = field

		case TextField:
			vj, ok := values[j].(string)
			if ok {
				err := field.populate(vj)
				if err != nil {
					return &PopulatedModel{}, err
				}
				mi.Fields[j] = field
			} else {
				return &PopulatedModel{}, errors.New("(population failed for TextField) The value provided does not meet the string type requirement for String Field.")
			}
		case IntegerField:
			vj, ok := values[j].(int)
			if ok {
				field.populate(vj)
				mi.Fields[j] = field
			} else {
				return &PopulatedModel{}, errors.New("(population failed for IntegerField) The value provided does not meet the integer type requirement for Integer Field.")
			}
		}
	}
	return &PopulatedModel{Model: mi}, nil
}

// Creates the table in the postgres database.
// if ifNotExists is true, the table will only be created if it does not already exist.
func (entity *Model) Create(ifNotExists bool) error {
	if !Manager.Connected() {
		return fmt.Errorf(Manager.RequiredConnectionError, "Create Table")
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
			statement += fmt.Sprintf("%s VARCHAR(32)", field.Name)
			if field.PRIMARY_KEY {
				statement += " PRIMARY KEY"
			}
			if field.UNIQUE {
				statement += " UNIQUE"
			}
		case IntegerField:
			statement += fmt.Sprintf("%s INT", field.Name)
			if field.PRIMARY_KEY {
				statement += " PRIMARY KEY"
			}
			if field.UNIQUE {
				statement += " UNIQUE"
			}
		case TextField:
			maxl := "MAX"
			if field.Maximum_Length != -1 {
				maxl = fmt.Sprint(field.Maximum_Length)
			}
			statement += fmt.Sprintf("%s VARCHAR(%s)", field.Name, maxl)
		}
		if j+1 != len(fields) {
			statement += ", "
		}
	}
	statement += ");"
	log.Statement(statement)
	_, err := Manager.DB.Exec(statement)
	return err
}

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
	log.Statement(statement)
	result, err := Manager.DB.Query(statement)
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
	return Result{Model: m, Query: q, Statement: statement, Objects: rc}, nil
}

func (m *Model) Delete(w map[string]interface{}) error {
	statement := "DELETE FROM "
	statement += m.Name
	statement += " " + where(w)
	statement += ";"
	log.Statement(statement)
	_, err := Manager.DB.Exec(statement)
	return err
}

type PopulatedModel struct {
	*Model
}

// Saves the current instance of Model into postgres (with values)
func (et *PopulatedModel) Save() error {
	entity := et.Model
	statement := "INSERT INTO "
	statement += entity.Name
	statement += " VALUES ("
	fields := entity.Fields
	for j := 0; j < len(fields); j++ {
		switch field := fields[j].(type) {
		case IDField:
			statement += "'" + field.value + "'"
		case IntegerField:
			statement += fmt.Sprint(field.value)
		case TextField:
			statement += "'" + field.value + "'"
		}
		if j+1 != len(fields) {
			statement += ","
		}
	}
	statement += ");"
	log.Statement(statement)
	_, err := Manager.DB.Exec(statement)
	return err
}
