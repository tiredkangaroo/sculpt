## Sculpt is a query-builder for go.
Sculpt builds queries to interact with raw PostgreSQL so you don't have to.

# Usage
The following provides documentation for the usage of Sculpt.
## Connections

Connecting to the Postgres Server:

```golang
err error := sculpt.Connect(username, password, database_name)
```

Disconnecting with the Postgres Server:

```golang
sculpt.Disconnect()
```

Checking the connection status:

```golang
connected bool := sculpt.Connected()
```
## Models
Models implement tables in PostgreSQL in go.

## Writing a Model

```golang
type ExampleModel struct {
  ID   string `kind:"IDField"` // must specify
  Name string // this defaults to TextField because the type is string
}
```
Kinds:
| Kind | Description |
|-|-|
| IDField | string with maxlength 32. auto-populated if left empty. |
| TextField| string with maxlength 4096. |
|IntegerField| integers only. |

At this point, you have only created the schema. You may want to create
the model now.

```golang
exampleModel := sculpt.NewModel(new(ExampleModel))
```

The NewModel function does not return an error, instead it panics in the
rare case there is an error.

panic messages and what they mean:

| Message | Meaning |
| ------ | ------|
| schema must be POINTER to struct | whatever you passed in is not a pointer |
| schema must be pointer to STRUCT | you passed in a pointer, but it wasn't a pointer to a struct |
| type for IDField must be string | you set the kind to IDField but the type of the structfield was not a string|
| type for TextField must be string | you set the kind to TextField but the type of the structfield was not a string|
| type for IntegerField must be int, int8, int16, int32, int64 | you set the kind to IntegerField but the type of the structfield was not any of the int's.|

After creating your model, you may want to save it (as a table) in postgres.
