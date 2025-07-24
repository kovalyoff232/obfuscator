package obfuscator

import "math/rand"

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// RandomIdentifier generates a random string to be used as a Go identifier.
func RandomIdentifier(prefix string) string {
	b := make([]byte, 12)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return prefix + "_" + string(b)
}
