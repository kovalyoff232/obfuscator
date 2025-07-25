package obfuscator

import "math/rand"

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// NewName generates a random string that is a valid Go identifier.
// It ensures the first character is a letter.
func NewName() string {
	const minLength = 8
	const maxLength = 16

	length := rand.Intn(maxLength-minLength+1) + minLength
	b := make([]byte, length)

	// First character must be a letter
	b[0] = letters[rand.Intn(len(letters))]

	// Subsequent characters can be letters or numbers
	for i := 1; i < length; i++ {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
}