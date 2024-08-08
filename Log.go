package sculpt

import (
	"fmt"
)

var bold = "\033[1m"
var normal = "\033[0m"
var redbgWhiteF = "\033[41;37m"
var yellowbgBlackF = "\033[43;30m"
var greenbgBlackF = "\033[42;30m"
var bluebgWhiteF = "\033[44;37m"

var currentLogLevel LogLevel

type LogLevel int

const (
	DEBUG LogLevel = 0
	INFO  LogLevel = 1
	WARN  LogLevel = 2
	ERROR LogLevel = 3
	NOLOG LogLevel = 4
)

// SetLogLevel sets the log level from level argument. Panics
// if level is not a LogLevel.
func SetLogLevel(level LogLevel) {
	switch level {
	case DEBUG, INFO, WARN, ERROR, NOLOG:
		currentLogLevel = level
	default:
		panic("Log level must be DEBUG, WARN, ERROR, or NOLOG.")
	}
}

func cprint(c string, r string, s string, a ...any) {
	fmt.Println(c+bold+r+normal, fmt.Sprintf(s, a...)+normal)
}

func LogDebug(application string, s string, a ...any) {
	if currentLogLevel <= DEBUG {
		cprint(bluebgWhiteF, application, s, a...)
	}
}

func LogInfo(application string, s string, a ...any) {
	if currentLogLevel <= INFO {
		cprint(bluebgWhiteF, application, s, a...)
	}
}

func LogWarn(s string, a ...any) {
	if currentLogLevel <= WARN {
		cprint(yellowbgBlackF, "WARNING", s, a...)
	}
}

func LogError(s string, a ...any) {
	if currentLogLevel <= WARN {
		cprint(redbgWhiteF, "ERROR", s, a...)
	}
}
