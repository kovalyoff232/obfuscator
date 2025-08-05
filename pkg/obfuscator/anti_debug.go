//go:build !linux
// +build !linux
package obfuscator
import (
	"go/ast"
	"go/token"
)
type AntiDebugPass struct{}
func (p *AntiDebugPass) Apply(obf *Obfuscator, fset *token.FileSet, file *ast.File) error {
	return nil
}
