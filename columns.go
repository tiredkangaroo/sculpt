package sculpt

import (
	"fmt"
	"reflect"
	"sculpt/internals/sql"
)

// registeredPrimaryKeys stores a map of model names to their primary key columns.
// It is used in order to validate references to other models, ensuring the type of
// the reference in Postgres and the existence of a primary key column in the Model.
var registeredPrimaryKeys = make(map[string]Column) // model name: primary key column

// Column represents a column in a Sculpt model.
type Column struct {
	name    string
	t       reflect.Type
	sqltype sql.Type

	primarykey    bool
	nullable      bool
	unique        bool
	autoincrement bool

	ondelete sql.OnDelete
}

// handleColumn handles a struct field from a Sculpt model's struct
// and returns a Column.
func handleColumn(f reflect.StructField) (Column, error) {
	var err error
	c := Column{}
	// name
	c.name = f.Name
	c.t = f.Type

	// nullable/Optional
	if isOptional(f.Type) {
		c.nullable = true
		valuemethod, _ := f.Type.MethodByName("Value")
		f.Type = valuemethod.Type.Out(0) // the type of the Optional (using the Value method)
	}

	// autoincrement
	c.autoincrement, err = boolFromString(f.Tag.Get("autoincrement"))
	if err != nil {
		return c, err
	}

	// sqltype
	c.sqltype = sql.TypeFromReflectType(f.Type, c.autoincrement)
	if c.sqltype == sql.InvalidType {
		return c, fmt.Errorf("unsupported type on column %s: %s. See documentation at: %s.", f.Name, f.Type.String(), DocsURL+"/columns")
	}

	// primary key
	if c.primarykey, err = boolFromString(f.Tag.Get("pk")); err != nil {
		return c, err
	}

	// unique
	if c.unique, err = boolFromString(f.Tag.Get("unique")); err != nil {
		return c, err
	}

	return c, nil
}

// boolFromString converts a string to a boolean.
func boolFromString(s string) (bool, error) {
	switch s {
	case "true":
		return true, nil
	case "false", "":
		return false, nil
	default:
		return false, fmt.Errorf("not boolean value: %s", s)
	}
}
