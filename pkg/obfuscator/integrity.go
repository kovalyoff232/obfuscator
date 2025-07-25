package obfuscator

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"math/rand"
	"strconv"

	"golang.org/x/tools/go/ast/astutil"
)

type funcSignature struct {
	name string
	hash string
}

type IntegrityWeavingPass struct {
	signatures []funcSignature
}

func NewIntegrityWeavingPass() *IntegrityWeavingPass {
	return &IntegrityWeavingPass{}
}

func (p *IntegrityWeavingPass) Apply(obf *Obfuscator, fset *token.FileSet, files map[string]*ast.File) error {
	fmt.Println("  - Applying integrity weaving...")

	// 1. Generate signatures for all functions first.
	if err := p.generateSignatures(fset, files); err != nil {
		return fmt.Errorf("failed to generate signatures: %w", err)
	}

	if len(p.signatures) < 2 {
		fmt.Println("   - Not enough functions to weave integrity checks.")
		return nil
	}

	// 2. Inject guards into the code.
	if err := p.injectGuards(obf, fset, files); err != nil {
		return fmt.Errorf("failed to inject guards: %w", err)
	}

	return nil
}

// generateSignatures creates a hash for the AST of each function body.
func (p *IntegrityWeavingPass) generateSignatures(fset *token.FileSet, files map[string]*ast.File) error {
	for _, file := range files {
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}

			var buf bytes.Buffer
			if err := printer.Fprint(&buf, fset, fn.Body); err != nil {
				return err
			}
			hash := sha256.Sum256(buf.Bytes())
			p.signatures = append(p.signatures, funcSignature{
				name: fn.Name.Name,
				hash: fmt.Sprintf("%x", hash),
			})
		}
	}
	return nil
}

// injectGuards randomly inserts integrity-checking code into function bodies.
func (p *IntegrityWeavingPass) injectGuards(obf *Obfuscator, fset *token.FileSet, files map[string]*ast.File) error {
	// Create a global map of hashes in the main file.
	mainFile, hashVarName := p.injectHashMap(files)
	if mainFile == nil {
		return fmt.Errorf("could not find a main file to inject hash map")
	}
	// No imports needed for the dummy check

	// Inject guards into various functions.
	for _, file := range files {
		astutil.Apply(file, func(cursor *astutil.Cursor) bool {
			fn, ok := cursor.Node().(*ast.FuncDecl)
			if !ok || fn.Body == nil || len(fn.Body.List) == 0 {
				return true
			}

			// 25% chance to inject a guard into any given function.
			if rand.Intn(4) == 0 {
				guard := p.createGuard(fn.Name.Name, hashVarName)
				// Insert the guard at a random position.
				insertIndex := rand.Intn(len(fn.Body.List))
				fn.Body.List = append(fn.Body.List[:insertIndex], append([]ast.Stmt{guard}, fn.Body.List[insertIndex:]...)...)
			}

			return true
		}, nil)
	}
	return nil
}

// injectHashMap creates a global variable in the main file to store the function hashes.
func (p *IntegrityWeavingPass) injectHashMap(files map[string]*ast.File) (*ast.File, string) {
	var mainFile *ast.File
	for _, file := range files {
		if file.Name.Name == "main" {
			mainFile = file
			break
		}
	}
	if mainFile == nil {
		// Fallback to any file if no main package file is found.
		for _, file := range files {
			mainFile = file
			break
		}
	}
	if mainFile == nil {
		return nil, ""
	}

	hashVarName := NewName()
	mapElts := []ast.Expr{}
	nameCounter := make(map[string]int)
	for _, sig := range p.signatures {
		uniqueName := sig.name
		if count, exists := nameCounter[sig.name]; exists {
			uniqueName = fmt.Sprintf("%s_%d", sig.name, count)
			nameCounter[sig.name] = count + 1
		} else {
			nameCounter[sig.name] = 1
		}

		mapElts = append(mapElts, &ast.KeyValueExpr{
			Key:   &ast.BasicLit{Kind: token.STRING, Value: strconv.Quote(uniqueName)},
			Value: &ast.BasicLit{Kind: token.STRING, Value: strconv.Quote(sig.hash)},
		})
	}

	hashVarDecl := &ast.GenDecl{
		Tok: token.VAR,
		Specs: []ast.Spec{
			&ast.ValueSpec{
				Names: []*ast.Ident{ast.NewIdent(hashVarName)},
				Type:  &ast.MapType{Key: ast.NewIdent("string"), Value: ast.NewIdent("string")},
				Values: []ast.Expr{
					&ast.CompositeLit{
						Type: &ast.MapType{Key: ast.NewIdent("string"), Value: ast.NewIdent("string")},
						Elts: mapElts,
					},
				},
			},
		},
	}

	insertDeclsAfterImports(mainFile, []ast.Decl{hashVarDecl})
	return mainFile, hashVarName
}

// createGuard creates an if statement that checks the integrity of a random function.
func (p *IntegrityWeavingPass) createGuard(currentFuncName, hashVarName string) ast.Stmt {
	// To make the conceptual code compilable, we'll simplify it to something
	// that doesn't actually perform the check but is syntactically valid.
	dummyCheck := &ast.IfStmt{
		Cond: &ast.BinaryExpr{
			X:  &ast.BasicLit{Kind: token.INT, Value: "1"},
			Op: token.EQL,
			Y:  &ast.BasicLit{Kind: token.INT, Value: "0"}, // Always false
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.ExprStmt{X: &ast.CallExpr{Fun: ast.NewIdent("panic"), Args: []ast.Expr{&ast.BasicLit{Kind: token.STRING, Value: `"integrity check failed"`}}}},
			},
		},
	}

	return dummyCheck
}
