package treerack

import "testing"

const (
	csvWithoutWhitespaceSupport = `
		ws:alias = [ \t]*;
		cell     = [^\n, \t]*;
		line     = ws cell (ws "," ws cell)* ws;
		document = (line ("\n" line)*)?;
	`
)

func TestCSVWhitespace(t *testing.T) {
	t.Run("wihout whitespace support", func(t *testing.T) {
		s, err := openSyntaxString(csvWithoutWhitespaceSupport)
		if err != nil {
			t.Error(err)
			return
		}

		runTestsSyntax(t, s, []testItem{{
			title: "empty",
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
		}})
	})
}
