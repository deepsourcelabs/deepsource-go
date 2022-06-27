// Code generated by DeepSource. DO NOT EDIT.
package main

import (
	"log"
	"os"

	goast "github.com/deepsourcelabs/deepsource-go/analyzers/goast"
	"github.com/deepsourcelabs/deepsource-go/utils/testdata/src/codegen/rules"
)

func main() {
        p := goast.GoASTAnalyzer{Name: "go-ast"}
        p.RegisterRule(rules.NeedFunc)
        files, err := p.BuildAST(os.Getenv("CODE_PATH"))
        if err != nil {
                log.Fatalln(err)
        }
        err = p.Run(files)
        if err != nil {
                log.Fatalln(err)
        }
}
