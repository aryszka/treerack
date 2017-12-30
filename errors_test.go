package treerack

import (
	"bytes"
	"reflect"
	"testing"
)

type errorTestItem struct {
	title  string
	syntax string
	doc    string
	perr   ParseError
}

func checkParseError(left, right ParseError) bool {
	left.registry = nil
	right.registry = nil
	return reflect.DeepEqual(left, right)
}

func testParseErrorItem(s *Syntax, test errorTestItem) func(t *testing.T) {
	return func(t *testing.T) {
		_, err := s.Parse(bytes.NewBufferString(test.doc))
		if err == nil {
			t.Fatal("failed to fail")
		}

		perr, ok := err.(*ParseError)
		if !ok {
			t.Fatal("invalid error type returned")
		}

		perr.Input = ""
		perr.registry = nil

		if !checkParseError(*perr, test.perr) {
			t.Error("invalid error returned")
			t.Log("got:     ", *perr)
			t.Log("expected:", test.perr)
		}
	}
}

func testParseError(t *testing.T, syntax string, tests []errorTestItem) {
	var s *Syntax
	if syntax != "" {
		var err error
		s, err = openSyntaxString(syntax)
		if err != nil {
			t.Fatal(err)
		}
	}

	for _, test := range tests {
		ts := s
		if test.syntax != "" {
			var err error
			ts, err = openSyntaxString(test.syntax)
			if err != nil {
				t.Fatal(err)
			}
		}

		if ts == nil {
			t.Fatal("no syntax defined")
		}

		t.Run(test.title, testParseErrorItem(ts, test))
	}
}

func TestError(t *testing.T) {
	testParseError(t, "", []errorTestItem{{
		title:  "single def, empty text",
		syntax: `a = "a"`,
		perr: ParseError{
			Definition: "a",
		},
	}, {
		title:  "single def, wrong text",
		syntax: `a = "a"`,
		doc:    "b",
		perr: ParseError{
			Definition: "a",
		},
	}, {
		title:  "single optional def, wrong text",
		syntax: `a = "a"?`,
		doc:    "b",
		perr: ParseError{
			Definition: "a",
		},
	}, {
		title:  "error on second line, second column",
		syntax: `a = [a\n]*`,
		doc:    "aa\nabaa\naa",
		perr: ParseError{
			Offset:     4,
			Line:       1,
			Column:     1,
			Definition: "a",
		},
	}, {
		title:  "multiple definitions",
		syntax: `a = "aa"; A:root = a`,
		doc:    "ab",
		perr: ParseError{
			Offset:     1,
			Column:     1,
			Definition: "a",
		},
	}, {
		title:  "choice, options succeed",
		syntax: `a = "12"; b = "1"; c:root = a | b`,
		doc:    "123",
		perr: ParseError{
			Offset:     2,
			Column:     2,
			Definition: "c",
		},
	}, {
		title:  "choice succeeds, document fails",
		syntax: `a = "12"; b = "1"; c:root = a | b`,
		doc:    "13",
		perr: ParseError{
			Offset:     1,
			Column:     1,
			Definition: "c",
		},
	}, {
		title:  "choice fails",
		syntax: `a = "12"; b = "2"; c:root = a | b`,
		doc:    "13",
		perr: ParseError{
			Offset:     1,
			Column:     1,
			Definition: "a",
		},
	}, {
		title:  "choice fails, longer option reported",
		syntax: `a = "12"; b = "134"; c:root = a | b`,
		doc:    "135",
		perr: ParseError{
			Offset:     2,
			Column:     2,
			Definition: "b",
		},
	}, {
		title:  "failing choice on the failing branch",
		syntax: `a = "123"; b:root = a | "13"`,
		doc:    "124",
		perr: ParseError{
			Offset:     2,
			Column:     2,
			Definition: "a",
		},
	}, {
		title:  "failing choice on a shorter branch",
		syntax: `a = "13"; b:root = "123" | a`,
		doc:    "124",
		perr: ParseError{
			Offset:     2,
			Column:     2,
			Definition: "b",
		},
	}, {
		title:  "longer failure on a later pass",
		syntax: `a = "12"; b = "34"; c = "1" b; d:root = a | c`,
		doc:    "135",
		perr: ParseError{
			Offset:     2,
			Column:     2,
			Definition: "b",
		},
	}, {
		title:  "char as a choice option",
		syntax: `a = "12"; b = [a] | [b]; c = a b`,
		doc:    "12c",
		perr: ParseError{
			Offset:     2,
			Column:     2,
			Definition: "b",
		},
	}})
}

func TestErrorRecursive(t *testing.T) {
	const syntax = `
		ws:ws = " ";
		symbol = [a-z]+;
		function-application = expression "(" expression? ")";
		expression = function-application | symbol;
		doc:root = (expression (";" expression)*)+;
	`

	testParseError(t, syntax, []errorTestItem{{
		title: "simple, open",
		doc:   "a(",
		perr: ParseError{
			Offset:     2,
			Column:     2,
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
			Offset:     5,
			Column:     5,
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
			Offset:     5,
			Column:     5,
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
	}})
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
	t.Skip()

	// 	const expected = `<input>:5:2:parse failed, parsing: string
	//
	// 		"c":3,
	// 	}<<<
	//
	// Parsing error on line: 5, column: 2, while parsing: string. Definition:
	//
	// 	string:nows = "\"" ([^\\"\b\f\n\r\t] | "\\" (["\\/bfnrt] | "u" [0-9a-f]{4}))* "\"";
	// `
	//
	// 	const doc = `{
	// 		"a":1,
	// 		"b":2,
	// 		"c":3,
	// 	}`
	//
	// 	s, err := openSyntaxFile("examples/json.treerack")
	// 	if err != nil {
	// 		t.Error(err)
	// 		return
	// 	}
	//
	// 	_, err = s.Parse(bytes.NewBufferString(doc))
	// 	perr, ok := err.(*ParseError)
	// 	if !ok {
	// 		t.Error("failed to return parse error")
	// 		return
	// 	}
	//
	// 	if perr.Verbose() != expected {
	// 		t.Error("failed to get the right error message")
	// 		t.Log("got:     ", perr.Verbose())
	// 		t.Log("expected:", expected)
	// 	}
}

func TestLongestFail(t *testing.T) {
	const syntax = `
		whitespace:ws           = [ \t];
		number:nows             = [0-9]+;
		symbol:nows             = [a-z]+;
		list-separator          = [,\n];
		function-application    = expression "(" (expression (list-separator+ expression)*)? ")";
		expression              = number | symbol | function-application;
		statement-separator     = [;\n];
		doc:root                = (expression (statement-separator+ expression)*)?
	`

	const doc = `f(a b c)`

	testParseError(t, syntax, []errorTestItem{{
		title: "fail on longest failing parser",
		doc:   doc,
		perr: ParseError{
			Offset:     4,
			Line:       0,
			Column:     4,
			Definition: "function-application",
		},
	}})
}

func TestFailPass(t *testing.T) {
	const syntax = `
		space:ws                = " ";
		symbol:nows             = [a-z]+;
		list-separator:failpass = ",";
		argument-list:failpass  = (symbol (list-separator+ symbol)*);
		function-application    = symbol "(" argument-list? ")";
	`

	const doc = `f(a b c)`

	testParseError(t, syntax, []errorTestItem{{
		title: "fail in outer definition",
		doc:   doc,
		perr: ParseError{
			Offset:     4,
			Line:       0,
			Column:     4,
			Definition: "function-application",
		},
	}})
}
