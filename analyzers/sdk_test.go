package analyzers

import (
	"errors"
	"os"
	"path"
	"reflect"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/deepsourcelabs/deepsource-go/analyzers/analysistest"
	"github.com/deepsourcelabs/deepsource-go/analyzers/build"
	"github.com/deepsourcelabs/deepsource-go/analyzers/processors"
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

		err := a.Run()
		if err != nil {
			t.Fatal(err)
		}

		report, err := a.Processor.Process(a.Stdout())
		if err != nil {
			t.Fatal(err)
		}

		err = a.SaveReport(report)
		if err != nil {
			t.Fatal(err)
		}

		err = analysistest.Run("./testdata/src/staticcheck")
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("Run csslint as DeepSource Analyzer", func(t *testing.T) {
		// set environment variables
		tempDir := t.TempDir()
		t.Setenv("TOOLBOX_PATH", tempDir)
		t.Setenv("REPO_ROOT", tempDir)

		issueCodeGenerator := func(content string) string {
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
			IssueCodeGenerator: issueCodeGenerator,
		}

		a := &CLIRunner{
			Name:      "csslint",
			Command:   "csslint",
			Args:      []string{"--format=compact", "."},
			Processor: &rp,
		}

		err := a.Run()
		if err != nil {
			t.Fatal(err)
		}

		report, err := a.Processor.Process(a.Stdout())
		if err != nil {
			t.Fatal(err)
		}

		err = a.SaveReport(report)
		if err != nil {
			t.Fatal(err)
		}

		err = analysistest.Run("./testdata/src/csslint")
		if err != nil {
			t.Fatal(err)
		}
	})
}

func TestUtils(t *testing.T) {
	t.Run("test TOML generation", func(t *testing.T) {
		// fetch parsed issues
		issues, err := build.ParseIssues("testdata/issues.toml")
		if err != nil {
			t.Fatal(err)
		}

		// generate TOML files
		err = build.BuildTOML(issues, "testdata/toml")
		if err != nil {
			t.Fatal(err)
		}

		// traverse directory
		files, err := os.ReadDir("testdata/toml")
		if err != nil {
			t.Fatal(err)
		}

		// parse issues from each TOML file
		var parsedIssue build.IssueMeta
		var parsedIssues []build.IssueMeta

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
