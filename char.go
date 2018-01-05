package treerack

type charParser struct {
	name   string
	id     int
	not    bool
	chars  []rune
	ranges [][]rune
}

type charBuilder struct {
	name string
	id   int
}

func (p *charParser) nodeName() string       { return p.name }
func (p *charParser) nodeID() int            { return p.id }
func (p *charParser) commitType() CommitType { return Alias }

func matchChar(chars []rune, ranges [][]rune, not bool, char rune) bool {
	for _, ci := range chars {
		if ci == char {
			return !not
		}
	}

	for _, ri := range ranges {
		if char >= ri[0] && char <= ri[1] {
			return !not
		}
	}

	return not
}

func (p *charParser) match(t rune) bool {
	return matchChar(p.chars, p.ranges, p.not, t)
}

func (p *charParser) parse(c *context) {
	if tok, ok := c.token(); !ok || !p.match(tok) {
		if c.offset > c.failOffset {
			c.failOffset = c.offset
			c.failingParser = nil
		}

		c.fail(c.offset)
		return
	}

	c.success(c.offset + 1)
}

func (b *charBuilder) nodeName() string { return b.name }
func (b *charBuilder) nodeID() int      { return b.id }

func (b *charBuilder) build(c *context) ([]*Node, bool) {
	return nil, false
}
