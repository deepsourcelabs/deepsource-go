package processors

import (
	"encoding/json"
	"strings"

	"github.com/deepsourcelabs/deepsource-go/sdk/types"
)

// sccIssue represents a staticcheck issue.
type sccIssue struct {
	Code     string           `json:"code"`
	Severity string           `json:"severity"`
	Location sccIssueLocation `json:"location"`
	Message  string           `json:"message"`
}

type sccIssueLocation struct {
	File   string `json:"file"`
	Line   int    `json:"line"`
	Column int    `json:"column"`
}

// StaticCheck processor returns a DeepSource compatible analysis report from staticcheck's results.
func StaticCheck(result interface{}) (types.AnalysisReport, error) {
	var issue sccIssue
	var issues []types.Issue

	// trim newline from stdout
	jsonStr := strings.TrimSuffix(result.(string), "\n")

	// parse output and generate issues
	lines := strings.Split(jsonStr, "\n")
	for _, l := range lines {
		err := json.Unmarshal([]byte(l), &issue)
		if err != nil {
			return types.AnalysisReport{}, err
		}

		// convert to a DeepSource issue
		dsIssue := convertIssue(issue)

		issues = append(issues, dsIssue)
	}

	// populate report
	report := types.AnalysisReport{
		Issues: issues,
	}

	// return report
	return report, nil
}

// convertIssue is a helper utility for converting a staticcheck issue to a DeepSource issue.
func convertIssue(issue sccIssue) types.Issue {
	convertedIssue := types.Issue{
		IssueCode: issue.Code,
		IssueText: issue.Message,
		Location: types.Location{
			Path: issue.Location.File,
			Position: types.Position{
				Begin: types.Coordinate{
					Line:   issue.Location.Line,
					Column: issue.Location.Column,
				},
			},
		},
	}
	return convertedIssue
}
