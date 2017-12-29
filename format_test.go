package treerack

import (
	"bytes"
	"fmt"
	"testing"
)

type formatDefinitionTestItem struct {
	title      string
	definition string
	syntax     string
	output     string
}

func testDefinitionFormatItem(t *testing.T, treerack *Syntax, f formatFlags, test formatDefinitionTestItem) func(t *testing.T) {
	return func(t *testing.T) {
		syntax := test.syntax
		if test.definition != "" {
			syntax = fmt.Sprintf("def = %s", test.definition)
		}

		nodes, err := treerack.Parse(bytes.NewBufferString(syntax))
		if err != nil {
			t.Fatal(err)
		}

		s := &Syntax{}
		if err := define(s, nodes); err != nil {
			t.Fatal(err)
		}

		def, ok := s.registry.definition("def")
		if !ok {
			t.Fatal("failed to register definition")
		}

		output := def.format(s.registry, f)
		if output != test.output {
			t.Error("invalid definition format")
			t.Log("got:     ", output)
			t.Log("expected:", test.output)
		}
	}
}

func testDefinitionFormat(t *testing.T, f formatFlags, tests []formatDefinitionTestItem) {
	treerack, err := bootSyntax()
	if err != nil {
		t.Fatal(err)
	}

	for _, test := range tests {
		t.Run(test.title, testDefinitionFormatItem(t, treerack, f, test))
	}
}

func TestCharFormat(t *testing.T) {
	testDefinitionFormat(t, formatNone, []formatDefinitionTestItem{{
		title:      "empty",
		definition: "[]",
		output:     "[]",
	}, {
		title:      "one char",
		definition: "[a]",
		output:     "[a]",
	}, {
		title:      "escaped char",
		definition: "[\\a]",
		output:     "[a]",
	}, {
		title:      "escaped control char",
		definition: "[\\^]",
		output:     "[\\^]",
	}, {
		title:      "escaped whitespace char",
		definition: "[\\n]",
		output:     "[\\n]",
	}, {
		title:      "escaped verbatim whitespace char",
		definition: "[\n]",
		output:     "[\\n]",
	}, {
		title:      "escaped range",
		definition: "[\\b-\\v]",
		output:     "[\\b-\\v]",
	}, {
		title:      "anything",
		definition: ".",
		output:     ".",
	}, {
		title:      "not something",
		definition: "[^abc]",
		output:     "[^abc]",
	}, {
		title:      "range",
		definition: "[a-z]",
		output:     "[a-z]",
	}, {
		title:      "range and char mixed",
		definition: "[a-z_\\-A-Z]",
		output:     "[_\\-a-zA-Z]",
	}})
}

func TestSequenceFormat(t *testing.T) {
	testDefinitionFormat(t, formatNone, []formatDefinitionTestItem{{
		title:  "empty char sequence",
		syntax: `def = ""`,
		output: `""`,
	}, {
		title:  "char sequence",
		syntax: `def = "abc"`,
		output: `"abc"`,
	}, {
		title:  "char sequence, escaped",
		syntax: `def = "\\n"`,
		output: `"\\n"`,
	}, {
		title:  "chars",
		syntax: `def = "abc" [a-z]`,
		output: `"abc" [a-z]`,
	}, {
		title:  "quantifiers, 0-or-more, single char",
		syntax: `def = "a"*`,
		output: `[a]*`,
	}, {
		title:  "quantifiers, 0-or-more",
		syntax: `def = "abc"*`,
		output: `"abc"*`,
	}, {
		title:  "quantifiers, 1-or-more, single char",
		syntax: `def = "a"+`,
		output: `[a]+`,
	}, {
		title:  "quantifiers, 1-or-more",
		syntax: `def = "abc"+`,
		output: `"abc"+`,
	}, {
		title:  "quantifiers, 0-or-one, single char",
		syntax: `def = "a"?`,
		output: `[a]?`,
	}, {
		title:  "quantifiers, 0-or-one",
		syntax: `def = "abc"?`,
		output: `"abc"?`,
	}, {
		title:  "quantifiers, exact number, single char",
		syntax: `def = "a"{3}`,
		output: `[a]{3}`,
	}, {
		title:  "quantifiers, exact number",
		syntax: `def = "abc"{3}`,
		output: `"abc"{3}`,
	}, {
		title:  "quantifiers, max, single char",
		syntax: `def = "a"{0, 3}`,
		output: `[a]{,3}`,
	}, {
		title:  "quantifiers, max",
		syntax: `def = "abc"{0, 3}`,
		output: `"abc"{,3}`,
	}, {
		title:  "quantifiers, min, single char",
		syntax: `def = "a"{3,}`,
		output: `[a]{3,}`,
	}, {
		title:  "quantifiers, min",
		syntax: `def = "abc"{3,}`,
		output: `"abc"{3,}`,
	}, {
		title:  "quantifiers, range, single char",
		syntax: `def = "a"{3, 9}`,
		output: `[a]{3,9}`,
	}, {
		title:  "quantifiers, range",
		syntax: `def = "abc"{3, 9}`,
		output: `"abc"{3,9}`,
	}, {
		title:  "symbols",
		syntax: `a = "a"; b = "b"; c = "c"; def = a b c`,
		output: "a b c",
	}, {
		title:  "choice in sequence, single char",
		syntax: `def = "a" ("b" | "c")`,
		output: `[a] ([b] | [c])`,
	}, {
		title:  "choice in sequence",
		syntax: `def = "abc" ("def" | "ghi")`,
		output: `"abc" ("def" | "ghi")`,
	}, {
		title:  "grouped quantifier, single char",
		syntax: `def = ("a" "b"){3}`,
		output: `([a] [b]){3}`,
	}, {
		title:  "grouped quantifier",
		syntax: `def = ("abc" "def"){3}`,
		output: `("abc" "def"){3}`,
	}})
}

func TestChoiceFormat(t *testing.T) {
	testDefinitionFormat(t, formatNone, []formatDefinitionTestItem{{
		title:  "choice of char sequences, single char",
		syntax: `def = "a" | "b" | "c"`,
		output: `[a] | [b] | [c]`,
	}, {
		title:  "choice of char sequences",
		syntax: `def = "abc" | "def" | "ghi"`,
		output: `"abc" | "def" | "ghi"`,
	}, {
		title:  "choice of inline sequences, single char",
		syntax: `def = "a" "b" | "c" "d" | "e" "f"`,
		output: `[a] [b] | [c] [d] | [e] [f]`,
	}, {
		title:  "choice of inline sequences",
		syntax: `def = "abc" "def" | "ghi" "jkl" | "mno" "pqr"`,
		output: `"abc" "def" | "ghi" "jkl" | "mno" "pqr"`,
	}, {
		title:  "choice of symbol",
		syntax: `a = "a"; b = "b"; c = "c"; def = a | b | c`,
		output: "a | b | c",
	}})
}

func TestMultiLine(t *testing.T) {
}

func TestLineSplit(t *testing.T) {
}
