package types

type Issue struct {
	IssueCode string `json:"issue_code"`
	IssueText string `json:"issue_text"`
	Location  struct {
		Path     string `json:"path"`
		Position struct {
			Begin struct {
				Line   int `json:"line"`
				Column int `json:"column"`
			} `json:"begin"`
			End struct {
				Line   int `json:"line"`
				Column int `json:"column"`
			} `json:"end"`
		} `json:"position"`
	} `json:"location"`
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
