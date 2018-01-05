package main

import (
	"log"
	"os"

	"github.com/aryszka/treerack"
)

func main() {
	s := &treerack.Syntax{}

	if err := s.ReadSyntax(os.Stdin); err != nil {
		log.Fatalln(err)
	}

	if err := s.Generate(treerack.GeneratorOptions{PackageName: "self"}, os.Stdout); err != nil {
		log.Fatalln(err)
	}
}
