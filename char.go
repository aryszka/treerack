package treerack

type charParser struct {
	name       string
	id         int
	not        bool
	chars      []rune
	ranges     [][]rune
	includedBy []int
}

func newChar(
	name string,
	not bool,
	chars []rune,
	ranges [][]rune,
) *charParser {
	return &charParser{
		name:   name,
		not:    not,
		chars:  chars,
		ranges: ranges,
	}
}

func (p *charParser) nodeName() string                  { return p.name }
func (p *charParser) setNodeName(n string)              { p.name = n }
func (p *charParser) nodeID() int                       { return p.id }
func (p *charParser) setID(id int)                      { p.id = id }
func (p *charParser) commitType() CommitType            { return Alias }
func (p *charParser) setCommitType(ct CommitType)       {}
func (p *charParser) validate(*registry, *idSet) error  { return nil }
func (p *charParser) normalize(*registry, *idSet) error { return nil }

func (p *charParser) init(r *registry) error { return nil }

func (p *charParser) setIncludedBy(r *registry, includedBy int, parsers *idSet) error {
	p.includedBy = appendIfMissing(p.includedBy, includedBy)
	return nil
}

func (p *charParser) parser(r *registry, parsers *idSet) (parser, error) {
	if parsers.has(p.id) {
		panic(cannotIncludeParsers(p.name))
	}

	if _, ok := r.parser(p.name); ok {
		return p, nil
	}

	r.setParser(p)
	return p, nil
}

func (p *charParser) builder() builder {
	return p
}

func matchChars(chars []rune, ranges [][]rune, not bool, char rune) bool {
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
	return matchChars(p.chars, p.ranges, p.not, t)
}

func (p *charParser) parse(c *context) {
	if tok, ok := c.token(); !ok || !p.match(tok) {
		c.fail(c.offset)
		return
	}

	c.success(c.offset + 1)
	for _, includedBy := range p.includedBy {
		c.store.setMatch(c.offset, includedBy, c.offset+1)
	}
}

func (p *charParser) build(c *context) ([]*Node, bool) {
	panic("called char build")
}
