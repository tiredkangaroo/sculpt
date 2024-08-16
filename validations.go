package sculpt

import (
	"fmt"
	"strings"
)

type Validator struct {
	Kind Field
	Func func(any) (bool, error)
}

func RegisterValidator(name string, v Validator) {
	registeredValidators[name] = v
}

var registeredValidators = make(map[string]Validator)

// Standard validators
var EmailValidator = Validator{
	Kind: TextField{},
	Func: func(a any) (bool, error) {
		email := a.(string)
		parts := strings.Split(email, "@")
		if len(parts) != 2 {
			return false, fmt.Errorf("there must be only one @")
		}
		part2 := parts[1]

		part2Parts := strings.Split(part2, ".")
		if len(part2Parts) <= 1 {
			return false, fmt.Errorf("a username and domain must be specfied")
		}

		return true, nil
	},
}
