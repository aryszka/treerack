package treerack

import "testing"

func TestKeyword(t *testing.T) {
	const syntax = `
		keywords:kw = "foo" | "bar";
		symbol:nokw = [a-z]+;
	`

	runTests(t, syntax, []testItem{{
		title: "keyword",
		text:  "foo",
		fail:  true,
	}, {
		title:          "not keyword",
		text:           "baz",
		ignorePosition: true,
		node: &Node{
			Name: "symbol",
		},
	}})
}
