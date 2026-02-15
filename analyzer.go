package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

// AnalyzeSource parses Go source from a string and returns findings.
func AnalyzeSource(path, src string) ([]Finding, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, src, parser.AllErrors)
	if err != nil {
		return nil, err
	}
	return analyze(fset, path, f), nil
}

// AnalyzeFile parses a Go file from disk and returns findings.
func AnalyzeFile(path string) ([]Finding, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, parser.AllErrors)
	if err != nil {
		return nil, err
	}
	return analyze(fset, path, f), nil
}

func analyze(fset *token.FileSet, path string, f *ast.File) []Finding {
	var out []Finding
	ast.Inspect(f, func(n ast.Node) bool {
		if n == nil {
			return false
		}
		if isLoop(n) {
			out = append(out, loopAllocs(fset, path, n)...)
			return false
		}
		if call, ok := n.(*ast.CallExpr); ok {
			if name := callName(call); strings.HasPrefix(name, "fmt.") {
				out = append(out, Finding{path, fset.Position(call.Pos()).Line,
					"fmt-alloc", "medium", name + " causes heap allocation"})
			}
		}
		if ret, ok := n.(*ast.ReturnStmt); ok {
			for _, r := range ret.Results {
				if ue, uok := r.(*ast.UnaryExpr); uok && ue.Op.String() == "&" {
					out = append(out, Finding{path, fset.Position(ue.Pos()).Line,
						"ptr-escape", "high", "returning pointer to local causes heap escape"})
				}
			}
		}
		return true
	})
	return out
}

func isLoop(n ast.Node) bool {
	switch n.(type) {
	case *ast.ForStmt, *ast.RangeStmt:
		return true
	}
	return false
}

func loopAllocs(fset *token.FileSet, path string, loop ast.Node) []Finding {
	var out []Finding
	ast.Inspect(loop, func(n ast.Node) bool {
		if n == nil || n == loop {
			return true
		}
		line := fset.Position(n.Pos()).Line
		switch node := n.(type) {
		case *ast.CallExpr:
			name := callName(node)
			if name == "make" || name == "new" {
				out = append(out, Finding{path, line, "loop-alloc", "high",
					name + "() in loop causes repeated heap allocation"})
			}
			if name == "append" {
				out = append(out, Finding{path, line, "loop-append", "medium",
					"append in loop may cause repeated allocations; pre-allocate"})
			}
			if strings.HasPrefix(name, "fmt.") {
				out = append(out, Finding{path, line, "loop-fmt", "high",
					name + " in loop causes repeated heap allocation"})
			}
		case *ast.UnaryExpr:
			if node.Op.String() == "&" {
				out = append(out, Finding{path, line, "loop-escape", "high",
					"address-of in loop causes allocation per iteration"})
			}
		}
		return true
	})
	return out
}

func callName(call *ast.CallExpr) string {
	switch fn := call.Fun.(type) {
	case *ast.Ident:
		return fn.Name
	case *ast.SelectorExpr:
		if x, ok := fn.X.(*ast.Ident); ok {
			return x.Name + "." + fn.Sel.Name
		}
	}
	return ""
}
