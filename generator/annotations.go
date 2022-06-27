package generator

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// analyzerRuleMap represents the pairing between analyzers and analyzers.
var analyzerRuleMap map[string][]string

func init() {
	analyzerRuleMap = make(map[string][]string)
}

// ParseAnnotations reads files from a directory and returns a list of issues.
func ParseAnnotations(directory, codegenPath string) (Issues, error) {
	var issues Issues
	// get filenames
	files, err := walkDir(directory)
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
	generatedCode := codeGenerator(analyzerRuleMap)
	generatedCode.Save(codegenPath)

	return issues, nil
}

// walkDir walks over a directory and returns a list of file names.
func walkDir(directory string) ([]string, error) {
	var files []string

	// check if directory exists before walking
	if _, err := os.Stat(directory); err != nil {
		return nil, err
	}

	err := filepath.WalkDir(directory, func(path string, fileInfo fs.DirEntry, err error) error {
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
func traverseAST(f *ast.File) (Issues, error) {
	// regular expression for matching annotation body
	exp, err := regexp.Compile(`(?s)(?P<annotation>.+)\nanalyzer = "(?P<analyzer>.+)"\nissue_code = "(?P<issue_code>.+)"\ncategory = "(?P<category>.+)"\ntitle = "(?P<title>.+)"\ndescription = """\n(?P<description>.*?)\n"""`)
	if err != nil {
		return Issues{}, err
	}

	var issues Issues

	// traverse AST
	ast.Inspect(f, func(n ast.Node) bool {
		// check if the node is a function
		if node, ok := n.(*ast.FuncDecl); ok {
			// extract comment from the node
			doc := node.Doc.Text()

			// exit early if the comment doesn't contain the "deepsource:rule" annotation
			if !strings.Contains(doc, "deepsource:rule") {
				return false
			}

			// handle both type of comments: a multi-line comment, or a single-line comment over multiple lines
			// trim the "// " prefix in the case of single-line comment over multiple lines
			var lines []string
			for _, line := range strings.Split(doc, "\n") {
				trimmed := strings.TrimPrefix(line, "// ")
				lines = append(lines, trimmed)
			}
			content := strings.Join(lines, "\n")

			// namedGroups is the map containing the content of the named groups of the regular expression
			namedGroups := make(map[string]string)

			// find matches using regular expressions
			match := exp.FindStringSubmatch(content)
			if len(match) > 0 {
				for i, name := range exp.SubexpNames() {
					if i != 0 && name != "" {
						namedGroups[name] = match[i]
					}
				}
			}

			if len(namedGroups) != 0 {
				issue := Issue{
					IssueCode:   namedGroups["issue_code"],
					Category:    namedGroups["category"],
					Title:       namedGroups["title"],
					Description: namedGroups["description"],
				}
				issues = append(issues, issue)

				// add analyzer-analyzer mapping to our global map
				analyzerName := namedGroups["analyzer"]
				analyzerRuleMap[analyzerName] = append(analyzerRuleMap[analyzerName], node.Name.String())
			}
		}

		return true
	})

	return issues, nil
}
