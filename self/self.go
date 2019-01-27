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
	name            string
	id              int
	commit          CommitType
	items           []builder
	ranges          [][]int
	generalizations []int
	allChars        bool
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
		for _, g := range b.generalizations {
			c.results.dropMatchTo(c.offset, g, to)
		}
	} else {
		if c.results.pending(c.offset, b.id) {
			return nil, false
		}
		c.results.markPending(c.offset, b.id)
		for _, g := range b.generalizations {
			c.results.markPending(c.offset, g)
		}
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
		for _, g := range b.generalizations {
			c.results.unmarkPending(from, g)
		}
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
	name            string
	id              int
	commit          CommitType
	options         []builder
	generalizations []int
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
		for _, g := range b.generalizations {
			c.results.dropMatchTo(c.offset, g, to)
		}
	} else {
		if c.results.pending(c.offset, b.id) {
			return nil, false
		}
		c.results.markPending(c.offset, b.id)
		for _, g := range b.generalizations {
			c.results.markPending(c.offset, g)
		}
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
		for _, g := range b.generalizations {
			c.results.unmarkPending(from, g)
		}
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
	var p15 = choiceParser{id: 15, commit: 66, name: "wschar", generalizations: []int{185, 186}}
	var p2 = sequenceParser{id: 2, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{15, 185, 186}}
	var p1 = charParser{id: 1, chars: []rune{32}}
	p2.items = []parser{&p1}
	var p4 = sequenceParser{id: 4, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{15, 185, 186}}
	var p3 = charParser{id: 3, chars: []rune{9}}
	p4.items = []parser{&p3}
	var p6 = sequenceParser{id: 6, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{15, 185, 186}}
	var p5 = charParser{id: 5, chars: []rune{10}}
	p6.items = []parser{&p5}
	var p8 = sequenceParser{id: 8, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{15, 185, 186}}
	var p7 = charParser{id: 7, chars: []rune{8}}
	p8.items = []parser{&p7}
	var p10 = sequenceParser{id: 10, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{15, 185, 186}}
	var p9 = charParser{id: 9, chars: []rune{12}}
	p10.items = []parser{&p9}
	var p12 = sequenceParser{id: 12, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{15, 185, 186}}
	var p11 = charParser{id: 11, chars: []rune{13}}
	p12.items = []parser{&p11}
	var p14 = sequenceParser{id: 14, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{15, 185, 186}}
	var p13 = charParser{id: 13, chars: []rune{11}}
	p14.items = []parser{&p13}
	p15.options = []parser{&p2, &p4, &p6, &p8, &p10, &p12, &p14}
	var p54 = sequenceParser{id: 54, commit: 72, name: "comment", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}, generalizations: []int{185, 186}}
	var p37 = choiceParser{id: 37, commit: 74, name: "comment-segment"}
	var p36 = sequenceParser{id: 36, commit: 74, name: "line-comment", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}, generalizations: []int{37}}
	var p33 = sequenceParser{id: 33, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var p31 = charParser{id: 31, chars: []rune{47}}
	var p32 = charParser{id: 32, chars: []rune{47}}
	p33.items = []parser{&p31, &p32}
	var p35 = sequenceParser{id: 35, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p34 = charParser{id: 34, not: true, chars: []rune{10}}
	p35.items = []parser{&p34}
	p36.items = []parser{&p33, &p35}
	var p30 = sequenceParser{id: 30, commit: 74, name: "block-comment", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {1, 1}, {0, -1}, {1, 1}}, generalizations: []int{37}}
	var p18 = sequenceParser{id: 18, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var p16 = charParser{id: 16, chars: []rune{47}}
	var p17 = charParser{id: 17, chars: []rune{42}}
	p18.items = []parser{&p16, &p17}
	var p26 = choiceParser{id: 26, commit: 10}
	var p23 = sequenceParser{id: 23, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{26}}
	var p20 = sequenceParser{id: 20, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p19 = charParser{id: 19, chars: []rune{42}}
	p20.items = []parser{&p19}
	var p22 = sequenceParser{id: 22, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p21 = charParser{id: 21, not: true, chars: []rune{47}}
	p22.items = []parser{&p21}
	p23.items = []parser{&p20, &p22}
	var p25 = sequenceParser{id: 25, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{26}}
	var p24 = charParser{id: 24, not: true, chars: []rune{42}}
	p25.items = []parser{&p24}
	p26.options = []parser{&p23, &p25}
	var p29 = sequenceParser{id: 29, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var p27 = charParser{id: 27, chars: []rune{42}}
	var p28 = charParser{id: 28, chars: []rune{47}}
	p29.items = []parser{&p27, &p28}
	p30.items = []parser{&p18, &p26, &p29}
	p37.options = []parser{&p36, &p30}
	var p53 = sequenceParser{id: 53, commit: 10, ranges: [][]int{{0, -1}, {0, 1}, {0, -1}, {1, 1}, {0, -1}, {0, 1}, {0, -1}, {1, 1}}}
	var p50 = choiceParser{id: 50, commit: 74, name: "ws-no-nl"}
	var p39 = sequenceParser{id: 39, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{50}}
	var p38 = charParser{id: 38, chars: []rune{32}}
	p39.items = []parser{&p38}
	var p41 = sequenceParser{id: 41, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{50}}
	var p40 = charParser{id: 40, chars: []rune{9}}
	p41.items = []parser{&p40}
	var p43 = sequenceParser{id: 43, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{50}}
	var p42 = charParser{id: 42, chars: []rune{8}}
	p43.items = []parser{&p42}
	var p45 = sequenceParser{id: 45, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{50}}
	var p44 = charParser{id: 44, chars: []rune{12}}
	p45.items = []parser{&p44}
	var p47 = sequenceParser{id: 47, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{50}}
	var p46 = charParser{id: 46, chars: []rune{13}}
	p47.items = []parser{&p46}
	var p49 = sequenceParser{id: 49, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{50}}
	var p48 = charParser{id: 48, chars: []rune{11}}
	p49.items = []parser{&p48}
	p50.options = []parser{&p39, &p41, &p43, &p45, &p47, &p49}
	var p52 = sequenceParser{id: 52, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p51 = charParser{id: 51, chars: []rune{10}}
	p52.items = []parser{&p51}
	p53.items = []parser{&p50, &p52, &p50, &p37}
	p54.items = []parser{&p37, &p53}
	p185.options = []parser{&p15, &p54}
	p186.options = []parser{&p185}
	var p187 = sequenceParser{id: 187, commit: 66, name: "syntax:wsroot", ranges: [][]int{{0, 1}, {0, -1}, {0, 1}, {0, 1}}}
	var p182 = sequenceParser{id: 182, commit: 2, ranges: [][]int{{1, 1}, {0, -1}}}
	var p178 = sequenceParser{id: 178, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p177 = charParser{id: 177, chars: []rune{59}}
	p178.items = []parser{&p177}
	var p181 = sequenceParser{id: 181, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p181.items = []parser{&p186, &p178}
	p182.items = []parser{&p178, &p181}
	var p176 = sequenceParser{id: 176, commit: 66, name: "definitions", ranges: [][]int{{1, 1}, {0, 1}}}
	var p169 = sequenceParser{id: 169, commit: 64, name: "definition", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}}}
	var p166 = sequenceParser{id: 166, commit: 74, name: "definition-name", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}}
	var p92 = sequenceParser{id: 92, commit: 72, name: "symbol", ranges: [][]int{{1, -1}, {1, -1}}, generalizations: []int{133, 123, 127}}
	var p91 = sequenceParser{id: 91, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p90 = charParser{id: 90, not: true, chars: []rune{92, 32, 10, 9, 8, 12, 13, 11, 47, 46, 91, 93, 34, 123, 125, 94, 43, 42, 63, 124, 40, 41, 58, 61, 59}}
	p91.items = []parser{&p90}
	p92.items = []parser{&p91}
	var p165 = sequenceParser{id: 165, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var p164 = sequenceParser{id: 164, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p163 = charParser{id: 163, chars: []rune{58}}
	p164.items = []parser{&p163}
	var p162 = choiceParser{id: 162, commit: 66, name: "flag"}
	var p139 = sequenceParser{id: 139, commit: 72, name: "alias", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{162}}
	var p134 = charParser{id: 134, chars: []rune{97}}
	var p135 = charParser{id: 135, chars: []rune{108}}
	var p136 = charParser{id: 136, chars: []rune{105}}
	var p137 = charParser{id: 137, chars: []rune{97}}
	var p138 = charParser{id: 138, chars: []rune{115}}
	p139.items = []parser{&p134, &p135, &p136, &p137, &p138}
	var p142 = sequenceParser{id: 142, commit: 72, name: "ws", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{162}}
	var p140 = charParser{id: 140, chars: []rune{119}}
	var p141 = charParser{id: 141, chars: []rune{115}}
	p142.items = []parser{&p140, &p141}
	var p147 = sequenceParser{id: 147, commit: 72, name: "nows", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{162}}
	var p143 = charParser{id: 143, chars: []rune{110}}
	var p144 = charParser{id: 144, chars: []rune{111}}
	var p145 = charParser{id: 145, chars: []rune{119}}
	var p146 = charParser{id: 146, chars: []rune{115}}
	p147.items = []parser{&p143, &p144, &p145, &p146}
	var p156 = sequenceParser{id: 156, commit: 72, name: "failpass", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{162}}
	var p148 = charParser{id: 148, chars: []rune{102}}
	var p149 = charParser{id: 149, chars: []rune{97}}
	var p150 = charParser{id: 150, chars: []rune{105}}
	var p151 = charParser{id: 151, chars: []rune{108}}
	var p152 = charParser{id: 152, chars: []rune{112}}
	var p153 = charParser{id: 153, chars: []rune{97}}
	var p154 = charParser{id: 154, chars: []rune{115}}
	var p155 = charParser{id: 155, chars: []rune{115}}
	p156.items = []parser{&p148, &p149, &p150, &p151, &p152, &p153, &p154, &p155}
	var p161 = sequenceParser{id: 161, commit: 72, name: "root", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{162}}
	var p157 = charParser{id: 157, chars: []rune{114}}
	var p158 = charParser{id: 158, chars: []rune{111}}
	var p159 = charParser{id: 159, chars: []rune{111}}
	var p160 = charParser{id: 160, chars: []rune{116}}
	p161.items = []parser{&p157, &p158, &p159, &p160}
	p162.options = []parser{&p139, &p142, &p147, &p156, &p161}
	p165.items = []parser{&p164, &p162}
	p166.items = []parser{&p92, &p165}
	var p168 = sequenceParser{id: 168, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p167 = charParser{id: 167, chars: []rune{61}}
	p168.items = []parser{&p167}
	var p133 = choiceParser{id: 133, commit: 66, name: "expression"}
	var p89 = choiceParser{id: 89, commit: 66, name: "terminal", generalizations: []int{133, 123, 127}}
	var p56 = sequenceParser{id: 56, commit: 72, name: "any-char", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{89, 133, 123, 127}}
	var p55 = charParser{id: 55, chars: []rune{46}}
	p56.items = []parser{&p55}
	var p75 = sequenceParser{id: 75, commit: 72, name: "char-class", ranges: [][]int{{1, 1}, {0, 1}, {0, -1}, {1, 1}, {1, 1}, {0, 1}, {0, -1}, {1, 1}}, generalizations: []int{89, 133, 123, 127}}
	var p71 = sequenceParser{id: 71, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p70 = charParser{id: 70, chars: []rune{91}}
	p71.items = []parser{&p70}
	var p58 = sequenceParser{id: 58, commit: 72, name: "class-not", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p57 = charParser{id: 57, chars: []rune{94}}
	p58.items = []parser{&p57}
	var p72 = choiceParser{id: 72, commit: 10}
	var p66 = choiceParser{id: 66, commit: 72, name: "class-char", generalizations: []int{72}}
	var p60 = sequenceParser{id: 60, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{66, 72}}
	var p59 = charParser{id: 59, not: true, chars: []rune{92, 91, 93, 94, 45}}
	p60.items = []parser{&p59}
	var p65 = sequenceParser{id: 65, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{66, 72}}
	var p62 = sequenceParser{id: 62, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p61 = charParser{id: 61, chars: []rune{92}}
	p62.items = []parser{&p61}
	var p64 = sequenceParser{id: 64, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p63 = charParser{id: 63, not: true}
	p64.items = []parser{&p63}
	p65.items = []parser{&p62, &p64}
	p66.options = []parser{&p60, &p65}
	var p69 = sequenceParser{id: 69, commit: 72, name: "char-range", ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{72}}
	var p68 = sequenceParser{id: 68, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p67 = charParser{id: 67, chars: []rune{45}}
	p68.items = []parser{&p67}
	p69.items = []parser{&p66, &p68, &p66}
	p72.options = []parser{&p66, &p69}
	var p74 = sequenceParser{id: 74, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p73 = charParser{id: 73, chars: []rune{93}}
	p74.items = []parser{&p73}
	p75.items = []parser{&p71, &p58, &p72, &p74}
	var p88 = sequenceParser{id: 88, commit: 72, name: "char-sequence", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {1, 1}, {0, -1}, {1, 1}}, generalizations: []int{89, 133, 123, 127}}
	var p85 = sequenceParser{id: 85, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p84 = charParser{id: 84, chars: []rune{34}}
	p85.items = []parser{&p84}
	var p83 = choiceParser{id: 83, commit: 72, name: "sequence-char"}
	var p77 = sequenceParser{id: 77, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{83}}
	var p76 = charParser{id: 76, not: true, chars: []rune{92, 34}}
	p77.items = []parser{&p76}
	var p82 = sequenceParser{id: 82, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{83}}
	var p79 = sequenceParser{id: 79, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p78 = charParser{id: 78, chars: []rune{92}}
	p79.items = []parser{&p78}
	var p81 = sequenceParser{id: 81, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p80 = charParser{id: 80, not: true}
	p81.items = []parser{&p80}
	p82.items = []parser{&p79, &p81}
	p83.options = []parser{&p77, &p82}
	var p87 = sequenceParser{id: 87, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p86 = charParser{id: 86, chars: []rune{34}}
	p87.items = []parser{&p86}
	p88.items = []parser{&p85, &p83, &p87}
	p89.options = []parser{&p56, &p75, &p88}
	var p97 = sequenceParser{id: 97, commit: 66, name: "group", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}}, generalizations: []int{133, 123, 127}}
	var p94 = sequenceParser{id: 94, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p93 = charParser{id: 93, chars: []rune{40}}
	p94.items = []parser{&p93}
	var p96 = sequenceParser{id: 96, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p95 = charParser{id: 95, chars: []rune{41}}
	p96.items = []parser{&p95}
	p97.items = []parser{&p94, &p186, &p133, &p186, &p96}
	var p126 = sequenceParser{id: 126, commit: 64, name: "sequence", ranges: [][]int{{1, 1}, {0, -1}}, generalizations: []int{133, 127}}
	var p124 = sequenceParser{id: 124, commit: 72, name: "item", ranges: [][]int{{1, 1}, {0, 1}, {1, 1}, {0, 1}}}
	var p123 = choiceParser{id: 123, commit: 10}
	p123.options = []parser{&p89, &p92, &p97}
	var p122 = choiceParser{id: 122, commit: 66, name: "quantity"}
	var p106 = sequenceParser{id: 106, commit: 64, name: "count-quantifier", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}}, generalizations: []int{122}}
	var p103 = sequenceParser{id: 103, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p102 = charParser{id: 102, chars: []rune{123}}
	p103.items = []parser{&p102}
	var p101 = sequenceParser{id: 101, commit: 64, name: "count", ranges: [][]int{{1, 1}}}
	var p100 = sequenceParser{id: 100, commit: 74, name: "number", ranges: [][]int{{1, -1}, {1, -1}}}
	var p99 = sequenceParser{id: 99, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p98 = charParser{id: 98, ranges: [][]rune{{48, 57}}}
	p99.items = []parser{&p98}
	p100.items = []parser{&p99}
	p101.items = []parser{&p100}
	var p105 = sequenceParser{id: 105, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p104 = charParser{id: 104, chars: []rune{125}}
	p105.items = []parser{&p104}
	p106.items = []parser{&p103, &p186, &p101, &p186, &p105}
	var p115 = sequenceParser{id: 115, commit: 64, name: "range-quantifier", ranges: [][]int{{1, 1}, {0, -1}, {0, 1}, {0, -1}, {1, 1}, {0, -1}, {0, 1}, {0, -1}, {1, 1}}, generalizations: []int{122}}
	var p110 = sequenceParser{id: 110, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p109 = charParser{id: 109, chars: []rune{123}}
	p110.items = []parser{&p109}
	var p107 = sequenceParser{id: 107, commit: 64, name: "range-from", ranges: [][]int{{1, 1}}}
	p107.items = []parser{&p100}
	var p112 = sequenceParser{id: 112, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p111 = charParser{id: 111, chars: []rune{44}}
	p112.items = []parser{&p111}
	var p108 = sequenceParser{id: 108, commit: 64, name: "range-to", ranges: [][]int{{1, 1}}}
	p108.items = []parser{&p100}
	var p114 = sequenceParser{id: 114, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p113 = charParser{id: 113, chars: []rune{125}}
	p114.items = []parser{&p113}
	p115.items = []parser{&p110, &p186, &p107, &p186, &p112, &p186, &p108, &p186, &p114}
	var p117 = sequenceParser{id: 117, commit: 72, name: "one-or-more", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{122}}
	var p116 = charParser{id: 116, chars: []rune{43}}
	p117.items = []parser{&p116}
	var p119 = sequenceParser{id: 119, commit: 72, name: "zero-or-more", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{122}}
	var p118 = charParser{id: 118, chars: []rune{42}}
	p119.items = []parser{&p118}
	var p121 = sequenceParser{id: 121, commit: 72, name: "zero-or-one", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{122}}
	var p120 = charParser{id: 120, chars: []rune{63}}
	p121.items = []parser{&p120}
	p122.options = []parser{&p106, &p115, &p117, &p119, &p121}
	p124.items = []parser{&p123, &p122}
	var p125 = sequenceParser{id: 125, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p125.items = []parser{&p186, &p124}
	p126.items = []parser{&p124, &p125}
	var p132 = sequenceParser{id: 132, commit: 64, name: "choice", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}, generalizations: []int{133}}
	var p127 = choiceParser{id: 127, commit: 66, name: "option"}
	p127.options = []parser{&p89, &p92, &p97, &p126}
	var p130 = sequenceParser{id: 130, commit: 2, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}}}
	var p129 = sequenceParser{id: 129, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p128 = charParser{id: 128, chars: []rune{124}}
	p129.items = []parser{&p128}
	p130.items = []parser{&p129, &p186, &p127}
	var p131 = sequenceParser{id: 131, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p131.items = []parser{&p186, &p130}
	p132.items = []parser{&p127, &p186, &p130, &p131}
	p133.options = []parser{&p89, &p92, &p97, &p126, &p132}
	p169.items = []parser{&p166, &p186, &p168, &p186, &p133}
	var p175 = sequenceParser{id: 175, commit: 2, ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var p173 = sequenceParser{id: 173, commit: 2, ranges: [][]int{{1, 1}, {0, -1}, {0, -1}, {1, 1}}}
	var p171 = sequenceParser{id: 171, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p170 = charParser{id: 170, chars: []rune{59}}
	p171.items = []parser{&p170}
	var p172 = sequenceParser{id: 172, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p172.items = []parser{&p186, &p171}
	p173.items = []parser{&p171, &p172, &p186, &p169}
	var p174 = sequenceParser{id: 174, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p174.items = []parser{&p186, &p173}
	p175.items = []parser{&p186, &p173, &p174}
	p176.items = []parser{&p169, &p175}
	var p184 = sequenceParser{id: 184, commit: 2, ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var p180 = sequenceParser{id: 180, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p179 = charParser{id: 179, chars: []rune{59}}
	p180.items = []parser{&p179}
	var p183 = sequenceParser{id: 183, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p183.items = []parser{&p186, &p180}
	p184.items = []parser{&p186, &p180, &p183}
	p187.items = []parser{&p182, &p186, &p176, &p184}
	p188.items = []parser{&p186, &p187, &p186}
	var b188 = sequenceBuilder{id: 188, commit: 32, name: "syntax", ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var b186 = choiceBuilder{id: 186, commit: 2}
	var b185 = choiceBuilder{id: 185, commit: 70, generalizations: []int{186}}
	var b15 = choiceBuilder{id: 15, commit: 66, generalizations: []int{185, 186}}
	var b2 = sequenceBuilder{id: 2, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{15, 185, 186}}
	var b1 = charBuilder{}
	b2.items = []builder{&b1}
	var b4 = sequenceBuilder{id: 4, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{15, 185, 186}}
	var b3 = charBuilder{}
	b4.items = []builder{&b3}
	var b6 = sequenceBuilder{id: 6, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{15, 185, 186}}
	var b5 = charBuilder{}
	b6.items = []builder{&b5}
	var b8 = sequenceBuilder{id: 8, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{15, 185, 186}}
	var b7 = charBuilder{}
	b8.items = []builder{&b7}
	var b10 = sequenceBuilder{id: 10, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{15, 185, 186}}
	var b9 = charBuilder{}
	b10.items = []builder{&b9}
	var b12 = sequenceBuilder{id: 12, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{15, 185, 186}}
	var b11 = charBuilder{}
	b12.items = []builder{&b11}
	var b14 = sequenceBuilder{id: 14, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{15, 185, 186}}
	var b13 = charBuilder{}
	b14.items = []builder{&b13}
	b15.options = []builder{&b2, &b4, &b6, &b8, &b10, &b12, &b14}
	var b54 = sequenceBuilder{id: 54, commit: 72, name: "comment", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}, generalizations: []int{185, 186}}
	var b37 = choiceBuilder{id: 37, commit: 74}
	var b36 = sequenceBuilder{id: 36, commit: 74, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}, generalizations: []int{37}}
	var b33 = sequenceBuilder{id: 33, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b31 = charBuilder{}
	var b32 = charBuilder{}
	b33.items = []builder{&b31, &b32}
	var b35 = sequenceBuilder{id: 35, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b34 = charBuilder{}
	b35.items = []builder{&b34}
	b36.items = []builder{&b33, &b35}
	var b30 = sequenceBuilder{id: 30, commit: 74, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {1, 1}, {0, -1}, {1, 1}}, generalizations: []int{37}}
	var b18 = sequenceBuilder{id: 18, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b16 = charBuilder{}
	var b17 = charBuilder{}
	b18.items = []builder{&b16, &b17}
	var b26 = choiceBuilder{id: 26, commit: 10}
	var b23 = sequenceBuilder{id: 23, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{26}}
	var b20 = sequenceBuilder{id: 20, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b19 = charBuilder{}
	b20.items = []builder{&b19}
	var b22 = sequenceBuilder{id: 22, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b21 = charBuilder{}
	b22.items = []builder{&b21}
	b23.items = []builder{&b20, &b22}
	var b25 = sequenceBuilder{id: 25, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{26}}
	var b24 = charBuilder{}
	b25.items = []builder{&b24}
	b26.options = []builder{&b23, &b25}
	var b29 = sequenceBuilder{id: 29, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b27 = charBuilder{}
	var b28 = charBuilder{}
	b29.items = []builder{&b27, &b28}
	b30.items = []builder{&b18, &b26, &b29}
	b37.options = []builder{&b36, &b30}
	var b53 = sequenceBuilder{id: 53, commit: 10, ranges: [][]int{{0, -1}, {0, 1}, {0, -1}, {1, 1}, {0, -1}, {0, 1}, {0, -1}, {1, 1}}}
	var b50 = choiceBuilder{id: 50, commit: 74}
	var b39 = sequenceBuilder{id: 39, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{50}}
	var b38 = charBuilder{}
	b39.items = []builder{&b38}
	var b41 = sequenceBuilder{id: 41, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{50}}
	var b40 = charBuilder{}
	b41.items = []builder{&b40}
	var b43 = sequenceBuilder{id: 43, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{50}}
	var b42 = charBuilder{}
	b43.items = []builder{&b42}
	var b45 = sequenceBuilder{id: 45, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{50}}
	var b44 = charBuilder{}
	b45.items = []builder{&b44}
	var b47 = sequenceBuilder{id: 47, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{50}}
	var b46 = charBuilder{}
	b47.items = []builder{&b46}
	var b49 = sequenceBuilder{id: 49, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{50}}
	var b48 = charBuilder{}
	b49.items = []builder{&b48}
	b50.options = []builder{&b39, &b41, &b43, &b45, &b47, &b49}
	var b52 = sequenceBuilder{id: 52, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b51 = charBuilder{}
	b52.items = []builder{&b51}
	b53.items = []builder{&b50, &b52, &b50, &b37}
	b54.items = []builder{&b37, &b53}
	b185.options = []builder{&b15, &b54}
	b186.options = []builder{&b185}
	var b187 = sequenceBuilder{id: 187, commit: 66, ranges: [][]int{{0, 1}, {0, -1}, {0, 1}, {0, 1}}}
	var b182 = sequenceBuilder{id: 182, commit: 2, ranges: [][]int{{1, 1}, {0, -1}}}
	var b178 = sequenceBuilder{id: 178, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b177 = charBuilder{}
	b178.items = []builder{&b177}
	var b181 = sequenceBuilder{id: 181, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b181.items = []builder{&b186, &b178}
	b182.items = []builder{&b178, &b181}
	var b176 = sequenceBuilder{id: 176, commit: 66, ranges: [][]int{{1, 1}, {0, 1}}}
	var b169 = sequenceBuilder{id: 169, commit: 64, name: "definition", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}}}
	var b166 = sequenceBuilder{id: 166, commit: 74, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}}
	var b92 = sequenceBuilder{id: 92, commit: 72, name: "symbol", ranges: [][]int{{1, -1}, {1, -1}}, generalizations: []int{133, 123, 127}}
	var b91 = sequenceBuilder{id: 91, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b90 = charBuilder{}
	b91.items = []builder{&b90}
	b92.items = []builder{&b91}
	var b165 = sequenceBuilder{id: 165, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b164 = sequenceBuilder{id: 164, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b163 = charBuilder{}
	b164.items = []builder{&b163}
	var b162 = choiceBuilder{id: 162, commit: 66}
	var b139 = sequenceBuilder{id: 139, commit: 72, name: "alias", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{162}}
	var b134 = charBuilder{}
	var b135 = charBuilder{}
	var b136 = charBuilder{}
	var b137 = charBuilder{}
	var b138 = charBuilder{}
	b139.items = []builder{&b134, &b135, &b136, &b137, &b138}
	var b142 = sequenceBuilder{id: 142, commit: 72, name: "ws", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{162}}
	var b140 = charBuilder{}
	var b141 = charBuilder{}
	b142.items = []builder{&b140, &b141}
	var b147 = sequenceBuilder{id: 147, commit: 72, name: "nows", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{162}}
	var b143 = charBuilder{}
	var b144 = charBuilder{}
	var b145 = charBuilder{}
	var b146 = charBuilder{}
	b147.items = []builder{&b143, &b144, &b145, &b146}
	var b156 = sequenceBuilder{id: 156, commit: 72, name: "failpass", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{162}}
	var b148 = charBuilder{}
	var b149 = charBuilder{}
	var b150 = charBuilder{}
	var b151 = charBuilder{}
	var b152 = charBuilder{}
	var b153 = charBuilder{}
	var b154 = charBuilder{}
	var b155 = charBuilder{}
	b156.items = []builder{&b148, &b149, &b150, &b151, &b152, &b153, &b154, &b155}
	var b161 = sequenceBuilder{id: 161, commit: 72, name: "root", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{162}}
	var b157 = charBuilder{}
	var b158 = charBuilder{}
	var b159 = charBuilder{}
	var b160 = charBuilder{}
	b161.items = []builder{&b157, &b158, &b159, &b160}
	b162.options = []builder{&b139, &b142, &b147, &b156, &b161}
	b165.items = []builder{&b164, &b162}
	b166.items = []builder{&b92, &b165}
	var b168 = sequenceBuilder{id: 168, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b167 = charBuilder{}
	b168.items = []builder{&b167}
	var b133 = choiceBuilder{id: 133, commit: 66}
	var b89 = choiceBuilder{id: 89, commit: 66, generalizations: []int{133, 123, 127}}
	var b56 = sequenceBuilder{id: 56, commit: 72, name: "any-char", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{89, 133, 123, 127}}
	var b55 = charBuilder{}
	b56.items = []builder{&b55}
	var b75 = sequenceBuilder{id: 75, commit: 72, name: "char-class", ranges: [][]int{{1, 1}, {0, 1}, {0, -1}, {1, 1}, {1, 1}, {0, 1}, {0, -1}, {1, 1}}, generalizations: []int{89, 133, 123, 127}}
	var b71 = sequenceBuilder{id: 71, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b70 = charBuilder{}
	b71.items = []builder{&b70}
	var b58 = sequenceBuilder{id: 58, commit: 72, name: "class-not", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b57 = charBuilder{}
	b58.items = []builder{&b57}
	var b72 = choiceBuilder{id: 72, commit: 10}
	var b66 = choiceBuilder{id: 66, commit: 72, name: "class-char", generalizations: []int{72}}
	var b60 = sequenceBuilder{id: 60, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{66, 72}}
	var b59 = charBuilder{}
	b60.items = []builder{&b59}
	var b65 = sequenceBuilder{id: 65, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{66, 72}}
	var b62 = sequenceBuilder{id: 62, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b61 = charBuilder{}
	b62.items = []builder{&b61}
	var b64 = sequenceBuilder{id: 64, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b63 = charBuilder{}
	b64.items = []builder{&b63}
	b65.items = []builder{&b62, &b64}
	b66.options = []builder{&b60, &b65}
	var b69 = sequenceBuilder{id: 69, commit: 72, name: "char-range", ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{72}}
	var b68 = sequenceBuilder{id: 68, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b67 = charBuilder{}
	b68.items = []builder{&b67}
	b69.items = []builder{&b66, &b68, &b66}
	b72.options = []builder{&b66, &b69}
	var b74 = sequenceBuilder{id: 74, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b73 = charBuilder{}
	b74.items = []builder{&b73}
	b75.items = []builder{&b71, &b58, &b72, &b74}
	var b88 = sequenceBuilder{id: 88, commit: 72, name: "char-sequence", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {1, 1}, {0, -1}, {1, 1}}, generalizations: []int{89, 133, 123, 127}}
	var b85 = sequenceBuilder{id: 85, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b84 = charBuilder{}
	b85.items = []builder{&b84}
	var b83 = choiceBuilder{id: 83, commit: 72, name: "sequence-char"}
	var b77 = sequenceBuilder{id: 77, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{83}}
	var b76 = charBuilder{}
	b77.items = []builder{&b76}
	var b82 = sequenceBuilder{id: 82, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{83}}
	var b79 = sequenceBuilder{id: 79, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b78 = charBuilder{}
	b79.items = []builder{&b78}
	var b81 = sequenceBuilder{id: 81, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b80 = charBuilder{}
	b81.items = []builder{&b80}
	b82.items = []builder{&b79, &b81}
	b83.options = []builder{&b77, &b82}
	var b87 = sequenceBuilder{id: 87, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b86 = charBuilder{}
	b87.items = []builder{&b86}
	b88.items = []builder{&b85, &b83, &b87}
	b89.options = []builder{&b56, &b75, &b88}
	var b97 = sequenceBuilder{id: 97, commit: 66, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}}, generalizations: []int{133, 123, 127}}
	var b94 = sequenceBuilder{id: 94, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b93 = charBuilder{}
	b94.items = []builder{&b93}
	var b96 = sequenceBuilder{id: 96, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b95 = charBuilder{}
	b96.items = []builder{&b95}
	b97.items = []builder{&b94, &b186, &b133, &b186, &b96}
	var b126 = sequenceBuilder{id: 126, commit: 64, name: "sequence", ranges: [][]int{{1, 1}, {0, -1}}, generalizations: []int{133, 127}}
	var b124 = sequenceBuilder{id: 124, commit: 72, name: "item", ranges: [][]int{{1, 1}, {0, 1}, {1, 1}, {0, 1}}}
	var b123 = choiceBuilder{id: 123, commit: 10}
	b123.options = []builder{&b89, &b92, &b97}
	var b122 = choiceBuilder{id: 122, commit: 66}
	var b106 = sequenceBuilder{id: 106, commit: 64, name: "count-quantifier", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}}, generalizations: []int{122}}
	var b103 = sequenceBuilder{id: 103, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b102 = charBuilder{}
	b103.items = []builder{&b102}
	var b101 = sequenceBuilder{id: 101, commit: 64, name: "count", ranges: [][]int{{1, 1}}}
	var b100 = sequenceBuilder{id: 100, commit: 74, ranges: [][]int{{1, -1}, {1, -1}}}
	var b99 = sequenceBuilder{id: 99, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b98 = charBuilder{}
	b99.items = []builder{&b98}
	b100.items = []builder{&b99}
	b101.items = []builder{&b100}
	var b105 = sequenceBuilder{id: 105, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b104 = charBuilder{}
	b105.items = []builder{&b104}
	b106.items = []builder{&b103, &b186, &b101, &b186, &b105}
	var b115 = sequenceBuilder{id: 115, commit: 64, name: "range-quantifier", ranges: [][]int{{1, 1}, {0, -1}, {0, 1}, {0, -1}, {1, 1}, {0, -1}, {0, 1}, {0, -1}, {1, 1}}, generalizations: []int{122}}
	var b110 = sequenceBuilder{id: 110, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b109 = charBuilder{}
	b110.items = []builder{&b109}
	var b107 = sequenceBuilder{id: 107, commit: 64, name: "range-from", ranges: [][]int{{1, 1}}}
	b107.items = []builder{&b100}
	var b112 = sequenceBuilder{id: 112, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b111 = charBuilder{}
	b112.items = []builder{&b111}
	var b108 = sequenceBuilder{id: 108, commit: 64, name: "range-to", ranges: [][]int{{1, 1}}}
	b108.items = []builder{&b100}
	var b114 = sequenceBuilder{id: 114, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b113 = charBuilder{}
	b114.items = []builder{&b113}
	b115.items = []builder{&b110, &b186, &b107, &b186, &b112, &b186, &b108, &b186, &b114}
	var b117 = sequenceBuilder{id: 117, commit: 72, name: "one-or-more", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{122}}
	var b116 = charBuilder{}
	b117.items = []builder{&b116}
	var b119 = sequenceBuilder{id: 119, commit: 72, name: "zero-or-more", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{122}}
	var b118 = charBuilder{}
	b119.items = []builder{&b118}
	var b121 = sequenceBuilder{id: 121, commit: 72, name: "zero-or-one", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{122}}
	var b120 = charBuilder{}
	b121.items = []builder{&b120}
	b122.options = []builder{&b106, &b115, &b117, &b119, &b121}
	b124.items = []builder{&b123, &b122}
	var b125 = sequenceBuilder{id: 125, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b125.items = []builder{&b186, &b124}
	b126.items = []builder{&b124, &b125}
	var b132 = sequenceBuilder{id: 132, commit: 64, name: "choice", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}, generalizations: []int{133}}
	var b127 = choiceBuilder{id: 127, commit: 66}
	b127.options = []builder{&b89, &b92, &b97, &b126}
	var b130 = sequenceBuilder{id: 130, commit: 2, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}}}
	var b129 = sequenceBuilder{id: 129, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b128 = charBuilder{}
	b129.items = []builder{&b128}
	b130.items = []builder{&b129, &b186, &b127}
	var b131 = sequenceBuilder{id: 131, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b131.items = []builder{&b186, &b130}
	b132.items = []builder{&b127, &b186, &b130, &b131}
	b133.options = []builder{&b89, &b92, &b97, &b126, &b132}
	b169.items = []builder{&b166, &b186, &b168, &b186, &b133}
	var b175 = sequenceBuilder{id: 175, commit: 2, ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var b173 = sequenceBuilder{id: 173, commit: 2, ranges: [][]int{{1, 1}, {0, -1}, {0, -1}, {1, 1}}}
	var b171 = sequenceBuilder{id: 171, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b170 = charBuilder{}
	b171.items = []builder{&b170}
	var b172 = sequenceBuilder{id: 172, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b172.items = []builder{&b186, &b171}
	b173.items = []builder{&b171, &b172, &b186, &b169}
	var b174 = sequenceBuilder{id: 174, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b174.items = []builder{&b186, &b173}
	b175.items = []builder{&b186, &b173, &b174}
	b176.items = []builder{&b169, &b175}
	var b184 = sequenceBuilder{id: 184, commit: 2, ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var b180 = sequenceBuilder{id: 180, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b179 = charBuilder{}
	b180.items = []builder{&b179}
	var b183 = sequenceBuilder{id: 183, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b183.items = []builder{&b186, &b180}
	b184.items = []builder{&b186, &b180, &b183}
	b187.items = []builder{&b182, &b186, &b176, &b184}
	b188.items = []builder{&b186, &b187, &b186}

	return parseInput(r, &p188, &b188)
}
