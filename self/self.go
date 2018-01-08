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
	var p129 = choiceParser{id: 129, commit: 66, name: "wschar", generalizations: []int{185, 186}}
	var p91 = sequenceParser{id: 91, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{129, 185, 186}}
	var p160 = charParser{id: 160, chars: []rune{32}}
	p91.items = []parser{&p160}
	var p45 = sequenceParser{id: 45, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{129, 185, 186}}
	var p55 = charParser{id: 55, chars: []rune{9}}
	p45.items = []parser{&p55}
	var p166 = sequenceParser{id: 166, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{129, 185, 186}}
	var p16 = charParser{id: 16, chars: []rune{10}}
	p166.items = []parser{&p16}
	var p152 = sequenceParser{id: 152, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{129, 185, 186}}
	var p133 = charParser{id: 133, chars: []rune{8}}
	p152.items = []parser{&p133}
	var p20 = sequenceParser{id: 20, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{129, 185, 186}}
	var p174 = charParser{id: 174, chars: []rune{12}}
	p20.items = []parser{&p174}
	var p71 = sequenceParser{id: 71, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{129, 185, 186}}
	var p23 = charParser{id: 23, chars: []rune{13}}
	p71.items = []parser{&p23}
	var p137 = sequenceParser{id: 137, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{129, 185, 186}}
	var p92 = charParser{id: 92, chars: []rune{11}}
	p137.items = []parser{&p92}
	p129.options = []parser{&p91, &p45, &p166, &p152, &p20, &p71, &p137}
	var p66 = sequenceParser{id: 66, commit: 72, name: "comment", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}, generalizations: []int{185, 186}}
	var p39 = choiceParser{id: 39, commit: 74, name: "comment-segment"}
	var p161 = sequenceParser{id: 161, commit: 74, name: "line-comment", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}, generalizations: []int{39}}
	var p107 = sequenceParser{id: 107, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var p179 = charParser{id: 179, chars: []rune{47}}
	var p118 = charParser{id: 118, chars: []rune{47}}
	p107.items = []parser{&p179, &p118}
	var p110 = sequenceParser{id: 110, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p123 = charParser{id: 123, not: true, chars: []rune{10}}
	p110.items = []parser{&p123}
	p161.items = []parser{&p107, &p110}
	var p24 = sequenceParser{id: 24, commit: 74, name: "block-comment", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {1, 1}, {0, -1}, {1, 1}}, generalizations: []int{39}}
	var p17 = sequenceParser{id: 17, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var p21 = charParser{id: 21, chars: []rune{47}}
	var p87 = charParser{id: 87, chars: []rune{42}}
	p17.items = []parser{&p21, &p87}
	var p142 = choiceParser{id: 142, commit: 10}
	var p56 = sequenceParser{id: 56, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{142}}
	var p46 = sequenceParser{id: 46, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p98 = charParser{id: 98, chars: []rune{42}}
	p46.items = []parser{&p98}
	var p106 = sequenceParser{id: 106, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p167 = charParser{id: 167, not: true, chars: []rune{47}}
	p106.items = []parser{&p167}
	p56.items = []parser{&p46, &p106}
	var p184 = sequenceParser{id: 184, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{142}}
	var p7 = charParser{id: 7, not: true, chars: []rune{42}}
	p184.items = []parser{&p7}
	p142.options = []parser{&p56, &p184}
	var p10 = sequenceParser{id: 10, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var p162 = charParser{id: 162, chars: []rune{42}}
	var p138 = charParser{id: 138, chars: []rune{47}}
	p10.items = []parser{&p162, &p138}
	p24.items = []parser{&p17, &p142, &p10}
	p39.options = []parser{&p161, &p24}
	var p8 = sequenceParser{id: 8, commit: 10, ranges: [][]int{{0, -1}, {0, 1}, {0, -1}, {1, 1}, {0, -1}, {0, 1}, {0, -1}, {1, 1}}}
	var p101 = choiceParser{id: 101, commit: 74, name: "ws-no-nl"}
	var p76 = sequenceParser{id: 76, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{101}}
	var p119 = charParser{id: 119, chars: []rune{32}}
	p76.items = []parser{&p119}
	var p180 = sequenceParser{id: 180, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{101}}
	var p154 = charParser{id: 154, chars: []rune{9}}
	p180.items = []parser{&p154}
	var p177 = sequenceParser{id: 177, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{101}}
	var p130 = charParser{id: 130, chars: []rune{8}}
	p177.items = []parser{&p130}
	var p40 = sequenceParser{id: 40, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{101}}
	var p175 = charParser{id: 175, chars: []rune{12}}
	p40.items = []parser{&p175}
	var p99 = sequenceParser{id: 99, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{101}}
	var p100 = charParser{id: 100, chars: []rune{13}}
	p99.items = []parser{&p100}
	var p28 = sequenceParser{id: 28, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{101}}
	var p41 = charParser{id: 41, chars: []rune{11}}
	p28.items = []parser{&p41}
	p101.options = []parser{&p76, &p180, &p177, &p40, &p99, &p28}
	var p181 = sequenceParser{id: 181, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p65 = charParser{id: 65, chars: []rune{10}}
	p181.items = []parser{&p65}
	p8.items = []parser{&p101, &p181, &p101, &p39}
	p66.items = []parser{&p39, &p8}
	p185.options = []parser{&p129, &p66}
	p186.options = []parser{&p185}
	var p187 = sequenceParser{id: 187, commit: 66, name: "syntax:wsroot", ranges: [][]int{{0, 1}, {0, -1}, {0, 1}, {0, 1}}}
	var p82 = sequenceParser{id: 82, commit: 2, ranges: [][]int{{1, 1}, {0, -1}}}
	var p173 = sequenceParser{id: 173, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p172 = charParser{id: 172, chars: []rune{59}}
	p173.items = []parser{&p172}
	var p81 = sequenceParser{id: 81, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p81.items = []parser{&p186, &p173}
	p82.items = []parser{&p173, &p81}
	var p64 = sequenceParser{id: 64, commit: 66, name: "definitions", ranges: [][]int{{1, 1}, {0, 1}}}
	var p131 = sequenceParser{id: 131, commit: 64, name: "definition", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}}}
	var p159 = sequenceParser{id: 159, commit: 74, name: "definition-name", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}}
	var p44 = sequenceParser{id: 44, commit: 72, name: "symbol", ranges: [][]int{{1, -1}, {1, -1}}, generalizations: []int{136, 29, 59}}
	var p103 = sequenceParser{id: 103, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p78 = charParser{id: 78, not: true, chars: []rune{92, 32, 10, 9, 8, 12, 13, 11, 47, 46, 91, 93, 34, 123, 125, 94, 43, 42, 63, 124, 40, 41, 58, 61, 59}}
	p103.items = []parser{&p78}
	p44.items = []parser{&p103}
	var p165 = sequenceParser{id: 165, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var p94 = sequenceParser{id: 94, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p141 = charParser{id: 141, chars: []rune{58}}
	p94.items = []parser{&p141}
	var p27 = choiceParser{id: 27, commit: 66, name: "flag"}
	var p158 = sequenceParser{id: 158, commit: 72, name: "alias", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{27}}
	var p176 = charParser{id: 176, chars: []rune{97}}
	var p145 = charParser{id: 145, chars: []rune{108}}
	var p60 = charParser{id: 60, chars: []rune{105}}
	var p163 = charParser{id: 163, chars: []rune{97}}
	var p5 = charParser{id: 5, chars: []rune{115}}
	p158.items = []parser{&p176, &p145, &p60, &p163, &p5}
	var p75 = sequenceParser{id: 75, commit: 72, name: "ws", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{27}}
	var p151 = charParser{id: 151, chars: []rune{119}}
	var p74 = charParser{id: 74, chars: []rune{115}}
	p75.items = []parser{&p151, &p74}
	var p121 = sequenceParser{id: 121, commit: 72, name: "nows", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{27}}
	var p111 = charParser{id: 111, chars: []rune{110}}
	var p54 = charParser{id: 54, chars: []rune{111}}
	var p68 = charParser{id: 68, chars: []rune{119}}
	var p127 = charParser{id: 127, chars: []rune{115}}
	p121.items = []parser{&p111, &p54, &p68, &p127}
	var p36 = sequenceParser{id: 36, commit: 72, name: "failpass", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{27}}
	var p6 = charParser{id: 6, chars: []rune{102}}
	var p171 = charParser{id: 171, chars: []rune{97}}
	var p108 = charParser{id: 108, chars: []rune{105}}
	var p32 = charParser{id: 32, chars: []rune{108}}
	var p61 = charParser{id: 61, chars: []rune{112}}
	var p69 = charParser{id: 69, chars: []rune{97}}
	var p49 = charParser{id: 49, chars: []rune{115}}
	var p164 = charParser{id: 164, chars: []rune{115}}
	p36.items = []parser{&p6, &p171, &p108, &p32, &p61, &p69, &p49, &p164}
	var p109 = sequenceParser{id: 109, commit: 72, name: "root", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{27}}
	var p128 = charParser{id: 128, chars: []rune{114}}
	var p9 = charParser{id: 9, chars: []rune{111}}
	var p183 = charParser{id: 183, chars: []rune{111}}
	var p47 = charParser{id: 47, chars: []rune{116}}
	p109.items = []parser{&p128, &p9, &p183, &p47}
	p27.options = []parser{&p158, &p75, &p121, &p36, &p109}
	p165.items = []parser{&p94, &p27}
	p159.items = []parser{&p44, &p165}
	var p122 = sequenceParser{id: 122, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p80 = charParser{id: 80, chars: []rune{61}}
	p122.items = []parser{&p80}
	var p136 = choiceParser{id: 136, commit: 66, name: "expression"}
	var p2 = choiceParser{id: 2, commit: 66, name: "terminal", generalizations: []int{136, 29, 59}}
	var p168 = sequenceParser{id: 168, commit: 72, name: "any-char", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{2, 136, 29, 59}}
	var p125 = charParser{id: 125, chars: []rune{46}}
	p168.items = []parser{&p125}
	var p13 = sequenceParser{id: 13, commit: 72, name: "char-class", ranges: [][]int{{1, 1}, {0, 1}, {0, -1}, {1, 1}, {1, 1}, {0, 1}, {0, -1}, {1, 1}}, generalizations: []int{2, 136, 29, 59}}
	var p134 = sequenceParser{id: 134, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p148 = charParser{id: 148, chars: []rune{91}}
	p134.items = []parser{&p148}
	var p153 = sequenceParser{id: 153, commit: 72, name: "class-not", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p42 = charParser{id: 42, chars: []rune{94}}
	p153.items = []parser{&p42}
	var p77 = choiceParser{id: 77, commit: 10}
	var p155 = choiceParser{id: 155, commit: 72, name: "class-char", generalizations: []int{77}}
	var p85 = sequenceParser{id: 85, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{155, 77}}
	var p135 = charParser{id: 135, not: true, chars: []rune{92, 91, 93, 94, 45}}
	p85.items = []parser{&p135}
	var p169 = sequenceParser{id: 169, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{155, 77}}
	var p12 = sequenceParser{id: 12, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p11 = charParser{id: 11, chars: []rune{92}}
	p12.items = []parser{&p11}
	var p146 = sequenceParser{id: 146, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p31 = charParser{id: 31, not: true}
	p146.items = []parser{&p31}
	p169.items = []parser{&p12, &p146}
	p155.options = []parser{&p85, &p169}
	var p72 = sequenceParser{id: 72, commit: 72, name: "char-range", ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{77}}
	var p50 = sequenceParser{id: 50, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p178 = charParser{id: 178, chars: []rune{45}}
	p50.items = []parser{&p178}
	p72.items = []parser{&p155, &p50, &p155}
	p77.options = []parser{&p155, &p72}
	var p52 = sequenceParser{id: 52, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p37 = charParser{id: 37, chars: []rune{93}}
	p52.items = []parser{&p37}
	p13.items = []parser{&p134, &p153, &p77, &p52}
	var p38 = sequenceParser{id: 38, commit: 72, name: "char-sequence", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {1, 1}, {0, -1}, {1, 1}}, generalizations: []int{2, 136, 29, 59}}
	var p51 = sequenceParser{id: 51, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p86 = charParser{id: 86, chars: []rune{34}}
	p51.items = []parser{&p86}
	var p126 = choiceParser{id: 126, commit: 72, name: "sequence-char"}
	var p88 = sequenceParser{id: 88, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{126}}
	var p1 = charParser{id: 1, not: true, chars: []rune{92, 34}}
	p88.items = []parser{&p1}
	var p102 = sequenceParser{id: 102, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{126}}
	var p67 = sequenceParser{id: 67, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p143 = charParser{id: 143, chars: []rune{92}}
	p67.items = []parser{&p143}
	var p73 = sequenceParser{id: 73, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p43 = charParser{id: 43, not: true}
	p73.items = []parser{&p43}
	p102.items = []parser{&p67, &p73}
	p126.options = []parser{&p88, &p102}
	var p149 = sequenceParser{id: 149, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p22 = charParser{id: 22, chars: []rune{34}}
	p149.items = []parser{&p22}
	p38.items = []parser{&p51, &p126, &p149}
	p2.options = []parser{&p168, &p13, &p38}
	var p14 = sequenceParser{id: 14, commit: 66, name: "group", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}}, generalizations: []int{136, 29, 59}}
	var p79 = sequenceParser{id: 79, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p57 = charParser{id: 57, chars: []rune{40}}
	p79.items = []parser{&p57}
	var p3 = sequenceParser{id: 3, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p124 = charParser{id: 124, chars: []rune{41}}
	p3.items = []parser{&p124}
	p14.items = []parser{&p79, &p186, &p136, &p186, &p3}
	var p26 = sequenceParser{id: 26, commit: 64, name: "sequence", ranges: [][]int{{1, 1}, {0, -1}}, generalizations: []int{136, 59}}
	var p4 = sequenceParser{id: 4, commit: 72, name: "item", ranges: [][]int{{1, 1}, {0, 1}, {1, 1}, {0, 1}}}
	var p29 = choiceParser{id: 29, commit: 10}
	p29.options = []parser{&p2, &p44, &p14}
	var p53 = choiceParser{id: 53, commit: 66, name: "quantity"}
	var p113 = sequenceParser{id: 113, commit: 64, name: "count-quantifier", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}}, generalizations: []int{53}}
	var p182 = sequenceParser{id: 182, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p104 = charParser{id: 104, chars: []rune{123}}
	p182.items = []parser{&p104}
	var p89 = sequenceParser{id: 89, commit: 64, name: "count", ranges: [][]int{{1, 1}}}
	var p112 = sequenceParser{id: 112, commit: 74, name: "number", ranges: [][]int{{1, -1}, {1, -1}}}
	var p150 = sequenceParser{id: 150, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p15 = charParser{id: 15, ranges: [][]rune{{48, 57}}}
	p150.items = []parser{&p15}
	p112.items = []parser{&p150}
	p89.items = []parser{&p112}
	var p18 = sequenceParser{id: 18, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p120 = charParser{id: 120, chars: []rune{125}}
	p18.items = []parser{&p120}
	p113.items = []parser{&p182, &p186, &p89, &p186, &p18}
	var p58 = sequenceParser{id: 58, commit: 64, name: "range-quantifier", ranges: [][]int{{1, 1}, {0, -1}, {0, 1}, {0, -1}, {1, 1}, {0, -1}, {0, 1}, {0, -1}, {1, 1}}, generalizations: []int{53}}
	var p170 = sequenceParser{id: 170, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p95 = charParser{id: 95, chars: []rune{123}}
	p170.items = []parser{&p95}
	var p90 = sequenceParser{id: 90, commit: 64, name: "range-from", ranges: [][]int{{1, 1}}}
	p90.items = []parser{&p112}
	var p139 = sequenceParser{id: 139, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p33 = charParser{id: 33, chars: []rune{44}}
	p139.items = []parser{&p33}
	var p156 = sequenceParser{id: 156, commit: 64, name: "range-to", ranges: [][]int{{1, 1}}}
	p156.items = []parser{&p112}
	var p48 = sequenceParser{id: 48, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p114 = charParser{id: 114, chars: []rune{125}}
	p48.items = []parser{&p114}
	p58.items = []parser{&p170, &p186, &p90, &p186, &p139, &p186, &p156, &p186, &p48}
	var p115 = sequenceParser{id: 115, commit: 72, name: "one-or-more", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{53}}
	var p105 = charParser{id: 105, chars: []rune{43}}
	p115.items = []parser{&p105}
	var p147 = sequenceParser{id: 147, commit: 72, name: "zero-or-more", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{53}}
	var p140 = charParser{id: 140, chars: []rune{42}}
	p147.items = []parser{&p140}
	var p34 = sequenceParser{id: 34, commit: 72, name: "zero-or-one", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{53}}
	var p157 = charParser{id: 157, chars: []rune{63}}
	p34.items = []parser{&p157}
	p53.options = []parser{&p113, &p58, &p115, &p147, &p34}
	p4.items = []parser{&p29, &p53}
	var p25 = sequenceParser{id: 25, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p25.items = []parser{&p186, &p4}
	p26.items = []parser{&p4, &p25}
	var p117 = sequenceParser{id: 117, commit: 64, name: "choice", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}, generalizations: []int{136}}
	var p59 = choiceParser{id: 59, commit: 66, name: "option"}
	p59.options = []parser{&p2, &p44, &p14, &p26}
	var p93 = sequenceParser{id: 93, commit: 2, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}}}
	var p35 = sequenceParser{id: 35, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p144 = charParser{id: 144, chars: []rune{124}}
	p35.items = []parser{&p144}
	p93.items = []parser{&p35, &p186, &p59}
	var p116 = sequenceParser{id: 116, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p116.items = []parser{&p186, &p93}
	p117.items = []parser{&p59, &p186, &p93, &p116}
	p136.options = []parser{&p2, &p44, &p14, &p26, &p117}
	p131.items = []parser{&p159, &p186, &p122, &p186, &p136}
	var p63 = sequenceParser{id: 63, commit: 2, ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var p97 = sequenceParser{id: 97, commit: 2, ranges: [][]int{{1, 1}, {0, -1}, {0, -1}, {1, 1}}}
	var p70 = sequenceParser{id: 70, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p132 = charParser{id: 132, chars: []rune{59}}
	p70.items = []parser{&p132}
	var p96 = sequenceParser{id: 96, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p96.items = []parser{&p186, &p70}
	p97.items = []parser{&p70, &p96, &p186, &p131}
	var p62 = sequenceParser{id: 62, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p62.items = []parser{&p186, &p97}
	p63.items = []parser{&p186, &p97, &p62}
	p64.items = []parser{&p131, &p63}
	var p84 = sequenceParser{id: 84, commit: 2, ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var p30 = sequenceParser{id: 30, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p19 = charParser{id: 19, chars: []rune{59}}
	p30.items = []parser{&p19}
	var p83 = sequenceParser{id: 83, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p83.items = []parser{&p186, &p30}
	p84.items = []parser{&p186, &p30, &p83}
	p187.items = []parser{&p82, &p186, &p64, &p84}
	p188.items = []parser{&p186, &p187, &p186}
	var b188 = sequenceBuilder{id: 188, commit: 32, name: "syntax", ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var b186 = choiceBuilder{id: 186, commit: 2}
	var b185 = choiceBuilder{id: 185, commit: 70}
	var b129 = choiceBuilder{id: 129, commit: 66}
	var b91 = sequenceBuilder{id: 91, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b160 = charBuilder{}
	b91.items = []builder{&b160}
	var b45 = sequenceBuilder{id: 45, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b55 = charBuilder{}
	b45.items = []builder{&b55}
	var b166 = sequenceBuilder{id: 166, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b16 = charBuilder{}
	b166.items = []builder{&b16}
	var b152 = sequenceBuilder{id: 152, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b133 = charBuilder{}
	b152.items = []builder{&b133}
	var b20 = sequenceBuilder{id: 20, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b174 = charBuilder{}
	b20.items = []builder{&b174}
	var b71 = sequenceBuilder{id: 71, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b23 = charBuilder{}
	b71.items = []builder{&b23}
	var b137 = sequenceBuilder{id: 137, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b92 = charBuilder{}
	b137.items = []builder{&b92}
	b129.options = []builder{&b91, &b45, &b166, &b152, &b20, &b71, &b137}
	var b66 = sequenceBuilder{id: 66, commit: 72, name: "comment", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}}
	var b39 = choiceBuilder{id: 39, commit: 74}
	var b161 = sequenceBuilder{id: 161, commit: 74, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}}
	var b107 = sequenceBuilder{id: 107, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b179 = charBuilder{}
	var b118 = charBuilder{}
	b107.items = []builder{&b179, &b118}
	var b110 = sequenceBuilder{id: 110, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b123 = charBuilder{}
	b110.items = []builder{&b123}
	b161.items = []builder{&b107, &b110}
	var b24 = sequenceBuilder{id: 24, commit: 74, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {1, 1}, {0, -1}, {1, 1}}}
	var b17 = sequenceBuilder{id: 17, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b21 = charBuilder{}
	var b87 = charBuilder{}
	b17.items = []builder{&b21, &b87}
	var b142 = choiceBuilder{id: 142, commit: 10}
	var b56 = sequenceBuilder{id: 56, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b46 = sequenceBuilder{id: 46, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b98 = charBuilder{}
	b46.items = []builder{&b98}
	var b106 = sequenceBuilder{id: 106, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b167 = charBuilder{}
	b106.items = []builder{&b167}
	b56.items = []builder{&b46, &b106}
	var b184 = sequenceBuilder{id: 184, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b7 = charBuilder{}
	b184.items = []builder{&b7}
	b142.options = []builder{&b56, &b184}
	var b10 = sequenceBuilder{id: 10, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b162 = charBuilder{}
	var b138 = charBuilder{}
	b10.items = []builder{&b162, &b138}
	b24.items = []builder{&b17, &b142, &b10}
	b39.options = []builder{&b161, &b24}
	var b8 = sequenceBuilder{id: 8, commit: 10, ranges: [][]int{{0, -1}, {0, 1}, {0, -1}, {1, 1}, {0, -1}, {0, 1}, {0, -1}, {1, 1}}}
	var b101 = choiceBuilder{id: 101, commit: 74}
	var b76 = sequenceBuilder{id: 76, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b119 = charBuilder{}
	b76.items = []builder{&b119}
	var b180 = sequenceBuilder{id: 180, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b154 = charBuilder{}
	b180.items = []builder{&b154}
	var b177 = sequenceBuilder{id: 177, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b130 = charBuilder{}
	b177.items = []builder{&b130}
	var b40 = sequenceBuilder{id: 40, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b175 = charBuilder{}
	b40.items = []builder{&b175}
	var b99 = sequenceBuilder{id: 99, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b100 = charBuilder{}
	b99.items = []builder{&b100}
	var b28 = sequenceBuilder{id: 28, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b41 = charBuilder{}
	b28.items = []builder{&b41}
	b101.options = []builder{&b76, &b180, &b177, &b40, &b99, &b28}
	var b181 = sequenceBuilder{id: 181, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b65 = charBuilder{}
	b181.items = []builder{&b65}
	b8.items = []builder{&b101, &b181, &b101, &b39}
	b66.items = []builder{&b39, &b8}
	b185.options = []builder{&b129, &b66}
	b186.options = []builder{&b185}
	var b187 = sequenceBuilder{id: 187, commit: 66, ranges: [][]int{{0, 1}, {0, -1}, {0, 1}, {0, 1}}}
	var b82 = sequenceBuilder{id: 82, commit: 2, ranges: [][]int{{1, 1}, {0, -1}}}
	var b173 = sequenceBuilder{id: 173, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b172 = charBuilder{}
	b173.items = []builder{&b172}
	var b81 = sequenceBuilder{id: 81, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b81.items = []builder{&b186, &b173}
	b82.items = []builder{&b173, &b81}
	var b64 = sequenceBuilder{id: 64, commit: 66, ranges: [][]int{{1, 1}, {0, 1}}}
	var b131 = sequenceBuilder{id: 131, commit: 64, name: "definition", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}}}
	var b159 = sequenceBuilder{id: 159, commit: 74, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}}
	var b44 = sequenceBuilder{id: 44, commit: 72, name: "symbol", ranges: [][]int{{1, -1}, {1, -1}}}
	var b103 = sequenceBuilder{id: 103, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b78 = charBuilder{}
	b103.items = []builder{&b78}
	b44.items = []builder{&b103}
	var b165 = sequenceBuilder{id: 165, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b94 = sequenceBuilder{id: 94, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b141 = charBuilder{}
	b94.items = []builder{&b141}
	var b27 = choiceBuilder{id: 27, commit: 66}
	var b158 = sequenceBuilder{id: 158, commit: 72, name: "alias", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b176 = charBuilder{}
	var b145 = charBuilder{}
	var b60 = charBuilder{}
	var b163 = charBuilder{}
	var b5 = charBuilder{}
	b158.items = []builder{&b176, &b145, &b60, &b163, &b5}
	var b75 = sequenceBuilder{id: 75, commit: 72, name: "ws", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b151 = charBuilder{}
	var b74 = charBuilder{}
	b75.items = []builder{&b151, &b74}
	var b121 = sequenceBuilder{id: 121, commit: 72, name: "nows", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b111 = charBuilder{}
	var b54 = charBuilder{}
	var b68 = charBuilder{}
	var b127 = charBuilder{}
	b121.items = []builder{&b111, &b54, &b68, &b127}
	var b36 = sequenceBuilder{id: 36, commit: 72, name: "failpass", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b6 = charBuilder{}
	var b171 = charBuilder{}
	var b108 = charBuilder{}
	var b32 = charBuilder{}
	var b61 = charBuilder{}
	var b69 = charBuilder{}
	var b49 = charBuilder{}
	var b164 = charBuilder{}
	b36.items = []builder{&b6, &b171, &b108, &b32, &b61, &b69, &b49, &b164}
	var b109 = sequenceBuilder{id: 109, commit: 72, name: "root", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b128 = charBuilder{}
	var b9 = charBuilder{}
	var b183 = charBuilder{}
	var b47 = charBuilder{}
	b109.items = []builder{&b128, &b9, &b183, &b47}
	b27.options = []builder{&b158, &b75, &b121, &b36, &b109}
	b165.items = []builder{&b94, &b27}
	b159.items = []builder{&b44, &b165}
	var b122 = sequenceBuilder{id: 122, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b80 = charBuilder{}
	b122.items = []builder{&b80}
	var b136 = choiceBuilder{id: 136, commit: 66}
	var b2 = choiceBuilder{id: 2, commit: 66}
	var b168 = sequenceBuilder{id: 168, commit: 72, name: "any-char", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b125 = charBuilder{}
	b168.items = []builder{&b125}
	var b13 = sequenceBuilder{id: 13, commit: 72, name: "char-class", ranges: [][]int{{1, 1}, {0, 1}, {0, -1}, {1, 1}, {1, 1}, {0, 1}, {0, -1}, {1, 1}}}
	var b134 = sequenceBuilder{id: 134, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b148 = charBuilder{}
	b134.items = []builder{&b148}
	var b153 = sequenceBuilder{id: 153, commit: 72, name: "class-not", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b42 = charBuilder{}
	b153.items = []builder{&b42}
	var b77 = choiceBuilder{id: 77, commit: 10}
	var b155 = choiceBuilder{id: 155, commit: 72, name: "class-char"}
	var b85 = sequenceBuilder{id: 85, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b135 = charBuilder{}
	b85.items = []builder{&b135}
	var b169 = sequenceBuilder{id: 169, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b12 = sequenceBuilder{id: 12, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b11 = charBuilder{}
	b12.items = []builder{&b11}
	var b146 = sequenceBuilder{id: 146, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b31 = charBuilder{}
	b146.items = []builder{&b31}
	b169.items = []builder{&b12, &b146}
	b155.options = []builder{&b85, &b169}
	var b72 = sequenceBuilder{id: 72, commit: 72, name: "char-range", ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b50 = sequenceBuilder{id: 50, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b178 = charBuilder{}
	b50.items = []builder{&b178}
	b72.items = []builder{&b155, &b50, &b155}
	b77.options = []builder{&b155, &b72}
	var b52 = sequenceBuilder{id: 52, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b37 = charBuilder{}
	b52.items = []builder{&b37}
	b13.items = []builder{&b134, &b153, &b77, &b52}
	var b38 = sequenceBuilder{id: 38, commit: 72, name: "char-sequence", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {1, 1}, {0, -1}, {1, 1}}}
	var b51 = sequenceBuilder{id: 51, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b86 = charBuilder{}
	b51.items = []builder{&b86}
	var b126 = choiceBuilder{id: 126, commit: 72, name: "sequence-char"}
	var b88 = sequenceBuilder{id: 88, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b1 = charBuilder{}
	b88.items = []builder{&b1}
	var b102 = sequenceBuilder{id: 102, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b67 = sequenceBuilder{id: 67, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b143 = charBuilder{}
	b67.items = []builder{&b143}
	var b73 = sequenceBuilder{id: 73, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b43 = charBuilder{}
	b73.items = []builder{&b43}
	b102.items = []builder{&b67, &b73}
	b126.options = []builder{&b88, &b102}
	var b149 = sequenceBuilder{id: 149, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b22 = charBuilder{}
	b149.items = []builder{&b22}
	b38.items = []builder{&b51, &b126, &b149}
	b2.options = []builder{&b168, &b13, &b38}
	var b14 = sequenceBuilder{id: 14, commit: 66, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}}}
	var b79 = sequenceBuilder{id: 79, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b57 = charBuilder{}
	b79.items = []builder{&b57}
	var b3 = sequenceBuilder{id: 3, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b124 = charBuilder{}
	b3.items = []builder{&b124}
	b14.items = []builder{&b79, &b186, &b136, &b186, &b3}
	var b26 = sequenceBuilder{id: 26, commit: 64, name: "sequence", ranges: [][]int{{1, 1}, {0, -1}}}
	var b4 = sequenceBuilder{id: 4, commit: 72, name: "item", ranges: [][]int{{1, 1}, {0, 1}, {1, 1}, {0, 1}}}
	var b29 = choiceBuilder{id: 29, commit: 10}
	b29.options = []builder{&b2, &b44, &b14}
	var b53 = choiceBuilder{id: 53, commit: 66}
	var b113 = sequenceBuilder{id: 113, commit: 64, name: "count-quantifier", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}}}
	var b182 = sequenceBuilder{id: 182, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b104 = charBuilder{}
	b182.items = []builder{&b104}
	var b89 = sequenceBuilder{id: 89, commit: 64, name: "count", ranges: [][]int{{1, 1}}}
	var b112 = sequenceBuilder{id: 112, commit: 74, ranges: [][]int{{1, -1}, {1, -1}}}
	var b150 = sequenceBuilder{id: 150, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b15 = charBuilder{}
	b150.items = []builder{&b15}
	b112.items = []builder{&b150}
	b89.items = []builder{&b112}
	var b18 = sequenceBuilder{id: 18, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b120 = charBuilder{}
	b18.items = []builder{&b120}
	b113.items = []builder{&b182, &b186, &b89, &b186, &b18}
	var b58 = sequenceBuilder{id: 58, commit: 64, name: "range-quantifier", ranges: [][]int{{1, 1}, {0, -1}, {0, 1}, {0, -1}, {1, 1}, {0, -1}, {0, 1}, {0, -1}, {1, 1}}}
	var b170 = sequenceBuilder{id: 170, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b95 = charBuilder{}
	b170.items = []builder{&b95}
	var b90 = sequenceBuilder{id: 90, commit: 64, name: "range-from", ranges: [][]int{{1, 1}}}
	b90.items = []builder{&b112}
	var b139 = sequenceBuilder{id: 139, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b33 = charBuilder{}
	b139.items = []builder{&b33}
	var b156 = sequenceBuilder{id: 156, commit: 64, name: "range-to", ranges: [][]int{{1, 1}}}
	b156.items = []builder{&b112}
	var b48 = sequenceBuilder{id: 48, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b114 = charBuilder{}
	b48.items = []builder{&b114}
	b58.items = []builder{&b170, &b186, &b90, &b186, &b139, &b186, &b156, &b186, &b48}
	var b115 = sequenceBuilder{id: 115, commit: 72, name: "one-or-more", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b105 = charBuilder{}
	b115.items = []builder{&b105}
	var b147 = sequenceBuilder{id: 147, commit: 72, name: "zero-or-more", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b140 = charBuilder{}
	b147.items = []builder{&b140}
	var b34 = sequenceBuilder{id: 34, commit: 72, name: "zero-or-one", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b157 = charBuilder{}
	b34.items = []builder{&b157}
	b53.options = []builder{&b113, &b58, &b115, &b147, &b34}
	b4.items = []builder{&b29, &b53}
	var b25 = sequenceBuilder{id: 25, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b25.items = []builder{&b186, &b4}
	b26.items = []builder{&b4, &b25}
	var b117 = sequenceBuilder{id: 117, commit: 64, name: "choice", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}}
	var b59 = choiceBuilder{id: 59, commit: 66}
	b59.options = []builder{&b2, &b44, &b14, &b26}
	var b93 = sequenceBuilder{id: 93, commit: 2, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}}}
	var b35 = sequenceBuilder{id: 35, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b144 = charBuilder{}
	b35.items = []builder{&b144}
	b93.items = []builder{&b35, &b186, &b59}
	var b116 = sequenceBuilder{id: 116, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b116.items = []builder{&b186, &b93}
	b117.items = []builder{&b59, &b186, &b93, &b116}
	b136.options = []builder{&b2, &b44, &b14, &b26, &b117}
	b131.items = []builder{&b159, &b186, &b122, &b186, &b136}
	var b63 = sequenceBuilder{id: 63, commit: 2, ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var b97 = sequenceBuilder{id: 97, commit: 2, ranges: [][]int{{1, 1}, {0, -1}, {0, -1}, {1, 1}}}
	var b70 = sequenceBuilder{id: 70, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b132 = charBuilder{}
	b70.items = []builder{&b132}
	var b96 = sequenceBuilder{id: 96, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b96.items = []builder{&b186, &b70}
	b97.items = []builder{&b70, &b96, &b186, &b131}
	var b62 = sequenceBuilder{id: 62, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b62.items = []builder{&b186, &b97}
	b63.items = []builder{&b186, &b97, &b62}
	b64.items = []builder{&b131, &b63}
	var b84 = sequenceBuilder{id: 84, commit: 2, ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var b30 = sequenceBuilder{id: 30, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b19 = charBuilder{}
	b30.items = []builder{&b19}
	var b83 = sequenceBuilder{id: 83, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b83.items = []builder{&b186, &b30}
	b84.items = []builder{&b186, &b30, &b83}
	b187.items = []builder{&b82, &b186, &b64, &b84}
	b188.items = []builder{&b186, &b187, &b186}

	return parseInput(r, &p188, &b188)
}
