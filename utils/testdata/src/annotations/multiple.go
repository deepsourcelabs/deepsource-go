package annotations

/*deepsource:analyzer
plugin = "go-ast"
issue_code = "NU001"
category = "style"
title = "notused"
description = """
## markdown
```
// comment inside code
multi-line code
```
"""
*/
func NotUsedRule() error {
	return nil
}

/*deepsource:analyzer
plugin = "go-ast"
issue_code = "E001"
category = "bug-risk"
title = "handle error"
description = """
## markdown
```
// comment inside code
multi-line code
```
"""
*/
func HandleErrorRule() error {
	return nil
}
