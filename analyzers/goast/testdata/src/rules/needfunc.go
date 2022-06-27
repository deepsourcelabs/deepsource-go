package rules

import (
	"go/ast"

	"github.com/deepsourcelabs/deepsource-go/analyzers"
)

func NeedFunc(n ast.Node) ([]analyzers.Diagnostic, error) {
	var diags []analyzers.Diagnostic

	if node, ok := n.(*ast.FuncDecl); ok {
		diagnostic := analyzers.Diagnostic{
			Line:           node.Pos(),
			Message:        "found function!",
			SuggestedFixes: []string{"do nothing"},
		}
		diags = append(diags, diagnostic)
	}

	return diags, nil
}
