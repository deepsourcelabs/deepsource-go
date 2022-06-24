package utils

import (
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/dave/jennifer/jen"
	"github.com/pelletier/go-toml/v2"
)

// pluginAnalyzerMap represents the pairing between plugins and analyzers.
var pluginAnalyzerMap map[string][]string
var issues []map[string]string

// ParseAnnotations reads files from a directory and returns a list of issues.
func ParseAnnotations(dir, codegenPath string) ([]map[string]string, error) {
	pluginAnalyzerMap = make(map[string][]string)

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
func traverseAST(f *ast.File) ([]map[string]string, error) {
	// regular expression for matching annotation body
	exp, err := regexp.Compile(`(?s)(?P<annotation>.+)\nplugin = "(?P<plugin>.+)"\nissue_code = "(?P<issue_code>.+)"\ncategory = "(?P<category>.+)"\ntitle = "(?P<title>.+)"\ndescription = """\n(?P<description>.*?)\n"""`)
	if err != nil {
		return nil, err
	}

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

				// add plugin-analyzer mapping to our global map
				pluginAnalyzerMap[result["plugin"]] = append(pluginAnalyzerMap[result["plugin"]], node.Name.String())
			}
		}

		return true
	})

	return issues, nil
}

// WriteIssues writes issues extracted from ParseAnnotations to the respective TOML files (issue_code.toml)
func WriteIssues(issues []map[string]string, dir string) error {
	for _, result := range issues {
		if result["issue_code"] != "" {
			fname := dir + result["issue_code"] + ".toml"

			f, err := os.OpenFile(fname, os.O_WRONLY|os.O_CREATE, 0600)
			if err != nil {
				return err
			}

			err = writeTOML(result, f)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// generateTOMLContent generates the TOML content for an issue using the result map.
func generateTOMLContent(result map[string]string) ([]byte, error) {
	type IssueTOML struct {
		IssueCode   string `toml:"issue_code"`
		Category    string `toml:"category"`
		Title       string `toml:"title"`
		Description string `toml:"description"`
	}

	var i IssueTOML

	// only generate content if the issue code is not empty
	if result["issue_code"] != "" {
		i.Title = result["title"]
		i.Description = result["description"]
		i.IssueCode = result["issue_code"]
		i.Category = result["category"]

		content, err := toml.Marshal(i)
		if err != nil {
			return nil, err
		}

		return content, err
	}

	// return an error if the issue code is empty
	return nil, errors.New("issue code is empty")
}

// writeTOML writes the TOML content for an issue to a TOML file.
func writeTOML(result map[string]string, w io.Writer) error {
	content, err := generateTOMLContent(result)
	if err != nil {
		return err
	}

	_, err = w.Write(content)
	if err != nil {
		return err
	}

	return nil
}

// codeGenerator generates Go code for running the analysis. The code generator takes care of setting up boilerplate code (registering plugins, etc.)
func codeGenerator(pluginAnalyzerMap map[string][]string) *jen.File {
	// create new file
	f := jen.NewFile("main")
	f.PackageComment("Code generated by DeepSource. DO NOT EDIT.")

	var statements []jen.Code

	// generate the plugin import statement
	pluginImport := jen.Id("p").Op(":=").Qual("github.com/deepsourcelabs/deepsource-go/plugins", "GoASTPlugin").Values(
		jen.Dict{
			jen.Id("Name"): jen.Lit("go-ast"),
		},
	)

	statements = append(statements, pluginImport)

	// generate code for registering analyzers to a plugin
	for pluginName, pluginAnalyzers := range pluginAnalyzerMap {
		if pluginName != "" {
			for _, analyzer := range pluginAnalyzers {
				registerCode := jen.Id("p").Op(".").Id("RegisterRule").Call(jen.Id(analyzer))

				statements = append(statements, registerCode)
			}
		}
	}

	// generate code for calling BuildAST on a plugin
	buildAST := jen.List(jen.Id("files"), jen.Err()).Op(":=").Id("p").Op(".").Id("BuildAST").Call(jen.Lit("test_dir"))
	statements = append(statements, buildAST)

	// generate code for handling errors
	errCheck := jen.If(
		jen.Err().Op("!=").Nil(),
	).Block(
		jen.Qual("log", "Fatalln").Call(jen.Err()),
	)
	statements = append(statements, errCheck)

	// generate code for calling Run on a plugin
	run := jen.Err().Op("=").Id("p").Op(".").Id("Run").Call(jen.Id("files"))
	statements = append(statements, run)
	statements = append(statements, errCheck)

	// enclose all statements inside main function
	f.Func().Id("main").Params().Block(
		statements...,
	)

	return f
}
