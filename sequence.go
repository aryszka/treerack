package treerack

type sequenceDefinition struct {
	name       string
	id         int
	commit     CommitType
	items      []SequenceItem
	includedBy []int
}

type sequenceParser struct {
	name       string
	id         int
	commit     CommitType
	items      []parser
	ranges     [][]int
	includedBy []int
}

type sequenceBuilder struct {
	name string
	id int
	commit CommitType
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
func (d *sequenceDefinition) commitType() CommitType { return d.commit }

func (d *sequenceDefinition) includeItems() bool {
	return len(d.items) == 1 && d.items[0].Min == 1 && d.items[0].Max == 1
}

func (d *sequenceDefinition) init(r *registry) error {
	for _, item := range d.items {
		if item.Min == 0 && item.Max == 0 {
			item.Min, item.Max = 1, 1
		} else if item.Max == 0 {
			item.Max = -1
		}
	}

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

	sp.items = items
	sp.ranges = ranges
	return sp, nil
}

func (d *sequenceDefinition) builder() builder {
	return &sequenceBuilder{}
}

func (p *sequenceParser) nodeName() string { return p.name }
func (p *sequenceParser) nodeID() int      { return p.id }

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
		c.store.setMatch(from, includedBy, to)
	}

	c.store.setMatch(from, p.id, to)
	c.success(to)
	c.include(from, p.id)
}

func (b *sequenceBuilder) nodeName() string { return b.name }
func (b *sequenceBuilder) nodeID() int      { return b.id }

func (b *sequenceBuilder) build(*context) ([]*Node, bool) {
	return nil, false
}
