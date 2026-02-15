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
				sev, conf := ScoreFinding(call, "fmt-alloc")
				out = append(out, Finding{path, fset.Position(call.Pos()).Line,
					"fmt-alloc", sev.String(), conf.String(), name + " causes heap allocation"})
			}
		}
		if ret, ok := n.(*ast.ReturnStmt); ok {
			for _, r := range ret.Results {
				if ue, uok := r.(*ast.UnaryExpr); uok && ue.Op.String() == "&" {
					sev, conf := ScoreFinding(ue, "ptr-escape")
					out = append(out, Finding{path, fset.Position(ue.Pos()).Line,
						"ptr-escape", sev.String(), conf.String(), "returning pointer to local causes heap escape"})
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
func loopAllocs(fset *token.FileSet, path string, loop ast.Node) []Finding {
	var out []Finding
	ast.Inspect(loop, func(n ast.Node) bool {
		if n == nil {
			return false
		}
		switch x := n.(type) {
		case *ast.CallExpr:
			name := callName(x)
			if name == "make" || name == "new" {
				sev, conf := ScoreFinding(x, "loop-alloc")
				out = append(out, Finding{path, fset.Position(x.Pos()).Line,
					"loop-alloc", sev.String(), conf.String(),
					name + "() in loop causes repeated heap allocation"})
			}
			if strings.HasPrefix(name, "fmt.") {
				sev, conf := ScoreFinding(x, "loop-fmt")
				out = append(out, Finding{path, fset.Position(x.Pos()).Line,
					"loop-fmt", sev.String(), conf.String(),
					name + " in loop causes repeated heap allocation"})
			}
			if name == "append" {
				sev, conf := ScoreFinding(x, "loop-append")
				out = append(out, Finding{path, fset.Position(x.Pos()).Line,
					"loop-append", sev.String(), conf.String(),
					"append in loop without pre-allocation"})
			}
		case *ast.UnaryExpr:
			if x.Op.String() == "&" {
				sev, conf := ScoreFinding(x, "loop-escape")
				out = append(out, Finding{path, fset.Position(x.Pos()).Line,
					"loop-escape", sev.String(), conf.String(),
					"&var inside loop causes allocation per iteration"})
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

	return ""
}
