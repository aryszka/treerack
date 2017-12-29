package treerack

import (
	"bytes"
	"reflect"
	"testing"
)

func checkParseError(left, right ParseError) bool {
	left.registry = nil
	right.registry = nil
	return reflect.DeepEqual(left, right)
}

func TestError(t *testing.T) {
	type testItem struct {
		title  string
		syntax string
		text   string
		perr   ParseError
	}

	for _, test := range []testItem{{
		title:  "single def, empty text",
		syntax: `a = "a"`,
		perr: ParseError{
			Definition: "a",
		},
	}, {
		title:  "single def, wrong text",
		syntax: `a = "a"`,
		text:   "b",
		perr: ParseError{
			Definition: "a",
		},
	}, {
		title:  "single optional def, wrong text",
		syntax: `a = "a"?`,
		text:   "b",
		perr: ParseError{
			Definition: "a",
		},
	}, {
		title:  "error on second line, second column",
		syntax: `a = [a\n]*`,
		text:   "aa\nabaa\naa",
		perr: ParseError{
			Offset:     4,
			Line:       1,
			Column:     1,
			Definition: "a",
		},
	}, {
		title:  "multiple definitions",
		syntax: `a = "aa"; A:root = a`,
		text:   "ab",
		perr: ParseError{
			Offset:     1,
			Column:     1,
			Definition: "a",
		},
	}, {
		title:  "choice, options succeed",
		syntax: `a = "12"; b = "1"; c:root = a | b`,
		text:   "123",
		perr: ParseError{
			Offset:     2,
			Column:     2,
			Definition: "c",
		},
	}, {
		title:  "choice succeeds, document fails",
		syntax: `a = "12"; b = "1"; c:root = a | b`,
		text:   "13",
		perr: ParseError{
			Offset:     1,
			Column:     1,
			Definition: "c",
		},
	}, {
		title:  "choice fails",
		syntax: `a = "12"; b = "2"; c:root = a | b`,
		text:   "13",
		perr: ParseError{
			Offset:     1,
			Column:     1,
			Definition: "a",
		},
	}, {
		title:  "choice fails, longer option reported",
		syntax: `a = "12"; b = "134"; c:root = a | b`,
		text:   "135",
		perr: ParseError{
			Offset:     2,
			Column:     2,
			Definition: "b",
		},
	}, {
		title:  "failing choice on the failing branch",
		syntax: `a = "123"; b:root = a | "13"`,
		text:   "124",
		perr: ParseError{
			Offset:     2,
			Column:     2,
			Definition: "a",
		},
	}, {
		title:  "failing choice on a shorter branch",
		syntax: `a = "13"; b:root = "123" | a`,
		text:   "124",
		perr: ParseError{
			Offset:     2,
			Column:     2,
			Definition: "b",
		},
	}, {
		title:  "longer failure on a later pass",
		syntax: `a = "12"; b = "34"; c = "1" b; d:root = a | c`,
		text:   "135",
		perr: ParseError{
			Offset:     2,
			Column:     2,
			Definition: "b",
		},
	}, {
		title:  "char as a choice option",
		syntax: `a = "12"; b = [a] | [b]; c = a b`,
		text:   "12c",
		perr: ParseError{
			Offset:     2,
			Column:     2,
			Definition: "b",
		},
	}} {
		t.Run(test.title, func(t *testing.T) {
			s, err := openSyntaxString(test.syntax)
			if err != nil {
				t.Error(err)
				return
			}

			_, err = s.Parse(bytes.NewBufferString(test.text))
			if err == nil {
				t.Error("failed to fail")
				return
			}

			perr, ok := err.(*ParseError)
			if !ok {
				t.Error("invalid error returned", err)
				return
			}

			if perr.Input != "<input>" {
				t.Error("invalid default input name")
				return
			}

			perr.Input = ""
			if !checkParseError(*perr, test.perr) {
				t.Error("failed to return the right error")
			}
		})
	}
}

func TestErrorRecursive(t *testing.T) {
	const syntax = `
		ws:ws = " ";
		symbol = [a-z]+;
		function-application = expression "(" expression? ")";
		expression = function-application | symbol;
		doc:root = (expression (";" expression)*)+;
	`

	s, err := openSyntaxString(syntax)
	if err != nil {
		t.Fatal(err)
	}

	for _, test := range []struct {
		title string
		doc   string
		perr  ParseError
	}{{
		title: "simple, open",
		doc:   "a(",
		perr: ParseError{
			Offset:     1,
			Column:     1,
			Definition: "function-application",
		},
	}, {
		title: "simple, close",
		doc:   "a)",
		perr: ParseError{
			Offset:     1,
			Column:     1,
			Definition: "function-application",
		},
	}, {
		title: "inner, open",
		doc:   "a(b()",
		perr: ParseError{
			Offset:     1,
			Column:     1,
			Definition: "function-application",
		},
	}, {
		title: "inner, close",
		doc:   "a(b))",
		perr: ParseError{
			Offset:     4,
			Column:     4,
			Definition: "function-application",
		},
	}, {
		title: "outer, open",
		doc:   "a()b(",
		perr: ParseError{
			Offset:     4,
			Column:     4,
			Definition: "function-application",
		},
	}, {
		title: "outer, close",
		doc:   "a()b)",
		perr: ParseError{
			Offset:     4,
			Column:     4,
			Definition: "function-application",
		},
	}} {
		t.Run(test.title, func(t *testing.T) {
			_, err := s.Parse(bytes.NewBufferString(test.doc))
			if err == nil {
				t.Fatal("failed to fail")
			}

			perr, ok := err.(*ParseError)
			if !ok {
				t.Fatal("invalid error type")
			}

			perr.Input = ""

			if !checkParseError(*perr, test.perr) {
				t.Error("failed to return the right error")
				t.Log("got:     ", *perr)
				t.Log("expected:", test.perr)
			}
		})
	}
}

func TestErrorMessage(t *testing.T) {
	const expected = "foo:4:10:parse failed, parsing: bar"

	perr := &ParseError{
		Input:      "foo",
		Offset:     42,
		Line:       3,
		Column:     9,
		Definition: "bar",
	}

	message := perr.Error()
	if message != expected {
		t.Error("failed to return the right error message")
		t.Log("got:     ", message)
		t.Log("expected:", expected)
	}
}

func TestErrorVerbose(t *testing.T) {
	const expected = `<input>:5:2:parse failed, parsing: string`

	const doc = `{
		"a":1,
		"b":2,
		"c":3,
	}`

	s, err := openSyntaxFile("examples/json.treerack")
	if err != nil {
		t.Error(err)
		return
	}

	_, err = s.Parse(bytes.NewBufferString(doc))
	perr, ok := err.(*ParseError)
	if !ok {
		t.Error("failed to return parse error")
		return
	}

	if perr.Error() != expected {
		t.Error("failed to get the right error message")
		t.Log("got:     ", perr.Error())
		t.Log("expected:", expected)
	}
}
