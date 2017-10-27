package treerack

import "testing"

func TestSExpr(t *testing.T) {
	runTestsFile(t, "sexpr.parser", []testItem{{
		title: "number",
		text:  "42",
		nodes: []*Node{{
			Name: "number",
		}},
		ignorePosition: true,
	}, {
		title: "string",
		text:  "\"foo\"",
		nodes: []*Node{{
			Name: "string",
		}},
		ignorePosition: true,
	}, {
		title: "symbol",
		text:  "foo",
		nodes: []*Node{{
			Name: "symbol",
		}},
		ignorePosition: true,
	}, {
		title: "nil",
		text:  "()",
		nodes: []*Node{{
			Name: "list",
		}},
		ignorePosition: true,
	}, {
		title: "list",
		text:  "(foo bar baz)",
		nodes: []*Node{{
			Name: "list",
			Nodes: []*Node{{
				Name: "symbol",
			}, {
				Name: "symbol",
			}, {
				Name: "symbol",
			}},
		}},
		ignorePosition: true,
	}, {
		title: "embedded list",
		text:  "(foo (bar (baz)) qux)",
		nodes: []*Node{{
			Name: "list",
			Nodes: []*Node{{
				Name: "symbol",
			}, {
				Name: "list",
				Nodes: []*Node{{
					Name: "symbol",
				}, {
					Name: "list",
					Nodes: []*Node{{
						Name: "symbol",
					}},
				}},
			}, {
				Name: "symbol",
			}},
		}},
		ignorePosition: true,
	}})
}
