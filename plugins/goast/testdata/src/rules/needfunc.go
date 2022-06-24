package rules

import (
	"go/ast"

	"github.com/deepsourcelabs/deepsource-go/types"
)

func NeedFunc(n ast.Node) ([]types.Diagnostic, error) {
	var diags []types.Diagnostic

	if node, ok := n.(*ast.FuncDecl); ok {
		diagnostic := types.Diagnostic{
			Line:           node.Pos(),
			Message:        "found function!",
			SuggestedFixes: []string{"do nothing"},
		}
		diags = append(diags, diagnostic)
	}

	return diags, nil
}
