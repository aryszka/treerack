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

func runTestsGetSyntax(t *testing.T, getSyntax func(t *testing.T) *Syntax, tests []testItem) {
	var s *Syntax
	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			if s == nil {
				s = getSyntax(t)
			}

			if t.Failed() {
				return
			}

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

func runTestsSyntax(t *testing.T, s *Syntax, tests []testItem) {
	runTestsGetSyntax(t, func(*testing.T) *Syntax { return s }, tests)
}

func runTests(t *testing.T, syntax string, tests []testItem) {
	getSyntax := func(t *testing.T) *Syntax {
		s, err := openSyntaxString(syntax)
		if err != nil {
			t.Error(err)
		}

		return s
	}

	runTestsGetSyntax(t, getSyntax, tests)
}

func runTestsFile(t *testing.T, file string, tests []testItem) {
	getSyntax := func(t *testing.T) *Syntax {
		s, err := openSyntaxFile(file)
		if err != nil {
			t.Error(err)
		}

		return s
	}

	runTestsGetSyntax(t, getSyntax, tests)
}
