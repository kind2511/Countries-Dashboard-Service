package utils

import (
	"math/rand"
	"time"
)

// initialize the package
func init() {
	// Seed rand with time, making it nondeterministic
	rand.New(rand.NewSource(time.Now().UnixNano()))
}

// Runes list of characters to use for ID
var Runes = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789")

// GenerateUID returns a Unique Identifier that is 'n' long
func GenerateUID(n int) string {
	// Make a slice that is 'n' long
	b := make([]rune, n)
	//loop through the slice and insert a random character at each index
	for i := range b {
		b[i] = Runes[rand.Intn(len(Runes))]
	}
	// concatenate the slice to string
	return string(b)

}
