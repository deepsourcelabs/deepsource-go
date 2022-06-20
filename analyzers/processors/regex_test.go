package processors

import (
	"testing"

	"github.com/deepsourcelabs/deepsource-go/analyzers/types"
	"github.com/go-test/deep"
)

func mockIssueCodeGenerator(content string) string {
	issueMap := map[string]string{
		"empty-rules":      "E001",
		"errors":           "E002",
		"known-properties": "K001",
	}
	if issueMap[content] == "" {
		return "U001"
	}
	return issueMap[content]
}

func TestRegex_PopulateIssue(t *testing.T) {
	mockIssue := types.Issue{
		IssueCode: "E001",
		IssueText: "Warning - Rule is empty.",
	}
	mockIssue.Location.Path = "/home/testdir/file1.css"
	mockIssue.Location.Position.Begin.Line = 1
	mockIssue.Location.Position.Begin.Column = 1

	// mock issue for staticcheck
	mockStaticCheckIssue := types.Issue{
		IssueCode: "U1000",
		IssueText: "func trigger is unused ",
	}
	mockStaticCheckIssue.Location.Path = "staticcheck/staticcheck.go"
	mockStaticCheckIssue.Location.Position.Begin.Line = 5
	mockStaticCheckIssue.Location.Position.Begin.Column = 6

	cases := []struct {
		description        string
		line               string
		pattern            string
		issueCodeGenerator IssueCodeGenerator
		expected           types.Issue
	}{
		{"issue code generator must work", "/home/testdir/file1.css: line 1, col 1, Warning - Rule is empty. (empty-rules)", `(?P<filename>.+): line (?P<line>\d+), col (?P<column>\d+), (?P<message>.+) \((?P<issue_code>.+)\)`, mockIssueCodeGenerator, mockIssue},
		{"processor must work without issue code generator", "staticcheck/staticcheck.go:5:6: func trigger is unused (U1000)", `(?P<filename>.+):(?P<line>\d+):(?P<column>\d+): (?P<message>.+)\((?P<issue_code>\w+)\)`, nil, mockStaticCheckIssue},
	}

	for _, tc := range cases {
		actual, err := populateIssue(tc.line, tc.pattern, tc.issueCodeGenerator)
		if err != nil {
			t.Error(err)
		}

		if diff := deep.Equal(actual, tc.expected); diff != nil {
			t.Errorf("description: %s, %s", tc.description, diff)
		}
	}
}
