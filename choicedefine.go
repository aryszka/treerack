package treerack

import (
	"fmt"
	"io"
)

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

func (d *choiceDefinition) nodeName() string            { return d.name }
func (d *choiceDefinition) setName(n string)            { d.name = n }
func (d *choiceDefinition) nodeID() int                 { return d.id }
func (d *choiceDefinition) setID(id int)                { d.id = id }
func (d *choiceDefinition) commitType() CommitType      { return d.commit }
func (d *choiceDefinition) setCommitType(ct CommitType) { d.commit = ct }
func (d *choiceDefinition) preinit()                    {}

func newChoice(name string, ct CommitType, options []string) *choiceDefinition {
	return &choiceDefinition{
		name:    name,
		commit:  ct,
		options: options,
	}
}

func (d *choiceDefinition) validate(r *registry) error {
	if d.validated {
		return nil
	}

	d.validated = true
	for i := range d.options {
		o, ok := r.definition[d.options[i]]
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
		def := r.definition[o]
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

		optionDef := r.definition[d.options[i]]
		if optionDef.commitType()&userDefined != 0 {
			chars = append(chars, []rune(optionDef.nodeName())...)
		} else {
			chars = append(chars, []rune(optionDef.format(r, f))...)
		}
	}

	return string(chars)
}

func (p *choiceParser) generate(w io.Writer, done map[string]bool) error {
	if done[p.name] {
		return nil
	}

	done[p.name] = true

	var err error
	fprintf := func(f string, args ...interface{}) {
		if err != nil {
			return
		}

		_, err = fmt.Fprintf(w, f, args...)
	}

	fprintf("var p%d = choiceParser{", p.id)
	fprintf("id: %d, commit: %d,", p.id, p.commit)

	if p.commitType()&userDefined != 0 {
		fprintf("name: \"%s\",", p.name)
	}

	if len(p.generalizations) > 0 {
		fprintf("generalizations: []int{")
		for i := range p.generalizations {
			fprintf("%d,", p.generalizations[i])
		}

		fprintf("},")
	}

	fprintf("};")

	if len(p.options) > 0 {
		for i := range p.options {
			if err := p.options[i].(generator).generate(w, done); err != nil {
				return err
			}
		}

		fprintf("p%d.options = []parser{", p.id)
		for i := range p.options {
			fprintf("&p%d,", p.options[i].nodeID())
		}

		fprintf("};")
	}

	return err
}

func (b *choiceBuilder) generate(w io.Writer, done map[string]bool) error {
	if done[b.name] {
		return nil
	}

	done[b.name] = true

	var err error
	fprintf := func(f string, args ...interface{}) {
		if err != nil {
			return
		}

		_, err = fmt.Fprintf(w, f, args...)
	}

	fprintf("var b%d = choiceBuilder{", b.id)
	fprintf("id: %d, commit: %d,", b.id, b.commit)

	if b.commit&Alias == 0 {
		fprintf("name: \"%s\",", b.name)
	}

	fprintf("};")

	if len(b.options) > 0 {
		for i := range b.options {
			if err := b.options[i].(generator).generate(w, done); err != nil {
				return err
			}
		}

		fprintf("b%d.options = []builder{", b.id)
		for i := range b.options {
			fprintf("&b%d,", b.options[i].nodeID())
		}

		fprintf("};")
	}

	return err
}
