package obfuscator

import (
	"bytes"
	"go/parser"
	"go/printer"
	"go/token"
	"strings"
	"testing"
)

func TestXorEncrypt(t *testing.T) {
	originalData := []byte("hello, world!")
	key := byte(0xAB)

	encryptedData := xorEncrypt(originalData, key)

	if bytes.Equal(originalData, encryptedData) {
		t.Errorf("Encrypted data should not be the same as original data")
	}

	decryptedData := xorEncrypt(encryptedData, key)

	if !bytes.Equal(originalData, decryptedData) {
		t.Errorf("Decrypted data does not match original data. Got %q, want %q", decryptedData, originalData)
	}
}

func TestEncryptStringsIntegration(t *testing.T) {
	src := `
package main

import "fmt"

func main() {
	fmt.Println("hello, world!")
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "source.go", src, 0)
	if err != nil {
		t.Fatalf("Failed to parse source: %v", err)
	}

	if err := EncryptStrings(file); err != nil {
		t.Fatalf("EncryptStrings failed: %v", err)
	}

	var buf bytes.Buffer
	if err := printer.Fprint(&buf, fset, file); err != nil {
		t.Fatalf("Failed to print AST: %v", err)
	}

	got := buf.String()

	// Check 1: The original string should be gone
	if strings.Contains(got, `"hello, world!"`) {
		t.Errorf("Original string literal should not be present in the output, but it was found")
	}

	// Check 2: The decrypt function call should be present
	if !strings.Contains(got, `o_d([]byte{`) {
		t.Errorf("Call to decrypt function o_d was not found in the output")
	}

	// Check 3: The decrypt function declaration should be injected
	if !strings.Contains(got, `func o_d(data []byte, key byte) string {`) {
		t.Errorf("Decrypt function o_d was not injected into the source code")
	}

	// Check 4: The main logic should still be there
	if !strings.Contains(got, `fmt.Println(o_d`) {
		t.Errorf("The main logic (fmt.Println) seems to be missing or altered incorrectly")
	}
}

func TestEncryptStringsNoStrings(t *testing.T) {
	src := `
package main

func main() {
	a := 1 + 1
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "source.go", src, 0)
	if err != nil {
		t.Fatalf("Failed to parse source: %v", err)
	}

	if err := EncryptStrings(file); err != nil {
		t.Fatalf("EncryptStrings failed: %v", err)
	}

	var buf bytes.Buffer
	if err := printer.Fprint(&buf, fset, file); err != nil {
		t.Fatalf("Failed to print AST: %v", err)
	}

	got := buf.String()

	// If no strings were encrypted, the decrypt function should NOT be injected.
	// This is the most important check.
	if strings.Contains(got, "func o_d") {
		t.Errorf("Decrypt function should not be injected when there are no strings to encrypt")
	}

	// Check that the original code is mostly unchanged.
	if !strings.Contains(got, "a := 1 + 1") {
		t.Errorf("Original code was altered unexpectedly")
	}
}
