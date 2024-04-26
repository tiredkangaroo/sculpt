## Sculpt is a query-builder for go.
Sculpt builds queries to interact with raw PostgreSQL so you don't have to.

# Usage
The following provides documentation for the usage of Sculpt.
## Connections

Connecting to the Postgres Server:

`err := sculpt.Manager.Connect(username, password, database_name)`

Disconnecting to the Postgres Server:

`err := sculpt.Manager.Disconnect()`

## Supported Fields

### ID Field
`sculpt.models.IDField{}`

ID Fields are required to have a maximum of 12 digits.

```
PRIMARY_KEY bool
UNIQUE      bool //optional, default: false
Name        string
Auto        bool   //auto-populate with a random id?
```
### Text Field
`sculpt.models.TextField{}`

Options:
```
PRIMARY_KEY    bool   //optional, default: false
UNIQUE         bool   //optional, default: false
Name           string
Minimum_Length int   //optional
Maximum_Length int   //optional
```
### Integer Field
`sculpt.models.IntegerField{}`

Options:
```
PRIMARY_KEY bool //optional, default: false
UNIQUE      bool //optional, default: false
Name        string
```

## Operations
### Creating a Model Instance:

```
myModelInstance := sculpt.models.Model{
  Name: "your_model_name"
  Fields: []interface{}{
    //your fields here
  }
}

```

### Creating a Table from a Model Instance:

```
err := myModelInstance.Create()
```

Options:

(REQUIRED) `ifNotExists bool` 

Create the table in the database only if it does not already exist.

### Creating a Row from a Model Instance:
```
newRow, err := myModelInstance.New()
```

Options:
(REQUIRED) Every field in order must be passed into the function.


### Saving a New Row:

`err := newRow.Save()`

### Querying a Row from a Model Instance:

```
myModelInstance.Get(sculpt.models.Query{})
```
Options for `models.Query{}`:

If empty, it will return everything.


DISTINCT `bool`

- will return only distinct values for the first column.

- if you do not pass in a column into Columns AND distinct is true, it will return all distinct rows

Columns  `[]string` 

- what columns to return in the map (first column is used for distinct if DISTINCT is true)

Where    `map[string]interface{}` 

- return the rows where the value for a column is what is specified (only exact matches) <a href="https://www.w3schools.com/sql/sql_where.asp">Explanation</a>

Order_By `string`

- order by the name of the column

### Deleting a Row from a Model Instance:

```
myModelInstance.Delete()
```
If empty, it will delete everything.

Options:

w `map[string]interface{}`

- delete where column name (key) is equal to value (value)

#### WARNING:
SANITIZE ALL USER INPUTS. SCULPT DOES NOT SANITIZE INPUTS.
