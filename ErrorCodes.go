package sculpt

import (
	"errors"
	"fmt"
)

var ErrorCodes = map[string]string{
	"T90R": "%s Missing Field for Population: %s",
	"R41P": "%s Incorrect Type for Population of Field: %s",
	"L35Z": "%s RequiredConnectionError: There must be connected database to run the %s operation.",
}

func ErrorFromCode(c string, addi string) error {
	return errors.New(fmt.Sprintf(ErrorCodes[c], c, addi))
}
