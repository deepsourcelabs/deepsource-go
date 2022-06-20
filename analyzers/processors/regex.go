package processors

import (
	"bytes"
	"errors"
	"regexp"
	"strconv"
	"strings"

	"github.com/deepsourcelabs/deepsource-go/analyzers/types"
)

// IssueCodeGenerator is used when an analyzer doesn't support issue codes. IssueCodeGenerator reads the content of the "issue_code" named group and returns an appropriate issue code. If not implemented, it fallbacks to using the content as the issue code.
type IssueCodeGenerator func(string) string

// RegexProcessor utilizes regular expressions for processing.
type RegexProcessor struct {
	Pattern            string
	IssueCodeGenerator IssueCodeGenerator
}

func (r *RegexProcessor) Process(buf bytes.Buffer) (types.AnalysisReport, error) {
	var issues []types.Issue

	// trim newline from buffer output
	lines := strings.Split(buf.String(), "\n")

	for _, line := range lines {
		// trim spaces
		line = strings.TrimSpace(line)
		if line == "" {
			break
		}

		issue, err := populateIssue(line, r.Pattern, r.IssueCodeGenerator)
		if err != nil {
			return types.AnalysisReport{}, err
		}

		issues = append(issues, issue)
	}

	// populate report
	report := types.AnalysisReport{
		Issues: issues,
	}

	// return report
	return report, nil
}

// populateIssue returns an issue with the help of a regular expression based pattern and an issue code generator.
func populateIssue(line string, pattern string, issueCodeGenerator IssueCodeGenerator) (types.Issue, error) {
	// compile regular expression
	exp, err := regexp.Compile(pattern)
	if err != nil {
		return types.Issue{}, err
	}

	// get groups
	groupNames := exp.SubexpNames()

	var issue types.Issue
	groups := exp.FindAllStringSubmatch(strings.TrimSuffix(line, "\n"), -1)
	if len(groups) == 0 {
		return types.Issue{}, errors.New("failed to parse message")
	}

	for groupIdx, content := range groups[0] {
		groupName := groupNames[groupIdx]

		// populate issue using named groups
		switch groupName {
		case "filename":
			issue.Location.Path = content
		case "line":
			line, err := strconv.Atoi(content)
			if err != nil {
				return types.Issue{}, err
			}
			issue.Location.Position.Begin.Line = line
		case "column":
			col, err := strconv.Atoi(content)
			if err != nil {
				return types.Issue{}, err
			}
			issue.Location.Position.Begin.Column = col
		case "message":
			issue.IssueText = content
		case "issue_code":
			if issueCodeGenerator == nil {
				issue.IssueCode = content
			} else {
				issue.IssueCode = issueCodeGenerator(content)
			}
		default:
			continue
		}
	}

	return issue, nil
}
