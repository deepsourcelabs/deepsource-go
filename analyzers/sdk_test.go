package analyzers

import (
	"encoding/json"
	"errors"
	"os"
	"path"
	"reflect"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/deepsourcelabs/deepsource-go/analyzers/analysistest"
	"github.com/deepsourcelabs/deepsource-go/analyzers/processors"
	"github.com/deepsourcelabs/deepsource-go/analyzers/types"
	"github.com/deepsourcelabs/deepsource-go/analyzers/utils"
)

func TestAnalyzer(t *testing.T) {
	t.Run("Run staticcheck as DeepSource Analyzer", func(t *testing.T) {
		// set environment variables
		tempDir := t.TempDir()
		t.Setenv("TOOLBOX_PATH", tempDir)
		t.Setenv("REPO_ROOT", tempDir)

		rp := processors.RegexProcessor{
			Pattern: `(?P<filename>.+):(?P<line>\d+):(?P<column>\d+): (?P<message>.+)\((?P<issue_code>\w+)\)`,
		}

		a := &CLIRunner{
			Name:      "staticcheck",
			Command:   "staticcheck",
			Args:      []string{"-f", "text", "./testdata/src/staticcheck/..."},
			Processor: &rp,
		}

		err := testRunner(a, tempDir, "testdata/src/staticcheck/staticcheck.go")
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("Run csslint as DeepSource Analyzer", func(t *testing.T) {
		// set environment variables
		tempDir := t.TempDir()
		t.Setenv("TOOLBOX_PATH", tempDir)
		t.Setenv("REPO_ROOT", tempDir)

		issueProcessor := func(content string) string {
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

		rp := processors.RegexProcessor{
			Pattern:            `(?P<filename>.+): line (?P<line>\d+), col (?P<column>\d+), (?P<message>.+) \((?P<issue_code>.+)\)`,
			IssueCodeProcessor: issueProcessor,
		}

		a := &CLIRunner{
			Name:      "csslint",
			Command:   "csslint",
			Args:      []string{"--format=compact", "."},
			Processor: &rp,
		}

		err := testRunner(a, tempDir, "testdata/src/csslint/csslint.css")
		if err != nil {
			t.Fatal(err)
		}
	})
}

func testRunner(a *CLIRunner, tempDir string, triggerFilename string) error {
	err := a.Run()
	if err != nil {
		return err
	}

	processedReport, err := a.Processor.Process(a.Stdout())
	if err != nil {
		return err
	}

	// save report
	err = a.SaveReport(processedReport)
	if err != nil {
		return err
	}

	// read the generated report
	generatedFile := path.Join(tempDir, "analysis_report.json")
	reportContent, err := os.ReadFile(generatedFile)
	if err != nil {
		return err
	}

	var report types.AnalysisReport
	err = json.Unmarshal(reportContent, &report)
	if err != nil {
		return err
	}

	// do a verification check for the generated report
	err = analysistest.Verify(report, triggerFilename)
	if err != nil {
		return err
	}

	// cleanup after test
	err = os.Remove(generatedFile)
	if err != nil {
		return err
	}

	return nil
}

func TestUtils(t *testing.T) {
	t.Run("test TOML generation", func(t *testing.T) {
		// fetch parsed issues
		issues, err := utils.ParseIssues("testdata/issues.toml")
		if err != nil {
			t.Fatal(err)
		}

		// generate TOML files
		err = utils.BuildTOML(issues, "testdata/toml")
		if err != nil {
			t.Fatal(err)
		}

		// traverse directory
		files, err := os.ReadDir("testdata/toml")
		if err != nil {
			t.Fatal(err)
		}

		// parse issues from each TOML file
		var parsedIssue utils.IssueMeta
		var parsedIssues []utils.IssueMeta

		for _, f := range files {
			filePath := path.Join("testdata/toml", f.Name())
			_, err = toml.DecodeFile(filePath, &parsedIssue)
			if err != nil {
				t.Fatal(err)
			}
			parsedIssues = append(parsedIssues, parsedIssue)
		}

		// check if the parsed issues and the issues present in the parent TOML are equal
		if !reflect.DeepEqual(issues, parsedIssues) {
			t.Fatal(errors.New("mismatch between issues in parent TOML file and parsed issues"))
		}

		// cleanup TOMLs
		err = os.RemoveAll("testdata/toml")
		if err != nil {
			t.Fatal(err)
		}
	})
}
