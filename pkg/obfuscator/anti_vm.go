//go:build !linux
// +build !linux
package obfuscator
import (
	"go/ast"
	"go/token"
)
type AntiVMPass struct{}
func (p *AntiVMPass) Apply(obf *Obfuscator, fset *token.FileSet, file *ast.File) error {
	return nil
}
