package parse

type choiceDefinition struct {
	name     string
	commit   CommitType
	elements []string
}

type choiceParser struct {
	name       string
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

func (d *choiceDefinition) parser(r *registry, path []string) (parser, error) {
	p, ok := r.parser(d.name)
	if ok {
		return p, nil
	}

	cp := &choiceParser{
		name:   d.name,
		commit: d.commit,
	}

	r.setParser(cp)

	var elements []parser
	path = append(path, d.name)
	for _, e := range d.elements {
		element, ok := r.parser(e)
		if ok {
			elements = append(elements, element)
			element.setIncludedBy(cp, path)
			continue
		}

		elementDefinition, ok := r.definition(e)
		if !ok {
			return nil, parserNotFound(e)
		}

		element, err := elementDefinition.parser(r, path)
		if err != nil {
			return nil, err
		}

		element.setIncludedBy(cp, path)
		elements = append(elements, element)
	}

	cp.elements = elements
	return cp, nil
}

func (d *choiceDefinition) commitType() CommitType {
	return d.commit
}

func (p *choiceParser) nodeName() string { return p.name }

func (p *choiceParser) setIncludedBy(includedBy parser, path []string) {
	if stringsContain(path, p.name) {
		return
	}

	p.includedBy = append(p.includedBy, includedBy)
}

func (p *choiceParser) cacheIncluded(c *context, n *Node) {
	if !c.excluded(n.from, p.name) {
		return
	}

	nc := newNode(p.name, p.commit, n.from, n.to)
	nc.append(n)
	c.cache.set(nc.from, p.name, nc)

	for _, includedBy := range p.includedBy {
		includedBy.cacheIncluded(c, nc)
	}
}

func (p *choiceParser) parse(t Trace, c *context) {
	// t = t.Extend(p.name)
	// t.Out1("parsing choice", c.offset)

	if p.commit&Documentation != 0 {
		// t.Out1("fail, doc")
		c.fail(c.offset)
		return
	}

	if _, ok := c.fromCache(p.name); ok {
		// t.Out1("found in cache, match:", m)
		return
	}

	if c.excluded(c.offset, p.name) {
		// t.Out1("excluded")
		c.fail(c.offset)
		return
	}

	c.exclude(c.offset, p.name)
	defer c.include(c.offset, p.name)

	node := newNode(p.name, p.commit, c.offset, c.offset)
	var match bool

	for {
		elements := p.elements
		var foundMatch bool

		for len(elements) > 0 {
			elements[0].parse(t, c)
			elements = elements[1:]
			c.offset = node.from

			if !c.match || match && c.node.tokenLength() <= node.tokenLength() {
				continue
			}

			match = true
			foundMatch = true
			node = newNode(p.name, p.commit, c.offset, c.offset)
			node.append(c.node)

			c.cache.set(node.from, p.name, node)
			for _, includedBy := range p.includedBy {
				includedBy.cacheIncluded(c, node)
			}
		}

		if !foundMatch {
			break
		}
	}

	if match {
		// t.Out1("choice, success")
		c.success(node)
		return
	}

	// t.Out1("fail")
	c.cache.set(node.from, p.name, nil)
	c.fail(node.from)
}
