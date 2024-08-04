package sculpt

import (
	"fmt"
)

func MissingFieldForPopulation(fieldName string) error {
	return fmt.Errorf("missing field for population: %s", fieldName)
}

func OperationRequiresDatabaseConnection(operationName string) error {
	return fmt.Errorf("operation '%s' requires a database connection.", operationName)
}

func ModelTypeMismatch(modelType string, got string) error {
	return fmt.Errorf("model is formed with type %s not %s", modelType, got)
}
func FieldTypeMismatch(fieldName string, requiredType string) error {
	return fmt.Errorf("field %s requires type %s but was given the wrong type.", fieldName, requiredType)
}

func ReadFile(filename string, error error) error {
	return fmt.Errorf("reading file %s failed with: %s.", filename, error.Error())
}
func ValidatorIsNil(validatorName string, columnName string) error {
	return fmt.Errorf("validator %s is nil. cannot validate column %s.", validatorName, columnName)
}

func ValidatorHasNoFunc(validatorName string, columnName string) error {
	return fmt.Errorf("validator %s has no function attached. cannot validate column %s.", validatorName, columnName)
}

func ValidatorCannotBeUsedForKind(validatorName string, validatorKind interface{}, columnName string, columnKind interface{}) error {
	return fmt.Errorf("validator %s's %T kind cannot be used to validator column %s's %T kind", validatorName, validatorKind, columnName, columnKind)
}
func ValidationFailed(validatorName string, columnName string, value any, err error) error {
	return fmt.Errorf("validator %s failed population on column %s with value %s with error: %s", validatorName, columnName, value, err.Error())
}

func AnyToSQLString(value any) (string, error) {
	switch v := value.(type) {
	case string:
		return fmt.Sprintf("'%s'", v), nil
	case int:
		return fmt.Sprintf("%d", v), nil
	default:
		return "", fmt.Errorf("type not supported")
	}
}
