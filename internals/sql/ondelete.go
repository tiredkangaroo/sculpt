package sql

type OnDelete uint8

const (
	CASCADE OnDelete = iota
	SETNULL
	RESTRICT
	NOACTION
)

func OnDeleteFromString(s string) OnDelete {
	switch s {
	case "CASCADE":
		return CASCADE
	case "SET NULL":
		return SETNULL
	case "RESTRICT":
		return RESTRICT
	case "NO ACTION":
		return NOACTION
	default:
		return NOACTION
	}
}

func (o OnDelete) String() string {
	switch o {
	case CASCADE:
		return "CASCADE"
	case SETNULL:
		return "SET NULL"
	case RESTRICT:
		return "RESTRICT"
	case NOACTION:
		return "NO ACTION"
	default:
		return ""
	}
}
