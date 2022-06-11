package analyzers

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/deepsourcelabs/deepsource-go/analyzers/types"
	"github.com/deepsourcelabs/deepsource-go/analyzers/utils"
)

type StaticCheckProcessor struct{}

// StaticCheck processor returns a DeepSource-compatible analysis report from staticcheck's results.
func (s *StaticCheckProcessor) Process(buf bytes.Buffer) (types.AnalysisReport, error) {
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
			return types.AnalysisReport{}, errors.New("failed to parse output string")
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

		// compile regular expression for parsing staticcheck message

		// group descriptions:
		// 0: complete string
		// 1: partial message string
		// 2: issue code
		// 3: parentheses
		messageExp, err := regexp.Compile("(.+)[(](.+)(.+)")
		if err != nil {
			return types.AnalysisReport{}, err
		}
		messageGroups := messageExp.FindAllStringSubmatch(groups[0][4], -1)
		if len(messageGroups) == 0 {
			return types.AnalysisReport{}, errors.New("failed to parse message")
		}

		// populate issue
		issue := types.Issue{
			IssueCode: messageGroups[0][2],
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

func TestStaticCheck(t *testing.T) {
	t.Run("staticcheck analyzer", func(t *testing.T) {
		a := CLIAnalyzer{
			Name:      "staticcheck",
			Command:   "staticcheck",
			Args:      []string{"-f", "text", "./testdata/triggers/staticcheck/..."},
			Processor: &StaticCheckProcessor{},
		}

		err := a.Run()
		if err != nil {
			t.Fatal(err)
		}

		processedReport, err := a.Processor.Process(a.Stdout())
		if err != nil {
			t.Fatal(err)
		}

		// save report
		err = utils.SaveReport(processedReport, "testdata/triggers/staticcheck/issues.json", "json")
		if err != nil {
			t.Fatal(err)
		}

		// read the generated report
		reportContent, err := os.ReadFile("testdata/triggers/staticcheck/issues.json")
		if err != nil {
			t.Fatal(err)
		}

		var report types.AnalysisReport
		err = json.Unmarshal(reportContent, &report)
		if err != nil {
			t.Fatal(err)
		}

		// do a verification check for the generated report
		err = utils.Verify(report, "testdata/triggers/staticcheck/staticcheck.go")
		if err != nil {
			t.Fatal(err)
		}

		// cleanup after test
		err = os.Remove("testdata/triggers/staticcheck/issues.json")
		if err != nil {
			t.Fatal(err)
		}

		// test TOML generation
		err = a.GenerateTOML("testdata/issues.toml", "testdata/toml")
		if err != nil {
			t.Fatal(err)
		}

		// cleanup TOMLs
		err = os.RemoveAll("testdata/toml")
		if err != nil {
			t.Fatal(err)
		}
	})
}
