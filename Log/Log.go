package log

import (
	"fmt"
)

var BOLD = "\033[1m"
var NORMAL = "\033[0m"

var REDBWHITEF = "\033[37;101m"

func Statement(query string) {
	fmt.Printf("%sDatabase:%s %s\n", BOLD, NORMAL, query)
}
func Error(error string) {
	fmt.Printf("%s%sERROR%s: %s\n", REDBWHITEF, BOLD, NORMAL, error)
}
