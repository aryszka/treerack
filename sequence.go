package treerack

import "strconv"

type sequenceDefinition struct {
	name            string
	id              int
	commit          CommitType
	originalItems   []SequenceItem
	items           []SequenceItem
	itemDefs        []definition
	ranges          [][]int
	generalizations []int
	sbuilder        *sequenceBuilder
	sparser         *sequenceParser
	allChars        bool
	validated       bool
	initialized     bool
}

type sequenceParser struct {
	name            string
	id              int
	commit          CommitType
	items           []parser
	ranges          [][]int
	generalizations []int
	allChars        bool
}

type sequenceBuilder struct {
	name     string
	id       int
	commit   CommitType
	items    []builder
	ranges   [][]int
	allChars bool
}

func newSequence(name string, ct CommitType, items []SequenceItem) *sequenceDefinition {
	original := make([]SequenceItem, len(items))
	for i := range items {
		original[i] = items[i]
	}

	return &sequenceDefinition{
		name:          name,
		commit:        ct,
		items:         items,
		originalItems: original,
	}
}

func (d *sequenceDefinition) nodeName() string            { return d.name }
func (d *sequenceDefinition) setName(n string)            { d.name = n }
func (d *sequenceDefinition) nodeID() int                 { return d.id }
func (d *sequenceDefinition) setID(id int)                { d.id = id }
func (d *sequenceDefinition) commitType() CommitType      { return d.commit }
func (d *sequenceDefinition) setCommitType(ct CommitType) { d.commit = ct }

func normalizeItemRange(item SequenceItem) SequenceItem {
	if item.Min == 0 && item.Max == 0 {
		item.Min, item.Max = 1, 1
		return item
	}

	if item.Min <= 0 {
		item.Min = 0
	}

	if item.Max <= 0 {
		item.Max = -1
	}

	return item
}

func (d *sequenceDefinition) initRanges() {
	for i, item := range d.items {
		item = normalizeItemRange(item)
		d.items[i] = item
		d.ranges = append(d.ranges, []int{item.Min, item.Max})
	}
}

func (d *sequenceDefinition) preinit() {
	d.initRanges()
}

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

func (d *sequenceDefinition) createBuilder() {
	d.sbuilder = &sequenceBuilder{
		name:   d.name,
		id:     d.id,
		commit: d.commit,
		ranges: d.ranges,
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
	d.createBuilder()
	d.initItems(r)
}

func (d *sequenceDefinition) addGeneralization(g int) {
	if intsContain(d.generalizations, g) {
		return
	}

	d.generalizations = append(d.generalizations, g)
}

func (d *sequenceDefinition) createParser() {
	d.sparser = &sequenceParser{
		name:            d.name,
		id:              d.id,
		commit:          d.commit,
		generalizations: d.generalizations,
		allChars:        d.allChars,
		ranges:          d.ranges,
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

func (d *sequenceDefinition) isCharSequence(r *registry) bool {
	for i := range d.originalItems {
		item := normalizeItemRange(d.originalItems[i])
		if item.Min != 1 || item.Max != 1 {
			return false
		}

		itemDef, _ := r.definition(d.originalItems[i].Name)
		c, ok := itemDef.(*charParser)
		if !ok || !c.isSingleChar() {
			return false
		}
	}

	return true
}

func (d *sequenceDefinition) format(r *registry, f formatFlags) string {
	if d.isCharSequence(r) {
		var chars []rune
		for i := range d.originalItems {
			itemDef, _ := r.definition(d.originalItems[i].Name)
			c, _ := itemDef.(*charParser)
			chars = append(chars, c.chars[0])
		}

		chars = escape(charClassEscape, []rune(charClassBanned), chars)
		return string(append([]rune{'"'}, append(chars, '"')...))
	}

	var chars []rune
	for i := range d.originalItems {
		if len(chars) > 0 {
			chars = append(chars, ' ')
		}

		item := normalizeItemRange(d.originalItems[i])
		needsQuantifier := item.Min != 1 || item.Max != 1

		itemDef, _ := r.definition(item.Name)
		isSymbol := itemDef.commitType()&userDefined != 0

		ch, isChoice := itemDef.(*choiceDefinition)
		isChoiceOfMultiple := isChoice && len(ch.options) > 1

		seq, isSequence := itemDef.(*sequenceDefinition)
		isSequenceOfMultiple := isSequence && len(seq.originalItems) > 1 && !seq.isCharSequence(r)

		needsGrouping := isChoiceOfMultiple || isSequenceOfMultiple

		if isSymbol {
			chars = append(chars, []rune(itemDef.nodeName())...)
		} else {
			if needsGrouping {
				chars = append(chars, '(')
			}

			chars = append(chars, []rune(itemDef.format(r, f))...)

			if needsGrouping {
				chars = append(chars, ')')
			}
		}

		if !needsQuantifier {
			continue
		}

		if item.Min == 0 && item.Max == 1 {
			chars = append(chars, '?')
			continue
		}

		if item.Min == 0 && item.Max < 0 {
			chars = append(chars, '*')
			continue
		}

		if item.Min == 1 && item.Max < 0 {
			chars = append(chars, '+')
			continue
		}

		chars = append(chars, '{')

		if item.Min == item.Max {
			chars = append(chars, []rune(strconv.Itoa(item.Min))...)
		} else {
			if item.Min > 0 {
				chars = append(chars, []rune(strconv.Itoa(item.Min))...)
			}

			chars = append(chars, ',')

			if item.Max >= 0 {
				chars = append(chars, []rune(strconv.Itoa(item.Max))...)
			}
		}

		chars = append(chars, '}')
	}

	return string(chars)
}

func (p *sequenceParser) nodeName() string       { return p.name }
func (p *sequenceParser) nodeID() int            { return p.id }
func (p *sequenceParser) commitType() CommitType { return p.commit }

func (p *sequenceParser) parse(c *context) {
	if !p.allChars {
		if c.results.pending(c.offset, p.id) {
			c.fail(c.offset)
			return
		}

		c.results.markPending(c.offset, p.id)
	}

	itemIndex := 0
	var currentCount int
	from := c.offset
	to := c.offset
	var parsed bool

	for itemIndex < len(p.items) {
		p.items[itemIndex].parse(c)
		if !c.matchLast {
			if currentCount < p.ranges[itemIndex][0] {
				if c.failingParser == nil &&
					p.commitType()&userDefined != 0 &&
					p.commitType()&Whitespace == 0 &&
					p.commitType()&FailPass == 0 {
					c.failingParser = p
				}

				c.fail(from)
				if !p.allChars {
					c.results.unmarkPending(from, p.id)
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

		if !parsed || p.ranges[itemIndex][1] > 0 && currentCount == p.ranges[itemIndex][1] {
			itemIndex++
			currentCount = 0
		}
	}

	for _, g := range p.generalizations {
		if c.results.pending(from, g) {
			c.results.setMatch(from, g, to)
		}
	}

	if to > c.failOffset {
		c.failOffset = -1
		c.failingParser = nil
	}

	c.results.setMatch(from, p.id, to)
	c.success(to)
	if !p.allChars {
		c.results.unmarkPending(from, p.id)
	}
}

func (b *sequenceBuilder) nodeName() string { return b.name }
func (b *sequenceBuilder) nodeID() int      { return b.id }

func (b *sequenceBuilder) build(c *context) ([]*Node, bool) {
	to, ok := c.results.longestMatch(c.offset, b.id)
	if !ok {
		return nil, false
	}

	from := c.offset
	parsed := to > from

	if b.allChars {
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
	} else if parsed {
		c.results.dropMatchTo(c.offset, b.id, to)
	} else {
		if c.results.pending(c.offset, b.id) {
			return nil, false
		}

		c.results.markPending(c.offset, b.id)
	}

	var (
		itemIndex    int
		currentCount int
		nodes        []*Node
	)

	for itemIndex < len(b.items) {
		itemFrom := c.offset
		n, ok := b.items[itemIndex].build(c)
		if !ok {
			itemIndex++
			currentCount = 0
			continue
		}

		if c.offset > itemFrom {
			nodes = append(nodes, n...)
			currentCount++

			if b.ranges[itemIndex][1] > 0 && currentCount == b.ranges[itemIndex][1] {
				itemIndex++
				currentCount = 0
			}

			continue
		}

		if currentCount < b.ranges[itemIndex][0] {
			for i := 0; i < b.ranges[itemIndex][0]-currentCount; i++ {
				nodes = append(nodes, n...)
			}
		}

		itemIndex++
		currentCount = 0
	}

	if !parsed {
		c.results.unmarkPending(from, b.id)
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
