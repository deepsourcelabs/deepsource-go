package utils

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/deepsourcelabs/deepsource-go/types"
)

// pluginAnalyzerMap represents the pairing between plugins and analyzers.
var pluginAnalyzerMap map[string][]string

// ParseAnnotations reads files from a directory and returns a list of issues.
func ParseAnnotations(dir, codegenPath string) ([]types.Issue, error) {
	pluginAnalyzerMap = make(map[string][]string)
	var issues []types.Issue

	// get filenames
	files, err := walkDir(dir)
	if err != nil {
		return nil, err
	}

	// traverse AST and parse annotations
	fset := token.NewFileSet()
	for _, f := range files {
		content, err := os.ReadFile(f)
		if err != nil {
			return nil, err
		}

		src := string(content)

		f, err := parser.ParseFile(fset, "", src, parser.ParseComments)
		if err != nil {
			return nil, err
		}

		parsedIssues, err := traverseAST(f)
		if err != nil {
			return nil, err
		}

		issues = append(issues, parsedIssues...)
	}

	// get generated code and save to codegenPath
	generatedCode := codeGenerator(pluginAnalyzerMap)
	generatedCode.Save(codegenPath)

	return issues, nil
}

// walkDir walks over a directory and returns a list of file names.
func walkDir(dir string) ([]string, error) {
	var files []string

	err := filepath.WalkDir(dir, func(path string, fileInfo fs.DirEntry, err error) error {
		// check if it is a directory
		if !fileInfo.IsDir() {
			files = append(files, path)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return files, nil
}

// traverseAST traverses the AST and parses annotations. It returns a list of issues.
func traverseAST(f *ast.File) ([]types.Issue, error) {
	// regular expression for matching annotation body
	exp, err := regexp.Compile(`(?s)(?P<annotation>.+)\nplugin = "(?P<plugin>.+)"\nissue_code = "(?P<issue_code>.+)"\ncategory = "(?P<category>.+)"\ntitle = "(?P<title>.+)"\ndescription = """\n(?P<description>.*?)\n"""`)
	if err != nil {
		return nil, err
	}

	var issues []types.Issue

	// traverse AST
	ast.Inspect(f, func(n ast.Node) bool {
		// check if the node is a function
		if node, ok := n.(*ast.FuncDecl); ok {
			// result is the map containing the content of the named groups of the regular expression
			result := make(map[string]string)

			// extract comment from the node
			doc := node.Doc.Text()

			// check if the comment contains the "deepsource:analyzer" annotation
			if strings.Contains(doc, "deepsource:analyzer") {
				// handle both type of comments: a multi-line comment, or a single-line comment over multiple lines
				// trim the "// " prefix in the case of single-line comment over multiple lines
				var lines []string
				for _, line := range strings.Split(doc, "\n") {
					trimmed := strings.TrimPrefix(line, "// ")
					lines = append(lines, trimmed)
				}
				content := strings.Join(lines, "\n")

				// find matches using regular expressions
				match := exp.FindStringSubmatch(content)
				if len(match) > 0 {
					for i, name := range exp.SubexpNames() {
						if i != 0 && name != "" {
							result[name] = match[i]
						}
					}
				}

				if len(result) != 0 {
					issue := types.Issue{
						IssueCode:   result["issue_code"],
						Category:    result["category"],
						Title:       result["title"],
						Description: result["description"],
					}
					issues = append(issues, issue)

					// add plugin-analyzer mapping to our global map
					pluginAnalyzerMap[result["plugin"]] = append(pluginAnalyzerMap[result["plugin"]], node.Name.String())
				}
			}
		}

		return true
	})

	return issues, nil
}
