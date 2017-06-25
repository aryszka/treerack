package parse

type sequenceDefinition struct {
	name   string
	commit CommitType
	items  []SequenceItem
}

type sequenceParser struct {
	name       string
	commit     CommitType
	items      []parser
	ranges     [][]int
	includedBy []parser
}

func newSequence(name string, ct CommitType, items []SequenceItem) *sequenceDefinition {
	return &sequenceDefinition{
		name:   name,
		commit: ct,
		items:  items,
	}
}

func (d *sequenceDefinition) nodeName() string { return d.name }

func (d *sequenceDefinition) parser(r *registry, path []string) (parser, error) {
	if stringsContain(path, d.name) {
		panic(cannotIncludeParsers(d.name))
	}

	p, ok := r.parser(d.name)
	if ok {
		return p, nil
	}

	sp := &sequenceParser{
		name:   d.name,
		commit: d.commit,
	}

	r.setParser(sp)

	var (
		items  []parser
		ranges [][]int
	)

	path = append(path, d.name)
	for _, item := range d.items {
		if item.Min == 0 && item.Max == 0 {
			item.Min, item.Max = 1, 1
		} else if item.Max == 0 {
			item.Max = -1
		}

		pi, ok := r.parser(item.Name)
		if ok {
			items = append(items, pi)
			ranges = append(ranges, []int{item.Min, item.Max})
			continue
		}

		itemDefinition, ok := r.definition(item.Name)
		if !ok {
			return nil, parserNotFound(item.Name)
		}

		pi, err := itemDefinition.parser(r, path)
		if err != nil {
			return nil, err
		}

		items = append(items, pi)
		ranges = append(ranges, []int{item.Min, item.Max})
	}

	// for single items, acts like a choice
	if len(items) == 1 && ranges[0][0] == 1 && ranges[0][1] == 1 {
		items[0].setIncludedBy(sp, path)
	}

	sp.items = items
	sp.ranges = ranges
	return sp, nil
}

func (d *sequenceDefinition) commitType() CommitType {
	return d.commit
}

func (p *sequenceParser) nodeName() string { return p.name }

func (p *sequenceParser) setIncludedBy(includedBy parser, path []string) {
	if stringsContain(path, p.name) {
		return
	}

	p.includedBy = append(p.includedBy, includedBy)
}

func (p *sequenceParser) cacheIncluded(c *context, n *Node) {
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

func (p *sequenceParser) parse(t Trace, c *context) {
	t = t.Extend(p.name)
	t.Out1("parsing sequence", c.offset)

	if p.commit&Documentation != 0 {
		t.Out1("fail, doc")
		c.fail(c.offset)
		return
	}

	if c.excluded(c.offset, p.name) {
		t.Out1("excluded")
		c.fail(c.offset)
		return
	}

	c.exclude(c.offset, p.name)
	defer c.include(c.offset, p.name)

	items := p.items
	ranges := p.ranges
	var currentCount int
	node := newNode(p.name, p.commit, c.offset, c.offset)

	for len(items) > 0 {
		m, ok := c.fromCache(items[0].nodeName())
		if ok {
			t.Out1("sequence item found in cache, match:", m, items[0].nodeName(), c.offset)
		} else {
			items[0].parse(t, c)
			m = c.match
		}

		if !m {
			if currentCount < ranges[0][0] {
				t.Out1("fail, item failed")
				c.cache.set(node.from, p.name, nil)
				c.fail(node.from)
				return
			}

			items = items[1:]
			ranges = ranges[1:]
			currentCount = 0
			continue
		}

		if c.node.tokenLength() > 0 {
			node.append(c.node)
			currentCount++
		}

		if c.node.tokenLength() == 0 || ranges[0][1] >= 0 && currentCount == ranges[0][1] {
			items = items[1:]
			ranges = ranges[1:]
			currentCount = 0
		}
	}

	t.Out1("success, items parsed")

	c.cache.set(node.from, p.name, node)
	for _, includedBy := range p.includedBy {
		includedBy.cacheIncluded(c, node)
	}

	c.success(node)
}
