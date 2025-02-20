package sculpt

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/tiredkangaroo/sculpt/internals/sql"
)

var validators = make(map[string]Validator)

// validator func: func (v T, a string) error

// Validator is a struct that provides methods to validate a value before insertion.
// If the value is a sculpt.Optional[T], the validator's input type must be T. If the value
// is nil, the validator will not be called.
type Validator struct {
	// f is the function that will be called to validate the value.
	f reflect.Value

	// t is the type of the value that will be validated.
	t reflect.Type

	// p contains all the types for the parameters of the validator.
	p []reflect.Type
}

// UseFor returns whether the validator can be used for the given type.
func (v Validator) UseFor(t reflect.Type) bool {
	return v.t == t
}

// Validate validates a value using the validator.
func (va Validator) Validate(v any, a ...reflect.Value) error {
	in := make([]reflect.Value, len(a)+1)
	in[0] = reflect.ValueOf(v)
	for i, arg := range a {
		in[i+1] = reflect.ValueOf(arg)
	}

	values := va.f.Call(in)
	err := values[0].Interface()
	if err == nil {
		return nil
	}
	return err.(error)
}

// RegisterValidator registers a new validator with the given name. It is possible
// to register multiple validators with the same name, but the last one registered
// will be used.
func RegisterValidator(name string, f any) error {
	if f == nil {
		return fmt.Errorf("validator function cannot be nil")
	}

	// validate function
	t := reflect.TypeOf(f)
	if t.Kind() != reflect.Func {
		return fmt.Errorf("validator must be a function")
	}
	if t.NumIn() == 0 {
		return fmt.Errorf("validator function must have at least one parameter")
	}
	if sql.TypeFromReflectType(t.In(0), false) == sql.InvalidType {
		return fmt.Errorf(
			"unsupported type for validator: %s. the first parameter for a validator function handles a value for a sculpt column",
			t.In(0))
	}
	if t.NumOut() != 1 || t.Out(0) != reflect.TypeOf((*error)(nil)).Elem() {
		return fmt.Errorf("validator function must return an error")
	}

	// p
	p := make([]reflect.Type, t.NumIn()-1)
	for i := range t.NumIn() - 1 {
		p[i] = t.In(i + 1)
		switch p[i].Kind() {
		case reflect.String, reflect.Int, reflect.Bool, reflect.Float64, reflect.Uint:
			// do nothing
		default:
			return fmt.Errorf("unsupported type for validator argument: %s", p[i])
		}
	}

	v := Validator{
		f: reflect.ValueOf(f),
		t: t.In(0),
	}
	validators[name] = v
	return nil
}

func validatorsFromTag(fieldt reflect.Type, tag string) (map[*Validator][]reflect.Value, error) {
	if tag == "" {
		return nil, nil
	}

	// `validators: "email, min:5, max:10, somevalidator:a,b,c"`
	v := map[*Validator][]reflect.Value{}
	rawvalidators := strings.Split(tag, ", ")
	for _, rawvalidator := range rawvalidators {
		rs := strings.Split(rawvalidator, ":")
		var name string
		var args []string
		switch len(rs) {
		case 1:
			name = rs[0]
		case 2:
			name = rs[0]
			args = strings.Split(rs[1], ",")
		default:
			return nil, fmt.Errorf("improperly formed validator tag")
		}

		validator, ok := validators[name]
		if !ok {
			return nil, fmt.Errorf("validator %s not found", rawvalidator)
		}
		if !validator.UseFor(fieldt) {
			return nil, fmt.Errorf("validator %s cannot be used for type %s", rawvalidator, fieldt.String())
		}
		if len(args) != len(validator.p) {
			return nil, fmt.Errorf("validator %s requires %d arguments, %d given", rawvalidator, len(validator.p), len(args))
		}

		// correct the arguments to the correct types
		argsv := make([]reflect.Value, len(args))
		for i, varg := range validator.p {
			switch varg.Kind() {
			case reflect.String:
				argsv[i] = reflect.ValueOf(args[i])
			case reflect.Int:
				v, err := strconv.Atoi(args[i])
				if err != nil {
					return nil, fmt.Errorf("argument %d for validator %s must be an integer", i, name)
				}
				argsv[i] = reflect.ValueOf(v)
			case reflect.Bool:
				v, err := strconv.ParseBool(args[i])
				if err != nil {
					return nil, fmt.Errorf("argument %d for validator %s must be a boolean", i, name)
				}
				argsv[i] = reflect.ValueOf(v)
			case reflect.Float64:
				v, err := strconv.ParseFloat(args[i], 64)
				if err != nil {
					return nil, fmt.Errorf("argument %d for validator %s must be a float", i, name)
				}
				argsv[i] = reflect.ValueOf(v)
			case reflect.Uint:
				v, err := strconv.ParseUint(args[i], 10, 64)
				if err != nil {
					return nil, fmt.Errorf("argument %d for validator %s must be a uint", i, name)
				}
				argsv[i] = reflect.ValueOf(uint(v))
			}
		}
		v[&validator] = argsv
	}
	return v, nil
}
