package treerack

import (
	"bytes"
	"testing"
	"time"
)

type testItem struct {
	title          string
	text           string
	fail           bool
	node           *Node
	nodes          []*Node
	ignorePosition bool
}

func runTestsSyntax(t *testing.T, s *Syntax, tests []testItem) {
	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			b := bytes.NewBufferString(test.text)

			start := time.Now()
			n, err := s.Parse(b)
			t.Log("parse duration:", time.Now().Sub(start))

			if test.fail && err == nil {
				t.Error("failed to fail")
				return
			} else if !test.fail && err != nil {
				t.Error(err)
				return
			} else if test.fail {
				return
			}

			if test.node != nil {
				checkNode(t, test.ignorePosition, n, test.node)
			} else {
				checkNodes(t, test.ignorePosition, n.Nodes, test.nodes)
			}
		})
	}
}

func runTests(t *testing.T, syntax string, tests []testItem) {
	s, err := openSyntaxString(syntax)
	if err != nil {
		t.Error(err)
		return
	}

	runTestsSyntax(t, s, tests)
}

func runTestsFile(t *testing.T, file string, tests []testItem) {
	s, err := openSyntaxFile(file)
	if err != nil {
		t.Error(err)
		return
	}

	runTestsSyntax(t, s, tests)
}
