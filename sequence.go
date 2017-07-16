package treerack

type sequenceDefinition struct {
	name   string
	id     int
	commit CommitType
	items  []SequenceItem
}

type sequenceParser struct {
	name       string
	id         int
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
func (d *sequenceDefinition) nodeID() int      { return d.id }
func (d *sequenceDefinition) setID(id int)     { d.id = id }

func (d *sequenceDefinition) parser(r *registry, parsers *idSet) (parser, error) {
	if parsers.has(d.id) {
		panic(cannotIncludeParsers(d.name))
	}

	p, ok := r.parser(d.name)
	if ok {
		return p, nil
	}

	sp := &sequenceParser{
		name:   d.name,
		id:     d.id,
		commit: d.commit,
	}

	r.setParser(sp)

	var (
		items  []parser
		ranges [][]int
	)

	parsers.set(d.id)
	defer parsers.unset(d.id)
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

		pi, err := itemDefinition.parser(r, parsers)
		if err != nil {
			return nil, err
		}

		items = append(items, pi)
		ranges = append(ranges, []int{item.Min, item.Max})
	}

	// for single items, acts like a choice
	if len(items) == 1 && ranges[0][0] == 1 && ranges[0][1] == 1 {
		items[0].setIncludedBy(sp, parsers)
	}

	sp.items = items
	sp.ranges = ranges
	return sp, nil
}

func (d *sequenceDefinition) commitType() CommitType {
	return d.commit
}

func (p *sequenceParser) nodeName() string { return p.name }
func (p *sequenceParser) nodeID() int      { return p.id }

func (p *sequenceParser) setIncludedBy(includedBy parser, parsers *idSet) {
	if parsers.has(p.id) {
		return
	}

	p.includedBy = append(p.includedBy, includedBy)
}

func (p *sequenceParser) storeIncluded(c *context, n *Node) {
	if !c.excluded(n.From, p.id) {
		return
	}

	nc := newNode(p.name, p.id, n.From, n.To, p.commit)
	nc.append(n)
	c.store.set(nc.From, p.id, nc)

	for _, includedBy := range p.includedBy {
		includedBy.storeIncluded(c, nc)
	}
}

func (p *sequenceParser) parse(t Trace, c *context) {
	// t = t.Extend(p.name)
	// t.Out1("parsing sequence", c.offset)

	if p.commit&Documentation != 0 {
		// t.Out1("fail, doc")
		c.fail(c.offset)
		return
	}

	if c.excluded(c.offset, p.id) {
		// t.Out1("excluded")
		c.fail(c.offset)
		return
	}

	c.exclude(c.offset, p.id)
	initialOffset := c.offset

	items := p.items
	ranges := p.ranges
	var currentCount int
	node := newNode(p.name, p.id, c.offset, c.offset, p.commit)

	for len(items) > 0 {
		var m bool
		// var ok bool
		// m, ok = c.fromStore(items[0].nodeID())
		// if ok {
		// 	// t.Out1("sequence item found in store, match:", m, items[0].nodeName(), c.offset)
		// } else {
		items[0].parse(t, c)
		m = c.match
		// }

		if !m {
			if currentCount < ranges[0][0] {
				// t.Out1("fail, item failed")
				// c.store.set(node.From, p.id, nil)
				c.fail(node.From)
				c.include(initialOffset, p.id)
				return
			}

			items = items[1:]
			ranges = ranges[1:]
			currentCount = 0
			continue
		}

		// nil as char
		if c.node == nil {
			node.appendChar(c.offset)
			currentCount++
		} else if c.node.tokenLength() > 0 {
			node.append(c.node)
			currentCount++
		}

		if c.node != nil && c.node.tokenLength() == 0 || ranges[0][1] >= 0 && currentCount == ranges[0][1] {
			items = items[1:]
			ranges = ranges[1:]
			currentCount = 0
		}
	}

	// t.Out1("success, items parsed")

	// c.store.set(node.From, p.id, node)
	for _, includedBy := range p.includedBy {
		includedBy.storeIncluded(c, node)
	}

	c.success(node)
	c.include(initialOffset, p.id)
}
