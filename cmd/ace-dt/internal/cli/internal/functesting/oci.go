package functesting

import (
	"math/rand"
	"strings"
)

// TempOCIRef creates a random OCI reference.
func TempOCIRef(registry string) string {
	r := rand.NewSource(rand.Int63())

	return registry + "/tmp/" + strings.ToLower(GenRandString(r, 8)) + ":" + strings.ToLower(GenRandString(r, 3))
}
