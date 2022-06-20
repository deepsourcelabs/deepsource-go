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
	"github.com/smacker/go-tree-sitter/bash"
	"github.com/smacker/go-tree-sitter/c"
	"github.com/smacker/go-tree-sitter/cpp"
	"github.com/smacker/go-tree-sitter/csharp"
	"github.com/smacker/go-tree-sitter/css"
	"github.com/smacker/go-tree-sitter/elm"
	"github.com/smacker/go-tree-sitter/golang"
	"github.com/smacker/go-tree-sitter/hcl"
	"github.com/smacker/go-tree-sitter/html"
	"github.com/smacker/go-tree-sitter/java"
	"github.com/smacker/go-tree-sitter/javascript"
	"github.com/smacker/go-tree-sitter/lua"
	"github.com/smacker/go-tree-sitter/ocaml"
	"github.com/smacker/go-tree-sitter/php"
	"github.com/smacker/go-tree-sitter/protobuf"
	"github.com/smacker/go-tree-sitter/python"
	"github.com/smacker/go-tree-sitter/ruby"
	"github.com/smacker/go-tree-sitter/rust"
	"github.com/smacker/go-tree-sitter/scala"
	"github.com/smacker/go-tree-sitter/svelte"
	"github.com/smacker/go-tree-sitter/toml"
	"github.com/smacker/go-tree-sitter/typescript/typescript"
	"github.com/smacker/go-tree-sitter/yaml"
)

// ParsedIssue represents an issue parsed using tree-sitter.
type ParsedIssue struct {
	IssueCode string
	Line      int
}

type ParsedIssues []ParsedIssue

const queryStr = "(comment) @comment"

func Run(directory string, toolboxPath string, ctx context.Context) error {
	// read the generated report from TOOLBOX_PATH
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
	err = verifyReport(report, directory, ctx)
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
	err := filepath.Walk(directory, func(path string, info fs.FileInfo, err error) error {
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
	if err != nil {
		return nil, err
	}

	return files, nil
}

// Verify compares the generated report and parsed issues using tree-sitter.
func verifyReport(report types.AnalysisReport, directory string, ctx context.Context) error {
	var parsedIssues ParsedIssues

	// get filenames
	files, err := getFilenames(directory)
	if err != nil {
		return err
	}

	parser := sitter.NewParser()

	// walk through each file and get issues
	for _, filename := range files {
		// set language for the parser
		lang, err := getLanguage(filename)
		if err != nil {
			return err
		}
		parser.SetLanguage(lang)

		// read the report
		content, err := os.ReadFile(filename)
		if err != nil {
			return err
		}

		// generate the tree using tree-sitter
		tree, err := parser.ParseCtx(ctx, nil, content)
		if err != nil {
			return err
		}

		// create a query for fetching comments
		query, err := sitter.NewQuery([]byte(queryStr), lang)
		if err != nil {
			return err
		}

		// execute query on root node
		qc := sitter.NewQueryCursor()
		defer qc.Close()
		n := tree.RootNode()
		qc.Exec(query, n)

		// iterate over matches
		for {
			// fetch a match
			m, ok := qc.NextMatch()
			if !ok {
				break
			}

			// We iterate over the query captures for each match. This traversal consists of various steps:
			// 1. get the node content
			// 2. check if the node has a "raise" annotation
			//	2.1 if true, the node's content is matched with a regular expression. the regular expression matches comments having a raise annotation.
			// 	2.2 the submatches are fetched using the regular expression.
			// 	2.3 if there exists some submatches, then:
			//		2.3.1 split the node content into substrings
			//		2.3.2 check if there exists at least 2 substrings. a valid annotation contains at least 2 substrings: "raise" and issue codes separated by a delimiter (,)
			//		2.3.3 if true, the issue codes are separated on the basis of the delimiter (,)
			//		2.3.4 the issue is then populated with the issue code and line numbers
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
						// get substrings
						substrings := exp.Split(nodeContent, -1)

						// the annotation must have at least 2 substrings: "raise" and issue codes separated by a delimiter (,)
						if len(substrings) > 1 {

							// fetch issue codes by splitting on the basis of the delimiter
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
	case ".sh":
		return bash.GetLanguage(), nil
	case ".c":
		return c.GetLanguage(), nil
	case ".cpp":
		return cpp.GetLanguage(), nil
	case ".cs":
		return csharp.GetLanguage(), nil
	case ".css":
		return css.GetLanguage(), nil
	case ".elm":
		return elm.GetLanguage(), nil
	case ".go":
		return golang.GetLanguage(), nil
	case ".hcl":
		return hcl.GetLanguage(), nil
	case ".html":
		return html.GetLanguage(), nil
	case ".java":
		return java.GetLanguage(), nil
	case ".js":
		return javascript.GetLanguage(), nil
	case ".lua":
		return lua.GetLanguage(), nil
	case ".ml":
		return ocaml.GetLanguage(), nil
	case ".php":
		return php.GetLanguage(), nil
	case ".pb", ".proto":
		return protobuf.GetLanguage(), nil
	case ".py":
		return python.GetLanguage(), nil
	case ".rb":
		return ruby.GetLanguage(), nil
	case ".rs":
		return rust.GetLanguage(), nil
	case ".scala":
		return scala.GetLanguage(), nil
	case ".svelte":
		return svelte.GetLanguage(), nil
	case ".toml":
		return toml.GetLanguage(), nil
	case ".ts":
		return typescript.GetLanguage(), nil
	case ".yaml":
		return yaml.GetLanguage(), nil
	default:
		return nil, errors.New("language not supported")
	}
}

// compareReport is a helper which checks if the parsed issues are identical to the issues present in the report.
func compareReport(parsedIssues ParsedIssues, report types.AnalysisReport) bool {
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
