package main

import (
	"log"
	"os"

	"github.com/aryszka/treerack"
)

const (
	syntaxPath = "syntax.treerack"
	outputPath = "parser.go"
)

func main() {
	s := &treerack.Syntax{}

	if err := s.ReadSyntax(os.Stdin); err != nil {
		log.Fatalln(err)
	}

	if err := s.Generate(os.Stdout); err != nil {
		log.Fatalln(err)
	}
}
