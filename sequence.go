package treerack

type sequenceDefinition struct {
	name       string
	id         int
	commit     CommitType
	items      []SequenceItem
	includedBy []int
	ranges     [][]int
	sbuilder   *sequenceBuilder
	allChars   bool
}

type sequenceParser struct {
	name       string
	id         int
	commit     CommitType
	items      []parser
	ranges     [][]int
	includedBy []int
	allChars   bool
}

type sequenceBuilder struct {
	name       string
	id         int
	commit     CommitType
	items      []builder
	ranges     [][]int
	includedBy *idSet
	allChars   bool
}

func newSequence(name string, ct CommitType, items []SequenceItem) *sequenceDefinition {
	return &sequenceDefinition{
		name:   name,
		commit: ct,
		items:  items,
	}
}

func (d *sequenceDefinition) nodeName() string       { return d.name }
func (d *sequenceDefinition) nodeID() int            { return d.id }
func (d *sequenceDefinition) setID(id int)           { d.id = id }
func (d *sequenceDefinition) commitType() CommitType { return d.commit }

func (d *sequenceDefinition) includeItems() bool {
	return len(d.items) == 1 && d.items[0].Min == 1 && d.items[0].Max == 1
}

func (d *sequenceDefinition) init(r *registry) error {
	if d.sbuilder == nil {
		d.sbuilder = &sequenceBuilder{
			name:       d.name,
			id:         d.id,
			commit:     d.commit,
			includedBy: &idSet{},
		}
	}

	allChars := true
	for _, item := range d.items {
		if item.Min == 0 && item.Max == 0 {
			item.Min, item.Max = 1, 1
		} else if item.Max == 0 {
			item.Max = -1
		}

		d.ranges = append(d.ranges, []int{item.Min, item.Max})

		def, ok := r.definition(item.Name)
		if !ok {
			return parserNotFound(item.Name)
		}

		d.sbuilder.items = append(d.sbuilder.items, def.builder())

		if allChars {
			if _, isChar := def.(*charParser); !isChar {
				allChars = false
			}
		}
	}

	d.sbuilder.ranges = d.ranges
	d.sbuilder.allChars = allChars
	d.allChars = allChars

	if !d.includeItems() {
		return nil
	}

	parsers := &idSet{}
	parsers.set(d.id)
	return setItemsIncludedBy(r, sequenceItemNames(d.items), d.id, parsers)
}

func (d *sequenceDefinition) setIncludedBy(r *registry, includedBy int, parsers *idSet) error {
	if parsers.has(d.id) {
		return nil
	}

	d.includedBy = appendIfMissing(d.includedBy, includedBy)

	if d.sbuilder == nil {
		d.sbuilder = &sequenceBuilder{
			name:       d.name,
			id:         d.id,
			commit:     d.commit,
			includedBy: &idSet{},
		}
	}

	d.sbuilder.includedBy.set(includedBy)

	if !d.includeItems() {
		return nil
	}

	parsers.set(d.id)
	return setItemsIncludedBy(r, sequenceItemNames(d.items), includedBy, parsers)
}

func (d *sequenceDefinition) parser(r *registry, parsers *idSet) (parser, error) {
	// TODO: what is this for? test with sequence containing a sequence through a choice
	if parsers.has(d.id) {
		panic(cannotIncludeParsers(d.name))
	}

	p, ok := r.parser(d.name)
	if ok {
		return p, nil
	}

	sp := &sequenceParser{
		name:       d.name,
		id:         d.id,
		commit:     d.commit,
		includedBy: d.includedBy,
		allChars:   d.allChars,
	}

	r.setParser(sp)

	var items []parser
	parsers.set(d.id)
	defer parsers.unset(d.id)
	for _, item := range d.items {
		pi, ok := r.parser(item.Name)
		if ok {
			items = append(items, pi)
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
	}

	sp.items = items
	sp.ranges = d.ranges
	return sp, nil
}

func (d *sequenceDefinition) builder() builder {
	if d.sbuilder == nil {
		d.sbuilder = &sequenceBuilder{
			name:       d.name,
			id:         d.id,
			commit:     d.commit,
			includedBy: &idSet{},
		}
	}

	return d.sbuilder
}

func (p *sequenceParser) nodeName() string { return p.name }
func (p *sequenceParser) nodeID() int      { return p.id }

func (p *sequenceParser) parse(t Trace, c *context) {
	// t = t.Extend(p.name)
	// t.Out1("parsing sequence", c.offset)

	// if p.commit&Documentation != 0 {
	// 	// t.Out1("fail, doc")
	// 	c.fail(c.offset)
	// 	return
	// }

	if !p.allChars {
		if c.excluded(c.offset, p.id) {
			// t.Out1("fail, excluded")
			c.fail(c.offset)
			return
		}

		c.exclude(c.offset, p.id)
	}

	itemIndex := 0
	var currentCount int
	from := c.offset
	to := c.offset
	var parsed bool

	for itemIndex < len(p.items) {
		// TODO: is it ok to parse before max range check? what if max=0
		p.items[itemIndex].parse(t, c)
		if !c.match {
			if currentCount < p.ranges[itemIndex][0] {
				// c.store.setNoMatch(from, p.id)
				c.fail(from)

				if !p.allChars {
					c.include(from, p.id)
				}

				// t.Out1("fail, not enough items")
				return
			}

			itemIndex++
			currentCount = 0
			continue
		}

		parsed = c.offset > to
		if parsed {
			currentCount++
		}

		to = c.offset

		if !parsed || p.ranges[itemIndex][1] >= 0 && currentCount == p.ranges[itemIndex][1] {
			itemIndex++
			currentCount = 0
		}
	}

	if !p.allChars {
		for _, includedBy := range p.includedBy {
			if c.excluded(from, includedBy) {
				// t.Out1("storing included", includedBy)
				c.store.setMatch(from, includedBy, to)
			}
		}
	}

	// t.Out1("success")
	c.store.setMatch(from, p.id, to)
	c.success(to)

	if !p.allChars {
		c.include(from, p.id)
	}
}

func (b *sequenceBuilder) nodeName() string { return b.name }
func (b *sequenceBuilder) nodeID() int      { return b.id }

func (b *sequenceBuilder) build(c *context) ([]*Node, bool) {
	to, ok := c.store.takeMatch(c.offset, b.id, b.includedBy)
	if !ok {
		return nil, false
	}

	if b.allChars {
		from := c.offset
		c.offset = to
		if b.commit&Alias != 0 {
			return nil, true
		}

		return []*Node{{
			Name:   b.name,
			From:   from,
			To:     to,
			tokens: c.tokens,
		}}, true
	}

	from := c.offset
	var (
		itemIndex    int
		currentCount int
		nodes        []*Node
	)

	for itemIndex < len(b.items) {
		itemFrom := c.offset
		n, ok := b.items[itemIndex].build(c)
		if !ok {
			if currentCount < b.ranges[itemIndex][0] {
				panic(b.name + ": damaged parse result")
			}

			itemIndex++
			currentCount = 0
			continue
		}

		parsed := c.offset > itemFrom
		if parsed {
			nodes = append(nodes, n...)
			currentCount++
		}

		if !parsed || b.ranges[itemIndex][1] >= 0 && currentCount == b.ranges[itemIndex][1] {
			itemIndex++
			currentCount = 0
		}
	}

	if b.commit&Alias != 0 {
		return nodes, true
	}

	return []*Node{{
		Name:   b.name,
		From:   from,
		To:     to,
		Nodes:  nodes,
		tokens: c.tokens,
	}}, true
}
