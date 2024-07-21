package sculpt

import (
	"fmt"
	"os"
)

var defined_models map[string]Model

func LoadModels(filename string) error {
	contents, err := os.ReadFile(filename)
	if err != nil {
		return ReadFile(filename, err)
	}
	fmt.Println(string(contents))
	return nil
}
