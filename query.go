package sculpt

import "fmt"

type Query struct {
	// Columns specifies which columns to return.
	Columns []string

	// Distinct specifies if the columns should be unique.
	Distinct bool

	// Conditions specifies the conditions for the rows being returned.
	Conditions []Condition
}

func EqualTo(name string, value any) Condition {
	v, err := AnyToSQLString(value)
	if err != nil {
		panic(err.Error())
	}
	return fmt.Sprintf("%s = %s", name, v)
}

func GreaterThan(name string, value any) Condition {
	v, err := AnyToSQLString(value)
	if err != nil {
		panic(err.Error())
	}
	return fmt.Sprintf("%s > %s", name, v)
}

func LessThan(name string, value any) Condition {
	v, err := AnyToSQLString(value)
	if err != nil {
		panic(err.Error())
	}
	return fmt.Sprintf("%s < %s", name, v)
}

func GreaterEqualOrEqualTo(name string, value any) Condition {
	v, err := AnyToSQLString(value)
	if err != nil {
		panic(err.Error())
	}
	return fmt.Sprintf("%s >= %s", name, v)
}

func LessThanOrEqualTo(name string, value any) Condition {
	v, err := AnyToSQLString(value)
	if err != nil {
		panic(err.Error())
	}
	return fmt.Sprintf("%s <= %s", name, v)
}

func NotEqualTo(name string, value any) Condition {
	v, err := AnyToSQLString(value)
	if err != nil {
		panic(err.Error())
	}
	return fmt.Sprintf("%s <> %s", name, v)
}

func Between(name string, range1 any, range2 any) Condition {
	v, err := AnyToSQLString(range1)
	if err != nil {
		panic(err.Error())
	}
	v2, err := AnyToSQLString(range2)
	if err != nil {
		panic(err.Error())
	}
	return fmt.Sprintf("%s BETWEEN %s AND %s", name, v, v2)
}

func Like(name string, value string) Condition {
	return fmt.Sprintf("%s LIKE %s", name, value)
}

func In(name string, value ...any) Condition {
	statement := fmt.Sprintf("%s IN (", name)
	for i, val := range value {
		v, err := AnyToSQLString(val)
		if err != nil {
			panic(err.Error())
		}
		statement += v
		if i+1 < len(value) {
			statement += ", "
		}
	}
	statement += ")"
	return statement
}
