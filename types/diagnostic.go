package types

import "go/token"

type Diagnostic struct {
	Line           token.Pos
	Message        string
	SuggestedFixes []string
}
