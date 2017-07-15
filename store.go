package treerack

type storeEntry struct {
	match   *idSet
	noMatch *idSet
	nodes   []*Node
}

type store struct {
	entries []*storeEntry
}

func (c *store) get(offset int, id int) (*Node, bool, bool) {
	if len(c.entries) <= offset {
		return nil, false, false
	}

	tc := c.entries[offset]
	if tc == nil {
		return nil, false, false
	}

	if tc.noMatch.has(id) {
		return nil, false, true
	}

	if !tc.match.has(id) {
		return nil, false, false
	}

	for _, n := range tc.nodes {
		if n.id == id {
			return n, true, true
		}
	}

	return nil, false, false
}

func (c *store) set(offset int, id int, n *Node) {
	var tc *storeEntry
	if len(c.entries) > offset {
		tc = c.entries[offset]
	} else {
		if cap(c.entries) > offset {
			c.entries = c.entries[:offset+1]
		} else {
			c.entries = c.entries[:cap(c.entries)]
			for len(c.entries) <= offset {
				c.entries = append(c.entries, nil)
			}
		}

		tc = &storeEntry{
			match:   &idSet{},
			noMatch: &idSet{},
		}

		c.entries[offset] = tc
	}

	if n == nil {
		if tc.match.has(id) {
			return
		}

		tc.noMatch.set(id)
		return
	}

	tc.match.set(id)
	for i, ni := range tc.nodes {
		if ni.id == id {
			if n.tokenLength() > ni.tokenLength() {
				tc.nodes[i] = n
			}

			return
		}
	}

	tc.nodes = append(tc.nodes, n)
}
