package sculpt

import (
	"database/sql"
	"reflect"
)

// Optional is used for columns that can be nil.
type Optional[T any] struct {
	n sql.Null[T]
}

// Nil returns whether the optional is nil.
func (o Optional[T]) Nil() bool {
	return !o.n.Valid
}

// Value returns the value of the optional. It panics if the optional
// is nil.
func (o Optional[T]) Value() T {
	if o.Nil() {
		panic("optional is nil")
	}
	return o.n.V
}

func (o *Optional[T]) Set(v T) {
	o.n.Valid = true
	o.n.V = v
}

func (o *Optional[T]) Scan(v any) error {
	return o.n.Scan(v)
}

// OptionalValue returns a non-nil optional value.
func OptionalValue[T any](v T) Optional[T] {
	return Optional[T]{n: sql.Null[T]{V: v, Valid: true}}
}

// isOptional checks if a type is an Optional.
func isOptional(t reflect.Type) bool {
	// Ensure t is a struct
	if t.Kind() != reflect.Struct {
		return false
	}

	// Verify the struct has exactly two fields
	if t.NumField() != 1 {
		return false
	}

	// Check fields "v" and "a" exist and have the expected types
	nField := t.Field(0)
	if nField.Name != "n" && nField.Type.Kind() != reflect.Struct && nField.Type.Implements(reflect.TypeOf((*sql.Scanner)(nil)).Elem()) {
		return false
	}

	// Check the methods of the struct
	nilMethod, ok1 := t.MethodByName("Nil")
	if !ok1 || nilMethod.Type.NumIn() != 1 || nilMethod.Type.NumOut() != 1 || nilMethod.Type.Out(0).Kind() != reflect.Bool {
		return false
	}

	// why is there a parameter in for Value? because the Value method has
	// a generic parameter
	valueMethod, ok2 := t.MethodByName("Value")
	if !ok2 || valueMethod.Type.NumIn() != 1 || valueMethod.Type.NumOut() != 1 {
		return false
	}

	// why reflect.PointerTo? because the Set method is a pointer receiver
	// and we need to get the method from the pointer receiver
	//
	// why setMethod.Type.NumIn() != 2? because the Set method has two arguments
	// the value, and the generic
	setMethod, ok3 := reflect.PointerTo(t).MethodByName("Set")
	if !ok3 || setMethod.Type.NumIn() != 2 || setMethod.Type.NumOut() != 0 {
		return false
	}

	// If all checks pass, it matches the Optional[T] pattern
	return true
}

func optionalType(t reflect.Type) reflect.Type {
	return t.Method(1).Type.Out(0)
}
