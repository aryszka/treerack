package treerack

type choiceDefinition struct {
	name        string
	id          int
	commit      CommitType
	elements    []string
	elementDefs []definition
	includedBy  []int
	cbuilder    *choiceBuilder
	cparser     *choiceParser
	validated   bool
	initialized bool
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

func (d *choiceDefinition) nodeName() string            { return d.name }
func (d *choiceDefinition) setNodeName(n string)        { d.name = n }
func (d *choiceDefinition) nodeID() int                 { return d.id }
func (d *choiceDefinition) setID(id int)                { d.id = id }
func (d *choiceDefinition) commitType() CommitType      { return d.commit }
func (d *choiceDefinition) setCommitType(ct CommitType) { d.commit = ct }

func (d *choiceDefinition) validate(r *registry) error {
	if d.validated {
		return nil
	}

	d.validated = true
	for i := range d.elements {
		e, ok := r.definitions[d.elements[i]]
		if !ok {
			return parserNotFound(d.elements[i])
		}

		if err := e.validate(r); err != nil {
			return err
		}
	}

	return nil
}

func (d *choiceDefinition) ensureBuilder() {
	if d.cbuilder != nil {
		return
	}

	d.cbuilder = &choiceBuilder{
		name:       d.name,
		id:         d.id,
		commit:     d.commit,
		includedBy: &idSet{},
	}
}

func (d *choiceDefinition) initElements(r *registry) {
	for _, e := range d.elements {
		def := r.definitions[e]
		d.elementDefs = append(d.elementDefs, def)
		def.init(r)
		d.cbuilder.elements = append(d.cbuilder.elements, def.builder())
		def.setIncludedBy(d.id)
	}
}

func (d *choiceDefinition) init(r *registry) {
	if d.initialized {
		return
	}

	d.initialized = true
	d.ensureBuilder()
	d.initElements(r)
}

func (d *choiceDefinition) setIncludedBy(includedBy int) {
	if intsContain(d.includedBy, includedBy) {
		return
	}

	d.includedBy = append(d.includedBy, includedBy)
	d.ensureBuilder()
	d.cbuilder.includedBy.set(includedBy)
	for _, e := range d.elementDefs {
		e.setIncludedBy(includedBy)
	}
}

func (d *choiceDefinition) createParser() {
	d.cparser = &choiceParser{
		name:       d.name,
		id:         d.id,
		commit:     d.commit,
		includedBy: d.includedBy,
	}
}

func (d *choiceDefinition) createElementParsers() {
	for _, def := range d.elementDefs {
		element := def.parser()
		d.cparser.elements = append(d.cparser.elements, element)
	}
}

func (d *choiceDefinition) parser() parser {
	if d.cparser != nil {
		return d.cparser
	}

	d.createParser()
	d.createElementParsers()
	return d.cparser
}

func (d *choiceDefinition) builder() builder { return d.cbuilder }
func (p *choiceParser) nodeName() string     { return p.name }
func (p *choiceParser) nodeID() int          { return p.id }

func (p *choiceParser) parse(c *context) {
	if c.fromStore(p.id) {
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
	var elementIndex int
	var foundMatch bool

	for {
		foundMatch = false
		elementIndex = 0

		// TODO:
		// - avoid double parsing by setting first-from-store in the context, prepare in advance to
		// know whether it can be it's own item
		// - it is also important to figure why disabling the failed elements breaks the parsing

		for elementIndex < len(p.elements) {
			p.elements[elementIndex].parse(c)
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
