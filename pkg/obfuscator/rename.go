package obfuscator

import (
	"go/ast"
	"math/rand"
	"time"
)

// newName generates a random string of a random length between 8 and 16 characters.
// The character set includes lower and upper case letters, and numbers.
func newName() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const minLength = 8
	const maxLength = 16

	length := rand.Intn(maxLength-minLength+1) + minLength
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func isSafeToRename(ident *ast.Ident) bool {
	if ident == nil || ident.Obj == nil {
		return false
	}
	// Do not rename exported identifiers, the main/init functions, or the blank identifier.
	return !ident.IsExported() && ident.Name != "main" && ident.Name != "init" && ident.Name != "_"
}

// RenameIdentifiers traverses the AST and renames all non-exported identifiers.
func RenameIdentifiers(node *ast.File) {
	// Seed the random number generator to ensure different names on each run.
	rand.Seed(time.Now().UnixNano())

	// A map to store the new name for each object.
	nameMap := make(map[*ast.Object]string)

	// First pass: find all declarations and assign a new name.
	ast.Inspect(node, func(n ast.Node) bool {
		ident, ok := n.(*ast.Ident)
		if !ok {
			return true
		}

		// If this identifier is a declaration, it's a candidate for renaming.
		if ident.Obj != nil && ident.Obj.Decl == n {
			if isSafeToRename(ident) {
				// Ensure we don't accidentally generate the same name for different objects.
				// While unlikely with random strings, this is a safeguard.
				if _, exists := nameMap[ident.Obj]; !exists {
					nameMap[ident.Obj] = newName()
				}
			}
		}
		return true
	})

	// Second pass: apply the new names to all uses of the identifiers.
	ast.Inspect(node, func(n ast.Node) bool {
		ident, ok := n.(*ast.Ident)
		if !ok {
			return true
		}
		if ident.Obj != nil {
			if newName, ok := nameMap[ident.Obj]; ok {
				ident.Name = newName
			}
		}
		return true
	})
}
