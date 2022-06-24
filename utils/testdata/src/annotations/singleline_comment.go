package annotations

// deepsource:analyzer
// plugin = "go-ast"
// issue_code = "P001"
// category = "performance"
// title = "Multiple appends can be combined into a single statement"
// description = """
// ## markdown
// ```
// // comment inside code
// multi-line code
// ```
// """
func MultipleAppendRule() error {
	return nil
}
