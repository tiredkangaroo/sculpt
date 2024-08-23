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

func NoColumnAssociatedWithFieldInSchema(modelName string, modelFieldName string) error {
	return fmt.Errorf("we could not find a column associated with %s in model %s", modelFieldName, modelName)
}

func ModelTypeMismatch(modelType string, got string) error {
	return fmt.Errorf("model is formed with type %s not %s", modelType, got)
}

func FieldTypeMismatch(fieldName string, requiredType string) error {
	return fmt.Errorf("field %s requires type %s but was given the wrong type.", fieldName, requiredType)
}

func FieldTypeMismatchGot(fieldName string, requiredType string, gotType string) error {
	return fmt.Errorf("field %s requires type %s but was given type %s.", fieldName, requiredType, gotType)
}

func ReadFile(filename string, error error) error {
	return fmt.Errorf("reading file %s failed with: %s.", filename, error.Error())
}
func ValidatorIsNil(validatorName string, columnName string) error {
	return fmt.Errorf("validator %s is nil. cannot validate column %s.", validatorName, columnName)
}

func ValidatorDoesNotExist(vn string, columnName string) error {
	return fmt.Errorf("column %s has validator %s but it is not registered.", columnName, vn)
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
func RowsAreEmpty() error {
	return fmt.Errorf("while querying the sql database, the pointer to rows was nil.")
}
func UnknownTypeFromDatabase(_type string) error {
	return fmt.Errorf("while parsing the sql database, we got an unknown type %s", _type)
}

func StructNotInRegistry(operation string, structName string) error {
	return fmt.Errorf("attempted to %s with unregistered model %s", operation, structName)
}

func QueryOnColumnThatDoesNotExist(modelName string, columnName string) error {
	return fmt.Errorf("attempted to query on column %s, but it does not exist on model %s", columnName, modelName)
}
