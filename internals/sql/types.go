package sql

import (
	"reflect"

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
	case UUIDType:
		return "uuid"
	default:
		return "invalid"
	}
}

func TypeFromGoType[T any](serial bool) Type {
	t := reflect.TypeFor[T]()
	return TypeFromReflectType(t, serial)
}

func ToReflectType(t Type) reflect.Type {
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
	case UUIDType:
		return reflect.TypeFor[uuid.UUID]()
	default:
		return nil
	}
}

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
	case reflect.TypeFor[uuid.UUID]():
		return UUIDType
	}
	return InvalidType
}
