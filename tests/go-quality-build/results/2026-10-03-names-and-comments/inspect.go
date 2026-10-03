package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"reflect"
	"strings"
)

func main() {
	path := os.Args[1]
	data, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	tree, err := parser.ParseFile(token.NewFileSet(), path, data, 0)
	if err != nil {
		panic(err)
	}
	if len(os.Args) > 2 && os.Args[2] == "ledger" {
		ast.Inspect(tree, func(node ast.Node) bool {
			if id, ok := node.(*ast.Ident); ok {
				switch id.Name {
				case "parseRow":
					id.Name = "doIt"
				case "rawValue":
					id.Name = "raw"
				}
			}
			return true
		})
	}
	var canonical bytes.Buffer
	filter := func(name string, value reflect.Value) bool {
		switch name {
		case "Obj", "Scope", "Unresolved", "Doc", "Comments":
			return false
		}
		return value.Type() != reflect.TypeOf(token.Pos(0))
	}
	if err := ast.Fprint(&canonical, nil, tree, filter); err != nil {
		panic(err)
	}
	fingerprint := sha256.Sum256(canonical.Bytes())
	documented, err := parser.ParseFile(token.NewFileSet(), path, data, parser.ParseComments)
	if err != nil {
		panic(err)
	}
	words := 0
	for _, group := range documented.Comments {
		words += len(strings.Fields(group.Text()))
	}
	output := struct {
		Fingerprint string `json:"ast_sha256"`
		Words       int    `json:"comment_words"`
	}{
		hex.EncodeToString(fingerprint[:]), words,
	}
	encoded, err := json.Marshal(output)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(encoded))
}
