/*
This file was generated with treerack (https://github.com/aryszka/treerack).

The contents of this file fall under different licenses.

The code between the "// head" and "// eo head" lines falls under the same
license as the source code of treerack (https://github.com/aryszka/treerack),
unless explicitly stated otherwise, if treerack's license allows changing the
license of this source code.

Treerack's license: MIT https://opensource.org/licenses/MIT
where YEAR=2017, COPYRIGHT HOLDER=Arpad Ryszka (arpad.ryszka@gmail.com)

The rest of the content of this file falls under the same license as the one
that the user of treerack generating this file declares for it, or it is
unlicensed.
*/

package self

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"
)

type charParser struct {
	name   string
	id     int
	not    bool
	chars  []rune
	ranges [][]rune
}
type charBuilder struct {
	name string
	id   int
}

func (p *charParser) nodeName() string {
	return p.name
}
func (p *charParser) nodeID() int {
	return p.id
}
func (p *charParser) commitType() CommitType {
	return Alias
}
func matchChar(chars []rune, ranges [][]rune, not bool, char rune) bool {
	for _, ci := range chars {
		if ci == char {
			return !not
		}
	}
	for _, ri := range ranges {
		if char >= ri[0] && char <= ri[1] {
			return !not
		}
	}
	return not
}
func (p *charParser) match(t rune) bool {
	return matchChar(p.chars, p.ranges, p.not, t)
}
func (p *charParser) parse(c *context) {
	if tok, ok := c.token(); !ok || !p.match(tok) {
		if c.offset > c.failOffset {
			c.failOffset = c.offset
			c.failingParser = nil
		}
		c.fail(c.offset)
		return
	}
	c.success(c.offset + 1)
}
func (b *charBuilder) nodeName() string {
	return b.name
}
func (b *charBuilder) nodeID() int {
	return b.id
}
func (b *charBuilder) build(c *context) ([]*Node, bool) {
	return nil, false
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

func (p *sequenceParser) nodeName() string {
	return p.name
}
func (p *sequenceParser) nodeID() int {
	return p.id
}
func (p *sequenceParser) commitType() CommitType {
	return p.commit
}
func (p *sequenceParser) parse(c *context) {
	if !p.allChars {
		if c.results.pending(c.offset, p.id) {
			c.fail(c.offset)
			return
		}
		c.results.markPending(c.offset, p.id)
	}
	var (
		currentCount int
		parsed       bool
	)
	itemIndex := 0
	from := c.offset
	to := c.offset
	for itemIndex < len(p.items) {
		p.items[itemIndex].parse(c)
		if !c.matchLast {
			if currentCount >= p.ranges[itemIndex][0] {
				itemIndex++
				currentCount = 0
				continue
			}
			if c.failingParser == nil && p.commit&userDefined != 0 && p.commit&Whitespace == 0 && p.commit&FailPass == 0 {
				c.failingParser = p
			}
			c.fail(from)
			if !p.allChars {
				c.results.unmarkPending(from, p.id)
			}
			return
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
func (b *sequenceBuilder) nodeName() string {
	return b.name
}
func (b *sequenceBuilder) nodeID() int {
	return b.id
}
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
		return []*Node{{Name: b.name, From: from, To: to, tokens: c.tokens}}, true
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
	return []*Node{{Name: b.name, From: from, To: to, Nodes: nodes, tokens: c.tokens}}, true
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

func (p *choiceParser) nodeName() string {
	return p.name
}
func (p *choiceParser) nodeID() int {
	return p.id
}
func (p *choiceParser) commitType() CommitType {
	return p.commit
}
func (p *choiceParser) parse(c *context) {
	if c.fromResults(p) {
		return
	}
	if c.results.pending(c.offset, p.id) {
		c.fail(c.offset)
		return
	}
	c.results.markPending(c.offset, p.id)
	var (
		match         bool
		optionIndex   int
		foundMatch    bool
		failingParser parser
	)
	from := c.offset
	to := c.offset
	initialFailOffset := c.failOffset
	initialFailingParser := c.failingParser
	failOffset := initialFailOffset
	for {
		foundMatch = false
		optionIndex = 0
		for optionIndex < len(p.options) {
			p.options[optionIndex].parse(c)
			optionIndex++
			if !c.matchLast {
				if c.failOffset > failOffset {
					failOffset = c.failOffset
					failingParser = c.failingParser
				}
			}
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
		if failOffset > to {
			c.failOffset = failOffset
			c.failingParser = failingParser
		} else if to > initialFailOffset {
			c.failOffset = -1
			c.failingParser = nil
		} else {
			c.failOffset = initialFailOffset
			c.failingParser = initialFailingParser
		}
		c.success(to)
		c.results.unmarkPending(from, p.id)
		return
	}
	if failOffset > initialFailOffset {
		c.failOffset = failOffset
		c.failingParser = failingParser
		if c.failingParser == nil && p.commitType()&userDefined != 0 && p.commitType()&Whitespace == 0 && p.commitType()&FailPass == 0 {
			c.failingParser = p
		}
	}
	c.results.setNoMatch(from, p.id)
	c.fail(from)
	c.results.unmarkPending(from, p.id)
}
func (b *choiceBuilder) nodeName() string {
	return b.name
}
func (b *choiceBuilder) nodeID() int {
	return b.id
}
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
	return []*Node{{Name: b.name, From: from, To: to, Nodes: n, tokens: c.tokens}}, true
}

type idSet struct{ ids []uint }

func divModBits(id int) (int, int) {
	return id / strconv.IntSize, id % strconv.IntSize
}
func (s *idSet) set(id int) {
	d, m := divModBits(id)
	if d >= len(s.ids) {
		if d < cap(s.ids) {
			s.ids = s.ids[:d+1]
		} else {
			s.ids = s.ids[:cap(s.ids)]
			for i := cap(s.ids); i <= d; i++ {
				s.ids = append(s.ids, 0)
			}
		}
	}
	s.ids[d] |= 1 << uint(m)
}
func (s *idSet) unset(id int) {
	d, m := divModBits(id)
	if d >= len(s.ids) {
		return
	}
	s.ids[d] &^= 1 << uint(m)
}
func (s *idSet) has(id int) bool {
	d, m := divModBits(id)
	if d >= len(s.ids) {
		return false
	}
	return s.ids[d]&(1<<uint(m)) != 0
}

type results struct {
	noMatch   []*idSet
	match     [][]int
	isPending [][]int
}

func ensureOffsetInts(ints [][]int, offset int) [][]int {
	if len(ints) > offset {
		return ints
	}
	if cap(ints) > offset {
		ints = ints[:offset+1]
		return ints
	}
	ints = ints[:cap(ints)]
	for i := len(ints); i <= offset; i++ {
		ints = append(ints, nil)
	}
	return ints
}
func ensureOffsetIDs(ids []*idSet, offset int) []*idSet {
	if len(ids) > offset {
		return ids
	}
	if cap(ids) > offset {
		ids = ids[:offset+1]
		return ids
	}
	ids = ids[:cap(ids)]
	for i := len(ids); i <= offset; i++ {
		ids = append(ids, nil)
	}
	return ids
}
func (r *results) setMatch(offset, id, to int) {
	r.match = ensureOffsetInts(r.match, offset)
	for i := 0; i < len(r.match[offset]); i += 2 {
		if r.match[offset][i] != id || r.match[offset][i+1] != to {
			continue
		}
		return
	}
	r.match[offset] = append(r.match[offset], id, to)
}
func (r *results) setNoMatch(offset, id int) {
	if len(r.match) > offset {
		for i := 0; i < len(r.match[offset]); i += 2 {
			if r.match[offset][i] != id {
				continue
			}
			return
		}
	}
	r.noMatch = ensureOffsetIDs(r.noMatch, offset)
	if r.noMatch[offset] == nil {
		r.noMatch[offset] = &idSet{}
	}
	r.noMatch[offset].set(id)
}
func (r *results) hasMatchTo(offset, id, to int) bool {
	if len(r.match) <= offset {
		return false
	}
	for i := 0; i < len(r.match[offset]); i += 2 {
		if r.match[offset][i] != id {
			continue
		}
		if r.match[offset][i+1] == to {
			return true
		}
	}
	return false
}
func (r *results) longestMatch(offset, id int) (int, bool) {
	if len(r.match) <= offset {
		return 0, false
	}
	var found bool
	to := -1
	for i := 0; i < len(r.match[offset]); i += 2 {
		if r.match[offset][i] != id {
			continue
		}
		if r.match[offset][i+1] > to {
			to = r.match[offset][i+1]
		}
		found = true
	}
	return to, found
}
func (r *results) longestResult(offset, id int) (int, bool, bool) {
	if len(r.noMatch) > offset && r.noMatch[offset] != nil && r.noMatch[offset].has(id) {
		return 0, false, true
	}
	to, ok := r.longestMatch(offset, id)
	return to, ok, ok
}
func (r *results) dropMatchTo(offset, id, to int) {
	for i := 0; i < len(r.match[offset]); i += 2 {
		if r.match[offset][i] != id {
			continue
		}
		if r.match[offset][i+1] == to {
			r.match[offset][i] = -1
			return
		}
	}
}
func (r *results) resetPending() {
	r.isPending = nil
}
func (r *results) pending(offset, id int) bool {
	if len(r.isPending) <= id {
		return false
	}
	for i := range r.isPending[id] {
		if r.isPending[id][i] == offset {
			return true
		}
	}
	return false
}
func (r *results) markPending(offset, id int) {
	r.isPending = ensureOffsetInts(r.isPending, id)
	for i := range r.isPending[id] {
		if r.isPending[id][i] == -1 {
			r.isPending[id][i] = offset
			return
		}
	}
	r.isPending[id] = append(r.isPending[id], offset)
}
func (r *results) unmarkPending(offset, id int) {
	for i := range r.isPending[id] {
		if r.isPending[id][i] == offset {
			r.isPending[id][i] = -1
			break
		}
	}
}

type context struct {
	reader        io.RuneReader
	offset        int
	readOffset    int
	consumed      int
	failOffset    int
	failingParser parser
	readErr       error
	eof           bool
	results       *results
	tokens        []rune
	matchLast     bool
}

func newContext(r io.RuneReader) *context {
	return &context{reader: r, results: &results{}, failOffset: -1}
}
func (c *context) read() bool {
	if c.eof || c.readErr != nil {
		return false
	}
	token, n, err := c.reader.ReadRune()
	if err != nil {
		if err == io.EOF {
			if n == 0 {
				c.eof = true
				return false
			}
		} else {
			c.readErr = err
			return false
		}
	}
	c.readOffset++
	if token == unicode.ReplacementChar {
		c.readErr = ErrInvalidUnicodeCharacter
		return false
	}
	c.tokens = append(c.tokens, token)
	return true
}
func (c *context) token() (rune, bool) {
	if c.offset == c.readOffset {
		if !c.read() {
			return 0, false
		}
	}
	return c.tokens[c.offset], true
}
func (c *context) fromResults(p parser) bool {
	to, m, ok := c.results.longestResult(c.offset, p.nodeID())
	if !ok {
		return false
	}
	if m {
		c.success(to)
	} else {
		c.fail(c.offset)
	}
	return true
}
func (c *context) success(to int) {
	c.offset = to
	c.matchLast = true
	if to > c.consumed {
		c.consumed = to
	}
}
func (c *context) fail(offset int) {
	c.offset = offset
	c.matchLast = false
}
func findLine(tokens []rune, offset int) (line, column int) {
	tokens = tokens[:offset]
	for i := range tokens {
		column++
		if tokens[i] == '\n' {
			column = 0
			line++
		}
	}
	return
}
func (c *context) parseError(p parser) error {
	definition := p.nodeName()
	flagIndex := strings.Index(definition, ":")
	if flagIndex > 0 {
		definition = definition[:flagIndex]
	}
	if c.failingParser == nil {
		c.failOffset = c.consumed
	}
	line, col := findLine(c.tokens, c.failOffset)
	return &ParseError{Offset: c.failOffset, Line: line, Column: col, Definition: definition}
}
func (c *context) finalizeParse(root parser) error {
	fp := c.failingParser
	if fp == nil {
		fp = root
	}
	to, match, found := c.results.longestResult(0, root.nodeID())
	if !found || !match || found && match && to < c.readOffset {
		return c.parseError(fp)
	}
	c.read()
	if c.eof {
		return nil
	}
	if c.readErr != nil {
		return c.readErr
	}
	return c.parseError(root)
}

type Node struct {
	Name     string
	Nodes    []*Node
	From, To int
	tokens   []rune
}

func (n *Node) Tokens() []rune {
	return n.tokens
}
func (n *Node) String() string {
	return fmt.Sprintf("%s:%d:%d:%s", n.Name, n.From, n.To, n.Text())
}
func (n *Node) Text() string {
	return string(n.Tokens()[n.From:n.To])
}

type CommitType int

const (
	None  CommitType = 0
	Alias CommitType = 1 << iota
	Whitespace
	NoWhitespace
	FailPass
	Root
	userDefined
)

type formatFlags int

const (
	formatNone   formatFlags = 0
	formatPretty formatFlags = 1 << iota
	formatIncludeComments
)

type ParseError struct {
	Input      string
	Offset     int
	Line       int
	Column     int
	Definition string
}
type parser interface {
	nodeName() string
	nodeID() int
	commitType() CommitType
	parse(*context)
}
type builder interface {
	nodeName() string
	nodeID() int
	build(*context) ([]*Node, bool)
}

var ErrInvalidUnicodeCharacter = errors.New("invalid unicode character")

func (pe *ParseError) Error() string {
	return fmt.Sprintf("%s:%d:%d:parse failed, parsing: %s", pe.Input, pe.Line+1, pe.Column+1, pe.Definition)
}
func parseInput(r io.Reader, p parser, b builder) (*Node, error) {
	c := newContext(bufio.NewReader(r))
	p.parse(c)
	if c.readErr != nil {
		return nil, c.readErr
	}
	if err := c.finalizeParse(p); err != nil {
		if perr, ok := err.(*ParseError); ok {
			perr.Input = "<input>"
		}
		return nil, err
	}
	c.offset = 0
	c.results.resetPending()
	n, _ := b.build(c)
	return n[0], nil
}

func Parse(r io.Reader) (*Node, error) {
	var p188 = sequenceParser{id: 188, commit: 32, ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var p186 = choiceParser{id: 186, commit: 2}
	var p185 = choiceParser{id: 185, commit: 70, name: "wsc", generalizations: []int{186}}
	var p5 = choiceParser{id: 5, commit: 66, name: "wschar", generalizations: []int{185, 186}}
	var p112 = sequenceParser{id: 112, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{5, 185, 186}}
	var p125 = charParser{id: 125, chars: []rune{32}}
	p112.items = []parser{&p125}
	var p173 = sequenceParser{id: 173, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{5, 185, 186}}
	var p102 = charParser{id: 102, chars: []rune{9}}
	p173.items = []parser{&p102}
	var p51 = sequenceParser{id: 51, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{5, 185, 186}}
	var p153 = charParser{id: 153, chars: []rune{10}}
	p51.items = []parser{&p153}
	var p126 = sequenceParser{id: 126, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{5, 185, 186}}
	var p141 = charParser{id: 141, chars: []rune{8}}
	p126.items = []parser{&p141}
	var p40 = sequenceParser{id: 40, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{5, 185, 186}}
	var p113 = charParser{id: 113, chars: []rune{12}}
	p40.items = []parser{&p113}
	var p142 = sequenceParser{id: 142, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{5, 185, 186}}
	var p60 = charParser{id: 60, chars: []rune{13}}
	p142.items = []parser{&p60}
	var p41 = sequenceParser{id: 41, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{5, 185, 186}}
	var p103 = charParser{id: 103, chars: []rune{11}}
	p41.items = []parser{&p103}
	p5.options = []parser{&p112, &p173, &p51, &p126, &p40, &p142, &p41}
	var p73 = sequenceParser{id: 73, commit: 72, name: "comment", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}, generalizations: []int{185, 186}}
	var p97 = choiceParser{id: 97, commit: 74, name: "comment-segment"}
	var p104 = sequenceParser{id: 104, commit: 74, name: "line-comment", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}, generalizations: []int{97}}
	var p149 = sequenceParser{id: 149, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var p42 = charParser{id: 42, chars: []rune{47}}
	var p17 = charParser{id: 17, chars: []rune{47}}
	p149.items = []parser{&p42, &p17}
	var p146 = sequenceParser{id: 146, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p32 = charParser{id: 32, not: true, chars: []rune{10}}
	p146.items = []parser{&p32}
	p104.items = []parser{&p149, &p146}
	var p62 = sequenceParser{id: 62, commit: 74, name: "block-comment", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {1, 1}, {0, -1}, {1, 1}}, generalizations: []int{97}}
	var p92 = sequenceParser{id: 92, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var p171 = charParser{id: 171, chars: []rune{47}}
	var p52 = charParser{id: 52, chars: []rune{42}}
	p92.items = []parser{&p171, &p52}
	var p8 = choiceParser{id: 8, commit: 10}
	var p154 = sequenceParser{id: 154, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{8}}
	var p118 = sequenceParser{id: 118, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p127 = charParser{id: 127, chars: []rune{42}}
	p118.items = []parser{&p127}
	var p61 = sequenceParser{id: 61, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p121 = charParser{id: 121, not: true, chars: []rune{47}}
	p61.items = []parser{&p121}
	p154.items = []parser{&p118, &p61}
	var p22 = sequenceParser{id: 22, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{8}}
	var p155 = charParser{id: 155, not: true, chars: []rune{42}}
	p22.items = []parser{&p155}
	p8.options = []parser{&p154, &p22}
	var p53 = sequenceParser{id: 53, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var p79 = charParser{id: 79, chars: []rune{42}}
	var p158 = charParser{id: 158, chars: []rune{47}}
	p53.items = []parser{&p79, &p158}
	p62.items = []parser{&p92, &p8, &p53}
	p97.options = []parser{&p104, &p62}
	var p80 = sequenceParser{id: 80, commit: 10, ranges: [][]int{{0, -1}, {0, 1}, {0, -1}, {1, 1}, {0, -1}, {0, 1}, {0, -1}, {1, 1}}}
	var p135 = choiceParser{id: 135, commit: 74, name: "ws-no-nl"}
	var p86 = sequenceParser{id: 86, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{135}}
	var p105 = charParser{id: 105, chars: []rune{32}}
	p86.items = []parser{&p105}
	var p25 = sequenceParser{id: 25, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{135}}
	var p35 = charParser{id: 35, chars: []rune{9}}
	p25.items = []parser{&p35}
	var p165 = sequenceParser{id: 165, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{135}}
	var p93 = charParser{id: 93, chars: []rune{8}}
	p165.items = []parser{&p93}
	var p9 = sequenceParser{id: 9, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{135}}
	var p98 = charParser{id: 98, chars: []rune{12}}
	p9.items = []parser{&p98}
	var p134 = sequenceParser{id: 134, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{135}}
	var p147 = charParser{id: 147, chars: []rune{13}}
	p134.items = []parser{&p147}
	var p119 = sequenceParser{id: 119, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{135}}
	var p87 = charParser{id: 87, chars: []rune{11}}
	p119.items = []parser{&p87}
	p135.options = []parser{&p86, &p25, &p165, &p9, &p134, &p119}
	var p136 = sequenceParser{id: 136, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p122 = charParser{id: 122, chars: []rune{10}}
	p136.items = []parser{&p122}
	p80.items = []parser{&p135, &p136, &p135, &p97}
	p73.items = []parser{&p97, &p80}
	p185.options = []parser{&p5, &p73}
	p186.options = []parser{&p185}
	var p187 = sequenceParser{id: 187, commit: 66, name: "syntax:wsroot", ranges: [][]int{{0, 1}, {0, -1}, {0, 1}, {0, 1}}}
	var p29 = sequenceParser{id: 29, commit: 2, ranges: [][]int{{1, 1}, {0, -1}}}
	var p184 = sequenceParser{id: 184, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p140 = charParser{id: 140, chars: []rune{59}}
	p184.items = []parser{&p140}
	var p28 = sequenceParser{id: 28, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p28.items = []parser{&p186, &p184}
	p29.items = []parser{&p184, &p28}
	var p59 = sequenceParser{id: 59, commit: 66, name: "definitions", ranges: [][]int{{1, 1}, {0, 1}}}
	var p172 = sequenceParser{id: 172, commit: 64, name: "definition", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}}}
	var p91 = sequenceParser{id: 91, commit: 74, name: "definition-name", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}}
	var p10 = sequenceParser{id: 10, commit: 72, name: "symbol", ranges: [][]int{{1, -1}, {1, -1}}, generalizations: []int{116, 63, 77}}
	var p69 = sequenceParser{id: 69, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p100 = charParser{id: 100, not: true, chars: []rune{92, 32, 10, 9, 8, 12, 13, 11, 47, 46, 91, 93, 34, 123, 125, 94, 43, 42, 63, 124, 40, 41, 58, 61, 59}}
	p69.items = []parser{&p100}
	p10.items = []parser{&p69}
	var p4 = sequenceParser{id: 4, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var p66 = sequenceParser{id: 66, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p133 = charParser{id: 133, chars: []rune{58}}
	p66.items = []parser{&p133}
	var p20 = choiceParser{id: 20, commit: 66, name: "flag"}
	var p170 = sequenceParser{id: 170, commit: 72, name: "alias", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{20}}
	var p169 = charParser{id: 169, chars: []rune{97}}
	var p64 = charParser{id: 64, chars: []rune{108}}
	var p45 = charParser{id: 45, chars: []rune{105}}
	var p111 = charParser{id: 111, chars: []rune{97}}
	var p95 = charParser{id: 95, chars: []rune{115}}
	p170.items = []parser{&p169, &p64, &p45, &p111, &p95}
	var p23 = sequenceParser{id: 23, commit: 72, name: "ws", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{20}}
	var p101 = charParser{id: 101, chars: []rune{119}}
	var p15 = charParser{id: 15, chars: []rune{115}}
	p23.items = []parser{&p101, &p15}
	var p178 = sequenceParser{id: 178, commit: 72, name: "nows", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{20}}
	var p48 = charParser{id: 48, chars: []rune{110}}
	var p137 = charParser{id: 137, chars: []rune{111}}
	var p144 = charParser{id: 144, chars: []rune{119}}
	var p65 = charParser{id: 65, chars: []rune{115}}
	p178.items = []parser{&p48, &p137, &p144, &p65}
	var p14 = sequenceParser{id: 14, commit: 72, name: "failpass", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{20}}
	var p117 = charParser{id: 117, chars: []rune{102}}
	var p124 = charParser{id: 124, chars: []rune{97}}
	var p49 = charParser{id: 49, chars: []rune{105}}
	var p179 = charParser{id: 179, chars: []rune{108}}
	var p56 = charParser{id: 56, chars: []rune{112}}
	var p34 = charParser{id: 34, chars: []rune{97}}
	var p24 = charParser{id: 24, chars: []rune{115}}
	var p96 = charParser{id: 96, chars: []rune{115}}
	p14.items = []parser{&p117, &p124, &p49, &p179, &p56, &p34, &p24, &p96}
	var p50 = sequenceParser{id: 50, commit: 72, name: "root", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{20}}
	var p152 = charParser{id: 152, chars: []rune{114}}
	var p85 = charParser{id: 85, chars: []rune{111}}
	var p72 = charParser{id: 72, chars: []rune{111}}
	var p78 = charParser{id: 78, chars: []rune{116}}
	p50.items = []parser{&p152, &p85, &p72, &p78}
	p20.options = []parser{&p170, &p23, &p178, &p14, &p50}
	p4.items = []parser{&p66, &p20}
	p91.items = []parser{&p10, &p4}
	var p180 = sequenceParser{id: 180, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p145 = charParser{id: 145, chars: []rune{61}}
	p180.items = []parser{&p145}
	var p116 = choiceParser{id: 116, commit: 66, name: "expression"}
	var p174 = choiceParser{id: 174, commit: 66, name: "terminal", generalizations: []int{116, 63, 77}}
	var p130 = sequenceParser{id: 130, commit: 72, name: "any-char", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{174, 116, 63, 77}}
	var p166 = charParser{id: 166, chars: []rune{46}}
	p130.items = []parser{&p166}
	var p129 = sequenceParser{id: 129, commit: 72, name: "char-class", ranges: [][]int{{1, 1}, {0, 1}, {0, -1}, {1, 1}, {1, 1}, {0, 1}, {0, -1}, {1, 1}}, generalizations: []int{174, 116, 63, 77}}
	var p114 = sequenceParser{id: 114, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p82 = charParser{id: 82, chars: []rune{91}}
	p114.items = []parser{&p82}
	var p18 = sequenceParser{id: 18, commit: 72, name: "class-not", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p88 = charParser{id: 88, chars: []rune{94}}
	p18.items = []parser{&p88}
	var p182 = choiceParser{id: 182, commit: 10}
	var p55 = choiceParser{id: 55, commit: 72, name: "class-char", generalizations: []int{182}}
	var p36 = sequenceParser{id: 36, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{55, 182}}
	var p181 = charParser{id: 181, not: true, chars: []rune{92, 91, 93, 94, 45}}
	p36.items = []parser{&p181}
	var p6 = sequenceParser{id: 6, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{55, 182}}
	var p74 = sequenceParser{id: 74, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p37 = charParser{id: 37, chars: []rune{92}}
	p74.items = []parser{&p37}
	var p94 = sequenceParser{id: 94, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p167 = charParser{id: 167, not: true}
	p94.items = []parser{&p167}
	p6.items = []parser{&p74, &p94}
	p55.options = []parser{&p36, &p6}
	var p12 = sequenceParser{id: 12, commit: 72, name: "char-range", ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{182}}
	var p75 = sequenceParser{id: 75, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p68 = charParser{id: 68, chars: []rune{45}}
	p75.items = []parser{&p68}
	p12.items = []parser{&p55, &p75, &p55}
	p182.options = []parser{&p55, &p12}
	var p128 = sequenceParser{id: 128, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p160 = charParser{id: 160, chars: []rune{93}}
	p128.items = []parser{&p160}
	p129.items = []parser{&p114, &p18, &p182, &p128}
	var p156 = sequenceParser{id: 156, commit: 72, name: "char-sequence", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {1, 1}, {0, -1}, {1, 1}}, generalizations: []int{174, 116, 63, 77}}
	var p164 = sequenceParser{id: 164, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p163 = charParser{id: 163, chars: []rune{34}}
	p164.items = []parser{&p163}
	var p162 = choiceParser{id: 162, commit: 72, name: "sequence-char"}
	var p131 = sequenceParser{id: 131, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{162}}
	var p159 = charParser{id: 159, not: true, chars: []rune{92, 34}}
	p131.items = []parser{&p159}
	var p26 = sequenceParser{id: 26, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{162}}
	var p7 = sequenceParser{id: 7, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p33 = charParser{id: 33, chars: []rune{92}}
	p7.items = []parser{&p33}
	var p161 = sequenceParser{id: 161, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p1 = charParser{id: 1, not: true}
	p161.items = []parser{&p1}
	p26.items = []parser{&p7, &p161}
	p162.options = []parser{&p131, &p26}
	var p89 = sequenceParser{id: 89, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p99 = charParser{id: 99, chars: []rune{34}}
	p89.items = []parser{&p99}
	p156.items = []parser{&p164, &p162, &p89}
	p174.options = []parser{&p130, &p129, &p156}
	var p90 = sequenceParser{id: 90, commit: 66, name: "group", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}}, generalizations: []int{116, 63, 77}}
	var p183 = sequenceParser{id: 183, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p2 = charParser{id: 2, chars: []rune{40}}
	p183.items = []parser{&p2}
	var p175 = sequenceParser{id: 175, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p70 = charParser{id: 70, chars: []rune{41}}
	p175.items = []parser{&p70}
	p90.items = []parser{&p183, &p186, &p116, &p186, &p175}
	var p44 = sequenceParser{id: 44, commit: 64, name: "sequence", ranges: [][]int{{1, 1}, {0, -1}}, generalizations: []int{116, 77}}
	var p39 = sequenceParser{id: 39, commit: 72, name: "item", ranges: [][]int{{1, 1}, {0, 1}, {1, 1}, {0, 1}}}
	var p63 = choiceParser{id: 63, commit: 10}
	p63.options = []parser{&p174, &p10, &p90}
	var p76 = choiceParser{id: 76, commit: 66, name: "quantity"}
	var p143 = sequenceParser{id: 143, commit: 64, name: "count-quantifier", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}}, generalizations: []int{76}}
	var p13 = sequenceParser{id: 13, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p83 = charParser{id: 83, chars: []rune{123}}
	p13.items = []parser{&p83}
	var p106 = sequenceParser{id: 106, commit: 64, name: "count", ranges: [][]int{{1, 1}}}
	var p120 = sequenceParser{id: 120, commit: 74, name: "number", ranges: [][]int{{1, -1}, {1, -1}}}
	var p150 = sequenceParser{id: 150, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p38 = charParser{id: 38, ranges: [][]rune{{48, 57}}}
	p150.items = []parser{&p38}
	p120.items = []parser{&p150}
	p106.items = []parser{&p120}
	var p84 = sequenceParser{id: 84, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p71 = charParser{id: 71, chars: []rune{125}}
	p84.items = []parser{&p71}
	p143.items = []parser{&p13, &p186, &p106, &p186, &p84}
	var p132 = sequenceParser{id: 132, commit: 64, name: "range-quantifier", ranges: [][]int{{1, 1}, {0, -1}, {0, 1}, {0, -1}, {1, 1}, {0, -1}, {0, 1}, {0, -1}, {1, 1}}, generalizations: []int{76}}
	var p81 = sequenceParser{id: 81, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p176 = charParser{id: 176, chars: []rune{123}}
	p81.items = []parser{&p176}
	var p148 = sequenceParser{id: 148, commit: 64, name: "range-from", ranges: [][]int{{1, 1}}}
	p148.items = []parser{&p120}
	var p46 = sequenceParser{id: 46, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p168 = charParser{id: 168, chars: []rune{44}}
	p46.items = []parser{&p168}
	var p19 = sequenceParser{id: 19, commit: 64, name: "range-to", ranges: [][]int{{1, 1}}}
	p19.items = []parser{&p120}
	var p27 = sequenceParser{id: 27, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p123 = charParser{id: 123, chars: []rune{125}}
	p27.items = []parser{&p123}
	p132.items = []parser{&p81, &p186, &p148, &p186, &p46, &p186, &p19, &p186, &p27}
	var p47 = sequenceParser{id: 47, commit: 72, name: "one-or-more", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{76}}
	var p107 = charParser{id: 107, chars: []rune{43}}
	p47.items = []parser{&p107}
	var p157 = sequenceParser{id: 157, commit: 72, name: "zero-or-more", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{76}}
	var p151 = charParser{id: 151, chars: []rune{42}}
	p157.items = []parser{&p151}
	var p3 = sequenceParser{id: 3, commit: 72, name: "zero-or-one", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{76}}
	var p108 = charParser{id: 108, chars: []rune{63}}
	p3.items = []parser{&p108}
	p76.options = []parser{&p143, &p132, &p47, &p157, &p3}
	p39.items = []parser{&p63, &p76}
	var p43 = sequenceParser{id: 43, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p43.items = []parser{&p186, &p39}
	p44.items = []parser{&p39, &p43}
	var p110 = sequenceParser{id: 110, commit: 64, name: "choice", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}, generalizations: []int{116}}
	var p77 = choiceParser{id: 77, commit: 66, name: "option"}
	p77.options = []parser{&p174, &p10, &p90, &p44}
	var p115 = sequenceParser{id: 115, commit: 2, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}}}
	var p177 = sequenceParser{id: 177, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p11 = charParser{id: 11, chars: []rune{124}}
	p177.items = []parser{&p11}
	p115.items = []parser{&p177, &p186, &p77}
	var p109 = sequenceParser{id: 109, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p109.items = []parser{&p186, &p115}
	p110.items = []parser{&p77, &p186, &p115, &p109}
	p116.options = []parser{&p174, &p10, &p90, &p44, &p110}
	p172.items = []parser{&p91, &p186, &p180, &p186, &p116}
	var p58 = sequenceParser{id: 58, commit: 2, ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var p139 = sequenceParser{id: 139, commit: 2, ranges: [][]int{{1, 1}, {0, -1}, {0, -1}, {1, 1}}}
	var p21 = sequenceParser{id: 21, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p16 = charParser{id: 16, chars: []rune{59}}
	p21.items = []parser{&p16}
	var p138 = sequenceParser{id: 138, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p138.items = []parser{&p186, &p21}
	p139.items = []parser{&p21, &p138, &p186, &p172}
	var p57 = sequenceParser{id: 57, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p57.items = []parser{&p186, &p139}
	p58.items = []parser{&p186, &p139, &p57}
	p59.items = []parser{&p172, &p58}
	var p31 = sequenceParser{id: 31, commit: 2, ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var p54 = sequenceParser{id: 54, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p67 = charParser{id: 67, chars: []rune{59}}
	p54.items = []parser{&p67}
	var p30 = sequenceParser{id: 30, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p30.items = []parser{&p186, &p54}
	p31.items = []parser{&p186, &p54, &p30}
	p187.items = []parser{&p29, &p186, &p59, &p31}
	p188.items = []parser{&p186, &p187, &p186}
	var b188 = sequenceBuilder{id: 188, commit: 32, name: "syntax", ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var b186 = choiceBuilder{id: 186, commit: 2}
	var b185 = choiceBuilder{id: 185, commit: 70}
	var b5 = choiceBuilder{id: 5, commit: 66}
	var b112 = sequenceBuilder{id: 112, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b125 = charBuilder{}
	b112.items = []builder{&b125}
	var b173 = sequenceBuilder{id: 173, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b102 = charBuilder{}
	b173.items = []builder{&b102}
	var b51 = sequenceBuilder{id: 51, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b153 = charBuilder{}
	b51.items = []builder{&b153}
	var b126 = sequenceBuilder{id: 126, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b141 = charBuilder{}
	b126.items = []builder{&b141}
	var b40 = sequenceBuilder{id: 40, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b113 = charBuilder{}
	b40.items = []builder{&b113}
	var b142 = sequenceBuilder{id: 142, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b60 = charBuilder{}
	b142.items = []builder{&b60}
	var b41 = sequenceBuilder{id: 41, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b103 = charBuilder{}
	b41.items = []builder{&b103}
	b5.options = []builder{&b112, &b173, &b51, &b126, &b40, &b142, &b41}
	var b73 = sequenceBuilder{id: 73, commit: 72, name: "comment", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}}
	var b97 = choiceBuilder{id: 97, commit: 74}
	var b104 = sequenceBuilder{id: 104, commit: 74, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}}
	var b149 = sequenceBuilder{id: 149, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b42 = charBuilder{}
	var b17 = charBuilder{}
	b149.items = []builder{&b42, &b17}
	var b146 = sequenceBuilder{id: 146, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b32 = charBuilder{}
	b146.items = []builder{&b32}
	b104.items = []builder{&b149, &b146}
	var b62 = sequenceBuilder{id: 62, commit: 74, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {1, 1}, {0, -1}, {1, 1}}}
	var b92 = sequenceBuilder{id: 92, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b171 = charBuilder{}
	var b52 = charBuilder{}
	b92.items = []builder{&b171, &b52}
	var b8 = choiceBuilder{id: 8, commit: 10}
	var b154 = sequenceBuilder{id: 154, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b118 = sequenceBuilder{id: 118, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b127 = charBuilder{}
	b118.items = []builder{&b127}
	var b61 = sequenceBuilder{id: 61, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b121 = charBuilder{}
	b61.items = []builder{&b121}
	b154.items = []builder{&b118, &b61}
	var b22 = sequenceBuilder{id: 22, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b155 = charBuilder{}
	b22.items = []builder{&b155}
	b8.options = []builder{&b154, &b22}
	var b53 = sequenceBuilder{id: 53, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b79 = charBuilder{}
	var b158 = charBuilder{}
	b53.items = []builder{&b79, &b158}
	b62.items = []builder{&b92, &b8, &b53}
	b97.options = []builder{&b104, &b62}
	var b80 = sequenceBuilder{id: 80, commit: 10, ranges: [][]int{{0, -1}, {0, 1}, {0, -1}, {1, 1}, {0, -1}, {0, 1}, {0, -1}, {1, 1}}}
	var b135 = choiceBuilder{id: 135, commit: 74}
	var b86 = sequenceBuilder{id: 86, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b105 = charBuilder{}
	b86.items = []builder{&b105}
	var b25 = sequenceBuilder{id: 25, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b35 = charBuilder{}
	b25.items = []builder{&b35}
	var b165 = sequenceBuilder{id: 165, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b93 = charBuilder{}
	b165.items = []builder{&b93}
	var b9 = sequenceBuilder{id: 9, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b98 = charBuilder{}
	b9.items = []builder{&b98}
	var b134 = sequenceBuilder{id: 134, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b147 = charBuilder{}
	b134.items = []builder{&b147}
	var b119 = sequenceBuilder{id: 119, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b87 = charBuilder{}
	b119.items = []builder{&b87}
	b135.options = []builder{&b86, &b25, &b165, &b9, &b134, &b119}
	var b136 = sequenceBuilder{id: 136, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b122 = charBuilder{}
	b136.items = []builder{&b122}
	b80.items = []builder{&b135, &b136, &b135, &b97}
	b73.items = []builder{&b97, &b80}
	b185.options = []builder{&b5, &b73}
	b186.options = []builder{&b185}
	var b187 = sequenceBuilder{id: 187, commit: 66, ranges: [][]int{{0, 1}, {0, -1}, {0, 1}, {0, 1}}}
	var b29 = sequenceBuilder{id: 29, commit: 2, ranges: [][]int{{1, 1}, {0, -1}}}
	var b184 = sequenceBuilder{id: 184, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b140 = charBuilder{}
	b184.items = []builder{&b140}
	var b28 = sequenceBuilder{id: 28, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b28.items = []builder{&b186, &b184}
	b29.items = []builder{&b184, &b28}
	var b59 = sequenceBuilder{id: 59, commit: 66, ranges: [][]int{{1, 1}, {0, 1}}}
	var b172 = sequenceBuilder{id: 172, commit: 64, name: "definition", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}}}
	var b91 = sequenceBuilder{id: 91, commit: 74, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}}
	var b10 = sequenceBuilder{id: 10, commit: 72, name: "symbol", ranges: [][]int{{1, -1}, {1, -1}}}
	var b69 = sequenceBuilder{id: 69, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b100 = charBuilder{}
	b69.items = []builder{&b100}
	b10.items = []builder{&b69}
	var b4 = sequenceBuilder{id: 4, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b66 = sequenceBuilder{id: 66, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b133 = charBuilder{}
	b66.items = []builder{&b133}
	var b20 = choiceBuilder{id: 20, commit: 66}
	var b170 = sequenceBuilder{id: 170, commit: 72, name: "alias", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b169 = charBuilder{}
	var b64 = charBuilder{}
	var b45 = charBuilder{}
	var b111 = charBuilder{}
	var b95 = charBuilder{}
	b170.items = []builder{&b169, &b64, &b45, &b111, &b95}
	var b23 = sequenceBuilder{id: 23, commit: 72, name: "ws", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b101 = charBuilder{}
	var b15 = charBuilder{}
	b23.items = []builder{&b101, &b15}
	var b178 = sequenceBuilder{id: 178, commit: 72, name: "nows", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b48 = charBuilder{}
	var b137 = charBuilder{}
	var b144 = charBuilder{}
	var b65 = charBuilder{}
	b178.items = []builder{&b48, &b137, &b144, &b65}
	var b14 = sequenceBuilder{id: 14, commit: 72, name: "failpass", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b117 = charBuilder{}
	var b124 = charBuilder{}
	var b49 = charBuilder{}
	var b179 = charBuilder{}
	var b56 = charBuilder{}
	var b34 = charBuilder{}
	var b24 = charBuilder{}
	var b96 = charBuilder{}
	b14.items = []builder{&b117, &b124, &b49, &b179, &b56, &b34, &b24, &b96}
	var b50 = sequenceBuilder{id: 50, commit: 72, name: "root", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b152 = charBuilder{}
	var b85 = charBuilder{}
	var b72 = charBuilder{}
	var b78 = charBuilder{}
	b50.items = []builder{&b152, &b85, &b72, &b78}
	b20.options = []builder{&b170, &b23, &b178, &b14, &b50}
	b4.items = []builder{&b66, &b20}
	b91.items = []builder{&b10, &b4}
	var b180 = sequenceBuilder{id: 180, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b145 = charBuilder{}
	b180.items = []builder{&b145}
	var b116 = choiceBuilder{id: 116, commit: 66}
	var b174 = choiceBuilder{id: 174, commit: 66}
	var b130 = sequenceBuilder{id: 130, commit: 72, name: "any-char", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b166 = charBuilder{}
	b130.items = []builder{&b166}
	var b129 = sequenceBuilder{id: 129, commit: 72, name: "char-class", ranges: [][]int{{1, 1}, {0, 1}, {0, -1}, {1, 1}, {1, 1}, {0, 1}, {0, -1}, {1, 1}}}
	var b114 = sequenceBuilder{id: 114, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b82 = charBuilder{}
	b114.items = []builder{&b82}
	var b18 = sequenceBuilder{id: 18, commit: 72, name: "class-not", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b88 = charBuilder{}
	b18.items = []builder{&b88}
	var b182 = choiceBuilder{id: 182, commit: 10}
	var b55 = choiceBuilder{id: 55, commit: 72, name: "class-char"}
	var b36 = sequenceBuilder{id: 36, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b181 = charBuilder{}
	b36.items = []builder{&b181}
	var b6 = sequenceBuilder{id: 6, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b74 = sequenceBuilder{id: 74, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b37 = charBuilder{}
	b74.items = []builder{&b37}
	var b94 = sequenceBuilder{id: 94, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b167 = charBuilder{}
	b94.items = []builder{&b167}
	b6.items = []builder{&b74, &b94}
	b55.options = []builder{&b36, &b6}
	var b12 = sequenceBuilder{id: 12, commit: 72, name: "char-range", ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b75 = sequenceBuilder{id: 75, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b68 = charBuilder{}
	b75.items = []builder{&b68}
	b12.items = []builder{&b55, &b75, &b55}
	b182.options = []builder{&b55, &b12}
	var b128 = sequenceBuilder{id: 128, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b160 = charBuilder{}
	b128.items = []builder{&b160}
	b129.items = []builder{&b114, &b18, &b182, &b128}
	var b156 = sequenceBuilder{id: 156, commit: 72, name: "char-sequence", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {1, 1}, {0, -1}, {1, 1}}}
	var b164 = sequenceBuilder{id: 164, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b163 = charBuilder{}
	b164.items = []builder{&b163}
	var b162 = choiceBuilder{id: 162, commit: 72, name: "sequence-char"}
	var b131 = sequenceBuilder{id: 131, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b159 = charBuilder{}
	b131.items = []builder{&b159}
	var b26 = sequenceBuilder{id: 26, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b7 = sequenceBuilder{id: 7, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b33 = charBuilder{}
	b7.items = []builder{&b33}
	var b161 = sequenceBuilder{id: 161, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b1 = charBuilder{}
	b161.items = []builder{&b1}
	b26.items = []builder{&b7, &b161}
	b162.options = []builder{&b131, &b26}
	var b89 = sequenceBuilder{id: 89, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b99 = charBuilder{}
	b89.items = []builder{&b99}
	b156.items = []builder{&b164, &b162, &b89}
	b174.options = []builder{&b130, &b129, &b156}
	var b90 = sequenceBuilder{id: 90, commit: 66, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}}}
	var b183 = sequenceBuilder{id: 183, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b2 = charBuilder{}
	b183.items = []builder{&b2}
	var b175 = sequenceBuilder{id: 175, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b70 = charBuilder{}
	b175.items = []builder{&b70}
	b90.items = []builder{&b183, &b186, &b116, &b186, &b175}
	var b44 = sequenceBuilder{id: 44, commit: 64, name: "sequence", ranges: [][]int{{1, 1}, {0, -1}}}
	var b39 = sequenceBuilder{id: 39, commit: 72, name: "item", ranges: [][]int{{1, 1}, {0, 1}, {1, 1}, {0, 1}}}
	var b63 = choiceBuilder{id: 63, commit: 10}
	b63.options = []builder{&b174, &b10, &b90}
	var b76 = choiceBuilder{id: 76, commit: 66}
	var b143 = sequenceBuilder{id: 143, commit: 64, name: "count-quantifier", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}}}
	var b13 = sequenceBuilder{id: 13, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b83 = charBuilder{}
	b13.items = []builder{&b83}
	var b106 = sequenceBuilder{id: 106, commit: 64, name: "count", ranges: [][]int{{1, 1}}}
	var b120 = sequenceBuilder{id: 120, commit: 74, ranges: [][]int{{1, -1}, {1, -1}}}
	var b150 = sequenceBuilder{id: 150, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b38 = charBuilder{}
	b150.items = []builder{&b38}
	b120.items = []builder{&b150}
	b106.items = []builder{&b120}
	var b84 = sequenceBuilder{id: 84, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b71 = charBuilder{}
	b84.items = []builder{&b71}
	b143.items = []builder{&b13, &b186, &b106, &b186, &b84}
	var b132 = sequenceBuilder{id: 132, commit: 64, name: "range-quantifier", ranges: [][]int{{1, 1}, {0, -1}, {0, 1}, {0, -1}, {1, 1}, {0, -1}, {0, 1}, {0, -1}, {1, 1}}}
	var b81 = sequenceBuilder{id: 81, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b176 = charBuilder{}
	b81.items = []builder{&b176}
	var b148 = sequenceBuilder{id: 148, commit: 64, name: "range-from", ranges: [][]int{{1, 1}}}
	b148.items = []builder{&b120}
	var b46 = sequenceBuilder{id: 46, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b168 = charBuilder{}
	b46.items = []builder{&b168}
	var b19 = sequenceBuilder{id: 19, commit: 64, name: "range-to", ranges: [][]int{{1, 1}}}
	b19.items = []builder{&b120}
	var b27 = sequenceBuilder{id: 27, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b123 = charBuilder{}
	b27.items = []builder{&b123}
	b132.items = []builder{&b81, &b186, &b148, &b186, &b46, &b186, &b19, &b186, &b27}
	var b47 = sequenceBuilder{id: 47, commit: 72, name: "one-or-more", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b107 = charBuilder{}
	b47.items = []builder{&b107}
	var b157 = sequenceBuilder{id: 157, commit: 72, name: "zero-or-more", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b151 = charBuilder{}
	b157.items = []builder{&b151}
	var b3 = sequenceBuilder{id: 3, commit: 72, name: "zero-or-one", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b108 = charBuilder{}
	b3.items = []builder{&b108}
	b76.options = []builder{&b143, &b132, &b47, &b157, &b3}
	b39.items = []builder{&b63, &b76}
	var b43 = sequenceBuilder{id: 43, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b43.items = []builder{&b186, &b39}
	b44.items = []builder{&b39, &b43}
	var b110 = sequenceBuilder{id: 110, commit: 64, name: "choice", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}}
	var b77 = choiceBuilder{id: 77, commit: 66}
	b77.options = []builder{&b174, &b10, &b90, &b44}
	var b115 = sequenceBuilder{id: 115, commit: 2, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}}}
	var b177 = sequenceBuilder{id: 177, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b11 = charBuilder{}
	b177.items = []builder{&b11}
	b115.items = []builder{&b177, &b186, &b77}
	var b109 = sequenceBuilder{id: 109, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b109.items = []builder{&b186, &b115}
	b110.items = []builder{&b77, &b186, &b115, &b109}
	b116.options = []builder{&b174, &b10, &b90, &b44, &b110}
	b172.items = []builder{&b91, &b186, &b180, &b186, &b116}
	var b58 = sequenceBuilder{id: 58, commit: 2, ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var b139 = sequenceBuilder{id: 139, commit: 2, ranges: [][]int{{1, 1}, {0, -1}, {0, -1}, {1, 1}}}
	var b21 = sequenceBuilder{id: 21, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b16 = charBuilder{}
	b21.items = []builder{&b16}
	var b138 = sequenceBuilder{id: 138, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b138.items = []builder{&b186, &b21}
	b139.items = []builder{&b21, &b138, &b186, &b172}
	var b57 = sequenceBuilder{id: 57, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b57.items = []builder{&b186, &b139}
	b58.items = []builder{&b186, &b139, &b57}
	b59.items = []builder{&b172, &b58}
	var b31 = sequenceBuilder{id: 31, commit: 2, ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var b54 = sequenceBuilder{id: 54, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b67 = charBuilder{}
	b54.items = []builder{&b67}
	var b30 = sequenceBuilder{id: 30, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b30.items = []builder{&b186, &b54}
	b31.items = []builder{&b186, &b54, &b30}
	b187.items = []builder{&b29, &b186, &b59, &b31}
	b188.items = []builder{&b186, &b187, &b186}

	return parseInput(r, &p188, &b188)
}
