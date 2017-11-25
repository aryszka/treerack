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
		title:      "choice, longer option fails",
		syntax:     `a = "12"; b = "1"; c:root = a | b`,
		text:       "13",
		offset:     1,
		column:     1,
		definition: "a",
	}, {
		title:      "choice, shorter option fails",
		syntax:     `a = "2"; b = "12"; c:root = a | b`,
		text:       "123",
		offset:     0,
		column:     0,
		definition: "1",
	}, {
		title:      "choice, both options fail",
		syntax:     `a = "12"; b = "2"; c:root = a | b`,
		text:       "13",
		offset:     1,
		column:     1,
		definition: "a",
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
