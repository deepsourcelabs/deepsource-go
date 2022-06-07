package triggers

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/golang"

	"github.com/deepsourcelabs/deepsource-go/sdk/types"
)

// ParsedIssue represents an issue parsed using tree-sitter.
type ParsedIssue struct {
	IssueCode string
	Line      int
}

// Verify compares the generated report and parsed issues using tree-sitter.
func Verify(report types.AnalysisReport, filename string) error {
	parser := sitter.NewParser()

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

	var parsedIssues []ParsedIssue

	// iterate over matches
	for {
		m, ok := qc.NextMatch()
		if !ok {
			break
		}

		for _, c := range m.Captures {
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

	// if number of issues don't match, exit early.
	if len(parsedIssues) != len(report.Issues) {
		return errors.New("mismatch between the number of reported issues and parsed issues")
	}

	// compare the report's issues and parsed issues
	for i, issue := range report.Issues {
		if (parsedIssues[i].Line != issue.Location.Position.Begin.Line) && (parsedIssues[i].IssueCode != issue.IssueCode) {
			return errors.New("mismatch between parsed issue and report issue")
		}
	}

	return nil
}

// getLanguage is a helper for fetching tree-sitter language based on the file's extension.
func getLanguage(filename string) (*sitter.Language, error) {
	extension := filepath.Ext(filename)

	switch extension {
	case ".go":
		return golang.GetLanguage(), nil
	default:
		return nil, errors.New("language not supported")
	}
}
