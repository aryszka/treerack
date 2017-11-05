package treerack

import "testing"

func TestNodeString(t *testing.T) {
	t.Run("valid node", func(t *testing.T) {
		n := &Node{
			Name:   "A",
			From:   0,
			To:     3,
			tokens: []rune("abc"),
		}

		if n.String() != "A:0:3:abc" {
			t.Error("invalid node string")
		}
	})

	t.Run("empty node", func(t *testing.T) {
		n := &Node{
			Name: "A",
		}

		if n.String() != "A:0:0:" {
			t.Error("invalid node string")
		}
	})
}
