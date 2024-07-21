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

func FieldTypeMismatch(fieldName string, requiredType string) error {
	return fmt.Errorf("field %s requires type %s but was given the wrong type.", fieldName, requiredType)
}

func ReadFile(filename string, error error) error {
	return fmt.Errorf("reading file %s failed with: %s.", filename, error.Error())
}
