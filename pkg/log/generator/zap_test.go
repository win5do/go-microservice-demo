package main

import (
	"go/ast"
	goparser "go/parser"
	"go/token"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetImportPkg(t *testing.T) {
	r, err := getImportPkg("go.uber.org/zap")
	require.NoError(t, err)
	t.Log(r)
}

func TestParseFunc(t *testing.T) {
	src := `
package test

func bar(a, b, c int) int {
	x := 1
	x = s.Info(a, b, c)
	return x
}
`
	fset := token.NewFileSet()
	r, err := goparser.ParseFile(fset, "foo.go", src, 0)
	require.NoError(t, err)
	err = ast.Print(fset, r)
	require.NoError(t, err)
	body := r.Decls[0].(*ast.FuncDecl).Body.List[0]
	err = ast.Print(fset, body)
	require.NoError(t, err)
}

func TestParseExpr(t *testing.T) {
	fset := token.NewFileSet()
	r, err := goparser.ParseExpr(`zap.S().Info(a, b, c)`)
	require.NoError(t, err)
	err = ast.Print(fset, r)
	require.NoError(t, err)
}

func TestWalkAst(t *testing.T) {
	f, err := goparser.ParseFile(token.NewFileSet(), filepath.Join(os.Getenv("GOPATH"), "/pkg/mod/go.uber.org/zap@v1.10.0/sugar.go"), nil, goparser.Mode(0))
	require.NoError(t, err)
	_, err = walkAst(f)
	require.NoError(t, err)
}
