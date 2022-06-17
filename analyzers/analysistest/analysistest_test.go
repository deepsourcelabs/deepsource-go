package analysistest

import (
	"testing"

	"github.com/deepsourcelabs/deepsource-go/analyzers/types"
	"github.com/go-test/deep"
)

type reportFactory struct {
	values []reportMeta
}

// reportMeta contains the metadata for an issue present in the report.
type reportMeta struct {
	IssueCode string
	Line      int
}

// Generate generates a report based on reportMeta values in reportFactory.
func (r *reportFactory) Generate() types.AnalysisReport {
	var report types.AnalysisReport
	var issues []types.Issue

	for _, rm := range r.values {
		issue := types.Issue{
			IssueCode: rm.IssueCode,
		}

		issue.Location.Position.Begin.Line = rm.Line

		issues = append(issues, issue)
	}

	report.Issues = issues
	return report
}

func TestCompareReport(t *testing.T) {
	var factory reportFactory
	factory.values = []reportMeta{
		{
			IssueCode: "E001",
			Line:      1,
		},
		{
			IssueCode: "E002",
			Line:      4,
		},
	}

	// mock factory with dissimilar issue codes
	var factoryMismatch reportFactory
	factoryMismatch.values = []reportMeta{
		{
			IssueCode: "E010",
			Line:      1,
		},
		{
			IssueCode: "E002",
			Line:      4,
		},
	}

	cases := []struct {
		description string
		issues      ParsedIssues
		report      types.AnalysisReport
		expected    bool
	}{
		{"must return true for identical reports", []ParsedIssue{
			{
				IssueCode: "E001",
				Line:      1,
			},
			{
				IssueCode: "E002",
				Line:      4,
			},
		}, factory.Generate(), true},
		{"must return false for dissimilar reports", []ParsedIssue{
			{
				IssueCode: "E001",
				Line:      2,
			},
			{
				IssueCode: "E002",
				Line:      4,
			},
		}, factoryMismatch.Generate(), false},
	}

	for _, tc := range cases {
		actual := compareReport(tc.issues, tc.report)
		if diff := deep.Equal(actual, tc.expected); diff != nil {
			t.Errorf("description: %s, %s", tc.description, diff)
		}
	}
}
