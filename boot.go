package treerack

import "os"

func bootSyntax() (*Syntax, error) {
	f, _ := os.Open("syntax.treerack")
	defer f.Close()

	// never fails:
	doc, _ := parsegen(f)

	s := &Syntax{}
	return s, define(s, doc)
}
