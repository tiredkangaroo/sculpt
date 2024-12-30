package sculpt

import (
	"fmt"
	"reflect"

	"github.com/tiredkangaroo/sculpt/internals/sql"
)

// registeredPrimaryKeys stores a map of model names to their primary key columns.
// It is used in order to validate references to other models, ensuring the type of
// the reference in Postgres and the existence of a primary key column in the Model.
var registeredPrimaryKeys = make(map[string]Column) // model name: primary key column

// Column represents a column in a Sculpt model.
type Column struct {
	// name specifies the name of the column.
	name string

	// t is the runtime reflection type of the column.
	//
	// It is used when creating new instances of the model, so that the correct type
	// is set using reflect.Value.Set. If the column is an Optional[T], t remains as Optional[T],
	// since when using reflection to set the value, the field's type is still the Optional type.
	//
	// Note that it is possible to determine the reflected type of the column by checking the
	// sqltype. However, this is just a convenience field, since determining the value from sqltype
	// requires a call internals/sql.Type.ReflectType.
	t reflect.Type

	// sqltype is the SQL type of the column, as provided by internals/sql.TypeFromReflectType.
	sqltype sql.Type

	// primarykey specifies whether the column is a primary key. This information is obtained from
	// the struct tag "pk".
	primarykey bool

	// nullable specifies whether the column can be null. This information is obtained type of the
	// field used to create the column, and whether it is an Optional[T] (or at least follows the pattern
	// of an Optional[T], see isOptional for more information on the Optional[T] pattern).
	nullable bool

	// unique specifies whether the column is unique. This information is obtained from the struct tag "unique".
	unique bool

	// autoincrement specifies whether the column is autoincrement. This information is obtained from the struct
	// tag "autoincrement".
	autoincrement bool

	// ondelete specifies the ON DELETE action for the column. This information is obtained from the struct tag "ondelete".
	ondelete sql.OnDelete

	// validators is a map of validators for the column. The key is the validator, and the value is a slice of
	// reflect.Values that represent the validator input. This information is obtained from the struct tag "validators".
	validators map[*Validator][]reflect.Value
}

// handleColumn handles a struct field from a Sculpt model's struct
// and returns a Column.
func handleColumn(f reflect.StructField) (Column, error) {
	var err error
	c := Column{}
	// name
	c.name = f.Name

	// t
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
		return c, fmt.Errorf("unsupported type on column %s: %s. See documentation at: %s.", f.Name, f.Type.String(), DocsURL+"/columns.md")
	}

	// primary key
	if c.primarykey, err = boolFromString(f.Tag.Get("pk")); err != nil {
		return c, err
	}

	// unique
	if c.unique, err = boolFromString(f.Tag.Get("unique")); err != nil {
		return c, err
	}

	// validators
	if c.validators, err = validatorsFromTag(f.Type, f.Tag.Get("validators")); err != nil {
		return c, err
	}
	if len(c.validators) != 0 && c.autoincrement {
		return c, fmt.Errorf("cannot use validators on an autoincrement column")
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
