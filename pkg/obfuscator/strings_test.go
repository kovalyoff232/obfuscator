package obfuscator
import (
	"bytes"
	"go/parser"
	"go/printer"
	"go/token"
	"strings"
	"testing"
)
func TestStringEncryption_InlinedAESCTR(t *testing.T) {
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
	obf := &Obfuscator{WeavingKeyVarName: "weave_key_dummy"}
	pass := NewStringEncryptionPass()
	if err := pass.Apply(obf, fset, file); err != nil {
		t.Fatalf("StringEncryptionPass.Apply failed: %v", err)
	}
	var buf bytes.Buffer
	if err := printer.Fprint(&buf, fset, file); err != nil {
		t.Fatalf("Failed to print AST: %v", err)
	}
	out := buf.String()
	if strings.Contains(out, `"hello, world!"`) {
		t.Errorf("Original string literal should not be present in the output")
	}
	if !strings.Contains(out, `"crypto/aes"`) || !strings.Contains(out, `"crypto/cipher"`) {
		t.Errorf("Expected crypto/aes and crypto/cipher imports to be injected")
	}
	if !strings.Contains(out, "aes.NewCipher(") {
		t.Errorf("Expected aes.NewCipher call in inlined decryptor")
	}
	if !strings.Contains(out, "cipher.NewCTR(") {
		t.Errorf("Expected cipher.NewCTR call in inlined decryptor")
	}
	if !strings.Contains(out, ".XORKeyStream(") {
		t.Errorf("Expected XORKeyStream usage in inlined decryptor")
	}
	if !strings.Contains(out, "return string(") {
		t.Errorf("Expected returning string(...) from decryptor")
	}
}
func TestStringEncryption_NoStrings_NoDecryptor(t *testing.T) {
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
	obf := &Obfuscator{WeavingKeyVarName: "weave_key_dummy"}
	pass := NewStringEncryptionPass()
	if err := pass.Apply(obf, fset, file); err != nil {
		t.Fatalf("StringEncryptionPass.Apply failed: %v", err)
	}
	var buf bytes.Buffer
	if err := printer.Fprint(&buf, fset, file); err != nil {
		t.Fatalf("Failed to print AST: %v", err)
	}
	out := buf.String()
	if strings.Contains(out, `"crypto/aes"`) || strings.Contains(out, `"crypto/cipher"`) ||
		strings.Contains(out, "aes.NewCipher(") || strings.Contains(out, "cipher.NewCTR(") ||
		strings.Contains(out, ".XORKeyStream(") {
		t.Errorf("Decryptor artifacts should not be injected when no string literals are present")
	}
	if !strings.Contains(out, "a := 1 + 1") {
		t.Errorf("Original code was unexpectedly altered")
	}
}
