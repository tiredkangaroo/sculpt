package sculpt

import (
	"fmt"
)

var BOLD = "\033[1m"
var NORMAL = "\033[0m"

var REDBWHITEF = "\033[37;101m"

var Logging bool = true

func On() {
	Logging = true
}
func Off() {
	Logging = false
}

func Statement(query string) {
	if Logging == true {
		fmt.Printf("%sDatabase:%s %s\n", BOLD, NORMAL, query)
	}
}
func Error(error string) {
	fmt.Printf("%s%sERROR%s: %s\n", REDBWHITEF, BOLD, NORMAL, error)
}
