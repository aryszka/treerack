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
	var p112 = choiceParser{id: 112, commit: 66, name: "wschar", generalizations: []int{185, 186}}
	var p137 = sequenceParser{id: 137, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{112, 185, 186}}
	var p182 = charParser{id: 182, chars: []rune{32}}
	p137.items = []parser{&p182}
	var p164 = sequenceParser{id: 164, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{112, 185, 186}}
	var p10 = charParser{id: 10, chars: []rune{9}}
	p164.items = []parser{&p10}
	var p11 = sequenceParser{id: 11, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{112, 185, 186}}
	var p21 = charParser{id: 21, chars: []rune{10}}
	p11.items = []parser{&p21}
	var p45 = sequenceParser{id: 45, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{112, 185, 186}}
	var p159 = charParser{id: 159, chars: []rune{8}}
	p45.items = []parser{&p159}
	var p25 = sequenceParser{id: 25, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{112, 185, 186}}
	var p66 = charParser{id: 66, chars: []rune{12}}
	p25.items = []parser{&p66}
	var p67 = sequenceParser{id: 67, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{112, 185, 186}}
	var p12 = charParser{id: 12, chars: []rune{13}}
	p67.items = []parser{&p12}
	var p117 = sequenceParser{id: 117, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{112, 185, 186}}
	var p22 = charParser{id: 22, chars: []rune{11}}
	p117.items = []parser{&p22}
	p112.options = []parser{&p137, &p164, &p11, &p45, &p25, &p67, &p117}
	var p18 = sequenceParser{id: 18, commit: 72, name: "comment", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}, generalizations: []int{185, 186}}
	var p34 = choiceParser{id: 34, commit: 74, name: "comment-segment"}
	var p101 = sequenceParser{id: 101, commit: 74, name: "line-comment", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}, generalizations: []int{34}}
	var p17 = sequenceParser{id: 17, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var p118 = charParser{id: 118, chars: []rune{47}}
	var p6 = charParser{id: 6, chars: []rune{47}}
	p17.items = []parser{&p118, &p6}
	var p119 = sequenceParser{id: 119, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p165 = charParser{id: 165, not: true, chars: []rune{10}}
	p119.items = []parser{&p165}
	p101.items = []parser{&p17, &p119}
	var p68 = sequenceParser{id: 68, commit: 74, name: "block-comment", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {1, 1}, {0, -1}, {1, 1}}, generalizations: []int{34}}
	var p16 = sequenceParser{id: 16, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var p148 = charParser{id: 148, chars: []rune{47}}
	var p46 = charParser{id: 46, chars: []rune{42}}
	p16.items = []parser{&p148, &p46}
	var p23 = choiceParser{id: 23, commit: 10}
	var p129 = sequenceParser{id: 129, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{23}}
	var p56 = sequenceParser{id: 56, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p100 = charParser{id: 100, chars: []rune{42}}
	p56.items = []parser{&p100}
	var p5 = sequenceParser{id: 5, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p47 = charParser{id: 47, not: true, chars: []rune{47}}
	p5.items = []parser{&p47}
	p129.items = []parser{&p56, &p5}
	var p125 = sequenceParser{id: 125, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{23}}
	var p183 = charParser{id: 183, not: true, chars: []rune{42}}
	p125.items = []parser{&p183}
	p23.options = []parser{&p129, &p125}
	var p121 = sequenceParser{id: 121, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var p57 = charParser{id: 57, chars: []rune{42}}
	var p92 = charParser{id: 92, chars: []rune{47}}
	p121.items = []parser{&p57, &p92}
	p68.items = []parser{&p16, &p23, &p121}
	p34.options = []parser{&p101, &p68}
	var p41 = sequenceParser{id: 41, commit: 10, ranges: [][]int{{0, -1}, {0, 1}, {0, -1}, {1, 1}, {0, -1}, {0, 1}, {0, -1}, {1, 1}}}
	var p87 = choiceParser{id: 87, commit: 74, name: "ws-no-nl"}
	var p102 = sequenceParser{id: 102, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{87}}
	var p149 = charParser{id: 149, chars: []rune{32}}
	p102.items = []parser{&p149}
	var p76 = sequenceParser{id: 76, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{87}}
	var p138 = charParser{id: 138, chars: []rune{9}}
	p76.items = []parser{&p138}
	var p40 = sequenceParser{id: 40, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{87}}
	var p61 = charParser{id: 61, chars: []rune{8}}
	p40.items = []parser{&p61}
	var p95 = sequenceParser{id: 95, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{87}}
	var p150 = charParser{id: 150, chars: []rune{12}}
	p95.items = []parser{&p150}
	var p62 = sequenceParser{id: 62, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{87}}
	var p93 = charParser{id: 93, chars: []rune{13}}
	p62.items = []parser{&p93}
	var p107 = sequenceParser{id: 107, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{87}}
	var p19 = charParser{id: 19, chars: []rune{11}}
	p107.items = []parser{&p19}
	p87.options = []parser{&p102, &p76, &p40, &p95, &p62, &p107}
	var p169 = sequenceParser{id: 169, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p143 = charParser{id: 143, chars: []rune{10}}
	p169.items = []parser{&p143}
	p41.items = []parser{&p87, &p169, &p87, &p34}
	p18.items = []parser{&p34, &p41}
	p185.options = []parser{&p112, &p18}
	p186.options = []parser{&p185}
	var p187 = sequenceParser{id: 187, commit: 66, name: "syntax:wsroot", ranges: [][]int{{0, 1}, {0, -1}, {0, 1}, {0, 1}}}
	var p31 = sequenceParser{id: 31, commit: 2, ranges: [][]int{{1, 1}, {0, -1}}}
	var p174 = sequenceParser{id: 174, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p9 = charParser{id: 9, chars: []rune{59}}
	p174.items = []parser{&p9}
	var p30 = sequenceParser{id: 30, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p30.items = []parser{&p186, &p174}
	p31.items = []parser{&p174, &p30}
	var p142 = sequenceParser{id: 142, commit: 66, name: "definitions", ranges: [][]int{{1, 1}, {0, 1}}}
	var p173 = sequenceParser{id: 173, commit: 64, name: "definition", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}}}
	var p71 = sequenceParser{id: 71, commit: 74, name: "definition-name", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}}
	var p28 = sequenceParser{id: 28, commit: 72, name: "symbol", ranges: [][]int{{1, -1}, {1, -1}}, generalizations: []int{162, 44, 184}}
	var p177 = sequenceParser{id: 177, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p79 = charParser{id: 79, not: true, chars: []rune{92, 32, 10, 9, 8, 12, 13, 11, 47, 46, 91, 93, 34, 123, 125, 94, 43, 42, 63, 124, 40, 41, 58, 61, 59}}
	p177.items = []parser{&p79}
	p28.items = []parser{&p177}
	var p86 = sequenceParser{id: 86, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var p85 = sequenceParser{id: 85, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p152 = charParser{id: 152, chars: []rune{58}}
	p85.items = []parser{&p152}
	var p151 = choiceParser{id: 151, commit: 66, name: "flag"}
	var p50 = sequenceParser{id: 50, commit: 72, name: "alias", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{151}}
	var p81 = charParser{id: 81, chars: []rune{97}}
	var p14 = charParser{id: 14, chars: []rune{108}}
	var p97 = charParser{id: 97, chars: []rune{105}}
	var p35 = charParser{id: 35, chars: []rune{97}}
	var p171 = charParser{id: 171, chars: []rune{115}}
	p50.items = []parser{&p81, &p14, &p97, &p35, &p171}
	var p69 = sequenceParser{id: 69, commit: 72, name: "ws", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{151}}
	var p116 = charParser{id: 116, chars: []rune{119}}
	var p98 = charParser{id: 98, chars: []rune{115}}
	p69.items = []parser{&p116, &p98}
	var p84 = sequenceParser{id: 84, commit: 72, name: "nows", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{151}}
	var p99 = charParser{id: 99, chars: []rune{110}}
	var p90 = charParser{id: 90, chars: []rune{111}}
	var p106 = charParser{id: 106, chars: []rune{119}}
	var p126 = charParser{id: 126, chars: []rune{115}}
	p84.items = []parser{&p99, &p90, &p106, &p126}
	var p172 = sequenceParser{id: 172, commit: 72, name: "failpass", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{151}}
	var p180 = charParser{id: 180, chars: []rune{102}}
	var p135 = charParser{id: 135, chars: []rune{97}}
	var p163 = charParser{id: 163, chars: []rune{105}}
	var p15 = charParser{id: 15, chars: []rune{108}}
	var p127 = charParser{id: 127, chars: []rune{112}}
	var p70 = charParser{id: 70, chars: []rune{97}}
	var p60 = charParser{id: 60, chars: []rune{115}}
	var p167 = charParser{id: 167, chars: []rune{115}}
	p172.items = []parser{&p180, &p135, &p163, &p15, &p127, &p70, &p60, &p167}
	var p36 = sequenceParser{id: 36, commit: 72, name: "root", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{151}}
	var p147 = charParser{id: 147, chars: []rune{114}}
	var p128 = charParser{id: 128, chars: []rune{111}}
	var p103 = charParser{id: 103, chars: []rune{111}}
	var p136 = charParser{id: 136, chars: []rune{116}}
	p36.items = []parser{&p147, &p128, &p103, &p136}
	p151.options = []parser{&p50, &p69, &p84, &p172, &p36}
	p86.items = []parser{&p85, &p151}
	p71.items = []parser{&p28, &p86}
	var p181 = sequenceParser{id: 181, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p104 = charParser{id: 104, chars: []rune{61}}
	p181.items = []parser{&p104}
	var p162 = choiceParser{id: 162, commit: 66, name: "expression"}
	var p130 = choiceParser{id: 130, commit: 66, name: "terminal", generalizations: []int{162, 44, 184}}
	var p58 = sequenceParser{id: 58, commit: 72, name: "any-char", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{130, 162, 44, 184}}
	var p113 = charParser{id: 113, chars: []rune{46}}
	p58.items = []parser{&p113}
	var p114 = sequenceParser{id: 114, commit: 72, name: "char-class", ranges: [][]int{{1, 1}, {0, 1}, {0, -1}, {1, 1}, {1, 1}, {0, 1}, {0, -1}, {1, 1}}, generalizations: []int{130, 162, 44, 184}}
	var p73 = sequenceParser{id: 73, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p160 = charParser{id: 160, chars: []rune{91}}
	p73.items = []parser{&p160}
	var p175 = sequenceParser{id: 175, commit: 72, name: "class-not", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p108 = charParser{id: 108, chars: []rune{94}}
	p175.items = []parser{&p108}
	var p48 = choiceParser{id: 48, commit: 10}
	var p37 = choiceParser{id: 37, commit: 72, name: "class-char", generalizations: []int{48}}
	var p176 = sequenceParser{id: 176, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{37, 48}}
	var p88 = charParser{id: 88, not: true, chars: []rune{92, 91, 93, 94, 45}}
	p176.items = []parser{&p88}
	var p89 = sequenceParser{id: 89, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{37, 48}}
	var p42 = sequenceParser{id: 42, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p7 = charParser{id: 7, chars: []rune{92}}
	p42.items = []parser{&p7}
	var p139 = sequenceParser{id: 139, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p144 = charParser{id: 144, not: true}
	p139.items = []parser{&p144}
	p89.items = []parser{&p42, &p139}
	p37.options = []parser{&p176, &p89}
	var p82 = sequenceParser{id: 82, commit: 72, name: "char-range", ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{48}}
	var p8 = sequenceParser{id: 8, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p145 = charParser{id: 145, chars: []rune{45}}
	p8.items = []parser{&p145}
	p82.items = []parser{&p37, &p8, &p37}
	p48.options = []parser{&p37, &p82}
	var p38 = sequenceParser{id: 38, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p51 = charParser{id: 51, chars: []rune{93}}
	p38.items = []parser{&p51}
	p114.items = []parser{&p73, &p175, &p48, &p38}
	var p2 = sequenceParser{id: 2, commit: 72, name: "char-sequence", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {1, 1}, {0, -1}, {1, 1}}, generalizations: []int{130, 162, 44, 184}}
	var p1 = sequenceParser{id: 1, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p26 = charParser{id: 26, chars: []rune{34}}
	p1.items = []parser{&p26}
	var p96 = choiceParser{id: 96, commit: 72, name: "sequence-char"}
	var p77 = sequenceParser{id: 77, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{96}}
	var p20 = charParser{id: 20, not: true, chars: []rune{92, 34}}
	p77.items = []parser{&p20}
	var p75 = sequenceParser{id: 75, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{96}}
	var p49 = sequenceParser{id: 49, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p74 = charParser{id: 74, chars: []rune{92}}
	p49.items = []parser{&p74}
	var p154 = sequenceParser{id: 154, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p132 = charParser{id: 132, not: true}
	p154.items = []parser{&p132}
	p75.items = []parser{&p49, &p154}
	p96.options = []parser{&p77, &p75}
	var p155 = sequenceParser{id: 155, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p94 = charParser{id: 94, chars: []rune{34}}
	p155.items = []parser{&p94}
	p2.items = []parser{&p1, &p96, &p155}
	p130.options = []parser{&p58, &p114, &p2}
	var p27 = sequenceParser{id: 27, commit: 66, name: "group", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}}, generalizations: []int{162, 44, 184}}
	var p166 = sequenceParser{id: 166, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p110 = charParser{id: 110, chars: []rune{40}}
	p166.items = []parser{&p110}
	var p83 = sequenceParser{id: 83, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p52 = charParser{id: 52, chars: []rune{41}}
	p83.items = []parser{&p52}
	p27.items = []parser{&p166, &p186, &p162, &p186, &p83}
	var p124 = sequenceParser{id: 124, commit: 64, name: "sequence", ranges: [][]int{{1, 1}, {0, -1}}, generalizations: []int{162, 184}}
	var p65 = sequenceParser{id: 65, commit: 72, name: "item", ranges: [][]int{{1, 1}, {0, 1}, {1, 1}, {0, 1}}}
	var p44 = choiceParser{id: 44, commit: 10}
	p44.options = []parser{&p130, &p28, &p27}
	var p29 = choiceParser{id: 29, commit: 66, name: "quantity"}
	var p161 = sequenceParser{id: 161, commit: 64, name: "count-quantifier", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}}, generalizations: []int{29}}
	var p156 = sequenceParser{id: 156, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p170 = charParser{id: 170, chars: []rune{123}}
	p156.items = []parser{&p170}
	var p53 = sequenceParser{id: 53, commit: 64, name: "count", ranges: [][]int{{1, 1}}}
	var p146 = sequenceParser{id: 146, commit: 74, name: "number", ranges: [][]int{{1, -1}, {1, -1}}}
	var p63 = sequenceParser{id: 63, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p133 = charParser{id: 133, ranges: [][]rune{{48, 57}}}
	p63.items = []parser{&p133}
	p146.items = []parser{&p63}
	p53.items = []parser{&p146}
	var p39 = sequenceParser{id: 39, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p64 = charParser{id: 64, chars: []rune{125}}
	p39.items = []parser{&p64}
	p161.items = []parser{&p156, &p186, &p53, &p186, &p39}
	var p122 = sequenceParser{id: 122, commit: 64, name: "range-quantifier", ranges: [][]int{{1, 1}, {0, -1}, {0, 1}, {0, -1}, {1, 1}, {0, -1}, {0, 1}, {0, -1}, {1, 1}}, generalizations: []int{29}}
	var p13 = sequenceParser{id: 13, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p115 = charParser{id: 115, chars: []rune{123}}
	p13.items = []parser{&p115}
	var p80 = sequenceParser{id: 80, commit: 64, name: "range-from", ranges: [][]int{{1, 1}}}
	p80.items = []parser{&p146}
	var p55 = sequenceParser{id: 55, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p54 = charParser{id: 54, chars: []rune{44}}
	p55.items = []parser{&p54}
	var p43 = sequenceParser{id: 43, commit: 64, name: "range-to", ranges: [][]int{{1, 1}}}
	p43.items = []parser{&p146}
	var p78 = sequenceParser{id: 78, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p157 = charParser{id: 157, chars: []rune{125}}
	p78.items = []parser{&p157}
	p122.items = []parser{&p13, &p186, &p80, &p186, &p55, &p186, &p43, &p186, &p78}
	var p109 = sequenceParser{id: 109, commit: 72, name: "one-or-more", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{29}}
	var p111 = charParser{id: 111, chars: []rune{43}}
	p109.items = []parser{&p111}
	var p105 = sequenceParser{id: 105, commit: 72, name: "zero-or-more", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{29}}
	var p131 = charParser{id: 131, chars: []rune{42}}
	p105.items = []parser{&p131}
	var p72 = sequenceParser{id: 72, commit: 72, name: "zero-or-one", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{29}}
	var p59 = charParser{id: 59, chars: []rune{63}}
	p72.items = []parser{&p59}
	p29.options = []parser{&p161, &p122, &p109, &p105, &p72}
	p65.items = []parser{&p44, &p29}
	var p123 = sequenceParser{id: 123, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p123.items = []parser{&p186, &p65}
	p124.items = []parser{&p65, &p123}
	var p179 = sequenceParser{id: 179, commit: 64, name: "choice", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}, generalizations: []int{162}}
	var p184 = choiceParser{id: 184, commit: 66, name: "option"}
	p184.options = []parser{&p130, &p28, &p27, &p124}
	var p158 = sequenceParser{id: 158, commit: 2, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}}}
	var p24 = sequenceParser{id: 24, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p134 = charParser{id: 134, chars: []rune{124}}
	p24.items = []parser{&p134}
	p158.items = []parser{&p24, &p186, &p184}
	var p178 = sequenceParser{id: 178, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p178.items = []parser{&p186, &p158}
	p179.items = []parser{&p184, &p186, &p158, &p178}
	p162.options = []parser{&p130, &p28, &p27, &p124, &p179}
	p173.items = []parser{&p71, &p186, &p181, &p186, &p162}
	var p141 = sequenceParser{id: 141, commit: 2, ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var p4 = sequenceParser{id: 4, commit: 2, ranges: [][]int{{1, 1}, {0, -1}, {0, -1}, {1, 1}}}
	var p91 = sequenceParser{id: 91, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p153 = charParser{id: 153, chars: []rune{59}}
	p91.items = []parser{&p153}
	var p3 = sequenceParser{id: 3, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p3.items = []parser{&p186, &p91}
	p4.items = []parser{&p91, &p3, &p186, &p173}
	var p140 = sequenceParser{id: 140, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p140.items = []parser{&p186, &p4}
	p141.items = []parser{&p186, &p4, &p140}
	p142.items = []parser{&p173, &p141}
	var p33 = sequenceParser{id: 33, commit: 2, ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var p120 = sequenceParser{id: 120, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p168 = charParser{id: 168, chars: []rune{59}}
	p120.items = []parser{&p168}
	var p32 = sequenceParser{id: 32, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p32.items = []parser{&p186, &p120}
	p33.items = []parser{&p186, &p120, &p32}
	p187.items = []parser{&p31, &p186, &p142, &p33}
	p188.items = []parser{&p186, &p187, &p186}
	var b188 = sequenceBuilder{id: 188, commit: 32, name: "syntax", ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var b186 = choiceBuilder{id: 186, commit: 2}
	var b185 = choiceBuilder{id: 185, commit: 70}
	var b112 = choiceBuilder{id: 112, commit: 66}
	var b137 = sequenceBuilder{id: 137, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b182 = charBuilder{}
	b137.items = []builder{&b182}
	var b164 = sequenceBuilder{id: 164, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b10 = charBuilder{}
	b164.items = []builder{&b10}
	var b11 = sequenceBuilder{id: 11, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b21 = charBuilder{}
	b11.items = []builder{&b21}
	var b45 = sequenceBuilder{id: 45, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b159 = charBuilder{}
	b45.items = []builder{&b159}
	var b25 = sequenceBuilder{id: 25, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b66 = charBuilder{}
	b25.items = []builder{&b66}
	var b67 = sequenceBuilder{id: 67, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b12 = charBuilder{}
	b67.items = []builder{&b12}
	var b117 = sequenceBuilder{id: 117, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b22 = charBuilder{}
	b117.items = []builder{&b22}
	b112.options = []builder{&b137, &b164, &b11, &b45, &b25, &b67, &b117}
	var b18 = sequenceBuilder{id: 18, commit: 72, name: "comment", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}}
	var b34 = choiceBuilder{id: 34, commit: 74}
	var b101 = sequenceBuilder{id: 101, commit: 74, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}}
	var b17 = sequenceBuilder{id: 17, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b118 = charBuilder{}
	var b6 = charBuilder{}
	b17.items = []builder{&b118, &b6}
	var b119 = sequenceBuilder{id: 119, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b165 = charBuilder{}
	b119.items = []builder{&b165}
	b101.items = []builder{&b17, &b119}
	var b68 = sequenceBuilder{id: 68, commit: 74, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {1, 1}, {0, -1}, {1, 1}}}
	var b16 = sequenceBuilder{id: 16, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b148 = charBuilder{}
	var b46 = charBuilder{}
	b16.items = []builder{&b148, &b46}
	var b23 = choiceBuilder{id: 23, commit: 10}
	var b129 = sequenceBuilder{id: 129, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b56 = sequenceBuilder{id: 56, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b100 = charBuilder{}
	b56.items = []builder{&b100}
	var b5 = sequenceBuilder{id: 5, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b47 = charBuilder{}
	b5.items = []builder{&b47}
	b129.items = []builder{&b56, &b5}
	var b125 = sequenceBuilder{id: 125, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b183 = charBuilder{}
	b125.items = []builder{&b183}
	b23.options = []builder{&b129, &b125}
	var b121 = sequenceBuilder{id: 121, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b57 = charBuilder{}
	var b92 = charBuilder{}
	b121.items = []builder{&b57, &b92}
	b68.items = []builder{&b16, &b23, &b121}
	b34.options = []builder{&b101, &b68}
	var b41 = sequenceBuilder{id: 41, commit: 10, ranges: [][]int{{0, -1}, {0, 1}, {0, -1}, {1, 1}, {0, -1}, {0, 1}, {0, -1}, {1, 1}}}
	var b87 = choiceBuilder{id: 87, commit: 74}
	var b102 = sequenceBuilder{id: 102, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b149 = charBuilder{}
	b102.items = []builder{&b149}
	var b76 = sequenceBuilder{id: 76, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b138 = charBuilder{}
	b76.items = []builder{&b138}
	var b40 = sequenceBuilder{id: 40, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b61 = charBuilder{}
	b40.items = []builder{&b61}
	var b95 = sequenceBuilder{id: 95, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b150 = charBuilder{}
	b95.items = []builder{&b150}
	var b62 = sequenceBuilder{id: 62, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b93 = charBuilder{}
	b62.items = []builder{&b93}
	var b107 = sequenceBuilder{id: 107, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b19 = charBuilder{}
	b107.items = []builder{&b19}
	b87.options = []builder{&b102, &b76, &b40, &b95, &b62, &b107}
	var b169 = sequenceBuilder{id: 169, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b143 = charBuilder{}
	b169.items = []builder{&b143}
	b41.items = []builder{&b87, &b169, &b87, &b34}
	b18.items = []builder{&b34, &b41}
	b185.options = []builder{&b112, &b18}
	b186.options = []builder{&b185}
	var b187 = sequenceBuilder{id: 187, commit: 66, ranges: [][]int{{0, 1}, {0, -1}, {0, 1}, {0, 1}}}
	var b31 = sequenceBuilder{id: 31, commit: 2, ranges: [][]int{{1, 1}, {0, -1}}}
	var b174 = sequenceBuilder{id: 174, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b9 = charBuilder{}
	b174.items = []builder{&b9}
	var b30 = sequenceBuilder{id: 30, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b30.items = []builder{&b186, &b174}
	b31.items = []builder{&b174, &b30}
	var b142 = sequenceBuilder{id: 142, commit: 66, ranges: [][]int{{1, 1}, {0, 1}}}
	var b173 = sequenceBuilder{id: 173, commit: 64, name: "definition", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}}}
	var b71 = sequenceBuilder{id: 71, commit: 74, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}}
	var b28 = sequenceBuilder{id: 28, commit: 72, name: "symbol", ranges: [][]int{{1, -1}, {1, -1}}}
	var b177 = sequenceBuilder{id: 177, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b79 = charBuilder{}
	b177.items = []builder{&b79}
	b28.items = []builder{&b177}
	var b86 = sequenceBuilder{id: 86, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b85 = sequenceBuilder{id: 85, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b152 = charBuilder{}
	b85.items = []builder{&b152}
	var b151 = choiceBuilder{id: 151, commit: 66}
	var b50 = sequenceBuilder{id: 50, commit: 72, name: "alias", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b81 = charBuilder{}
	var b14 = charBuilder{}
	var b97 = charBuilder{}
	var b35 = charBuilder{}
	var b171 = charBuilder{}
	b50.items = []builder{&b81, &b14, &b97, &b35, &b171}
	var b69 = sequenceBuilder{id: 69, commit: 72, name: "ws", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b116 = charBuilder{}
	var b98 = charBuilder{}
	b69.items = []builder{&b116, &b98}
	var b84 = sequenceBuilder{id: 84, commit: 72, name: "nows", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b99 = charBuilder{}
	var b90 = charBuilder{}
	var b106 = charBuilder{}
	var b126 = charBuilder{}
	b84.items = []builder{&b99, &b90, &b106, &b126}
	var b172 = sequenceBuilder{id: 172, commit: 72, name: "failpass", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b180 = charBuilder{}
	var b135 = charBuilder{}
	var b163 = charBuilder{}
	var b15 = charBuilder{}
	var b127 = charBuilder{}
	var b70 = charBuilder{}
	var b60 = charBuilder{}
	var b167 = charBuilder{}
	b172.items = []builder{&b180, &b135, &b163, &b15, &b127, &b70, &b60, &b167}
	var b36 = sequenceBuilder{id: 36, commit: 72, name: "root", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b147 = charBuilder{}
	var b128 = charBuilder{}
	var b103 = charBuilder{}
	var b136 = charBuilder{}
	b36.items = []builder{&b147, &b128, &b103, &b136}
	b151.options = []builder{&b50, &b69, &b84, &b172, &b36}
	b86.items = []builder{&b85, &b151}
	b71.items = []builder{&b28, &b86}
	var b181 = sequenceBuilder{id: 181, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b104 = charBuilder{}
	b181.items = []builder{&b104}
	var b162 = choiceBuilder{id: 162, commit: 66}
	var b130 = choiceBuilder{id: 130, commit: 66}
	var b58 = sequenceBuilder{id: 58, commit: 72, name: "any-char", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b113 = charBuilder{}
	b58.items = []builder{&b113}
	var b114 = sequenceBuilder{id: 114, commit: 72, name: "char-class", ranges: [][]int{{1, 1}, {0, 1}, {0, -1}, {1, 1}, {1, 1}, {0, 1}, {0, -1}, {1, 1}}}
	var b73 = sequenceBuilder{id: 73, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b160 = charBuilder{}
	b73.items = []builder{&b160}
	var b175 = sequenceBuilder{id: 175, commit: 72, name: "class-not", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b108 = charBuilder{}
	b175.items = []builder{&b108}
	var b48 = choiceBuilder{id: 48, commit: 10}
	var b37 = choiceBuilder{id: 37, commit: 72, name: "class-char"}
	var b176 = sequenceBuilder{id: 176, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b88 = charBuilder{}
	b176.items = []builder{&b88}
	var b89 = sequenceBuilder{id: 89, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b42 = sequenceBuilder{id: 42, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b7 = charBuilder{}
	b42.items = []builder{&b7}
	var b139 = sequenceBuilder{id: 139, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b144 = charBuilder{}
	b139.items = []builder{&b144}
	b89.items = []builder{&b42, &b139}
	b37.options = []builder{&b176, &b89}
	var b82 = sequenceBuilder{id: 82, commit: 72, name: "char-range", ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b8 = sequenceBuilder{id: 8, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b145 = charBuilder{}
	b8.items = []builder{&b145}
	b82.items = []builder{&b37, &b8, &b37}
	b48.options = []builder{&b37, &b82}
	var b38 = sequenceBuilder{id: 38, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b51 = charBuilder{}
	b38.items = []builder{&b51}
	b114.items = []builder{&b73, &b175, &b48, &b38}
	var b2 = sequenceBuilder{id: 2, commit: 72, name: "char-sequence", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {1, 1}, {0, -1}, {1, 1}}}
	var b1 = sequenceBuilder{id: 1, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b26 = charBuilder{}
	b1.items = []builder{&b26}
	var b96 = choiceBuilder{id: 96, commit: 72, name: "sequence-char"}
	var b77 = sequenceBuilder{id: 77, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b20 = charBuilder{}
	b77.items = []builder{&b20}
	var b75 = sequenceBuilder{id: 75, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b49 = sequenceBuilder{id: 49, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b74 = charBuilder{}
	b49.items = []builder{&b74}
	var b154 = sequenceBuilder{id: 154, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b132 = charBuilder{}
	b154.items = []builder{&b132}
	b75.items = []builder{&b49, &b154}
	b96.options = []builder{&b77, &b75}
	var b155 = sequenceBuilder{id: 155, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b94 = charBuilder{}
	b155.items = []builder{&b94}
	b2.items = []builder{&b1, &b96, &b155}
	b130.options = []builder{&b58, &b114, &b2}
	var b27 = sequenceBuilder{id: 27, commit: 66, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}}}
	var b166 = sequenceBuilder{id: 166, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b110 = charBuilder{}
	b166.items = []builder{&b110}
	var b83 = sequenceBuilder{id: 83, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b52 = charBuilder{}
	b83.items = []builder{&b52}
	b27.items = []builder{&b166, &b186, &b162, &b186, &b83}
	var b124 = sequenceBuilder{id: 124, commit: 64, name: "sequence", ranges: [][]int{{1, 1}, {0, -1}}}
	var b65 = sequenceBuilder{id: 65, commit: 72, name: "item", ranges: [][]int{{1, 1}, {0, 1}, {1, 1}, {0, 1}}}
	var b44 = choiceBuilder{id: 44, commit: 10}
	b44.options = []builder{&b130, &b28, &b27}
	var b29 = choiceBuilder{id: 29, commit: 66}
	var b161 = sequenceBuilder{id: 161, commit: 64, name: "count-quantifier", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}}}
	var b156 = sequenceBuilder{id: 156, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b170 = charBuilder{}
	b156.items = []builder{&b170}
	var b53 = sequenceBuilder{id: 53, commit: 64, name: "count", ranges: [][]int{{1, 1}}}
	var b146 = sequenceBuilder{id: 146, commit: 74, ranges: [][]int{{1, -1}, {1, -1}}}
	var b63 = sequenceBuilder{id: 63, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b133 = charBuilder{}
	b63.items = []builder{&b133}
	b146.items = []builder{&b63}
	b53.items = []builder{&b146}
	var b39 = sequenceBuilder{id: 39, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b64 = charBuilder{}
	b39.items = []builder{&b64}
	b161.items = []builder{&b156, &b186, &b53, &b186, &b39}
	var b122 = sequenceBuilder{id: 122, commit: 64, name: "range-quantifier", ranges: [][]int{{1, 1}, {0, -1}, {0, 1}, {0, -1}, {1, 1}, {0, -1}, {0, 1}, {0, -1}, {1, 1}}}
	var b13 = sequenceBuilder{id: 13, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b115 = charBuilder{}
	b13.items = []builder{&b115}
	var b80 = sequenceBuilder{id: 80, commit: 64, name: "range-from", ranges: [][]int{{1, 1}}}
	b80.items = []builder{&b146}
	var b55 = sequenceBuilder{id: 55, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b54 = charBuilder{}
	b55.items = []builder{&b54}
	var b43 = sequenceBuilder{id: 43, commit: 64, name: "range-to", ranges: [][]int{{1, 1}}}
	b43.items = []builder{&b146}
	var b78 = sequenceBuilder{id: 78, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b157 = charBuilder{}
	b78.items = []builder{&b157}
	b122.items = []builder{&b13, &b186, &b80, &b186, &b55, &b186, &b43, &b186, &b78}
	var b109 = sequenceBuilder{id: 109, commit: 72, name: "one-or-more", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b111 = charBuilder{}
	b109.items = []builder{&b111}
	var b105 = sequenceBuilder{id: 105, commit: 72, name: "zero-or-more", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b131 = charBuilder{}
	b105.items = []builder{&b131}
	var b72 = sequenceBuilder{id: 72, commit: 72, name: "zero-or-one", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b59 = charBuilder{}
	b72.items = []builder{&b59}
	b29.options = []builder{&b161, &b122, &b109, &b105, &b72}
	b65.items = []builder{&b44, &b29}
	var b123 = sequenceBuilder{id: 123, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b123.items = []builder{&b186, &b65}
	b124.items = []builder{&b65, &b123}
	var b179 = sequenceBuilder{id: 179, commit: 64, name: "choice", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}}
	var b184 = choiceBuilder{id: 184, commit: 66}
	b184.options = []builder{&b130, &b28, &b27, &b124}
	var b158 = sequenceBuilder{id: 158, commit: 2, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}}}
	var b24 = sequenceBuilder{id: 24, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b134 = charBuilder{}
	b24.items = []builder{&b134}
	b158.items = []builder{&b24, &b186, &b184}
	var b178 = sequenceBuilder{id: 178, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b178.items = []builder{&b186, &b158}
	b179.items = []builder{&b184, &b186, &b158, &b178}
	b162.options = []builder{&b130, &b28, &b27, &b124, &b179}
	b173.items = []builder{&b71, &b186, &b181, &b186, &b162}
	var b141 = sequenceBuilder{id: 141, commit: 2, ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var b4 = sequenceBuilder{id: 4, commit: 2, ranges: [][]int{{1, 1}, {0, -1}, {0, -1}, {1, 1}}}
	var b91 = sequenceBuilder{id: 91, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b153 = charBuilder{}
	b91.items = []builder{&b153}
	var b3 = sequenceBuilder{id: 3, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b3.items = []builder{&b186, &b91}
	b4.items = []builder{&b91, &b3, &b186, &b173}
	var b140 = sequenceBuilder{id: 140, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b140.items = []builder{&b186, &b4}
	b141.items = []builder{&b186, &b4, &b140}
	b142.items = []builder{&b173, &b141}
	var b33 = sequenceBuilder{id: 33, commit: 2, ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var b120 = sequenceBuilder{id: 120, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b168 = charBuilder{}
	b120.items = []builder{&b168}
	var b32 = sequenceBuilder{id: 32, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b32.items = []builder{&b186, &b120}
	b33.items = []builder{&b186, &b120, &b32}
	b187.items = []builder{&b31, &b186, &b142, &b33}
	b188.items = []builder{&b186, &b187, &b186}

	return parseInput(r, &p188, &b188)
}
