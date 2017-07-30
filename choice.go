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
	name       string
	id         int
	commit     CommitType
	elements   []builder
	includedBy *idSet
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
			name:       d.name,
			id:         d.id,
			commit:     d.commit,
			includedBy: &idSet{},
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

	if d.cbuilder == nil {
		d.cbuilder = &choiceBuilder{
			name:       d.name,
			id:         d.id,
			commit:     d.commit,
			includedBy: &idSet{},
		}
	}

	d.cbuilder.includedBy.set(includedBy)

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
			name:       d.name,
			id:         d.id,
			commit:     d.commit,
			includedBy: &idSet{},
		}
	}

	return d.cbuilder
}

func (p *choiceParser) nodeName() string { return p.name }
func (p *choiceParser) nodeID() int      { return p.id }

func (p *choiceParser) parse(t Trace, c *context) {
	// t = t.Extend(p.name)
	// t.Out1("parsing choice", c.offset)

	// TODO: don't add documentation
	// if p.commit&Documentation != 0 {
	// 	// t.Out1("fail, doc")
	// 	c.fail(c.offset)
	// 	return
	// }

	if c.fromStore(p.id) {
		// t.Out1("found in store, match:")
		return
	}

	if c.excluded(c.offset, p.id) {
		// t.Out1("fail, excluded")
		c.fail(c.offset)
		return
	}

	c.exclude(c.offset, p.id)
	from := c.offset
	to := c.offset

	var match bool
	var elementIndex int
	var foundMatch bool

	var excludedIncluded []int

	for {
		foundMatch = false
		elementIndex = 0

		for elementIndex < len(p.elements) {
			p.elements[elementIndex].parse(t, c)
			elementIndex++

			if !c.match || match && c.offset <= to {
				c.offset = from
				continue
			}

			match = true
			foundMatch = true
			to = c.offset
			c.offset = from

			c.store.setMatch(from, p.id, to)
			if match {
				for _, includedBy := range excludedIncluded {
					c.store.setMatch(from, includedBy, to)
				}
			} else {
				for _, includedBy := range p.includedBy {
					if c.excluded(from, includedBy) {
						excludedIncluded = append(excludedIncluded, includedBy)
						c.store.setMatch(from, includedBy, to)
					}
				}
			}
		}

		if !foundMatch {
			break
		}
	}

	if match {
		c.success(to)
		c.include(from, p.id)
		// t.Out1("choice, success")
		return
	}

	// t.Out1("fail")
	c.store.setNoMatch(from, p.id)
	c.fail(from)
	c.include(from, p.id)
}

func (b *choiceBuilder) nodeName() string { return b.name }
func (b *choiceBuilder) nodeID() int      { return b.id }

func (b *choiceBuilder) build(c *context) ([]*Node, bool) {
	to, ok := c.store.takeMatch(c.offset, b.id, b.includedBy)
	if !ok {
		return nil, false
	}

	var element builder
	for _, e := range b.elements {
		if c.store.hasMatchTo(c.offset, e.nodeID(), to) {
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
