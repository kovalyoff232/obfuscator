package main
import (
	"crypto/subtle"
	"fmt"
	"os"
)
func getPassword() []byte {
	if v := os.Getenv("PAYLOAD_PASSWORD"); v != "" {
		return []byte(v)
	}
	fmt.Println("[warn] PAYLOAD_PASSWORD не задан. Используется пустой пароль (только для демо).")
	return []byte("")
}
func checkPassword(userInput string) bool {
	expected := getPassword()
	input := []byte(userInput)
	if len(expected) != len(input) {
		padded := make([]byte, len(expected))
		copy(padded, input)
		return subtle.ConstantTimeCompare(expected, padded) == 1
	}
	return subtle.ConstantTimeCompare(expected, input) == 1
}
func main() {
	fmt.Println("Payload program for obfuscator demonstration.")
	fmt.Println("---------------------------------------------")
	fmt.Print("Enter password: ")
	var userInput string
	fmt.Scanln(&userInput)
	if checkPassword(userInput) {
		fmt.Println("\n[+] Access granted. The secrets are yours.")
	} else {
		fmt.Println("\n[-] Access denied. Self-destruct sequence activated.")
	}
}
