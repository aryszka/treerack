package treerack

type choiceParser struct {
	name            string
	id              int
	commit          CommitType
	options         []parser
	generalizations []int
}

type choiceBuilder struct {
	name            string
	id              int
	commit          CommitType
	options         []builder
	generalizations []int
}

func (p *choiceParser) nodeName() string       { return p.name }
func (p *choiceParser) nodeID() int            { return p.id }
func (p *choiceParser) commitType() CommitType { return p.commit }

func (p *choiceParser) parse(c *context) {
	if c.fromResults(p) {
		return
	}

	if c.results.pending(c.offset, p.id) {
		c.fail(c.offset)
		return
	}

	c.results.markPending(c.offset, p.id)

	var (
		match         bool
		optionIndex   int
		foundMatch    bool
		failingParser parser
	)

	from := c.offset
	to := c.offset
	initialFailOffset := c.failOffset
	initialFailingParser := c.failingParser
	failOffset := initialFailOffset

	for {
		foundMatch = false
		optionIndex = 0

		for optionIndex < len(p.options) {
			p.options[optionIndex].parse(c)
			optionIndex++

			if !c.matchLast {
				if c.failOffset > failOffset {
					failOffset = c.failOffset
					failingParser = c.failingParser
				}
			}

			if !c.matchLast || match && c.offset <= to {
				c.offset = from
				continue
			}

			match = true
			foundMatch = true
			to = c.offset
			c.offset = from
			c.results.setMatch(from, p.id, to)
		}

		if !foundMatch {
			break
		}
	}

	if match {
		if failOffset > to {
			c.failOffset = failOffset
			c.failingParser = failingParser
		} else if to > initialFailOffset {
			c.failOffset = -1
			c.failingParser = nil
		} else {
			c.failOffset = initialFailOffset
			c.failingParser = initialFailingParser
		}

		c.success(to)
		c.results.unmarkPending(from, p.id)
		return
	}

	if failOffset > initialFailOffset {
		c.failOffset = failOffset
		c.failingParser = failingParser
		if c.failingParser == nil &&
			p.commitType()&userDefined != 0 &&
			p.commitType()&Whitespace == 0 &&
			p.commitType()&FailPass == 0 {
			c.failingParser = p
		}
	}

	c.results.setNoMatch(from, p.id)
	c.fail(from)
	c.results.unmarkPending(from, p.id)
}

func (b *choiceBuilder) nodeName() string { return b.name }
func (b *choiceBuilder) nodeID() int      { return b.id }

func (b *choiceBuilder) build(c *context) ([]*Node, bool) {
	to, ok := c.results.longestMatch(c.offset, b.id)
	if !ok {
		return nil, false
	}

	from := c.offset
	parsed := to > from

	if parsed {
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

	var option builder
	for _, o := range b.options {
		if c.results.hasMatchTo(c.offset, o.nodeID(), to) {
			option = o
			break
		}
	}

	n, _ := option.build(c)
	if !parsed {
		c.results.unmarkPending(from, b.id)
		for _, g := range b.generalizations {
			c.results.unmarkPending(from, g)
		}
	}

	if b.commit&Alias != 0 {
		return n, true
	}

	return []*Node{{
		Name:   b.name,
		From:   from,
		To:     to,
		Nodes:  n,
		tokens: c.tokens,
	}}, true
}
