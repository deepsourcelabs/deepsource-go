package processors

import (
	"bytes"
	"errors"
	"regexp"
	"strconv"
	"strings"

	"github.com/deepsourcelabs/deepsource-go/analyzers/types"
)

// UnixProcessor is a processor for unix-formatted strings.
type UnixProcessor struct{}

func (u *UnixProcessor) Unix(buf bytes.Buffer) (types.AnalysisReport, error) {
	var issues []types.Issue

	// trim newline from buffer output
	lines := strings.Split(buf.String(), "\n")

	for _, line := range lines {
		// trim spaces
		line = strings.TrimSpace(line)
		if line == "" {
			break
		}

		// compile regular expression for parsing unix format

		// group descriptions:
		// 0: complete string
		// 1: path
		// 2: line number
		// 3: column number
		// 4: message
		exp, err := regexp.Compile("(.+):(.):(.): (.+)")
		if err != nil {
			return types.AnalysisReport{}, err
		}

		// get groups
		groups := exp.FindAllStringSubmatch(strings.TrimSuffix(line, "\n"), -1)
		if len(groups) == 0 {
			return types.AnalysisReport{}, errors.New("failed to parse message")
		}

		// convert line and column numbers to int
		line, err := strconv.Atoi(groups[0][2])
		if err != nil {
			return types.AnalysisReport{}, err
		}

		col, err := strconv.Atoi(groups[0][3])
		if err != nil {
			return types.AnalysisReport{}, err
		}

		// populate issue
		issue := types.Issue{
			IssueCode: "",
			IssueText: groups[0][4],
			Location: types.Location{
				Path: groups[0][1],
				Position: types.Position{
					Begin: types.Coordinate{
						Line:   line,
						Column: col,
					},
				},
			},
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
