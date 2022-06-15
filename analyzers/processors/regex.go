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

		exp, err := regexp.Compile(r.Pattern)
		if err != nil {
			return types.AnalysisReport{}, err
		}

		// get groups
		groupNames := exp.SubexpNames()

		var issue types.Issue
		groups := exp.FindAllStringSubmatch(strings.TrimSuffix(line, "\n"), -1)
		for groupIdx, content := range groups[0] {
			groupName := groupNames[groupIdx]

			// populate issue using named groups
			switch groupName {
			case "filename":
				issue.Location.Path = content
			case "line":
				line, err := strconv.Atoi(content)
				if err != nil {
					return types.AnalysisReport{}, err
				}
				issue.Location.Position.Begin.Line = line
			case "column":
				col, err := strconv.Atoi(content)
				if err != nil {
					return types.AnalysisReport{}, err
				}
				issue.Location.Position.Begin.Column = col
			case "message":
				issue.IssueText = content
			case "issue_code":
				if r.IssueCodeGenerator == nil {
					issue.IssueCode = content
				} else {
					issue.IssueCode = r.IssueCodeGenerator(content)
				}
			default:
				continue
			}
		}
		if len(groups) == 0 {
			return types.AnalysisReport{}, errors.New("failed to parse message")
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
