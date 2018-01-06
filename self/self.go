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
func parse(r io.Reader, p parser, b builder) (*Node, error) {
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
	var p125 = choiceParser{id: 125, commit: 66, name: "wschar", generalizations: []int{185, 186}}
	var p84 = sequenceParser{id: 84, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{125, 185, 186}}
	var p105 = charParser{id: 105, chars: []rune{32}}
	p84.items = []parser{&p105}
	var p78 = sequenceParser{id: 78, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{125, 185, 186}}
	var p175 = charParser{id: 175, chars: []rune{9}}
	p78.items = []parser{&p175}
	var p97 = sequenceParser{id: 97, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{125, 185, 186}}
	var p85 = charParser{id: 85, chars: []rune{10}}
	p97.items = []parser{&p85}
	var p5 = sequenceParser{id: 5, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{125, 185, 186}}
	var p26 = charParser{id: 26, chars: []rune{8}}
	p5.items = []parser{&p26}
	var p145 = sequenceParser{id: 145, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{125, 185, 186}}
	var p41 = charParser{id: 41, chars: []rune{12}}
	p145.items = []parser{&p41}
	var p137 = sequenceParser{id: 137, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{125, 185, 186}}
	var p18 = charParser{id: 18, chars: []rune{13}}
	p137.items = []parser{&p18}
	var p181 = sequenceParser{id: 181, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{125, 185, 186}}
	var p98 = charParser{id: 98, chars: []rune{11}}
	p181.items = []parser{&p98}
	p125.options = []parser{&p84, &p78, &p97, &p5, &p145, &p137, &p181}
	var p168 = sequenceParser{id: 168, commit: 72, name: "comment", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}, generalizations: []int{185, 186}}
	var p160 = choiceParser{id: 160, commit: 74, name: "comment-segment"}
	var p42 = sequenceParser{id: 42, commit: 74, name: "line-comment", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}, generalizations: []int{160}}
	var p35 = sequenceParser{id: 35, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var p73 = charParser{id: 73, chars: []rune{47}}
	var p117 = charParser{id: 117, chars: []rune{47}}
	p35.items = []parser{&p73, &p117}
	var p138 = sequenceParser{id: 138, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p132 = charParser{id: 132, not: true, chars: []rune{10}}
	p138.items = []parser{&p132}
	p42.items = []parser{&p35, &p138}
	var p159 = sequenceParser{id: 159, commit: 74, name: "block-comment", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {1, 1}, {0, -1}, {1, 1}}, generalizations: []int{160}}
	var p166 = sequenceParser{id: 166, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var p182 = charParser{id: 182, chars: []rune{47}}
	var p113 = charParser{id: 113, chars: []rune{42}}
	p166.items = []parser{&p182, &p113}
	var p148 = choiceParser{id: 148, commit: 10}
	var p54 = sequenceParser{id: 54, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{148}}
	var p22 = sequenceParser{id: 22, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p146 = charParser{id: 146, chars: []rune{42}}
	p22.items = []parser{&p146}
	var p155 = sequenceParser{id: 155, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p147 = charParser{id: 147, not: true, chars: []rune{47}}
	p155.items = []parser{&p147}
	p54.items = []parser{&p22, &p155}
	var p106 = sequenceParser{id: 106, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{148}}
	var p34 = charParser{id: 34, not: true, chars: []rune{42}}
	p106.items = []parser{&p34}
	p148.options = []parser{&p54, &p106}
	var p50 = sequenceParser{id: 50, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var p114 = charParser{id: 114, chars: []rune{42}}
	var p121 = charParser{id: 121, chars: []rune{47}}
	p50.items = []parser{&p114, &p121}
	p159.items = []parser{&p166, &p148, &p50}
	p160.options = []parser{&p42, &p159}
	var p123 = sequenceParser{id: 123, commit: 10, ranges: [][]int{{0, -1}, {0, 1}, {0, -1}, {1, 1}, {0, -1}, {0, 1}, {0, -1}, {1, 1}}}
	var p79 = choiceParser{id: 79, commit: 74, name: "ws-no-nl"}
	var p12 = sequenceParser{id: 12, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{79}}
	var p167 = charParser{id: 167, chars: []rune{32}}
	p12.items = []parser{&p167}
	var p13 = sequenceParser{id: 13, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{79}}
	var p6 = charParser{id: 6, chars: []rune{9}}
	p13.items = []parser{&p6}
	var p86 = sequenceParser{id: 86, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{79}}
	var p172 = charParser{id: 172, chars: []rune{8}}
	p86.items = []parser{&p172}
	var p139 = sequenceParser{id: 139, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{79}}
	var p122 = charParser{id: 122, chars: []rune{12}}
	p139.items = []parser{&p122}
	var p14 = sequenceParser{id: 14, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{79}}
	var p183 = charParser{id: 183, chars: []rune{13}}
	p14.items = []parser{&p183}
	var p141 = sequenceParser{id: 141, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{79}}
	var p43 = charParser{id: 43, chars: []rune{11}}
	p141.items = []parser{&p43}
	p79.options = []parser{&p12, &p13, &p86, &p139, &p14, &p141}
	var p156 = sequenceParser{id: 156, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p64 = charParser{id: 64, chars: []rune{10}}
	p156.items = []parser{&p64}
	p123.items = []parser{&p79, &p156, &p79, &p160}
	p168.items = []parser{&p160, &p123}
	p185.options = []parser{&p125, &p168}
	p186.options = []parser{&p185}
	var p187 = sequenceParser{id: 187, commit: 66, name: "syntax:wsroot", ranges: [][]int{{0, 1}, {0, -1}, {0, 1}, {0, 1}}}
	var p70 = sequenceParser{id: 70, commit: 2, ranges: [][]int{{1, 1}, {0, -1}}}
	var p161 = sequenceParser{id: 161, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p25 = charParser{id: 25, chars: []rune{59}}
	p161.items = []parser{&p25}
	var p69 = sequenceParser{id: 69, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p69.items = []parser{&p186, &p161}
	p70.items = []parser{&p161, &p69}
	var p112 = sequenceParser{id: 112, commit: 66, name: "definitions", ranges: [][]int{{1, 1}, {0, 1}}}
	var p68 = sequenceParser{id: 68, commit: 64, name: "definition", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}}}
	var p90 = sequenceParser{id: 90, commit: 74, name: "definition-name", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}}
	var p162 = sequenceParser{id: 162, commit: 72, name: "symbol", ranges: [][]int{{1, -1}, {1, -1}}, generalizations: []int{165, 83, 103}}
	var p29 = sequenceParser{id: 29, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p135 = charParser{id: 135, not: true, chars: []rune{92, 32, 10, 9, 8, 12, 13, 11, 47, 46, 91, 93, 34, 123, 125, 94, 43, 42, 63, 124, 40, 41, 58, 61, 59}}
	p29.items = []parser{&p135}
	p162.items = []parser{&p29}
	var p136 = sequenceParser{id: 136, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var p89 = sequenceParser{id: 89, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p59 = charParser{id: 59, chars: []rune{58}}
	p89.items = []parser{&p59}
	var p56 = choiceParser{id: 56, commit: 66, name: "flag"}
	var p32 = sequenceParser{id: 32, commit: 72, name: "alias", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{56}}
	var p58 = charParser{id: 58, chars: []rune{97}}
	var p11 = charParser{id: 11, chars: []rune{108}}
	var p116 = charParser{id: 116, chars: []rune{105}}
	var p151 = charParser{id: 151, chars: []rune{97}}
	var p1 = charParser{id: 1, chars: []rune{115}}
	p32.items = []parser{&p58, &p11, &p116, &p151, &p1}
	var p2 = sequenceParser{id: 2, commit: 72, name: "ws", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{56}}
	var p40 = charParser{id: 40, chars: []rune{119}}
	var p88 = charParser{id: 88, chars: []rune{115}}
	p2.items = []parser{&p40, &p88}
	var p104 = sequenceParser{id: 104, commit: 72, name: "nows", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{56}}
	var p118 = charParser{id: 118, chars: []rune{110}}
	var p67 = charParser{id: 67, chars: []rune{111}}
	var p174 = charParser{id: 174, chars: []rune{119}}
	var p52 = charParser{id: 52, chars: []rune{115}}
	p104.items = []parser{&p118, &p67, &p174, &p52}
	var p171 = sequenceParser{id: 171, commit: 72, name: "failpass", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{56}}
	var p66 = charParser{id: 66, chars: []rune{102}}
	var p119 = charParser{id: 119, chars: []rune{97}}
	var p124 = charParser{id: 124, chars: []rune{105}}
	var p3 = charParser{id: 3, chars: []rune{108}}
	var p53 = charParser{id: 53, chars: []rune{112}}
	var p140 = charParser{id: 140, chars: []rune{97}}
	var p170 = charParser{id: 170, chars: []rune{115}}
	var p131 = charParser{id: 131, chars: []rune{115}}
	p171.items = []parser{&p66, &p119, &p124, &p3, &p53, &p140, &p170, &p131}
	var p27 = sequenceParser{id: 27, commit: 72, name: "root", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{56}}
	var p109 = charParser{id: 109, chars: []rune{114}}
	var p49 = charParser{id: 49, chars: []rune{111}}
	var p120 = charParser{id: 120, chars: []rune{111}}
	var p152 = charParser{id: 152, chars: []rune{116}}
	p27.items = []parser{&p109, &p49, &p120, &p152}
	p56.options = []parser{&p32, &p2, &p104, &p171, &p27}
	p136.items = []parser{&p89, &p56}
	p90.items = []parser{&p162, &p136}
	var p24 = sequenceParser{id: 24, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p17 = charParser{id: 17, chars: []rune{61}}
	p24.items = []parser{&p17}
	var p165 = choiceParser{id: 165, commit: 66, name: "expression"}
	var p37 = choiceParser{id: 37, commit: 66, name: "terminal", generalizations: []int{165, 83, 103}}
	var p87 = sequenceParser{id: 87, commit: 72, name: "any-char", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{37, 165, 83, 103}}
	var p176 = charParser{id: 176, chars: []rune{46}}
	p87.items = []parser{&p176}
	var p99 = sequenceParser{id: 99, commit: 72, name: "char-class", ranges: [][]int{{1, 1}, {0, 1}, {0, -1}, {1, 1}, {1, 1}, {0, 1}, {0, -1}, {1, 1}}, generalizations: []int{37, 165, 83, 103}}
	var p142 = sequenceParser{id: 142, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p51 = charParser{id: 51, chars: []rune{91}}
	p142.items = []parser{&p51}
	var p55 = sequenceParser{id: 55, commit: 72, name: "class-not", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p19 = charParser{id: 19, chars: []rune{94}}
	p55.items = []parser{&p19}
	var p158 = choiceParser{id: 158, commit: 10}
	var p28 = choiceParser{id: 28, commit: 72, name: "class-char", generalizations: []int{158}}
	var p65 = sequenceParser{id: 65, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{28, 158}}
	var p81 = charParser{id: 81, not: true, chars: []rune{92, 91, 93, 94, 45}}
	p65.items = []parser{&p81}
	var p178 = sequenceParser{id: 178, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{28, 158}}
	var p177 = sequenceParser{id: 177, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p157 = charParser{id: 157, chars: []rune{92}}
	p177.items = []parser{&p157}
	var p36 = sequenceParser{id: 36, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p82 = charParser{id: 82, not: true}
	p36.items = []parser{&p82}
	p178.items = []parser{&p177, &p36}
	p28.options = []parser{&p65, &p178}
	var p45 = sequenceParser{id: 45, commit: 72, name: "char-range", ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{158}}
	var p133 = sequenceParser{id: 133, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p44 = charParser{id: 44, chars: []rune{45}}
	p133.items = []parser{&p44}
	p45.items = []parser{&p28, &p133, &p28}
	p158.options = []parser{&p28, &p45}
	var p61 = sequenceParser{id: 61, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p92 = charParser{id: 92, chars: []rune{93}}
	p61.items = []parser{&p92}
	p99.items = []parser{&p142, &p55, &p158, &p61}
	var p80 = sequenceParser{id: 80, commit: 72, name: "char-sequence", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {1, 1}, {0, -1}, {1, 1}}, generalizations: []int{37, 165, 83, 103}}
	var p107 = sequenceParser{id: 107, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p7 = charParser{id: 7, chars: []rune{34}}
	p107.items = []parser{&p7}
	var p46 = choiceParser{id: 46, commit: 72, name: "sequence-char"}
	var p134 = sequenceParser{id: 134, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{46}}
	var p20 = charParser{id: 20, not: true, chars: []rune{92, 34}}
	p134.items = []parser{&p20}
	var p169 = sequenceParser{id: 169, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{46}}
	var p100 = sequenceParser{id: 100, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p38 = charParser{id: 38, chars: []rune{92}}
	p100.items = []parser{&p38}
	var p173 = sequenceParser{id: 173, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p62 = charParser{id: 62, not: true}
	p173.items = []parser{&p62}
	p169.items = []parser{&p100, &p173}
	p46.options = []parser{&p134, &p169}
	var p143 = sequenceParser{id: 143, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p93 = charParser{id: 93, chars: []rune{34}}
	p143.items = []parser{&p93}
	p80.items = []parser{&p107, &p46, &p143}
	p37.options = []parser{&p87, &p99, &p80}
	var p31 = sequenceParser{id: 31, commit: 66, name: "group", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}}, generalizations: []int{165, 83, 103}}
	var p184 = sequenceParser{id: 184, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p94 = charParser{id: 94, chars: []rune{40}}
	p184.items = []parser{&p94}
	var p95 = sequenceParser{id: 95, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p30 = charParser{id: 30, chars: []rune{41}}
	p95.items = []parser{&p30}
	p31.items = []parser{&p184, &p186, &p165, &p186, &p95}
	var p48 = sequenceParser{id: 48, commit: 64, name: "sequence", ranges: [][]int{{1, 1}, {0, -1}}, generalizations: []int{165, 103}}
	var p130 = sequenceParser{id: 130, commit: 72, name: "item", ranges: [][]int{{1, 1}, {0, 1}, {1, 1}, {0, 1}}}
	var p83 = choiceParser{id: 83, commit: 10}
	p83.options = []parser{&p37, &p162, &p31}
	var p180 = choiceParser{id: 180, commit: 66, name: "quantity"}
	var p149 = sequenceParser{id: 149, commit: 64, name: "count-quantifier", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}}, generalizations: []int{180}}
	var p126 = sequenceParser{id: 126, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p74 = charParser{id: 74, chars: []rune{123}}
	p126.items = []parser{&p74}
	var p163 = sequenceParser{id: 163, commit: 64, name: "count", ranges: [][]int{{1, 1}}}
	var p108 = sequenceParser{id: 108, commit: 74, name: "number", ranges: [][]int{{1, -1}, {1, -1}}}
	var p15 = sequenceParser{id: 15, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p144 = charParser{id: 144, ranges: [][]rune{{48, 57}}}
	p15.items = []parser{&p144}
	p108.items = []parser{&p15}
	p163.items = []parser{&p108}
	var p8 = sequenceParser{id: 8, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p9 = charParser{id: 9, chars: []rune{125}}
	p8.items = []parser{&p9}
	p149.items = []parser{&p126, &p186, &p163, &p186, &p8}
	var p23 = sequenceParser{id: 23, commit: 64, name: "range-quantifier", ranges: [][]int{{1, 1}, {0, -1}, {0, 1}, {0, -1}, {1, 1}, {0, -1}, {0, 1}, {0, -1}, {1, 1}}, generalizations: []int{180}}
	var p63 = sequenceParser{id: 63, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p150 = charParser{id: 150, chars: []rune{123}}
	p63.items = []parser{&p150}
	var p127 = sequenceParser{id: 127, commit: 64, name: "range-from", ranges: [][]int{{1, 1}}}
	p127.items = []parser{&p108}
	var p96 = sequenceParser{id: 96, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p39 = charParser{id: 39, chars: []rune{44}}
	p96.items = []parser{&p39}
	var p129 = sequenceParser{id: 129, commit: 64, name: "range-to", ranges: [][]int{{1, 1}}}
	p129.items = []parser{&p108}
	var p101 = sequenceParser{id: 101, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p21 = charParser{id: 21, chars: []rune{125}}
	p101.items = []parser{&p21}
	p23.items = []parser{&p63, &p186, &p127, &p186, &p96, &p186, &p129, &p186, &p101}
	var p128 = sequenceParser{id: 128, commit: 72, name: "one-or-more", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{180}}
	var p16 = charParser{id: 16, chars: []rune{43}}
	p128.items = []parser{&p16}
	var p57 = sequenceParser{id: 57, commit: 72, name: "zero-or-more", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{180}}
	var p102 = charParser{id: 102, chars: []rune{42}}
	p57.items = []parser{&p102}
	var p179 = sequenceParser{id: 179, commit: 72, name: "zero-or-one", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{180}}
	var p75 = charParser{id: 75, chars: []rune{63}}
	p179.items = []parser{&p75}
	p180.options = []parser{&p149, &p23, &p128, &p57, &p179}
	p130.items = []parser{&p83, &p180}
	var p47 = sequenceParser{id: 47, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p47.items = []parser{&p186, &p130}
	p48.items = []parser{&p130, &p47}
	var p77 = sequenceParser{id: 77, commit: 64, name: "choice", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}, generalizations: []int{165}}
	var p103 = choiceParser{id: 103, commit: 66, name: "option"}
	p103.options = []parser{&p37, &p162, &p31, &p48}
	var p164 = sequenceParser{id: 164, commit: 2, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}}}
	var p115 = sequenceParser{id: 115, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p10 = charParser{id: 10, chars: []rune{124}}
	p115.items = []parser{&p10}
	p164.items = []parser{&p115, &p186, &p103}
	var p76 = sequenceParser{id: 76, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p76.items = []parser{&p186, &p164}
	p77.items = []parser{&p103, &p186, &p164, &p76}
	p165.options = []parser{&p37, &p162, &p31, &p48, &p77}
	p68.items = []parser{&p90, &p186, &p24, &p186, &p165}
	var p111 = sequenceParser{id: 111, commit: 2, ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var p154 = sequenceParser{id: 154, commit: 2, ranges: [][]int{{1, 1}, {0, -1}, {0, -1}, {1, 1}}}
	var p60 = sequenceParser{id: 60, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p33 = charParser{id: 33, chars: []rune{59}}
	p60.items = []parser{&p33}
	var p153 = sequenceParser{id: 153, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p153.items = []parser{&p186, &p60}
	p154.items = []parser{&p60, &p153, &p186, &p68}
	var p110 = sequenceParser{id: 110, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p110.items = []parser{&p186, &p154}
	p111.items = []parser{&p186, &p154, &p110}
	p112.items = []parser{&p68, &p111}
	var p72 = sequenceParser{id: 72, commit: 2, ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var p91 = sequenceParser{id: 91, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p4 = charParser{id: 4, chars: []rune{59}}
	p91.items = []parser{&p4}
	var p71 = sequenceParser{id: 71, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p71.items = []parser{&p186, &p91}
	p72.items = []parser{&p186, &p91, &p71}
	p187.items = []parser{&p70, &p186, &p112, &p72}
	p188.items = []parser{&p186, &p187, &p186}
	var b188 = sequenceBuilder{id: 188, commit: 32, name: "syntax", ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var b186 = choiceBuilder{id: 186, commit: 2}
	var b185 = choiceBuilder{id: 185, commit: 70}
	var b125 = choiceBuilder{id: 125, commit: 66}
	var b84 = sequenceBuilder{id: 84, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b105 = charBuilder{}
	b84.items = []builder{&b105}
	var b78 = sequenceBuilder{id: 78, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b175 = charBuilder{}
	b78.items = []builder{&b175}
	var b97 = sequenceBuilder{id: 97, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b85 = charBuilder{}
	b97.items = []builder{&b85}
	var b5 = sequenceBuilder{id: 5, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b26 = charBuilder{}
	b5.items = []builder{&b26}
	var b145 = sequenceBuilder{id: 145, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b41 = charBuilder{}
	b145.items = []builder{&b41}
	var b137 = sequenceBuilder{id: 137, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b18 = charBuilder{}
	b137.items = []builder{&b18}
	var b181 = sequenceBuilder{id: 181, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b98 = charBuilder{}
	b181.items = []builder{&b98}
	b125.options = []builder{&b84, &b78, &b97, &b5, &b145, &b137, &b181}
	var b168 = sequenceBuilder{id: 168, commit: 72, name: "comment", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}}
	var b160 = choiceBuilder{id: 160, commit: 74}
	var b42 = sequenceBuilder{id: 42, commit: 74, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}}
	var b35 = sequenceBuilder{id: 35, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b73 = charBuilder{}
	var b117 = charBuilder{}
	b35.items = []builder{&b73, &b117}
	var b138 = sequenceBuilder{id: 138, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b132 = charBuilder{}
	b138.items = []builder{&b132}
	b42.items = []builder{&b35, &b138}
	var b159 = sequenceBuilder{id: 159, commit: 74, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {1, 1}, {0, -1}, {1, 1}}}
	var b166 = sequenceBuilder{id: 166, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b182 = charBuilder{}
	var b113 = charBuilder{}
	b166.items = []builder{&b182, &b113}
	var b148 = choiceBuilder{id: 148, commit: 10}
	var b54 = sequenceBuilder{id: 54, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b22 = sequenceBuilder{id: 22, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b146 = charBuilder{}
	b22.items = []builder{&b146}
	var b155 = sequenceBuilder{id: 155, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b147 = charBuilder{}
	b155.items = []builder{&b147}
	b54.items = []builder{&b22, &b155}
	var b106 = sequenceBuilder{id: 106, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b34 = charBuilder{}
	b106.items = []builder{&b34}
	b148.options = []builder{&b54, &b106}
	var b50 = sequenceBuilder{id: 50, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b114 = charBuilder{}
	var b121 = charBuilder{}
	b50.items = []builder{&b114, &b121}
	b159.items = []builder{&b166, &b148, &b50}
	b160.options = []builder{&b42, &b159}
	var b123 = sequenceBuilder{id: 123, commit: 10, ranges: [][]int{{0, -1}, {0, 1}, {0, -1}, {1, 1}, {0, -1}, {0, 1}, {0, -1}, {1, 1}}}
	var b79 = choiceBuilder{id: 79, commit: 74}
	var b12 = sequenceBuilder{id: 12, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b167 = charBuilder{}
	b12.items = []builder{&b167}
	var b13 = sequenceBuilder{id: 13, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b6 = charBuilder{}
	b13.items = []builder{&b6}
	var b86 = sequenceBuilder{id: 86, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b172 = charBuilder{}
	b86.items = []builder{&b172}
	var b139 = sequenceBuilder{id: 139, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b122 = charBuilder{}
	b139.items = []builder{&b122}
	var b14 = sequenceBuilder{id: 14, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b183 = charBuilder{}
	b14.items = []builder{&b183}
	var b141 = sequenceBuilder{id: 141, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b43 = charBuilder{}
	b141.items = []builder{&b43}
	b79.options = []builder{&b12, &b13, &b86, &b139, &b14, &b141}
	var b156 = sequenceBuilder{id: 156, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b64 = charBuilder{}
	b156.items = []builder{&b64}
	b123.items = []builder{&b79, &b156, &b79, &b160}
	b168.items = []builder{&b160, &b123}
	b185.options = []builder{&b125, &b168}
	b186.options = []builder{&b185}
	var b187 = sequenceBuilder{id: 187, commit: 66, ranges: [][]int{{0, 1}, {0, -1}, {0, 1}, {0, 1}}}
	var b70 = sequenceBuilder{id: 70, commit: 2, ranges: [][]int{{1, 1}, {0, -1}}}
	var b161 = sequenceBuilder{id: 161, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b25 = charBuilder{}
	b161.items = []builder{&b25}
	var b69 = sequenceBuilder{id: 69, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b69.items = []builder{&b186, &b161}
	b70.items = []builder{&b161, &b69}
	var b112 = sequenceBuilder{id: 112, commit: 66, ranges: [][]int{{1, 1}, {0, 1}}}
	var b68 = sequenceBuilder{id: 68, commit: 64, name: "definition", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}}}
	var b90 = sequenceBuilder{id: 90, commit: 74, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}}
	var b162 = sequenceBuilder{id: 162, commit: 72, name: "symbol", ranges: [][]int{{1, -1}, {1, -1}}}
	var b29 = sequenceBuilder{id: 29, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b135 = charBuilder{}
	b29.items = []builder{&b135}
	b162.items = []builder{&b29}
	var b136 = sequenceBuilder{id: 136, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b89 = sequenceBuilder{id: 89, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b59 = charBuilder{}
	b89.items = []builder{&b59}
	var b56 = choiceBuilder{id: 56, commit: 66}
	var b32 = sequenceBuilder{id: 32, commit: 72, name: "alias", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b58 = charBuilder{}
	var b11 = charBuilder{}
	var b116 = charBuilder{}
	var b151 = charBuilder{}
	var b1 = charBuilder{}
	b32.items = []builder{&b58, &b11, &b116, &b151, &b1}
	var b2 = sequenceBuilder{id: 2, commit: 72, name: "ws", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b40 = charBuilder{}
	var b88 = charBuilder{}
	b2.items = []builder{&b40, &b88}
	var b104 = sequenceBuilder{id: 104, commit: 72, name: "nows", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b118 = charBuilder{}
	var b67 = charBuilder{}
	var b174 = charBuilder{}
	var b52 = charBuilder{}
	b104.items = []builder{&b118, &b67, &b174, &b52}
	var b171 = sequenceBuilder{id: 171, commit: 72, name: "failpass", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b66 = charBuilder{}
	var b119 = charBuilder{}
	var b124 = charBuilder{}
	var b3 = charBuilder{}
	var b53 = charBuilder{}
	var b140 = charBuilder{}
	var b170 = charBuilder{}
	var b131 = charBuilder{}
	b171.items = []builder{&b66, &b119, &b124, &b3, &b53, &b140, &b170, &b131}
	var b27 = sequenceBuilder{id: 27, commit: 72, name: "root", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b109 = charBuilder{}
	var b49 = charBuilder{}
	var b120 = charBuilder{}
	var b152 = charBuilder{}
	b27.items = []builder{&b109, &b49, &b120, &b152}
	b56.options = []builder{&b32, &b2, &b104, &b171, &b27}
	b136.items = []builder{&b89, &b56}
	b90.items = []builder{&b162, &b136}
	var b24 = sequenceBuilder{id: 24, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b17 = charBuilder{}
	b24.items = []builder{&b17}
	var b165 = choiceBuilder{id: 165, commit: 66}
	var b37 = choiceBuilder{id: 37, commit: 66}
	var b87 = sequenceBuilder{id: 87, commit: 72, name: "any-char", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b176 = charBuilder{}
	b87.items = []builder{&b176}
	var b99 = sequenceBuilder{id: 99, commit: 72, name: "char-class", ranges: [][]int{{1, 1}, {0, 1}, {0, -1}, {1, 1}, {1, 1}, {0, 1}, {0, -1}, {1, 1}}}
	var b142 = sequenceBuilder{id: 142, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b51 = charBuilder{}
	b142.items = []builder{&b51}
	var b55 = sequenceBuilder{id: 55, commit: 72, name: "class-not", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b19 = charBuilder{}
	b55.items = []builder{&b19}
	var b158 = choiceBuilder{id: 158, commit: 10}
	var b28 = choiceBuilder{id: 28, commit: 72, name: "class-char"}
	var b65 = sequenceBuilder{id: 65, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b81 = charBuilder{}
	b65.items = []builder{&b81}
	var b178 = sequenceBuilder{id: 178, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b177 = sequenceBuilder{id: 177, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b157 = charBuilder{}
	b177.items = []builder{&b157}
	var b36 = sequenceBuilder{id: 36, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b82 = charBuilder{}
	b36.items = []builder{&b82}
	b178.items = []builder{&b177, &b36}
	b28.options = []builder{&b65, &b178}
	var b45 = sequenceBuilder{id: 45, commit: 72, name: "char-range", ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b133 = sequenceBuilder{id: 133, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b44 = charBuilder{}
	b133.items = []builder{&b44}
	b45.items = []builder{&b28, &b133, &b28}
	b158.options = []builder{&b28, &b45}
	var b61 = sequenceBuilder{id: 61, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b92 = charBuilder{}
	b61.items = []builder{&b92}
	b99.items = []builder{&b142, &b55, &b158, &b61}
	var b80 = sequenceBuilder{id: 80, commit: 72, name: "char-sequence", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {1, 1}, {0, -1}, {1, 1}}}
	var b107 = sequenceBuilder{id: 107, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b7 = charBuilder{}
	b107.items = []builder{&b7}
	var b46 = choiceBuilder{id: 46, commit: 72, name: "sequence-char"}
	var b134 = sequenceBuilder{id: 134, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b20 = charBuilder{}
	b134.items = []builder{&b20}
	var b169 = sequenceBuilder{id: 169, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b100 = sequenceBuilder{id: 100, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b38 = charBuilder{}
	b100.items = []builder{&b38}
	var b173 = sequenceBuilder{id: 173, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b62 = charBuilder{}
	b173.items = []builder{&b62}
	b169.items = []builder{&b100, &b173}
	b46.options = []builder{&b134, &b169}
	var b143 = sequenceBuilder{id: 143, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b93 = charBuilder{}
	b143.items = []builder{&b93}
	b80.items = []builder{&b107, &b46, &b143}
	b37.options = []builder{&b87, &b99, &b80}
	var b31 = sequenceBuilder{id: 31, commit: 66, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}}}
	var b184 = sequenceBuilder{id: 184, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b94 = charBuilder{}
	b184.items = []builder{&b94}
	var b95 = sequenceBuilder{id: 95, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b30 = charBuilder{}
	b95.items = []builder{&b30}
	b31.items = []builder{&b184, &b186, &b165, &b186, &b95}
	var b48 = sequenceBuilder{id: 48, commit: 64, name: "sequence", ranges: [][]int{{1, 1}, {0, -1}}}
	var b130 = sequenceBuilder{id: 130, commit: 72, name: "item", ranges: [][]int{{1, 1}, {0, 1}, {1, 1}, {0, 1}}}
	var b83 = choiceBuilder{id: 83, commit: 10}
	b83.options = []builder{&b37, &b162, &b31}
	var b180 = choiceBuilder{id: 180, commit: 66}
	var b149 = sequenceBuilder{id: 149, commit: 64, name: "count-quantifier", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}}}
	var b126 = sequenceBuilder{id: 126, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b74 = charBuilder{}
	b126.items = []builder{&b74}
	var b163 = sequenceBuilder{id: 163, commit: 64, name: "count", ranges: [][]int{{1, 1}}}
	var b108 = sequenceBuilder{id: 108, commit: 74, ranges: [][]int{{1, -1}, {1, -1}}}
	var b15 = sequenceBuilder{id: 15, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b144 = charBuilder{}
	b15.items = []builder{&b144}
	b108.items = []builder{&b15}
	b163.items = []builder{&b108}
	var b8 = sequenceBuilder{id: 8, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b9 = charBuilder{}
	b8.items = []builder{&b9}
	b149.items = []builder{&b126, &b186, &b163, &b186, &b8}
	var b23 = sequenceBuilder{id: 23, commit: 64, name: "range-quantifier", ranges: [][]int{{1, 1}, {0, -1}, {0, 1}, {0, -1}, {1, 1}, {0, -1}, {0, 1}, {0, -1}, {1, 1}}}
	var b63 = sequenceBuilder{id: 63, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b150 = charBuilder{}
	b63.items = []builder{&b150}
	var b127 = sequenceBuilder{id: 127, commit: 64, name: "range-from", ranges: [][]int{{1, 1}}}
	b127.items = []builder{&b108}
	var b96 = sequenceBuilder{id: 96, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b39 = charBuilder{}
	b96.items = []builder{&b39}
	var b129 = sequenceBuilder{id: 129, commit: 64, name: "range-to", ranges: [][]int{{1, 1}}}
	b129.items = []builder{&b108}
	var b101 = sequenceBuilder{id: 101, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b21 = charBuilder{}
	b101.items = []builder{&b21}
	b23.items = []builder{&b63, &b186, &b127, &b186, &b96, &b186, &b129, &b186, &b101}
	var b128 = sequenceBuilder{id: 128, commit: 72, name: "one-or-more", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b16 = charBuilder{}
	b128.items = []builder{&b16}
	var b57 = sequenceBuilder{id: 57, commit: 72, name: "zero-or-more", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b102 = charBuilder{}
	b57.items = []builder{&b102}
	var b179 = sequenceBuilder{id: 179, commit: 72, name: "zero-or-one", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b75 = charBuilder{}
	b179.items = []builder{&b75}
	b180.options = []builder{&b149, &b23, &b128, &b57, &b179}
	b130.items = []builder{&b83, &b180}
	var b47 = sequenceBuilder{id: 47, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b47.items = []builder{&b186, &b130}
	b48.items = []builder{&b130, &b47}
	var b77 = sequenceBuilder{id: 77, commit: 64, name: "choice", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}}
	var b103 = choiceBuilder{id: 103, commit: 66}
	b103.options = []builder{&b37, &b162, &b31, &b48}
	var b164 = sequenceBuilder{id: 164, commit: 2, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}}}
	var b115 = sequenceBuilder{id: 115, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b10 = charBuilder{}
	b115.items = []builder{&b10}
	b164.items = []builder{&b115, &b186, &b103}
	var b76 = sequenceBuilder{id: 76, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b76.items = []builder{&b186, &b164}
	b77.items = []builder{&b103, &b186, &b164, &b76}
	b165.options = []builder{&b37, &b162, &b31, &b48, &b77}
	b68.items = []builder{&b90, &b186, &b24, &b186, &b165}
	var b111 = sequenceBuilder{id: 111, commit: 2, ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var b154 = sequenceBuilder{id: 154, commit: 2, ranges: [][]int{{1, 1}, {0, -1}, {0, -1}, {1, 1}}}
	var b60 = sequenceBuilder{id: 60, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b33 = charBuilder{}
	b60.items = []builder{&b33}
	var b153 = sequenceBuilder{id: 153, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b153.items = []builder{&b186, &b60}
	b154.items = []builder{&b60, &b153, &b186, &b68}
	var b110 = sequenceBuilder{id: 110, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b110.items = []builder{&b186, &b154}
	b111.items = []builder{&b186, &b154, &b110}
	b112.items = []builder{&b68, &b111}
	var b72 = sequenceBuilder{id: 72, commit: 2, ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var b91 = sequenceBuilder{id: 91, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b4 = charBuilder{}
	b91.items = []builder{&b4}
	var b71 = sequenceBuilder{id: 71, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b71.items = []builder{&b186, &b91}
	b72.items = []builder{&b186, &b91, &b71}
	b187.items = []builder{&b70, &b186, &b112, &b72}
	b188.items = []builder{&b186, &b187, &b186}

	return parse(r, &p188, &b188)
}
