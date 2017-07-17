package treerack

type choiceDefinition struct {
	name     string
	id       int
	commit   CommitType
	elements []string
}

type choiceParser struct {
	name       string
	id         int
	commit     CommitType
	elements   []parser
	includedBy []parser
}

func newChoice(name string, ct CommitType, elements []string) *choiceDefinition {
	return &choiceDefinition{
		name:     name,
		commit:   ct,
		elements: elements,
	}
}

func (d *choiceDefinition) nodeName() string { return d.name }
func (d *choiceDefinition) nodeID() int      { return d.id }
func (d *choiceDefinition) setID(id int)     { d.id = id }

func (d *choiceDefinition) parser(r *registry, parsers *idSet) (parser, error) {
	p, ok := r.parser(d.name)
	if ok {
		return p, nil
	}

	cp := &choiceParser{
		name:   d.name,
		id:     d.id,
		commit: d.commit,
	}

	r.setParser(cp)

	var elements []parser
	parsers.set(d.id)
	defer parsers.unset(d.id)
	for _, e := range d.elements {
		element, ok := r.parser(e)
		if ok {
			elements = append(elements, element)
			element.setIncludedBy(cp, parsers)
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

		element.setIncludedBy(cp, parsers)
		elements = append(elements, element)
	}

	cp.elements = elements
	return cp, nil
}

func (d *choiceDefinition) commitType() CommitType {
	return d.commit
}

func (p *choiceParser) nodeName() string { return p.name }
func (p *choiceParser) nodeID() int      { return p.id }

func (p *choiceParser) setIncludedBy(includedBy parser, parsers *idSet) {
	if parsers.has(p.id) {
		return
	}

	p.includedBy = append(p.includedBy, includedBy)
}

func (p *choiceParser) storeIncluded(c *context, from, to int) {
	if !c.excluded(from, p.id) {
		return
	}

	c.store.set(from, p.id, true, to)

	for _, includedBy := range p.includedBy {
		includedBy.storeIncluded(c, from, to)
	}
}

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

	for {
		elementIndex := 0
		var foundMatch bool

		for elementIndex < len(p.elements) {
			p.elements[elementIndex].parse(t, c)
			elementIndex++
			nextTo := c.offset
			c.offset = from

			if !c.match || match && nextTo <= to {
				continue
			}

			match = true
			foundMatch = true
			to = nextTo

			c.store.set(from, p.id, true, to)
			for _, includedBy := range p.includedBy {
				includedBy.storeIncluded(c, from, to)
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

	c.store.set(from, p.id, false, 0)
	c.fail(from)
	c.include(from, p.id)
}
