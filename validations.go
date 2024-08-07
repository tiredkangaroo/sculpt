package sculpt

type Validator[T any] struct {
	Kind Field
	Func func(T) (bool, error)
}

var registeredValidators = make(map[string]Validator[any])
