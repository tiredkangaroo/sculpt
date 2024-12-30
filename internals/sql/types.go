package sql

import (
	"reflect"
	"time"

	"github.com/google/uuid"
)

type Type uint8

const (
	InvalidType Type = iota
	// SmallintType represents the smallint type in PostgreSQL.
	SmallintType
	// IntegerType represents the integer type in PostgreSQL.
	IntegerType
	// BigintType represents the bigint type in PostgreSQL.
	BigintType
	// SerialType represents the serial type in PostgreSQL.
	SerialType
	// SmallSerialType represents the smallserial type in PostgreSQL.
	SmallSerialType
	// BigSerialType represents the bigserial type in PostgreSQL.
	BigSerialType
	// RealType represents the single precision type in PostgreSQL.
	RealType
	// DoubleType represents the double precision type in PostgreSQL.
	DoubleType
	// TextType represents the text type in PostgreSQL.
	TextType
	// ByteaType represents the bytea type in PostgreSQL.
	ByteaType
	// BooleanType represents the boolean type in PostgreSQL.
	BooleanType
	// TimestampType represents the timestamp type in PostgreSQL.
	TimestampType
	// IntervalType represents the interval type in PostgreSQL.
	IntervalType
	// UUIDType represents the UUID type in PostgreSQL.
	UUIDType
)

func (t Type) String() string {
	switch t {
	case SmallintType:
		return "smallint"
	case IntegerType:
		return "integer"
	case BigintType:
		return "bigint"
	case SerialType:
		return "serial"
	case SmallSerialType:
		return "smallserial"
	case BigSerialType:
		return "bigserial"
	case RealType:
		return "real"
	case DoubleType:
		return "double precision"
	case TextType:
		return "text"
	case ByteaType:
		return "bytea"
	case BooleanType:
		return "boolean"
	case TimestampType:
		return "timestamptz"
	case IntervalType:
		return "interval"
	case UUIDType:
		return "uuid"
	default:
		return "invalid"
	}
}

// ReflectType returns a reflect.Type that corresponds to the SQL type. It is
// a reverse of TypeFromReflectType. If the type is invalid, it returns nil.
func (t Type) ReflectType() reflect.Type {
	switch t {
	case SmallintType, SmallSerialType:
		return reflect.TypeFor[int16]()
	case IntegerType, SerialType:
		return reflect.TypeFor[int32]()
	case BigintType, BigSerialType:
		return reflect.TypeFor[int64]()
	case RealType:
		return reflect.TypeFor[float32]()
	case DoubleType:
		return reflect.TypeFor[float64]()
	case TextType:
		return reflect.TypeFor[string]()
	case BooleanType:
		return reflect.TypeFor[bool]()
	case ByteaType:
		return reflect.TypeFor[[]byte]()
	case TimestampType:
		return reflect.TypeFor[time.Time]()
	case IntervalType:
		return reflect.TypeFor[time.Duration]()
	case UUIDType:
		return reflect.TypeFor[uuid.UUID]()
	default:
		return nil
	}
}

// TypeFromReflectType returns the SQL type that corresponds to the reflect.Type.
// If serial is true (psql SERIAL), it will return the serial type that corresponds
// to the reflect.Type. If the reflect.Type is not supported, it returns InvalidType.
func TypeFromReflectType(t reflect.Type, serial bool) Type {
	if serial {
		switch t.Kind() {
		case reflect.Int16:
			return SmallSerialType
		case reflect.Int32:
			return SerialType
		case reflect.Int, reflect.Int64:
			return BigSerialType
		default:
			return InvalidType
		}
	}
	switch t.Kind() {
	case reflect.Int16:
		return SmallintType
	case reflect.Int32:
		return IntegerType
	case reflect.Int, reflect.Int64:
		return BigintType
	case reflect.Float32:
		return RealType
	case reflect.Float64:
		return DoubleType
	case reflect.String:
		return TextType
	case reflect.Bool:
		return BooleanType
	}
	switch t {
	case reflect.TypeFor[[]byte]():
		return ByteaType
	case reflect.TypeFor[time.Time]():
		return TimestampType
	case reflect.TypeFor[time.Duration]():
		return IntervalType
	case reflect.TypeFor[uuid.UUID]():
		return UUIDType
	}
	return InvalidType
}
