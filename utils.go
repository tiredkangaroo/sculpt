package sculpt

import (
	"fmt"
	"math/rand/v2"
	"reflect"
	"strconv"
	"strings"
)

// GenerateUID generates a random unique id in format LLLL-NNNN. NOTE: THIS SYSTEM IS NOT CRYPTO GRAPHICALLY SECURE.
func generateUID() string {
	id := ""
	for range 10 {
		r := rand.IntN(25) + 1
		id += fmt.Sprintf("%c", ('A' - 1 + r))
	}
	id += fmt.Sprintf("%02d", rand.IntN(26))
	return id
}

func boolFromTag(tag_name string, tag reflect.StructTag, _default bool) bool {
	tag_value := tag.Get(tag_name)
	if tag_value == "" {
		return _default
	}
	v, err := strconv.ParseBool(tag_value)
	if err != nil {
		panic(tag_name + " field must be bool")
	}
	return v
}

func uintFromTag(tag_name string, tag reflect.StructTag, _default uint) uint {
	tag_value := tag.Get(tag_name)
	if tag_value == "" {
		return _default
	}
	v, err := strconv.ParseUint(tag_value, 10, 64)
	if err != nil {
		panic(tag_name + " field must be int")
	}
	return uint(v)
}

func arrayFromTag(tag_name string, tag reflect.StructTag) []string {
	tag_value := tag.Get(tag_name)
	if tag_value == "" {
		return []string{}
	}

	return strings.Split(tag_value, ",")
}

func kindToSQL(k Field) (string, error) {
	switch k.String() {
	case "IntegerField":
		return "int", nil
	case "TextField":
		return "varchar(4096)", nil
	default:
		return "", fmt.Errorf("kind is not recognized")
	}
}

func kindToRefectKind(k Field) reflect.Kind {
	switch k.String() {
	case "IntegerField":
		return reflect.Int
	case "TextField":
		return reflect.String
	}
	return reflect.TypeOf(k).Kind()
}

func sliceContains[T comparable](slice []T, item T) bool {
	for _, ti := range slice {
		if ti == item {
			return true
		}
	}
	return false
}

func compareSlices(slice1 []reflect.StructField, slice2 []reflect.StructField) (additions, deletions []reflect.StructField) {
	additionsMap := make(map[*reflect.StructField]struct{})
	deletionsMap := make(map[*reflect.StructField]struct{})

	for _, item := range slice1 {
		deletionsMap[&item] = struct{}{}
	}

	for _, item := range slice2 {
		it := &item
		if _, found := deletionsMap[it]; found {
			delete(deletionsMap, it)
		} else {
			additionsMap[it] = struct{}{}
		}
	}

	for item := range additionsMap {
		additions = append(additions, *item)
	}

	for item := range deletionsMap {
		deletions = append(deletions, *item)
	}

	return
}

func extraColumnProperties(column Column) (statement string) {
	if column.PRIMARY_KEY {
		statement += " PRIMARY KEY"
	}
	if column.UNIQUE && !column.PRIMARY_KEY { // primary key is always unique and not null
		statement += " UNIQUE"
	}
	if !column.NULLABLE && !column.PRIMARY_KEY {
		statement += " NOT NULL"
	}
	return
}

func anyToSQLString(value any) (string, error) {
	switch v := value.(type) {
	case string:
		return fmt.Sprintf("'%s'", v), nil
	case int:
		return fmt.Sprintf("%d", v), nil
	default:
		return "", fmt.Errorf("type not supported")
	}
}

func buildWhere(query Query) (statement string) {
	if len(query.Conditions) != 0 {
		statement += ` WHERE `
		for i, c := range query.Conditions {
			statement += c
			if i+1 < len(query.Conditions) {
				statement += ` AND `
			}
		}
	}
	return
}
