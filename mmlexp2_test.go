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
		// BUG:
		t.Skip()

		runTestsSyntax(t, s, []testItem{{
			title: "mixed indexer",
			text:  "a.b[c]",
			ignorePosition: true,
			nodes: []*Node{{
				Name: "expression-indexer",
				Nodes: []*Node{{
					Name: "symbol-indexer",
					Nodes: []*Node{{
						Name: "symbol",
					}, {
						Name: "symbol",
					}},
				}, {
					Name: "symbol",
				}},
			}},
		}})

		runTestsSyntax(t, s, []testItem{{
			title: "mixed indexer inverted",
			text:  "a[b].c",
			ignorePosition: true,
			nodes: []*Node{{
				Name: "symbol-indexer",
				Nodes: []*Node{{
					Name: "expression-indexer",
					Nodes: []*Node{{
						Name: "symbol",
					}, {
						Name: "symbol",
					}},
				}, {
					Name: "symbol",
				}},
			}},
		}})
	})
}
