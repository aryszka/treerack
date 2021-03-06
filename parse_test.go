package treerack

import (
	"bufio"
	"bytes"
	"testing"
)

func TestRecursion(t *testing.T) {
	runTests(
		t,
		`A = "a" | A "a"`,
		[]testItem{{
			title: "recursion in choice, right, left, commit",
			text:  "aaa",
			node: &Node{
				Name: "A",
				Nodes: []*Node{{
					Name: "A",
					Nodes: []*Node{{
						Name: "A",
					}},
				}},
			},
			ignorePosition: true,
		}},
	)

	runTests(
		t,
		`A = "a" | "a" A`,
		[]testItem{{
			title: "recursion in choice, right, right, commit",
			text:  "aaa",
			node: &Node{
				Name: "A",
				Nodes: []*Node{{
					Name: "A",
					Nodes: []*Node{{
						Name: "A",
					}},
				}},
			},
			ignorePosition: true,
		}},
	)

	runTests(
		t,
		`A = "a" A | "a"`,
		[]testItem{{
			title: "recursion in choice, left, right, commit",
			text:  "aaa",
			node: &Node{
				Name: "A",
				Nodes: []*Node{{
					Name: "A",
					Nodes: []*Node{{
						Name: "A",
					}},
				}},
			},
			ignorePosition: true,
		}},
	)

	runTests(
		t,
		`A = A "a" | "a"`,
		[]testItem{{
			title: "recursion in choice, left, left, commit",
			text:  "aaa",
			node: &Node{
				Name: "A",
				Nodes: []*Node{{
					Name: "A",
					Nodes: []*Node{{
						Name: "A",
					}},
				}},
			},
			ignorePosition: true,
		}},
	)

	runTests(
		t,
		`A':alias = "a" | A' "a"; A = A'`,
		[]testItem{{
			title: "recursion in choice, right, left, alias",
			text:  "aaa",
			node: &Node{
				Name: "A",
				To:   3,
			},
		}},
	)

	runTests(
		t,
		`A':alias = "a" | "a" A'; A = A'`,
		[]testItem{{
			title: "recursion in choice, right, right, alias",
			text:  "aaa",
			node: &Node{
				Name: "A",
				To:   3,
			},
		}},
	)

	runTests(
		t,
		`A':alias = "a" A' | "a"; A = A'`,
		[]testItem{{
			title: "recursion in choice, left, right, alias",
			text:  "aaa",
			node: &Node{
				Name: "A",
				To:   3,
			},
		}},
	)

	runTests(
		t,
		`A':alias = A' "a" | "a"; A = A'`,
		[]testItem{{
			title: "recursion in choice, left, left, alias",
			text:  "aaa",
			node: &Node{
				Name: "A",
				To:   3,
			},
		}},
	)

	runTests(
		t,
		`A = "a" | A*`,
		[]testItem{{
			title:          "recursive sequence in choice",
			text:           "aaaa",
			ignorePosition: true,
			node: &Node{
				Name: "A",
				Nodes: []*Node{{
					Name: "A",
				}, {
					Name: "A",
					Nodes: []*Node{{
						Name: "A",
					}, {
						Name: "A",
						Nodes: []*Node{{
							Name: "A",
						}, {
							Name: "A",
						}},
					}},
				}},
			},
		}},
	)
}

func TestSequence(t *testing.T) {
	runTests(
		t,
		`AB = "a" | "a"? "a"? "b" "b"`,
		[]testItem{{
			title: "sequence with optional items",
			text:  "abb",
			node: &Node{
				Name: "AB",
				To:   3,
			},
		}, {
			title: "sequence with optional items, none",
			text:  "bb",
			node: &Node{
				Name: "AB",
				To:   2,
			},
		}},
	)

	runTests(
		t,
		`A = "a" | (A?)*`,
		[]testItem{{
			title: "recursive sequence in choice with redundant quantifier",
			text:  "aaa",
			node: &Node{
				Name: "A",
				Nodes: []*Node{{
					Name: "A",
				}, {
					Name: "A",
				}, {
					Name: "A",
				}},
			},
			ignorePosition: true,
		}},
	)

	runTests(
		t,
		`A = ("a"*)*`,
		[]testItem{{
			title: "sequence with redundant quantifier",
			text:  "aaa",
			node: &Node{
				Name: "A",
				To:   3,
			},
		}},
	)

	runTests(
		t,
		`a = "a"?; A = a | a*`,
		[]testItem{{
			title:          "single or zero-or-more optional in choice",
			text:           "aaa",
			ignorePosition: true,
			node: &Node{
				Name: "A",
				Nodes: []*Node{{
					Name: "a",
				}, {
					Name: "a",
				}, {
					Name: "a",
				}},
			},
		}},
	)

	runTests(
		t,
		`a = "a"{3,5}`,
		[]testItem{{
			title: "less than min",
			text:  "aa",
			fail:  true,
		}, {
			title:          "just min",
			text:           "aaa",
			ignorePosition: true,
			node: &Node{
				Name: "a",
			},
		}, {
			title:          "less than max",
			text:           "aaaa",
			ignorePosition: true,
			node: &Node{
				Name: "a",
			},
		}, {
			title:          "just max",
			text:           "aaaaa",
			ignorePosition: true,
			node: &Node{
				Name: "a",
			},
		}, {
			title: "more than max",
			text:  "aaaaaa",
			fail:  true,
		}},
	)
}

func TestQuantifiers(t *testing.T) {
	runTests(
		t,
		`A = "a" "b"{0} "a"`,
		[]testItem{{
			title: "zero, considered as one",
			text:  "aba",
			node: &Node{
				Name: "A",
				To:   3,
			},
		}, {
			title: "zero, fail",
			text:  "aa",
			fail:  true,
		}},
	)

	runTests(
		t,
		`A = "a" "b"{1} "a"`,
		[]testItem{{
			title: "one, missing",
			text:  "aa",
			fail:  true,
		}, {
			title: "one",
			text:  "aba",
			node: &Node{
				Name: "A",
				To:   3,
			},
		}, {
			title: "one, too much",
			text:  "abba",
			fail:  true,
		}},
	)

	runTests(
		t,
		`A = "a" "b"{3} "a"`,
		[]testItem{{
			title: "three, missing",
			text:  "abba",
			fail:  true,
		}, {
			title: "three",
			text:  "abbba",
			node: &Node{
				Name: "A",
				To:   5,
			},
		}, {
			title: "three, too much",
			text:  "abbbba",
			fail:  true,
		}},
	)

	runTests(
		t,
		`A = "a" "b"{0,1} "a"`,
		[]testItem{{
			title: "zero or one explicit, missing",
			text:  "aa",
			node: &Node{
				Name: "A",
				To:   2,
			},
		}, {
			title: "zero or one explicit",
			text:  "aba",
			node: &Node{
				Name: "A",
				To:   3,
			},
		}, {
			title: "zero or one explicit, too much",
			text:  "abba",
			fail:  true,
		}},
	)

	runTests(
		t,
		`A = "a" "b"{,1} "a"`,
		[]testItem{{
			title: "zero or one explicit, omit zero, missing",
			text:  "aa",
			node: &Node{
				Name: "A",
				To:   2,
			},
		}, {
			title: "zero or one explicit, omit zero",
			text:  "aba",
			node: &Node{
				Name: "A",
				To:   3,
			},
		}, {
			title: "zero or one explicit, omit zero, too much",
			text:  "abba",
			fail:  true,
		}},
	)

	runTests(
		t,
		`A = "a" "b"? "a"`,
		[]testItem{{
			title: "zero or one explicit, shortcut, missing",
			text:  "aa",
			node: &Node{
				Name: "A",
				To:   2,
			},
		}, {
			title: "zero or one explicit, shortcut",
			text:  "aba",
			node: &Node{
				Name: "A",
				To:   3,
			},
		}, {
			title: "zero or one explicit, shortcut, too much",
			text:  "abba",
			fail:  true,
		}},
	)

	runTests(
		t,
		`A = "a" "b"{0,3} "a"`,
		[]testItem{{
			title: "zero or three, missing",
			text:  "aa",
			node: &Node{
				Name: "A",
				To:   2,
			},
		}, {
			title: "zero or three",
			text:  "abba",
			node: &Node{
				Name: "A",
				To:   4,
			},
		}, {
			title: "zero or three",
			text:  "abbba",
			node: &Node{
				Name: "A",
				To:   5,
			},
		}, {
			title: "zero or three, too much",
			text:  "abbbba",
			fail:  true,
		}},
	)

	runTests(
		t,
		`A = "a" "b"{,3} "a"`,
		[]testItem{{
			title: "zero or three, omit zero, missing",
			text:  "aa",
			node: &Node{
				Name: "A",
				To:   2,
			},
		}, {
			title: "zero or three, omit zero",
			text:  "abba",
			node: &Node{
				Name: "A",
				To:   4,
			},
		}, {
			title: "zero or three, omit zero",
			text:  "abbba",
			node: &Node{
				Name: "A",
				To:   5,
			},
		}, {
			title: "zero or three, omit zero, too much",
			text:  "abbbba",
			fail:  true,
		}},
	)

	runTests(
		t,
		`A = "a" "b"{1,3} "a"`,
		[]testItem{{
			title: "one or three, missing",
			text:  "aa",
			fail:  true,
		}, {
			title: "one or three",
			text:  "abba",
			node: &Node{
				Name: "A",
				To:   4,
			},
		}, {
			title: "one or three",
			text:  "abbba",
			node: &Node{
				Name: "A",
				To:   5,
			},
		}, {
			title: "one or three, too much",
			text:  "abbbba",
			fail:  true,
		}},
	)

	runTests(
		t,
		`A = "a" "b"{3,5} "a"`,
		[]testItem{{
			title: "three or five, missing",
			text:  "abba",
			fail:  true,
		}, {
			title: "three or five",
			text:  "abbbba",
			node: &Node{
				Name: "A",
				To:   6,
			},
		}, {
			title: "three or five",
			text:  "abbbbba",
			node: &Node{
				Name: "A",
				To:   7,
			},
		}, {
			title: "three or five, too much",
			text:  "abbbbbba",
			fail:  true,
		}},
	)

	runTests(
		t,
		`A = "a" "b"{0,} "a"`,
		[]testItem{{
			title: "zero or more, explicit, missing",
			text:  "aa",
			node: &Node{
				Name: "A",
				To:   2,
			},
		}, {
			title: "zero or more, explicit",
			text:  "abba",
			node: &Node{
				Name: "A",
				To:   4,
			},
		}},
	)

	runTests(
		t,
		`A = "a" "b"* "a"`,
		[]testItem{{
			title: "zero or more, shortcut, missing",
			text:  "aa",
			node: &Node{
				Name: "A",
				To:   2,
			},
		}, {
			title: "zero or more, shortcut",
			text:  "abba",
			node: &Node{
				Name: "A",
				To:   4,
			},
		}},
	)

	runTests(
		t,
		`A = "a" "b"{1,} "a"`,
		[]testItem{{
			title: "one or more, explicit, missing",
			text:  "aa",
			fail:  true,
		}, {
			title: "one or more, explicit",
			text:  "abba",
			node: &Node{
				Name: "A",
				To:   4,
			},
		}},
	)

	runTests(
		t,
		`A = "a" "b"+ "a"`,
		[]testItem{{
			title: "one or more, shortcut, missing",
			text:  "aa",
			fail:  true,
		}, {
			title: "one or more, shortcut",
			text:  "abba",
			node: &Node{
				Name: "A",
				To:   4,
			},
		}},
	)

	runTests(
		t,
		`A = "a" "b"{3,} "a"`,
		[]testItem{{
			title: "three or more, explicit, missing",
			text:  "abba",
			fail:  true,
		}, {
			title: "three or more, explicit",
			text:  "abbbba",
			node: &Node{
				Name: "A",
				To:   6,
			},
		}},
	)
}

func TestEmpty(t *testing.T) {
	runTests(
		t,
		`A = "1"`,
		[]testItem{{
			title: "empty primitive, fail",
			fail:  true,
		}},
	)

	runTests(
		t,
		`A = "1"?`,
		[]testItem{{
			title: "empty primitive, succeed",
		}},
	)

	runTests(
		t,
		`a = "1"?; A = a a`,
		[]testItem{{
			title: "empty document with quantifiers in the item",
			node: &Node{
				Name: "A",
				Nodes: []*Node{{
					Name: "a",
				}, {
					Name: "a",
				}},
			},
		}},
	)

	runTests(
		t,
		`a = "1"; A = a? a?`,
		[]testItem{{
			title: "empty document with quantifiers in the reference",
			node: &Node{
				Name: "A",
			},
		}},
	)

	runTests(
		t,
		`a = [a]*; a':alias = a; a'' = a' [^a]*`,
		[]testItem{{
			title:          "no a",
			text:           "b",
			ignorePosition: true,
			node: &Node{
				Name: "a''",
				Nodes: []*Node{{
					Name: "a",
				}},
			},
		}},
	)
}

func TestCharAsRoot(t *testing.T) {
	runTests(
		t,
		`A = "a"`,
		[]testItem{{
			title:          "char as root",
			text:           "a",
			ignorePosition: true,
			node: &Node{
				Name: "A",
			},
		}},
	)
}

func TestPartialRead(t *testing.T) {
	runTests(
		t,
		`A = "a"`,
		[]testItem{{
			title: "document finished before eof",
			text:  "ab",
			fail:  true,
		}},
	)

	runTests(
		t,
		`A = "a"*`,
		[]testItem{{
			title: "document finished before eof with reading past",
			text:  "ab",
			fail:  true,
		}},
	)
}

func TestChoiceSequencePriority(t *testing.T) {
	runTests(
		t,
		`A = "a" | "b" "c"`,
		[]testItem{{
			title: "ac",
			text:  "ac",
			fail:  true,
		}},
	)

	runTests(
		t,
		`A = "a" | "b" "c"`,
		[]testItem{{
			title:          "bc",
			text:           "bc",
			ignorePosition: true,
			node:           &Node{Name: "A"},
		}},
	)
}

func TestCharBuildNoop(t *testing.T) {
	c := newChar("foo", false, nil, nil)
	c.init(newRegistry())
	b := c.builder()
	ctx := newContext(bufio.NewReader(bytes.NewBuffer(nil)), nil)
	if n, ok := b.build(ctx); len(n) != 0 || ok {
		t.Error("char build not noop")
	}
}
