package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
)

func concatGo(w io.Writer, r ...io.Reader) error {
	var others []ast.Decl
	imports := &ast.GenDecl{
		Tok:    token.IMPORT,
		Lparen: 1,
		Rparen: 1,
	}

	importPaths := make(map[string]bool)
	for i := range r {
		f, err := parser.ParseFile(token.NewFileSet(), "", r[i], 0)
		if err != nil {
			return err
		}

		for j := range f.Decls {
			d := f.Decls[j]
			switch dd := d.(type) {
			case *ast.GenDecl:
				if dd.Tok != token.IMPORT {
					others = append(others, d)
					continue
				}

				for k := range dd.Specs {
					i := dd.Specs[k].(*ast.ImportSpec)
					path := i.Path.Value
					name := ""
					if i.Name != nil {
						name = i.Name.Name
					}

					key := "(" + name + ")" + path
					if importPaths[key] {
						continue
					}

					importPaths[key] = true
					imports.Specs = append(imports.Specs, i)
				}
			default:
				others = append(others, d)
			}
		}
	}

	return printer.Fprint(w, token.NewFileSet(), &ast.File{
		Name:  &ast.Ident{Name: ""},
		Decls: append([]ast.Decl{imports}, others...),
	})
}

func removePackage(gO string) string {
	// so we know that it's the first line
	nl := strings.Index(gO, "\n")
	return gO[nl+2:]
}

func main() {
	var packageName string
	flag.StringVar(&packageName, "package", "treerack", "package name of the generated code file")
	flag.Parse()

	var files []io.Reader
	for _, fn := range flag.Args() {
		f, err := os.Open(fn)
		if err != nil {
			log.Fatalln(err)
		}

		defer f.Close()
		files = append(files, f)
	}

	var headCode bytes.Buffer
	if err := concatGo(&headCode, files...); err != nil {
		log.Fatalln(err)
	}

	code := headCode.String()
	code = removePackage(code)
	quotedCode := strconv.Quote(code)

	fmt.Printf("package %s\n\n// generated with scripts/createhead.go\nconst headCode=%s", packageName, quotedCode)
}
