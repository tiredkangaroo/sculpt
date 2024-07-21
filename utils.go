package sculpt

import (
	"fmt"
	"math/rand/v2"
)

// GenerateUID generates a random unique id in format LLLL-NNNN. NOTE: THIS SYSTEM IS NOT CRYPTO GRAPHICALLY SECURE.
func GenerateUID() string {
	id := ""
	for range 10 {
		r := rand.IntN(25) + 1
		id += fmt.Sprintf("%c", ('A' - 1 + r))
	}
	id += fmt.Sprintf("%02d", rand.IntN(26))
	return id
}
