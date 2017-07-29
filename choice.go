package treerack

type choiceDefinition struct {
	name       string
	id         int
	commit     CommitType
	elements   []string
	includedBy []int
	cbuilder   *choiceBuilder
}

type choiceParser struct {
	name       string
	id         int
	commit     CommitType
	elements   []parser
	includedBy []int
}

type choiceBuilder struct {
	name     string
	id       int
	commit   CommitType
	elements []builder
}

func newChoice(name string, ct CommitType, elements []string) *choiceDefinition {
	return &choiceDefinition{
		name:     name,
		commit:   ct,
		elements: elements,
	}
}

func (d *choiceDefinition) nodeName() string       { return d.name }
func (d *choiceDefinition) nodeID() int            { return d.id }
func (d *choiceDefinition) setID(id int)           { d.id = id }
func (d *choiceDefinition) commitType() CommitType { return d.commit }

func (d *choiceDefinition) init(r *registry) error {
	if d.cbuilder == nil {
		d.cbuilder = &choiceBuilder{
			name:   d.name,
			id:     d.id,
			commit: d.commit,
		}
	}

	for _, e := range d.elements {
		def, ok := r.definition(e)
		if !ok {
			return parserNotFound(e)
		}

		d.cbuilder.elements = append(d.cbuilder.elements, def.builder())
	}

	parsers := &idSet{}
	parsers.set(d.id)
	return setItemsIncludedBy(r, d.elements, d.id, parsers)
}

func (d *choiceDefinition) setIncludedBy(r *registry, includedBy int, parsers *idSet) error {
	if parsers.has(d.id) {
		return nil
	}

	d.includedBy = appendIfMissing(d.includedBy, includedBy)
	parsers.set(d.id)
	return setItemsIncludedBy(r, d.elements, includedBy, parsers)
}

// TODO:
// - it may be possible to initialize the parsers non-recursively
// - maybe the whole definition, parser and builder can be united

func (d *choiceDefinition) parser(r *registry, parsers *idSet) (parser, error) {
	p, ok := r.parser(d.name)
	if ok {
		return p, nil
	}

	cp := &choiceParser{
		name:       d.name,
		id:         d.id,
		commit:     d.commit,
		includedBy: d.includedBy,
	}

	r.setParser(cp)

	var elements []parser
	parsers.set(d.id)
	defer parsers.unset(d.id)
	for _, e := range d.elements {
		element, ok := r.parser(e)
		if ok {
			elements = append(elements, element)
			continue
		}

		elementDefinition, ok := r.definition(e)
		if !ok {
			return nil, parserNotFound(e)
		}

		element, err := elementDefinition.parser(r, parsers)
		if err != nil {
			return nil, err
		}

		elements = append(elements, element)
	}

	cp.elements = elements
	return cp, nil
}

func (d *choiceDefinition) builder() builder {
	if d.cbuilder == nil {
		d.cbuilder = &choiceBuilder{
			name:   d.name,
			id:     d.id,
			commit: d.commit,
		}
	}

	return d.cbuilder
}

func (p *choiceParser) nodeName() string { return p.name }
func (p *choiceParser) nodeID() int      { return p.id }

func (p *choiceParser) parse(t Trace, c *context) {
	if p.commit&Documentation != 0 {
		c.fail(c.offset)
		return
	}

	if _, ok := c.fromStore(p.id); ok {
		return
	}

	if c.excluded(c.offset, p.id) {
		c.fail(c.offset)
		return
	}

	c.exclude(c.offset, p.id)
	from := c.offset
	to := c.offset

	var match bool
	var nextTo int
	var elementIndex int

	for {
		var foundMatch bool
		elementIndex = 0

		for elementIndex < len(p.elements) {
			p.elements[elementIndex].parse(t, c)
			elementIndex++
			nextTo = c.offset
			c.offset = from

			if !c.match || match && nextTo <= to {
				continue
			}

			match = true
			foundMatch = true
			to = nextTo

			c.store.setMatch(from, p.id, to)
			for _, includedBy := range p.includedBy {
				c.store.setMatch(from, includedBy, to)
			}
		}

		if !foundMatch {
			break
		}
	}

	if match {
		c.success(to)
		c.include(from, p.id)
		return
	}

	c.store.setNoMatch(from, p.id)
	c.fail(from)
	c.include(from, p.id)
}

func (b *choiceBuilder) nodeName() string { return b.name }
func (b *choiceBuilder) nodeID() int      { return b.id }

func (b *choiceBuilder) build(c *context) ([]*Node, bool) {
	to, ok := c.store.takeMatch(c.offset, b.id)
	if !ok {
		return nil, false
	}

	var element builder
	for _, e := range b.elements {
		elementTo, match, _ := c.store.getMatch(c.offset, e.nodeID())
		if match && elementTo == to {
			element = e
			break
		}
	}

	if element == nil {
		panic("damaged parse result")
	}

	from := c.offset

	n, ok := element.build(c)
	if !ok {
		panic("damaged parse result")
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
