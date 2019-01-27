package treerack

import (
	"testing"
)

func TestMMLExp3(t *testing.T) {
	s, err := openSyntaxFile("examples/mml-exp3.treerack")
	if err != nil {
		t.Error(err)
		return
	}

	t.Run("indexer", func(t *testing.T) {
		runTestsSyntax(t, s, []testItem{{
			title:          "assignment",
			text:           "fn f() a.b = c",
			ignorePosition: true,
			nodes: []*Node{{
				Name: "function-definition",
				Nodes: []*Node{{
					Name: "function-capture",
					Nodes: []*Node{{
						Name: "symbol",
					}, {
						Name: "assignment",
						Nodes: []*Node{{
							Name: "indexer",
							Nodes: []*Node{{
								Name: "symbol",
							}, {
								Name: "symbol-index",
								Nodes: []*Node{{
									Name: "symbol",
								}},
							}},
						}, {
							Name: "symbol",
						}},
					}},
				}},
			}},
		}})
	})
}
