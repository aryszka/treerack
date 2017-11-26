package treerack

import (
	"bytes"
	"testing"
)

func TestError(t *testing.T) {
	type testItem struct {
		title      string
		syntax     string
		text       string
		offset     int
		line       int
		column     int
		definition string
	}

	for _, test := range []testItem{{
		title:      "single def, empty text",
		syntax:     `a = "a"`,
		definition: "a",
	}, {
		title:      "single def, wrong text",
		syntax:     `a = "a"`,
		text:       "b",
		definition: "a",
	}, {
		title:      "single optional def, wrong text",
		syntax:     `a = "a"?`,
		text:       "b",
		definition: "a",
	}, {
		title:      "error on second line, second column",
		syntax:     `a = [a\n]*`,
		text:       "aa\nabaa\naa",
		offset:     4,
		line:       1,
		column:     1,
		definition: "a",
	}, {
		title:      "multiple definitions",
		syntax:     `a = "aa"; A:root = a`,
		text:       "ab",
		offset:     1,
		column:     1,
		definition: "a",
	}, {
		title:      "choice, options succeed",
		syntax:     `a = "12"; b = "1"; c:root = a | b`,
		text:       "123",
		offset:     2,
		column:     2,
		definition: "c",
	}, {
		title:      "choice succeeds, document fails",
		syntax:     `a = "12"; b = "1"; c:root = a | b`,
		text:       "13",
		offset:     1,
		column:     1,
		definition: "c",
	}, {
		title:      "choice fails",
		syntax:     `a = "12"; b = "2"; c:root = a | b`,
		text:       "13",
		offset:     1,
		column:     1,
		definition: "a",
	}, {
		title:      "choice fails, longer option reported",
		syntax:     `a = "12"; b = "134"; c:root = a | b`,
		text:       "135",
		offset:     2,
		column:     2,
		definition: "b",
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

			if perr.Offset != test.offset {
				t.Error("invalid error offset", perr.Offset, test.offset)
				return
			}

			if perr.Line != test.line {
				t.Error("invalid line index", perr.Line, test.line)
				return
			}

			if perr.Column != test.column {
				t.Error("invalid column index", perr.Column, test.column)
			}

			if perr.Definition != test.definition {
				t.Error("invalid definition", perr.Definition, test.definition)
			}
		})
	}
}

func TestErrorMessage(t *testing.T) {
	const expected = "foo:4:10:failed to parse definition: bar"

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
	const expected = `
`

	const doc = `{
		"a": 1,
		"b": 2,
		"c": 3,
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

	t.Log(perr.Error())
}
