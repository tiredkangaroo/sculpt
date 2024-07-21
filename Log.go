package sculpt

import (
	"fmt"
)

var BOLD = "\033[1m"
var NORMAL = "\033[0m"
var REDBWHITEF = "\033[37;101m"

var currentLogLevel LogLevel

type LogLevel int

const (
	DEBUG LogLevel = 0
	WARN  LogLevel = 1
	ERROR LogLevel = 2
	NOLOG LogLevel = 3
)

func LogStatement(query string) {
	if currentLogLevel <= DEBUG {
		fmt.Printf("%sDatabase:%s %s\n", BOLD, NORMAL, query)
	}
}

// SetLogLevel sets the log level from level argument. Panics
// if level is not a LogLevel.
func SetLogLevel(level LogLevel) {
	switch level {
	case DEBUG, WARN, ERROR, NOLOG:
		currentLogLevel = level
	default:
		panic("Log level must be DEBUG, WARN, ERROR, or NOLOG.")
	}
}

func LogError(error string) {
	if currentLogLevel <= ERROR {
		fmt.Printf("%s%sERROR%s: %s\n", REDBWHITEF, BOLD, NORMAL, error)
	}
}
