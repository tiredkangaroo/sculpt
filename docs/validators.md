# Sculpt Validators

Sculpt validators are used to validate the input data before
insertion into the database. It is specified using the `validators`
struct tag.

## Defining a Validator
You may register a validator using the `sculpt.RegisterValidator`
function.

The function takes in a name and a function.

### The Function
The function must have one or more input parameters, and one
output. The type of the output must be `error`.

#### The First Parameter
The first parameter of the validator function should be the
value to be validated.

When calling `sculpt.RegisterValidator`, the first
parameter will be validated to ensure it is a supported
type for a field on a `sculpt.Model`.

#### The Other Parameters
The next parameters are the arguments to the validator.
These are specified in the `validators` struct tag.

There are only a couple supported type for arguments other than
the first.

##### Supported Types for Arguments Other Than The first
- `string`, `bool`, `int`, `uint`, `float64`

## The `validators` Struct Tag

The `validators` struct tag is used to specify the validators
to be used for a field, and their arguments.

### Syntax
`validators:"v1:arg1,arg2, v2, v3:arg1"`

This will apply the validators `v1`, `v2`, and `v3` to the field.

- `v1` will be called with the value of the field, and the arguments
`arg1` and `arg2`.

- `v2` will be called with the value of the field, and no arguments.

- `v3` will be called with the value of the field, and the argument
`arg1`.

## Rules
- It is mandatory to specify a validator that has been registered,
therefore, it is recommended to register all validators before
create a `sculpt.Model` using `sculpt.New`.

- The arguments to a validator must be able to be made into the
type specified in the function signature of the validator.

- The order of the arguments in the `validators` struct tag must
match the order of the arguments in the validator function signature.
