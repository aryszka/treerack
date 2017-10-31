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

func (p *charParser) match(t rune) bool {
	for _, ci := range p.chars {
		if ci == t {
			return !p.not
		}
	}

	for _, ri := range p.ranges {
		if t >= ri[0] && t <= ri[1] {
			return !p.not
		}
	}

	return p.not
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
	// // TODO: how to remove this check
	// t, ok := c.token()
	// if !ok {
	// 	panic("damaged parser context")
	// }

	// if !p.match(t) {
	// 	return nil, false
	// }

	// // always alias
	// c.offset++
	// return nil, true
}
