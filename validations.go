package sculpt

type Validator[T Field] struct {
	Kind T
	Func func(T) (bool, error)
}

var registeredValidators = make(map[string]Validator[Field])
