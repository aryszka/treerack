package treerack

type choiceDefinition struct {
	name            string
	id              int
	commit          CommitType
	options         []string
	optionDefs      []definition
	generalizations []int
	cparser         *choiceParser
	cbuilder        *choiceBuilder
	validated       bool
	initialized     bool
}

type choiceParser struct {
	name            string
	id              int
	commit          CommitType
	options         []parser
	generalizations []int
}

type choiceBuilder struct {
	name    string
	id      int
	commit  CommitType
	options []builder
}

func newChoice(name string, ct CommitType, options []string) *choiceDefinition {
	return &choiceDefinition{
		name:    name,
		commit:  ct,
		options: options,
	}
}

func (d *choiceDefinition) nodeName() string            { return d.name }
func (d *choiceDefinition) setName(n string)            { d.name = n }
func (d *choiceDefinition) nodeID() int                 { return d.id }
func (d *choiceDefinition) setID(id int)                { d.id = id }
func (d *choiceDefinition) commitType() CommitType      { return d.commit }
func (d *choiceDefinition) setCommitType(ct CommitType) { d.commit = ct }
func (d *choiceDefinition) preinit()                    {}

func (d *choiceDefinition) validate(r *registry) error {
	if d.validated {
		return nil
	}

	d.validated = true
	for i := range d.options {
		o, ok := r.definitions[d.options[i]]
		if !ok {
			return parserNotFound(d.options[i])
		}

		if err := o.validate(r); err != nil {
			return err
		}
	}

	return nil
}

func (d *choiceDefinition) createBuilder() {
	d.cbuilder = &choiceBuilder{
		name:   d.name,
		id:     d.id,
		commit: d.commit,
	}
}

func (d *choiceDefinition) initOptions(r *registry) {
	for _, o := range d.options {
		def := r.definitions[o]
		d.optionDefs = append(d.optionDefs, def)
		def.init(r)
		d.cbuilder.options = append(d.cbuilder.options, def.builder())
		def.addGeneralization(d.id)
	}
}

func (d *choiceDefinition) init(r *registry) {
	if d.initialized {
		return
	}

	d.initialized = true
	d.createBuilder()
	d.initOptions(r)
}

func (d *choiceDefinition) addGeneralization(g int) {
	if intsContain(d.generalizations, g) {
		return
	}

	d.generalizations = append(d.generalizations, g)
	for _, o := range d.optionDefs {
		o.addGeneralization(g)
	}
}

func (d *choiceDefinition) createParser() {
	d.cparser = &choiceParser{
		name:            d.name,
		id:              d.id,
		commit:          d.commit,
		generalizations: d.generalizations,
	}
}

func (d *choiceDefinition) createOptionParsers() {
	for _, def := range d.optionDefs {
		option := def.parser()
		d.cparser.options = append(d.cparser.options, option)
	}
}

func (d *choiceDefinition) parser() parser {
	if d.cparser != nil {
		return d.cparser
	}

	d.createParser()
	d.createOptionParsers()
	return d.cparser
}

func (d *choiceDefinition) builder() builder { return d.cbuilder }

func (d *choiceDefinition) format(r *registry, f formatFlags) string {
	var chars []rune
	for i := range d.options {
		if i > 0 {
			chars = append(chars, []rune(" | ")...)
		}

		optionDef, _ := r.definition(d.options[i])
		if optionDef.commitType()&userDefined != 0 {
			chars = append(chars, []rune(optionDef.nodeName())...)
		} else {
			chars = append(chars, []rune(optionDef.format(r, f))...)
		}
	}

	return string(chars)
}

func (p *choiceParser) nodeName() string       { return p.name }
func (p *choiceParser) nodeID() int            { return p.id }
func (p *choiceParser) commitType() CommitType { return p.commit }

func (p *choiceParser) parse(c *context) {
	if c.fromResults(p) {
		return
	}

	if c.results.pending(c.offset, p.id) {
		c.fail(c.offset)
		return
	}

	// println("parsing choice", p.name, c.offset)

	c.results.markPending(c.offset, p.id)
	from := c.offset
	to := c.offset

	var match bool
	var optionIndex int
	var foundMatch bool

	// TODO:
	// - if there is a failure already, it should be left alone
	// - what if reading more means that the previous failurs don't count
	initialFailOffset := c.failOffset

	for {
		foundMatch = false
		optionIndex = 0

		for optionIndex < len(p.options) {
			p.options[optionIndex].parse(c)
			optionIndex++

			if !c.matchLast || match && c.offset <= to {
				c.offset = from
				continue
			}

			match = true
			foundMatch = true
			to = c.offset
			c.offset = from
			c.results.setMatch(from, p.id, to)
		}

		if !foundMatch {
			break
		}
	}

	if match {
		if to >= c.failOffset {
			c.failOffset = -1
			c.failingParser = nil
		}

		c.success(to)
		c.results.unmarkPending(from, p.id)
		return
	}

	// TODO:
	// - what if all of it pending?
	if c.failOffset > initialFailOffset && c.failingParser == nil {
		if p.commitType()&userDefined != 0 && p.commitType()&Whitespace == 0 {
			// println("recording choice failure", p.name, from, c.failOffset)
			c.failingParser = p
		}
	}

	c.results.setNoMatch(from, p.id)
	c.fail(from)
	c.results.unmarkPending(from, p.id)
}

func (b *choiceBuilder) nodeName() string { return b.name }
func (b *choiceBuilder) nodeID() int      { return b.id }

func (b *choiceBuilder) build(c *context) ([]*Node, bool) {
	to, ok := c.results.longestMatch(c.offset, b.id)
	if !ok {
		return nil, false
	}

	from := c.offset
	parsed := to > from

	if parsed {
		c.results.dropMatchTo(c.offset, b.id, to)
	} else {
		if c.results.pending(c.offset, b.id) {
			return nil, false
		}

		c.results.markPending(c.offset, b.id)
	}

	var option builder
	for _, o := range b.options {
		if c.results.hasMatchTo(c.offset, o.nodeID(), to) {
			option = o
			break
		}
	}

	n, _ := option.build(c)
	if !parsed {
		c.results.unmarkPending(from, b.id)
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
