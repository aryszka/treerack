package treerack

type sequenceDefinition struct {
	name        string
	id          int
	commit      CommitType
	items       []SequenceItem
	itemDefs    []definition
	includedBy  []int
	ranges      [][]int
	sbuilder    *sequenceBuilder
	sparser     *sequenceParser
	allChars    bool
	validated   bool
	initialized bool
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

func (d *sequenceDefinition) nodeName() string            { return d.name }
func (d *sequenceDefinition) setNodeName(n string)        { d.name = n }
func (d *sequenceDefinition) nodeID() int                 { return d.id }
func (d *sequenceDefinition) setID(id int)                { d.id = id }
func (d *sequenceDefinition) commitType() CommitType      { return d.commit }
func (d *sequenceDefinition) setCommitType(ct CommitType) { d.commit = ct }

func (d *sequenceDefinition) validate(r *registry) error {
	if d.validated {
		return nil
	}

	d.validated = true
	for i := range d.items {
		ii, ok := r.definition(d.items[i].Name)
		if !ok {
			return parserNotFound(d.items[i].Name)
		}

		if err := ii.validate(r); err != nil {
			return err
		}
	}

	return nil
}

func (d *sequenceDefinition) includeItems() bool {
	return len(d.items) == 1 && d.items[0].Max == 1
}

func (d *sequenceDefinition) ensureBuilder() {
	if d.sbuilder != nil {
		return
	}

	d.sbuilder = &sequenceBuilder{
		name:       d.name,
		id:         d.id,
		commit:     d.commit,
		includedBy: &idSet{},
	}
}

func (d *sequenceDefinition) initRanges() {
	for _, item := range d.items {
		if item.Min == 0 && item.Max == 0 {
			item.Min, item.Max = 1, 1
		} else if item.Max == 0 {
			item.Max = -1
		}

		d.ranges = append(d.ranges, []int{item.Min, item.Max})
	}
}

func (d *sequenceDefinition) initItems(r *registry) {
	allChars := true
	for _, item := range d.items {
		def := r.definitions[item.Name]
		d.itemDefs = append(d.itemDefs, def)
		def.init(r)
		d.sbuilder.items = append(d.sbuilder.items, def.builder())
		if allChars {
			if _, isChar := def.(*charParser); !isChar {
				allChars = false
			}
		}
	}

	d.sbuilder.allChars = allChars
	d.allChars = allChars
}

func (d *sequenceDefinition) init(r *registry) {
	if d.initialized {
		return
	}

	d.initialized = true
	d.initRanges()
	d.ensureBuilder()
	d.sbuilder.ranges = d.ranges
	d.initItems(r)
	if d.includeItems() {
		d.itemDefs[0].setIncludedBy(d.id)
	}
}

func (d *sequenceDefinition) setIncludedBy(includedBy int) {
	if intsContain(d.includedBy, includedBy) {
		return
	}

	d.includedBy = append(d.includedBy, includedBy)
	d.ensureBuilder()
	d.sbuilder.includedBy.set(includedBy)
	if d.includeItems() {
		d.itemDefs[0].setIncludedBy(includedBy)
	}
}

func (d *sequenceDefinition) createParser() {
	d.sparser = &sequenceParser{
		name:       d.name,
		id:         d.id,
		commit:     d.commit,
		includedBy: d.includedBy,
		allChars:   d.allChars,
		ranges:     d.ranges,
	}
}

func (d *sequenceDefinition) createItemParsers() {
	for _, item := range d.itemDefs {
		pi := item.parser()
		d.sparser.items = append(d.sparser.items, pi)
	}
}

func (d *sequenceDefinition) parser() parser {
	if d.sparser != nil {
		return d.sparser
	}

	d.createParser()
	d.createItemParsers()
	return d.sparser
}

func (d *sequenceDefinition) builder() builder { return d.sbuilder }
func (p *sequenceParser) nodeName() string     { return p.name }
func (p *sequenceParser) nodeID() int          { return p.id }

func (p *sequenceParser) parse(c *context) {
	if !p.allChars {
		if c.pending(c.offset, p.id) {
			c.fail(c.offset)
			return
		}

		c.markPending(c.offset, p.id)
	}

	itemIndex := 0
	var currentCount int
	from := c.offset
	to := c.offset
	var parsed bool

	for itemIndex < len(p.items) {
		// TODO: is it ok to parse before max range check? what if max=0
		p.items[itemIndex].parse(c)
		if !c.matchLast {
			if currentCount < p.ranges[itemIndex][0] {
				c.fail(from)

				if !p.allChars {
					c.unmarkPending(from, p.id)
				}

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
			if c.pending(from, includedBy) {
				c.results.setMatch(from, includedBy, to)
			}
		}
	}

	c.results.setMatch(from, p.id, to)
	c.success(to)

	if !p.allChars {
		c.unmarkPending(from, p.id)
	}
}

func (b *sequenceBuilder) nodeName() string { return b.name }
func (b *sequenceBuilder) nodeID() int      { return b.id }

func (b *sequenceBuilder) build(c *context) ([]*Node, bool) {
	to, ok := c.results.takeMatch(c.offset, b.id, b.includedBy)
	if !ok {
		return nil, false
	}

	// maybe something like this:
	if to-c.offset == 0 && b.commit&Alias != 0 {
		return nil, true
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

		// maybe can handle the commit type differently

		parsed := c.offset > itemFrom
		if parsed || len(n) > 0 {
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
