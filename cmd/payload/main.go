package main

import (
	"fmt"
)

const superSecretPassword = "SuperSecretPassword123!"

func checkPassword(userInput string) bool {
	var isValid bool
	if userInput == superSecretPassword {
		isValid = true
	} else {
		isValid = false
	}
	return isValid
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
