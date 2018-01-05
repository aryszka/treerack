package treerack

import "os"

func bootSyntax() (*Syntax, error) {
	f, _ := os.Open("syntax.treerack")
	defer f.Close()
	s := &Syntax{}
	return s, s.ReadSyntax(f)
}
