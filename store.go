package parse

type storedItem struct {
	name string
	node *Node
}

type storeEntry struct {
	match   []*storedItem
	noMatch []string
}

type store struct {
	tokens []*storeEntry
}

func (c *store) get(offset int, name string) (*Node, bool, bool) {
	if len(c.tokens) <= offset {
		return nil, false, false
	}

	tc := c.tokens[offset]
	if tc == nil {
		return nil, false, false
	}

	for _, i := range tc.noMatch {
		if i == name {
			return nil, false, true
		}
	}

	for _, i := range tc.match {
		if i.name == name {
			return i.node, true, true
		}
	}

	return nil, false, false
}

func (c *store) set(offset int, name string, n *Node) {
	var tc *storeEntry
	if len(c.tokens) > offset {
		tc = c.tokens[offset]
	} else {
		if cap(c.tokens) > offset {
			c.tokens = c.tokens[:offset+1]
		} else {
			c.tokens = c.tokens[:cap(c.tokens)]
			for len(c.tokens) <= offset {
				c.tokens = append(c.tokens, nil)
			}
		}

		tc = &storeEntry{}
		c.tokens[offset] = tc
	}

	if n == nil {
		for _, i := range tc.match {
			if i.name == name {
				return
			}
		}

		for _, i := range tc.noMatch {
			if i == name {
				return
			}
		}

		tc.noMatch = append(tc.noMatch, name)
		return
	}

	for _, i := range tc.match {
		if i.name == name {
			if n.tokenLength() > i.node.tokenLength() {
				i.node = n
			}

			return
		}
	}

	tc.match = append(tc.match, &storedItem{
		name: name,
		node: n,
	})
}
