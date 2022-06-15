package analysistest

import (
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/deepsourcelabs/deepsource-go/analyzers/types"
	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/css"
	"github.com/smacker/go-tree-sitter/golang"
)

// ParsedIssue represents an issue parsed using tree-sitter.
type ParsedIssue struct {
	IssueCode string
	Line      int
}

func Run(directory string) error {
	// read the generated report from TOOLBOX_PATH
	toolboxPath := os.Getenv("TOOLBOX_PATH")
	generatedFile := path.Join(toolboxPath, "analysis_report.json")
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
	err = verifyReport(report, directory)
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

// getFilenames returns the filenames for a directory.
func getFilenames(directory string) ([]string, error) {
	var files []string
	filepath.Walk(directory, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// if not a directory, append to files
		if !info.IsDir() {
			filename := filepath.Join(directory, info.Name())
			files = append(files, filename)
		}

		return nil
	})

	return files, nil
}

// Verify compares the generated report and parsed issues using tree-sitter.
func verifyReport(report types.AnalysisReport, directory string) error {
	var parsedIssues []ParsedIssue

	// get filenames
	files, err := getFilenames(directory)
	if err != nil {
		return err
	}

	parser := sitter.NewParser()

	// walk through each file and get issues
	for _, filename := range files {
		// get language
		lang, err := getLanguage(filename)
		if err != nil {
			return err
		}
		parser.SetLanguage(lang)

		// read report
		content, err := os.ReadFile(filename)
		if err != nil {
			return err
		}

		// generate tree
		ctx := context.Background()
		tree, err := parser.ParseCtx(ctx, nil, content)
		if err != nil {
			return err
		}

		// create a query for fetching comments
		queryStr := "(comment) @comment"
		query, err := sitter.NewQuery([]byte(queryStr), lang)
		if err != nil {
			return err
		}

		// execute query on root node
		qc := sitter.NewQueryCursor()
		n := tree.RootNode()
		qc.Exec(query, n)
		defer qc.Close()

		// iterate over matches
		for {
			m, ok := qc.NextMatch()
			if !ok {
				break
			}

			for _, c := range m.Captures {
				// get node content
				node := c.Node
				nodeContent := node.Content(content)

				// check if the comment contains raise annotation
				if strings.Contains(nodeContent, "raise") {
					// find match using expression
					exp := regexp.MustCompile(`.+ raise: `)
					submatches := exp.FindStringSubmatch(nodeContent)

					if len(submatches) != 0 {
						substrings := exp.Split(nodeContent, -1)
						if len(substrings) > 1 {
							issueCodes := strings.Split(substrings[1], ",")
							// add issue to parsedIssues
							for _, issueCode := range issueCodes {
								parsedIssue := ParsedIssue{IssueCode: strings.TrimSpace(issueCode), Line: int(node.StartPoint().Row) + 1}
								parsedIssues = append(parsedIssues, parsedIssue)
							}
						}
					}
				}
			}
		}
	}

	// if number of issues don't match, exit early.
	if len(parsedIssues) != len(report.Issues) {
		return errors.New("mismatch between the number of reported issues and parsed issues")
	}

	// compare the report's issues and parsed issues
	match := compareReport(parsedIssues, report)
	if !match {
		return errors.New("mismatch between parsed issue and report issue")
	}

	return nil
}

// getLanguage is a helper for fetching a tree-sitter language based on the file's extension.
func getLanguage(filename string) (*sitter.Language, error) {
	extension := filepath.Ext(filename)

	switch extension {
	case ".go":
		return golang.GetLanguage(), nil
	case ".css":
		return css.GetLanguage(), nil
	default:
		return nil, errors.New("language not supported")
	}
}

// compareReport is a helper which checks if the parsed issues are identical to the issues present in the report.
func compareReport(parsedIssues []ParsedIssue, report types.AnalysisReport) bool {
	// sort report and parsedIssues by line number
	sort.Slice(parsedIssues, func(i, j int) bool {
		return parsedIssues[i].Line < parsedIssues[j].Line
	})

	sort.Slice(report.Issues, func(i, j int) bool {
		return report.Issues[i].Location.Position.Begin.Line < report.Issues[j].Location.Position.Begin.Line
	})

	for i, issue := range report.Issues {
		if (parsedIssues[i].Line != issue.Location.Position.Begin.Line) && (parsedIssues[i].IssueCode != issue.IssueCode) {
			return false
		}
	}

	return true
}
