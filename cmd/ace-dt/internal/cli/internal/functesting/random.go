package functesting

import (
	"math/rand"
	"strings"
)

const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// GenRandString returns a random string of variable length (from low to hig in length).
func GenRandString(r rand.Source, length int) string {
	var b strings.Builder
	b.Grow(length)
	rng := rand.New(r)

	for i := 0; i < length; i++ {
		c := alphabet[rng.Intn(len(alphabet))]
		if err := b.WriteByte(c); err != nil {
			panic(err)
		}
	}

	return b.String()
}
