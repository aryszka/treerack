package treerack

import "testing"

const (
	csvWithoutWhitespaceSupport = `
		ws:alias        = [ \t];
		word-char:alias = [^\n, \t];
		cell            = (word-char (ws* word-char)*)?;
		rest-cell:alias = "," ws* cell;
		line            = cell (ws* rest-cell (ws* rest-cell)*)?;
		rest-line:alias = "\n" ws* line;
		document        = ws* (line (ws* rest-line (ws* rest-line)*)?)? ws*;
	`

	csvWithWhitespaceSupport = `
		ws:ws    = [ \t];
		cell     = [^\n, \t]*;
		line     = cell ("," cell)*;
		document = (line ("\n" line)*)?;
	`
)

func TestCSVWhitespace(t *testing.T) {
	tests := []testItem{{
		title:          "empty",
		ignorePosition: true,
		node: &Node{
			Name: "document",
		},
	}, {
		title:          "only a cell",
		text:           "abc",
		ignorePosition: true,
		node: &Node{
			Name: "document",
			Nodes: []*Node{{
				Name: "line",
				Nodes: []*Node{{
					Name: "cell",
				}},
			}},
		},
	}, {
		title:          "single line",
		text:           `a, b, c`,
		ignorePosition: true,
		node: &Node{
			Name: "document",
			Nodes: []*Node{{
				Name: "line",
				Nodes: []*Node{{
					Name: "cell",
				}, {
					Name: "cell",
				}, {
					Name: "cell",
				}},
			}},
		},
	}, {
		title: "regular csv",
		text: `a, b, c
			       d, e, f`,
		ignorePosition: true,
		node: &Node{
			Name: "document",
			Nodes: []*Node{{
				Name: "line",
				Nodes: []*Node{{
					Name: "cell",
				}, {
					Name: "cell",
				}, {
					Name: "cell",
				}},
			}, {
				Name: "line",
				Nodes: []*Node{{
					Name: "cell",
				}, {
					Name: "cell",
				}, {
					Name: "cell",
				}},
			}},
		},
	}, {
		title: "irregular csv",
		text: `a,, b, c, 
			       d, ,,,,`,
		ignorePosition: true,
		node: &Node{
			Name: "document",
			Nodes: []*Node{{
				Name: "line",
				Nodes: []*Node{{
					Name: "cell",
				}, {
					Name: "cell",
				}, {
					Name: "cell",
				}, {
					Name: "cell",
				}, {
					Name: "cell",
				}},
			}, {
				Name: "line",
				Nodes: []*Node{{
					Name: "cell",
				}, {
					Name: "cell",
				}, {
					Name: "cell",
				}, {
					Name: "cell",
				}, {
					Name: "cell",
				}, {
					Name: "cell",
				}},
			}},
		},
	}, {
		title:          "too many commas",
		text:           `a,,`,
		ignorePosition: true,
		node: &Node{
			Name: "document",
			Nodes: []*Node{{
				Name: "line",
				Nodes: []*Node{{
					Name: "cell",
				}, {
					Name: "cell",
				}, {
					Name: "cell",
				}},
			}},
		},
	}, {
		title:          "csv with tabs",
		text:           "a,\tb, c",
		ignorePosition: true,
		node: &Node{
			Name: "document",
			Nodes: []*Node{{
				Name: "line",
				Nodes: []*Node{{
					Name: "cell",
				}, {
					Name: "cell",
				}, {
					Name: "cell",
				}},
			}},
		},
	}, {
		title: "whitespace between lines",
		text:  " a, b, c \n d, e, f ",
		node: &Node{
			Name: "document",
			To:   19,
			Nodes: []*Node{{
				Name: "line",
				From: 1,
				To:   8,
				Nodes: []*Node{{
					Name: "cell",
					From: 1,
					To:   2,
				}, {
					Name: "cell",
					From: 4,
					To:   5,
				}, {
					Name: "cell",
					From: 7,
					To:   8,
				}},
			}, {
				Name: "line",
				From: 11,
				To:   18,
				Nodes: []*Node{{
					Name: "cell",
					From: 11,
					To:   12,
				}, {
					Name: "cell",
					From: 14,
					To:   15,
				}, {
					Name: "cell",
					From: 17,
					To:   18,
				}},
			}},
		},
	}, {
		title:          "just a space",
		text:           " ",
		ignorePosition: true,
		node: &Node{
			Name: "document",
		},
	}, {
		title: "cell with spaces in it",
		text:  "cell content 1/1, cell content 1/2\ncell content 2/1, cell content 2/2",
		node: &Node{
			Name: "document",
			To:   69,
			Nodes: []*Node{{
				Name: "line",
				To:   34,
				Nodes: []*Node{{
					Name: "cell",
					To:   16,
				}, {
					Name: "cell",
					From: 18,
					To:   34,
				}},
			}, {
				Name: "line",
				From: 35,
				To:   69,
				Nodes: []*Node{{
					Name: "cell",
					From: 35,
					To:   51,
				}, {
					Name: "cell",
					From: 53,
					To:   69,
				}},
			}},
		},
	}, {
		title:          "multiple empty lines",
		text:           "\n\n",
		ignorePosition: true,
		node: &Node{
			Name: "document",
			Nodes: []*Node{{
				Name: "line",
				Nodes: []*Node{{
					Name: "cell",
				}},
			}, {
				Name: "line",
				Nodes: []*Node{{
					Name: "cell",
				}},
			}, {
				Name: "line",
				Nodes: []*Node{{
					Name: "cell",
				}},
			}},
		},
	}}

	t.Run("without whitespace support", func(t *testing.T) {
		s, err := openSyntaxString(csvWithoutWhitespaceSupport)
		if err != nil {
			t.Error(err)
			return
		}

		runTestsSyntax(t, s, tests)
	})

	t.Run("with whitespace support", func(t *testing.T) {
		s, err := openSyntaxString(csvWithWhitespaceSupport)
		if err != nil {
			t.Error(err)
			return
		}

		runTestsSyntax(t, s, tests)
	})
}

func TestWhitespace(t *testing.T) {
	t.Run("nows flag", func(t *testing.T) {
		runTests(
			t,
			`
				space:ws = " ";
				symbol:nows = [a-zA-Z_] [a-zA-Z0-9_]* | "[" .+ "]";
				symbols = symbol*;
			`,
			[]testItem{{
				title:          "multiple symbols",
				text:           "a b c",
				ignorePosition: true,
				node: &Node{
					Name: "symbols",
					Nodes: []*Node{{
						Name: "symbol",
					}, {
						Name: "symbol",
					}, {
						Name: "symbol",
					}},
				},
			}},
		)
	})

	t.Run("whitespace with max items", func(t *testing.T) {
		runTests(
			t,
			`space:ws = " "; a = "a"{3,5}`,
			[]testItem{{
				title: "less than min",
				text:  "a a",
				fail:  true,
			}, {
				title:          "just min",
				text:           "a a a",
				ignorePosition: true,
				node: &Node{
					Name: "a",
				},
			}, {
				title:          "less than max",
				text:           "a a a a",
				ignorePosition: true,
				node: &Node{
					Name: "a",
				},
			}, {
				title:          "just max",
				text:           "a a a a a",
				ignorePosition: true,
				node: &Node{
					Name: "a",
				},
			}, {
				title: "more than max",
				text:  "a a a a a a",
				fail:  true,
			}},
		)
	})
}
