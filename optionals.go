package sculpt

import "reflect"

// Optional is used for columns that can be nil.
type Optional[T any] struct {
	v T
	a bool
}

// Nil returns whether the optional is nil.
func (o Optional[T]) Nil() bool {
	return !o.a
}

// Value returns the value of the optional.
func (o Optional[T]) Value() T {
	if o.Nil() {
		panic("optional is nil")
	}
	return o.v
}

func (o *Optional[T]) Set(v T) {
	o.a = false
	o.v = v
}

// OptionalValue returns a non-nil optional value.
func OptionalValue[T any](v T) Optional[T] {
	return Optional[T]{v: v, a: false}
}

// isOptional checks if a type is an Optional.
func isOptional(t reflect.Type) bool {
	// Ensure t is a struct
	if t.Kind() != reflect.Struct {
		return false
	}

	// Verify the struct has exactly two fields
	if t.NumField() != 2 {
		return false
	}

	// Check fields "v" and "a" exist and have the expected types
	vField := t.Field(0)
	aField := t.Field(1)
	if vField.Name != "v" || aField.Name != "a" || aField.Type.Kind() != reflect.Bool {
		return false
	}

	// Check the methods of the struct
	nilMethod, ok1 := t.MethodByName("Nil")
	if !ok1 || nilMethod.Type.NumIn() != 1 || nilMethod.Type.NumOut() != 1 || nilMethod.Type.Out(0).Kind() != reflect.Bool {
		return false
	}

	valueMethod, ok2 := t.MethodByName("Value")
	if !ok2 || valueMethod.Type.NumIn() != 1 || valueMethod.Type.NumOut() != 1 {
		return false
	}

	setMethod, ok3 := t.MethodByName("Set")
	if !ok3 || setMethod.Type.NumIn() != 2 || setMethod.Type.NumOut() != 0 {
		return false
	}

	// If all checks pass, it matches the Optional[T] pattern
	return true
}

func optionalType(t reflect.Type) reflect.Type {
	return t.Method(1).Type.Out(0)
}
