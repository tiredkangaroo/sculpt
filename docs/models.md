# Sculpt Models

Sculpt models provide a type-safe abstraction over Postgres
tables. They are defined using a simple struct, and passed
in as a generic argument to `sculpt.New`.

## Supported Types

The following types are supported:
| Type                                                  | Postgres Type |
| ----                                                  | ------------- |
| int16                                                 | `smallint`    |
| int32                                                 | `integer`     |
| int, int64                                            | `bigint`      |
| float32                                               | `real`        |
| float64                                               | `double`      |
| string                                                | `text`        |
| bool                                                  | `boolean`     |
| []byte                                                | `bytea`       |
| time.Time                                             | `timestamptz` |
| time.Duration                                         | `interval`    |
| uuid.UUID (from https://github.com/google/uuid)       | `uuid`        |


## Model Tags

Models can be tagged with the following tags:

`pk`: "true" | "false" (default: "false")
    - Indicates that the field is a primary key.

`autoincrement`: "true" | "false" (default: "false")
    - Indicates that the value of this field should
    be automatically incremented for every new insertion.

`unique`: "true" | "false" (default: "false")
    - Indicates that the field should have a unique
    constraint. A save operation will fail if the
    constraint is violated.
