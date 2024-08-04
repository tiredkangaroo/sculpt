package sculpt

import (
	"fmt"
	"math/rand/v2"
	"reflect"
	"strconv"
	"strings"
)

// GenerateUID generates a random unique id in format LLLL-NNNN. NOTE: THIS SYSTEM IS NOT CRYPTO GRAPHICALLY SECURE.
func GenerateUID() string {
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

func arrayFromTag(tag_name string, tag reflect.StructTag) []string {
	tag_value := tag.Get(tag_name)
	if tag_value == "" {
		return []string{}
	}

	return strings.Split(tag_value, ",")
}

func kindToSQL(k any) (string, error) {
	switch k.(type) {
	case IDField:
		return "varchar(32)", nil
	case IntegerField:
		return "int", nil
	case TextField:
		return "varchar(4096)", nil
	default:
		return "", fmt.Errorf("kind is not recognized")
	}
}
