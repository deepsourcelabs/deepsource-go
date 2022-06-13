package types

type Coordinate struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

type Position struct {
	Begin Coordinate `json:"begin"`
	End   Coordinate `json:"end"`
}

type Location struct {
	Path     string   `json:"path"`
	Position Position `json:"position"`
}

type Issue struct {
	IssueCode string   `json:"issue_code"`
	IssueText string   `json:"issue_text"`
	Location  Location `json:"location"`
}

// Location of an issue
type IssueLocation struct {
	Path     string   `json:"path"`
	Position Position `json:"position"`
}

type AnalysisError struct {
	HMessage string `json:"hmessage"`
	Level    int    `json:"level"`
}

type AnalysisReport struct {
	Issues    []Issue         `json:"issues"`
	Errors    []AnalysisError `json:"errors"`
	ExtraData interface{}     `json:"extra_data"`
}
