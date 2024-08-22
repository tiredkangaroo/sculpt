package sculpt

import (
	"fmt"
	"regexp"
	"strings"
)

type Validator struct {
	Kind Field
	Func func(any) error
}

func RegisterValidator(name string, v Validator) {
	registeredValidators[name] = v
}

var registeredValidators = map[string]Validator{
	"email": EmailValidator,
	"password": NewPasswordValidator(
		12,
		255,
		true,
		true,
		true,
	),
}

// Standard validators
var EmailValidator = Validator{
	Kind: TextField{},
	Func: func(a any) error {
		email := a.(string)
		parts := strings.Split(email, "@")
		if len(parts) != 2 {
			return fmt.Errorf("there must be only one @")
		}
		part2 := parts[1]

		part2Parts := strings.Split(part2, ".")
		if len(part2Parts) <= 1 {
			return fmt.Errorf("a username and domain must be specfied")
		}

		return nil
	},
}

func NewPasswordValidator(
	minimumLength uint16,
	maximumLength uint16,
	requireCapitalLetter bool,
	requireNumber bool,
	requireSpecialCharacter bool,
) Validator {
	hasNumberRC := regexp.MustCompile(`\d`)
	hasSpecialCharacterRC := regexp.MustCompile(`[!@#$%^&*()]`)
	hasCapitalLetterRC := regexp.MustCompile(`[A-Z]`)
	return Validator{
		Kind: TextField{},
		Func: func(value any) error {
			rawpassword := value.(string)
			rawPasswordLen := len(rawpassword)
			if rawPasswordLen < int(minimumLength) {
				return fmt.Errorf("password must be %d or more characters", minimumLength)
			}
			if rawPasswordLen > int(maximumLength) {
				return fmt.Errorf("password must not be greater than %d characters", maximumLength)
			}
			if requireNumber {
				hasNumber := hasNumberRC.MatchString(rawpassword)
				if !hasNumber {
					return fmt.Errorf("password must have at least one number")
				}
			}
			if requireSpecialCharacter {
				hasSpecialCharacter := hasSpecialCharacterRC.MatchString(rawpassword)
				if !hasSpecialCharacter {
					return fmt.Errorf("password must have at least one special character")
				}
			}
			if requireCapitalLetter {
				hasCapitalLetter := hasCapitalLetterRC.MatchString(rawpassword)
				if !hasCapitalLetter {
					return fmt.Errorf("password must have at least one capital letter")
				}
			}
			return nil
		},
	}
}
