package obfuscator

import (
	"go/ast"
	"strconv"
)

type renamer struct {
	nameMap     map[*ast.Object]string
	nameCounter int
}

func (r *renamer) newName() string {
	name := "o" + strconv.Itoa(r.nameCounter)
	r.nameCounter++
	return name
}

func isSafeToRename(ident *ast.Ident) bool {
	if ident == nil || ident.Obj == nil {
		return false
	}
	return !ident.IsExported() && ident.Name != "main" && ident.Name != "init" && ident.Name != "_"
}

func RenameIdentifiers(node *ast.File) {
	r := &renamer{nameMap: make(map[*ast.Object]string)}

	ast.Inspect(node, func(n ast.Node) bool {
		ident, ok := n.(*ast.Ident)
		if !ok {
			return true
		}

		if ident.Obj != nil && ident.Obj.Decl == n {
			if isSafeToRename(ident) {
				r.nameMap[ident.Obj] = r.newName()
			}
		}
		return true
	})

	ast.Inspect(node, func(n ast.Node) bool {
		ident, ok := n.(*ast.Ident)
		if !ok {
			return true
		}
		if ident.Obj != nil {
			if newName, ok := r.nameMap[ident.Obj]; ok {
				ident.Name = newName
			}
		}
		return true
	})
}
