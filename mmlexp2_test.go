package treerack

import (
	"testing"
)

func TestMMLExp2(t *testing.T) {
	s, err := openSyntaxFile("examples/mml-exp2.treerack")
	if err != nil {
		t.Error(err)
		return
	}

	t.Run("indexer", func(t *testing.T) {
		runTestsSyntax(t, s, []testItem{{
			title:          "mixed indexer",
			text:           "a.b[c]",
			ignorePosition: true,
			nodes: []*Node{{
				Name: "indexer",
				Nodes: []*Node{{
					Name: "symbol",
				}, {
					Name: "symbol-index",
					Nodes: []*Node{{
						Name: "symbol",
					}},
				}, {
					Name: "expression-index",
					Nodes: []*Node{{
						Name: "symbol",
					}},
				}},
			}},
		}})

		runTestsSyntax(t, s, []testItem{{
			title:          "mixed indexer inverted",
			text:           "a[b].c",
			ignorePosition: true,
			nodes: []*Node{{
				Name: "indexer",
				Nodes: []*Node{{
					Name: "symbol",
				}, {
					Name: "expression-index",
					Nodes: []*Node{{
						Name: "symbol",
					}},
				}, {
					Name: "symbol-index",
					Nodes: []*Node{{
						Name: "symbol",
					}},
				}},
			}},
		}})
	})
}
