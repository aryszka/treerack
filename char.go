package parse

// TODO: rename to token

type charParser struct {
	name       string
	commit     CommitType
	not        bool
	chars      []rune
	ranges     [][]rune
	includedBy []parser
}

func newChar(
	name string,
	ct CommitType,
	not bool,
	chars []rune,
	ranges [][]rune,
) *charParser {
	return &charParser{
		name:   name,
		commit: ct,
		not:    not,
		chars:  chars,
		ranges: ranges,
	}
}

func (p *charParser) nodeName() string { return p.name }

func (p *charParser) parser(r *registry, path []string) (parser, error) {
	if stringsContain(path, p.name) {
		panic(cannotIncludeParsers(p.name))
	}

	if _, ok := r.parser(p.name); ok {
		return p, nil
	}

	r.setParser(p)
	return p, nil
}

func (p *charParser) commitType() CommitType {
	return p.commit
}

func (p *charParser) setIncludedBy(including parser, path []string) {
	if stringsContain(path, p.name) {
		panic(cannotIncludeParsers(p.name))
	}

	p.includedBy = append(p.includedBy, including)
}

func (p *charParser) cacheIncluded(*context, *Node) {
	panic(cannotIncludeParsers(p.name))
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

func (p *charParser) parse(t Trace, c *context) {
	t = t.Extend(p.name)
	t.Out1("parsing char", c.offset)

	if p.commit&Documentation != 0 {
		t.Out1("fail, doc")
		c.fail(c.offset)
		return
	}

	if m, ok := c.fromCache(p.name); ok {
		t.Out1("found in cache, match:", m)
		return
	}

	if tok, ok := c.token(); ok && p.match(tok) {
		t.Out1("success", string(tok))
		n := newNode(p.name, p.commit, c.offset, c.offset+1)
		c.cache.set(c.offset, p.name, n)
		for _, including := range p.includedBy {
			including.cacheIncluded(c, n)
		}

		c.success(n)
		return
	} else {
		t.Out1("fail", string(tok))
		c.cache.set(c.offset, p.name, nil)
		c.fail(c.offset)
		return
	}
}
