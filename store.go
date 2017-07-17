package treerack

type storeEntry struct {
	match   *idSet
	noMatch *idSet
	nodes   []int
}

type store struct {
	entries []*storeEntry
}

func (c *store) get(offset int, id int) (int, bool, bool) {
	if len(c.entries) <= offset {
		return 0, false, false
	}

	tc := c.entries[offset]
	if tc == nil {
		return 0, false, false
	}

	if tc.noMatch.has(id) {
		return 0, false, true
	}

	if !tc.match.has(id) {
		return 0, false, false
	}

	for i := 0; i < len(tc.nodes); i += 2 {
		if tc.nodes[i] == id {
			return tc.nodes[i + 1], true, true
		}
	}

	return 0, false, false
}

func (c *store) set(offset int, id int, match bool, to int) {
	if len(c.entries) <= offset {
		if cap(c.entries) > offset {
			c.entries = c.entries[:offset+1]
		} else {
			c.entries = c.entries[:cap(c.entries)]
			for len(c.entries) <= offset {
				c.entries = append(c.entries, nil)
			}
		}
	}

	tc := c.entries[offset]
	if tc == nil {
		tc = &storeEntry{
			match:   &idSet{},
			noMatch: &idSet{},
		}

		c.entries[offset] = tc
	}

	if !match {
		if tc.match.has(id) {
			return
		}

		tc.noMatch.set(id)
		return
	}

	tc.match.set(id)
	for i := 0; i < len(tc.nodes); i += 2 {
		if tc.nodes[i] == id {
			if to > tc.nodes[i + 1] {
				tc.nodes[i + 1] = to
			}

			return
		}
	}

	tc.nodes = append(tc.nodes, id, to)
}

/*

[][][]int

id, length, where to start in the underlying layer, which list in the layer

attibutes:
- sequence: length, the items in the layer below
- choice: the item below

features:
- there can be sequences or choices under choices

in every position:
- store the valid choices with the underlying parsed nodes

3D table: layer, choice, sequence
stored choice identified by: offset, layer, choice index

*/

func (c *store) inc() {
}

func (c *store) dec() {
}

func (c *store) get2(offset, id int) (int, bool) {
	return 0, false
}

func (c *store) cache(offset, id int, match bool, length int) {
}

func (c *store) set2(offset, id int, match bool, length int) {
	/*
		c.cache(offset, id, match, length)
		levels := c.offsetLevels[offset]
		levels[c.currentLevel] = id
	*/
}

/*
package treerack

type storeEntry struct {
	match   *idSet
	noMatch *idSet
	nodes   []*Node
	offset int
}

type store struct {
	current *storeEntry
	currentIndex int
	entries []*storeEntry
}

func (s *store) find(offset int) *storeEntry {
	if s.current == nil {
		return nil
	}

	var seekPrev, seekNext bool
	for {
		switch {
		case s.current.offset == offset:
			return s.current
		case s.current.offset < offset:
			if seekPrev {
				return nil
			}

			seekNext = true
			s.currentIndex++
			if s.currentIndex == len(s.entries) {
				s.currentIndex = len(s.entries) - 1
				return nil
			}

			s.current = s.entries[s.currentIndex]
		case s.current.offset > offset:
			if seekNext {
				return nil
			}

			seekPrev = true
			s.currentIndex--
			if s.currentIndex == -1 {
				s.currentIndex = 0
				return nil
			}

			s.current = s.entries[s.currentIndex]
		}
	}
}

func (s *store) findCreate(offset int) *storeEntry {
	entry := s.find(offset)
	if entry != nil {
		return entry
	}

	entry = &storeEntry{
		match:   &idSet{},
		noMatch: &idSet{},
		offset: offset,
	}

	switch {
	case s.current != nil && s.current.offset > offset:
		s.entries = append(
			s.entries[:s.currentIndex],
			append([]*storeEntry{entry}, s.entries[s.currentIndex:]...)...,
		)
		s.current = entry
	case s.current != nil && s.current.offset < offset:
		s.entries = append(
			s.entries[:s.currentIndex + 1],
			append([]*storeEntry{entry}, s.entries[s.currentIndex + 1:]...)...,
		)
		s.current = entry
		s.currentIndex++
	default:
		s.entries = []*storeEntry{entry}
		s.current = entry
		s.currentIndex = 0
	}

	return entry
}

func (s *store) get(offset int, id int) (*Node, bool, bool) {
	entry := s.find(offset)
	if entry == nil {
		return nil, false, false
	}

	if entry == nil {
		return nil, false, false
	}

	if entry.noMatch.has(id) {
		return nil, false, true
	}

	if !entry.match.has(id) {
		return nil, false, false
	}

	for _, n := range entry.nodes {
		if n.id == id {
			return n, true, true
		}
	}

	return nil, false, false
}

func (s *store) set(offset int, id int, n *Node) {
	entry := s.findCreate(offset)

	if n == nil {
		if entry.match.has(id) {
			return
		}

		entry.noMatch.set(id)
		return
	}

	entry.match.set(id)
	for i, ni := range entry.nodes {
		if ni.id == id {
			if n.tokenLength() > ni.tokenLength() {
				entry.nodes[i] = n
			}

			return
		}
	}

	entry.nodes = append(entry.nodes, n)
}
*/
