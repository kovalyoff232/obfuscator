package main

import (
	"crypto/subtle"
	"fmt"
)

const superSecretPassword = "SuperSecretPassword123!"

func checkPassword(userInput string) bool {
	// Используем ConstantTimeCompare для защиты от атак по времени.
	// Важно, чтобы срезы байтов имели одинаковую длину.
	expected := []byte(superSecretPassword)
	input := []byte(userInput)

	// Если длина не совпадает, сравнение заведомо ложно,
	// но для сохранения постоянного времени мы сравниваем пароль с самим собой.
	if len(expected) != len(input) {
		return subtle.ConstantTimeCompare(expected, expected) == 0 // Всегда false
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
