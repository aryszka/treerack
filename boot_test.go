package treerack

import (
	"io"
	"os"
	"testing"
)

func parseWithSyntax(s *Syntax, f io.ReadSeeker) (*Node, error) {
	if _, err := f.Seek(0, 0); err != nil {
		return nil, err
	}

	return s.Parse(f)
}

func syntaxFromTree(n *Node) (*Syntax, error) {
	s := NewSyntax()
	if err := define(s, n); err != nil {
		return nil, err
	}

	if err := s.Init(); err != nil {
		return nil, err
	}

	return s, nil
}

func testParseFromTree(t *testing.T, n *Node, f io.ReadSeeker) *Node {
	s, err := syntaxFromTree(n)
	if err != nil {
		t.Error(err)
		return nil
	}

	nn, err := parseWithSyntax(s, f)
	if err != nil {
		t.Error(err)
		return nil
	}

	checkNode(t, nn, n)
	return nn
}

func TestBoot(t *testing.T) {
	b, err := createBoot()
	if err != nil {
		t.Error(err)
		return
	}

	f, err := os.Open("syntax.parser")
	if err != nil {
		t.Error(err)
		return
	}

	defer f.Close()

	n0, err := parseWithSyntax(b, f)
	if err != nil {
		t.Error(err)
		return
	}

	n1 := testParseFromTree(t, n0, f)
	if t.Failed() {
		return
	}

	testParseFromTree(t, n1, f)
}
