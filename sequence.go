package treerack

type sequenceParser struct {
	name            string
	id              int
	commit          CommitType
	items           []parser
	ranges          [][]int
	generalizations []int
	allChars        bool
}

type sequenceBuilder struct {
	name            string
	id              int
	commit          CommitType
	items           []builder
	ranges          [][]int
	generalizations []int
	allChars        bool
}

func (p *sequenceParser) nodeName() string       { return p.name }
func (p *sequenceParser) nodeID() int            { return p.id }
func (p *sequenceParser) commitType() CommitType { return p.commit }

func (p *sequenceParser) parse(c *context) {
	if !p.allChars {
		if c.results.pending(c.offset, p.id) {
			c.fail(c.offset)
			return
		}

		c.results.markPending(c.offset, p.id)
	}

	var (
		currentCount int
		parsed       bool
	)

	itemIndex := 0
	from := c.offset
	to := c.offset

	for itemIndex < len(p.items) {
		p.items[itemIndex].parse(c)
		if !c.matchLast {
			if currentCount >= p.ranges[itemIndex][0] {
				itemIndex++
				currentCount = 0
				continue
			}

			if c.failingParser == nil &&
				p.commit&userDefined != 0 &&
				p.commit&Whitespace == 0 &&
				p.commit&FailPass == 0 {
				c.failingParser = p
			}

			c.fail(from)
			if !p.allChars {
				c.results.unmarkPending(from, p.id)
			}

			return
		}

		parsed = c.offset > to
		if parsed {
			currentCount++
		}

		to = c.offset

		if !parsed || p.ranges[itemIndex][1] > 0 && currentCount == p.ranges[itemIndex][1] {
			itemIndex++
			currentCount = 0
		}
	}

	for _, g := range p.generalizations {
		if c.results.pending(from, g) {
			c.results.setMatch(from, g, to)
		}
	}

	if to > c.failOffset {
		c.failOffset = -1
		c.failingParser = nil
	}

	c.results.setMatch(from, p.id, to)
	c.success(to)
	if !p.allChars {
		c.results.unmarkPending(from, p.id)
	}
}

func (b *sequenceBuilder) nodeName() string { return b.name }
func (b *sequenceBuilder) nodeID() int      { return b.id }

func (b *sequenceBuilder) build(c *context) ([]*Node, bool) {
	to, ok := c.results.longestMatch(c.offset, b.id)
	if !ok {
		return nil, false
	}

	from := c.offset
	parsed := to > from

	if b.allChars {
		c.offset = to
		if b.commit&Alias != 0 {
			return nil, true
		}

		return []*Node{{
			Name:   b.name,
			From:   from,
			To:     to,
			tokens: c.tokens,
		}}, true
	} else if parsed {
		c.results.dropMatchTo(c.offset, b.id, to)
		for _, g := range b.generalizations {
			c.results.dropMatchTo(c.offset, g, to)
		}
	} else {
		if c.results.pending(c.offset, b.id) {
			return nil, false
		}

		c.results.markPending(c.offset, b.id)

		for _, g := range b.generalizations {
			c.results.markPending(c.offset, g)
		}
	}

	var (
		itemIndex    int
		currentCount int
		nodes        []*Node
	)

	for itemIndex < len(b.items) {
		itemFrom := c.offset
		n, ok := b.items[itemIndex].build(c)
		if !ok {
			itemIndex++
			currentCount = 0
			continue
		}

		if c.offset > itemFrom {
			nodes = append(nodes, n...)
			currentCount++

			if b.ranges[itemIndex][1] > 0 && currentCount == b.ranges[itemIndex][1] {
				itemIndex++
				currentCount = 0
			}

			continue
		}

		if currentCount < b.ranges[itemIndex][0] {
			for i := 0; i < b.ranges[itemIndex][0]-currentCount; i++ {
				nodes = append(nodes, n...)
			}
		}

		itemIndex++
		currentCount = 0
	}

	if !parsed {
		c.results.unmarkPending(from, b.id)
		for _, g := range b.generalizations {
			c.results.unmarkPending(from, g)
		}
	}

	if b.commit&Alias != 0 {
		return nodes, true
	}

	return []*Node{{
		Name:   b.name,
		From:   from,
		To:     to,
		Nodes:  nodes,
		tokens: c.tokens,
	}}, true
}
