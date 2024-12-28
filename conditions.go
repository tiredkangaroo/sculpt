package sculpt

import (
	"fmt"
)

// Condition represents a condition that can be used in a query.
type Condition struct {
	s string
	a []any
}

// EqualsTo returns a Condition that is true when the value of the column
// is equal to the given value.
func EqualsTo(name string, v any) Condition {
	c := Condition{
		// <_sculpt> keeps a placeholder in order to replace instances
		// of the substring with an integer, that then gets replaced
		// by pgx with v
		s: fmt.Sprintf("%s = $<_sculpt>", name),
		a: []any{v},
	}
	return c
}

// LessThan returns a Condition that is true when the value of the column
// is less than the given value.
func LessThan(name string, v any) Condition {
	c := Condition{
		s: fmt.Sprintf("%s < $<_sculpt>", name),
		a: []any{v},
	}
	return c
}

// LessThanOrEqualTo returns a Condition that is true when the value of the column
// is less than or equal to the given value.
func LessThanOrEqualTo(name string, v any) Condition {
	c := Condition{
		s: fmt.Sprintf("%s <= $<_sculpt>", name),
		a: []any{v},
	}
	return c
}

// GreaterThan returns a Condition that is true when the value of the column
// is greater than the given value.
func GreaterThan(name string, v any) Condition {
	c := Condition{
		s: fmt.Sprintf("%s > $<_sculpt>", name),
		a: []any{v},
	}
	return c
}

// GreaterThanOrEqualTo returns a Condition that is true when the value of the column
// is greater than or equal to the given value.
func GreaterThanOrEqualTo(name string, v any) Condition {
	c := Condition{
		s: fmt.Sprintf("%s >= $<_sculpt>", name),
		a: []any{v},
	}
	return c
}

// NotEqualsTo returns a Condition that is true when the value of the column
// is not equal to the given value.
func NotEqualsTo(name string, v any) Condition {
	c := Condition{
		s: fmt.Sprintf("%s <> $<_sculpt>", name),
		a: []any{v},
	}
	return c
}

// Like returns a Condition that is true when the value of the column
// is LIKE the given value.
func Like(name string, v any) Condition {
	c := Condition{
		s: fmt.Sprintf("%s LIKE $<_sculpt>", name),
		a: []any{v},
	}
	return c
}

// Between returns a Condition that is true when the value of the column
// is between the two given values.
func Between(name string, v1 any, v2 any) Condition {
	c := Condition{
		s: fmt.Sprintf("%s BETWEEN $<_sculpt> AND $<_sculpt>", name),
		a: []any{v1, v2},
	}
	return c
}

// In returns a Condition that is true when the value of the column
// is in the given values.
func In(name string, values ...any) Condition {
	c := Condition{
		s: fmt.Sprintf("%s IN (", name),
	}
	for i, v := range values {
		c.s += "$<_sculpt>"
		if i < len(values)-1 {
			c.s += ", "
		}
		c.a = append(c.a, v)
	}
	c.s += ")"
	return c
}

// Or returns a Condition that is the result of combining two Conditions where either
// of the two must be true for the combined Condition to be true.
func Or(c1 Condition, c2 Condition) Condition {
	c := Condition{
		s: fmt.Sprintf("%s OR %s", c1.s, c2.s),
		a: append(c1.a, c2.a...),
	}
	return c
}

// Not returns a Condition whose results is opposite that of the given Condition.
func Not(c Condition) Condition {
	return Condition{
		s: fmt.Sprintf("NOT %s", c.s),
		a: c.a,
	}
}
