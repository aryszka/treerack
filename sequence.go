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

func (p *sequenceParser) storeIncluded(c *context, from, to int) {
	if !c.excluded(from, p.id) {
		return
	}

	c.store.setMatch(from, p.id, to)

	for _, includedBy := range p.includedBy {
		includedBy.storeIncluded(c, from, to)
	}
}

func (p *sequenceParser) parse(t Trace, c *context) {
	if p.commit&Documentation != 0 {
		c.fail(c.offset)
		return
	}

	if c.excluded(c.offset, p.id) {
		c.fail(c.offset)
		return
	}

	// if c.store.hasNoMatch(c.offset, p.id) {
	// 	c.fail(c.offset)
	// }

	c.exclude(c.offset, p.id)

	itemIndex := 0
	var currentCount int
	from := c.offset
	to := c.offset

	for itemIndex < len(p.items) {
		p.items[itemIndex].parse(t, c)
		if !c.match {
			if currentCount < p.ranges[itemIndex][0] {
				// c.store.setNoMatch(from, p.id)
				c.fail(from)
				c.include(from, p.id)
				return
			}

			itemIndex++
			currentCount = 0
			continue
		}

		parsed := c.offset > to
		if parsed {
			currentCount++
		}

		to = c.offset

		if !parsed || p.ranges[itemIndex][1] >= 0 && currentCount == p.ranges[itemIndex][1] {
			itemIndex++
			currentCount = 0
		}
	}

	for _, includedBy := range p.includedBy {
		includedBy.storeIncluded(c, from, to)
	}

	c.store.setMatch(from, p.id, to)
	c.success(to)
	c.include(from, p.id)
}
