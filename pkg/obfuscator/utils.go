package obfuscator

import (
	"crypto/rand"
	"go/ast"
	"go/token"
	"math/big"
)

const (
	charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

// NewName generates a new cryptographically random identifier.
func NewName() string {
	b := make([]byte, 10)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			// Fallback to a less random method in case of error
			b[i] = charset[i%len(charset)]
		} else {
			b[i] = charset[n.Int64()]
		}
	}
	// Identifiers in Go cannot start with a number, but our charset doesn't have numbers.
	// We'll start it with a letter to be safe.
	return "o_" + string(b)
}

// isVarDeclared checks if a variable with the given name is already declared in the file.
func isVarDeclared(file *ast.File, varName string) bool {
	for _, decl := range file.Decls {
		if gd, ok := decl.(*ast.GenDecl); ok {
			for _, spec := range gd.Specs {
				if vs, ok := spec.(*ast.ValueSpec); ok {
					for _, name := range vs.Names {
						if name.Name == varName {
							return true
						}
					}
				}
			}
		}
	}
	return false
}

// insertDeclsAfterImports finds the last import declaration and inserts the new
// declarations right after it.
func insertDeclsAfterImports(file *ast.File, decls []ast.Decl) {
	lastImportIndex := -1
	for i, decl := range file.Decls {
		if gd, ok := decl.(*ast.GenDecl); ok && gd.Tok == token.IMPORT {
			lastImportIndex = i
		}
	}

	if lastImportIndex != -1 {
		// Insert after the last import
		file.Decls = append(file.Decls[:lastImportIndex+1], append(decls, file.Decls[lastImportIndex+1:]...)...)
	} else {
		// If no imports, add to the top of the file.
		file.Decls = append(decls, file.Decls...)
	}
}
