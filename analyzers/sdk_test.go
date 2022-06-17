package analyzers

import (
	"os"
	"path"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/deepsourcelabs/deepsource-go/analyzers/analysistest"
	"github.com/deepsourcelabs/deepsource-go/analyzers/build"
	"github.com/deepsourcelabs/deepsource-go/analyzers/processors"
	"github.com/go-test/deep"
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
		rootDir := t.TempDir()
		// fetch parsed issues
		f, err := os.Open("testdata/issues.toml")
		if err != nil {
			t.Fatal(err)
		}

		issues, err := build.FetchIssues(f)
		if err != nil {
			t.Fatal(err)
		}

		// generate TOML files
		err = build.BuildTOML(issues, rootDir)
		if err != nil {
			t.Fatal(err)
		}

		// traverse directory
		files, err := os.ReadDir(rootDir)
		if err != nil {
			t.Fatal(err)
		}

		// parse issues from each TOML file
		var parsedIssue build.IssueMeta
		var parsedIssues build.IssuesMeta

		for _, f := range files {
			filePath := path.Join(rootDir, f.Name())
			_, err = toml.DecodeFile(filePath, &parsedIssue)
			if err != nil {
				t.Fatal(err)
			}
			parsedIssues = append(parsedIssues, parsedIssue)
		}

		// check if the parsed issues and the issues present in the parent TOML are equal
		if diff := deep.Equal(issues, parsedIssues); diff != nil {
			t.Errorf("mismatch between parsed issues and report's issues: %s\n", diff)
		}

		// cleanup TOMLs
		err = os.RemoveAll("testdata/toml")
		if err != nil {
			t.Fatal(err)
		}
	})
}
