package treerack

import (
	"fmt"
	"testing"
)

func TestCharFormat(t *testing.T) {
	type testItem struct {
		title      string
		definition string
		output     string
	}

	for _, test := range []testItem{{
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
	}} {
		t.Run(test.title, func(t *testing.T) {
			defString := fmt.Sprintf("def = %s", test.definition)
			s, err := openSyntaxString(defString)
			if err != nil {
				t.Error(err)
				return
			}

			def, ok := s.registry.definition(childName("def", 0))
			if !ok {
				t.Error("invalid syntax")
				return
			}

			output := def.format(s.registry, formatNone)
			if output != test.output {
				t.Error("invalid output", output, test.output)
			}
		})
	}
}

func TestSequenceFormat(t *testing.T) {
	type testItem struct {
		title  string
		syntax string
		output string
	}

	for _, test := range []testItem{{
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
		title:  "quantifiers, 0-or-more",
		syntax: `def = "a"*`,
		output: `"a"*`,
	}, {
		title:  "quantifiers, 1-or-more",
		syntax: `def = "a"+`,
		output: `"a"+`,
	}, {
		title:  "quantifiers, 0-or-one",
		syntax: `def = "a"?`,
		output: `"a"?`,
	}, {
		title:  "quantifiers, exact number",
		syntax: `def = "a"{3}`,
		output: `"a"{3}`,
	}, {
		title:  "quantifiers, max",
		syntax: `def = "a"{0, 3}`,
		output: `"a"{,3}`,
	}, {
		title:  "quantifiers, min",
		syntax: `def = "a"{3,}`,
		output: `"a"{3,}`,
	}, {
		title:  "quantifiers, range",
		syntax: `def = "a"{3, 9}`,
		output: `"a"{3,9}`,
	}, {
		title:  "symbols",
		syntax: `a = "a"; b = "b"; c = "c"; def = a b c`,
		output: "a b c",
	}, {
		title:  "choice in sequence",
		syntax: `def = "a" ("b" | "c")`,
		output: `"a" ("b" | "c")`,
	}, {
		title:  "grouped quantifier",
		syntax: `def = ("a" "b"){3}`,
		output: `("a" "b"){3}`,
	}} {
		t.Run(test.title, func(t *testing.T) {
			s, err := openSyntaxString(test.syntax)
			if err != nil {
				t.Error(err)
				return
			}

			output := s.root.format(s.registry, formatNone)
			if output != test.output {
				t.Error("invalid output", output, test.output)
			}
		})
	}
}

func TestChoiceFormat(t *testing.T) {
	type testItem struct {
		title  string
		syntax string
		output string
	}

	for _, test := range []testItem{{
		title:  "choice of char sequences",
		syntax: `def = "a" | "b" | "c"`,
		output: `"a" | "b" | "c"`,
	}, {
		title:  "choice of inline sequences",
		syntax: `def = "a" "b" | "c" "d" | "e" "f"`,
		output: `"a" "b" | "c" "d" | "e" "f"`,
	}, {
		title:  "choice of symbol",
		syntax: `a = "a"; b = "b"; c = "c"; def = a | b | c`,
		output: "a | b | c",
	}} {
		t.Run(test.title, func(t *testing.T) {
			s, err := openSyntaxString(test.syntax)
			if err != nil {
				t.Error(err)
				return
			}

			output := s.root.format(s.registry, formatNone)
			if output != test.output {
				t.Error("invalid output", output, test.output)
			}
		})
	}
}
