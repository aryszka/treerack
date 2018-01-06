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
	var p124 = choiceParser{id: 124, commit: 66, name: "wschar", generalizations: []int{185, 186}}
	var p38 = sequenceParser{id: 38, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{124, 185, 186}}
	var p160 = charParser{id: 160, chars: []rune{32}}
	p38.items = []parser{&p160}
	var p181 = sequenceParser{id: 181, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{124, 185, 186}}
	var p32 = charParser{id: 32, chars: []rune{9}}
	p181.items = []parser{&p32}
	var p75 = sequenceParser{id: 75, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{124, 185, 186}}
	var p92 = charParser{id: 92, chars: []rune{10}}
	p75.items = []parser{&p92}
	var p83 = sequenceParser{id: 83, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{124, 185, 186}}
	var p61 = charParser{id: 61, chars: []rune{8}}
	p83.items = []parser{&p61}
	var p131 = sequenceParser{id: 131, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{124, 185, 186}}
	var p139 = charParser{id: 139, chars: []rune{12}}
	p131.items = []parser{&p139}
	var p182 = sequenceParser{id: 182, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{124, 185, 186}}
	var p12 = charParser{id: 12, chars: []rune{13}}
	p182.items = []parser{&p12}
	var p165 = sequenceParser{id: 165, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{124, 185, 186}}
	var p118 = charParser{id: 118, chars: []rune{11}}
	p165.items = []parser{&p118}
	p124.options = []parser{&p38, &p181, &p75, &p83, &p131, &p182, &p165}
	var p78 = sequenceParser{id: 78, commit: 72, name: "comment", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}, generalizations: []int{185, 186}}
	var p62 = choiceParser{id: 62, commit: 74, name: "comment-segment"}
	var p33 = sequenceParser{id: 33, commit: 74, name: "line-comment", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}, generalizations: []int{62}}
	var p113 = sequenceParser{id: 113, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var p112 = charParser{id: 112, chars: []rune{47}}
	var p70 = charParser{id: 70, chars: []rune{47}}
	p113.items = []parser{&p112, &p70}
	var p47 = sequenceParser{id: 47, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p76 = charParser{id: 76, not: true, chars: []rune{10}}
	p47.items = []parser{&p76}
	p33.items = []parser{&p113, &p47}
	var p101 = sequenceParser{id: 101, commit: 74, name: "block-comment", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {1, 1}, {0, -1}, {1, 1}}, generalizations: []int{62}}
	var p151 = sequenceParser{id: 151, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var p84 = charParser{id: 84, chars: []rune{47}}
	var p68 = charParser{id: 68, chars: []rune{42}}
	p151.items = []parser{&p84, &p68}
	var p171 = choiceParser{id: 171, commit: 10}
	var p100 = sequenceParser{id: 100, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{171}}
	var p111 = sequenceParser{id: 111, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p93 = charParser{id: 93, chars: []rune{42}}
	p111.items = []parser{&p93}
	var p143 = sequenceParser{id: 143, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p99 = charParser{id: 99, not: true, chars: []rune{47}}
	p143.items = []parser{&p99}
	p100.items = []parser{&p111, &p143}
	var p148 = sequenceParser{id: 148, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{171}}
	var p166 = charParser{id: 166, not: true, chars: []rune{42}}
	p148.items = []parser{&p166}
	p171.options = []parser{&p100, &p148}
	var p46 = sequenceParser{id: 46, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var p19 = charParser{id: 19, chars: []rune{42}}
	var p69 = charParser{id: 69, chars: []rune{47}}
	p46.items = []parser{&p19, &p69}
	p101.items = []parser{&p151, &p171, &p46}
	p62.options = []parser{&p33, &p101}
	var p13 = sequenceParser{id: 13, commit: 10, ranges: [][]int{{0, -1}, {0, 1}, {0, -1}, {1, 1}, {0, -1}, {0, 1}, {0, -1}, {1, 1}}}
	var p125 = choiceParser{id: 125, commit: 74, name: "ws-no-nl"}
	var p26 = sequenceParser{id: 26, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{125}}
	var p77 = charParser{id: 77, chars: []rune{32}}
	p26.items = []parser{&p77}
	var p172 = sequenceParser{id: 172, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{125}}
	var p94 = charParser{id: 94, chars: []rune{9}}
	p172.items = []parser{&p94}
	var p152 = sequenceParser{id: 152, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{125}}
	var p102 = charParser{id: 102, chars: []rune{8}}
	p152.items = []parser{&p102}
	var p85 = sequenceParser{id: 85, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{125}}
	var p63 = charParser{id: 63, chars: []rune{12}}
	p85.items = []parser{&p63}
	var p167 = sequenceParser{id: 167, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{125}}
	var p149 = charParser{id: 149, chars: []rune{13}}
	p167.items = []parser{&p149}
	var p144 = sequenceParser{id: 144, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{125}}
	var p168 = charParser{id: 168, chars: []rune{11}}
	p144.items = []parser{&p168}
	p125.options = []parser{&p26, &p172, &p152, &p85, &p167, &p144}
	var p27 = sequenceParser{id: 27, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p95 = charParser{id: 95, chars: []rune{10}}
	p27.items = []parser{&p95}
	p13.items = []parser{&p125, &p27, &p125, &p62}
	p78.items = []parser{&p62, &p13}
	p185.options = []parser{&p124, &p78}
	p186.options = []parser{&p185}
	var p187 = sequenceParser{id: 187, commit: 66, name: "syntax:wsroot", ranges: [][]int{{0, 1}, {0, -1}, {0, 1}, {0, 1}}}
	var p8 = sequenceParser{id: 8, commit: 2, ranges: [][]int{{1, 1}, {0, -1}}}
	var p11 = sequenceParser{id: 11, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p31 = charParser{id: 31, chars: []rune{59}}
	p11.items = []parser{&p31}
	var p7 = sequenceParser{id: 7, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p7.items = []parser{&p186, &p11}
	p8.items = []parser{&p11, &p7}
	var p45 = sequenceParser{id: 45, commit: 66, name: "definitions", ranges: [][]int{{1, 1}, {0, 1}}}
	var p42 = sequenceParser{id: 42, commit: 64, name: "definition", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}}}
	var p87 = sequenceParser{id: 87, commit: 74, name: "definition-name", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}}
	var p88 = sequenceParser{id: 88, commit: 72, name: "symbol", ranges: [][]int{{1, -1}, {1, -1}}, generalizations: []int{64, 72, 155}}
	var p57 = sequenceParser{id: 57, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p14 = charParser{id: 14, not: true, chars: []rune{92, 32, 10, 9, 8, 12, 13, 11, 47, 46, 91, 93, 34, 123, 125, 94, 43, 42, 63, 124, 40, 41, 58, 61, 59}}
	p57.items = []parser{&p14}
	p88.items = []parser{&p57}
	var p158 = sequenceParser{id: 158, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var p82 = sequenceParser{id: 82, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p146 = charParser{id: 146, chars: []rune{58}}
	p82.items = []parser{&p146}
	var p4 = choiceParser{id: 4, commit: 66, name: "flag"}
	var p97 = sequenceParser{id: 97, commit: 72, name: "alias", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{4}}
	var p128 = charParser{id: 128, chars: []rune{97}}
	var p163 = charParser{id: 163, chars: []rune{108}}
	var p150 = charParser{id: 150, chars: []rune{105}}
	var p106 = charParser{id: 106, chars: []rune{97}}
	var p138 = charParser{id: 138, chars: []rune{115}}
	p97.items = []parser{&p128, &p163, &p150, &p106, &p138}
	var p54 = sequenceParser{id: 54, commit: 72, name: "ws", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{4}}
	var p67 = charParser{id: 67, chars: []rune{119}}
	var p41 = charParser{id: 41, chars: []rune{115}}
	p54.items = []parser{&p67, &p41}
	var p129 = sequenceParser{id: 129, commit: 72, name: "nows", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{4}}
	var p30 = charParser{id: 30, chars: []rune{110}}
	var p24 = charParser{id: 24, chars: []rune{111}}
	var p65 = charParser{id: 65, chars: []rune{119}}
	var p16 = charParser{id: 16, chars: []rune{115}}
	p129.items = []parser{&p30, &p24, &p65, &p16}
	var p110 = sequenceParser{id: 110, commit: 72, name: "failpass", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{4}}
	var p179 = charParser{id: 179, chars: []rune{102}}
	var p121 = charParser{id: 121, chars: []rune{97}}
	var p73 = charParser{id: 73, chars: []rune{105}}
	var p130 = charParser{id: 130, chars: []rune{108}}
	var p17 = charParser{id: 17, chars: []rune{112}}
	var p180 = charParser{id: 180, chars: []rune{97}}
	var p74 = charParser{id: 74, chars: []rune{115}}
	var p184 = charParser{id: 184, chars: []rune{115}}
	p110.items = []parser{&p179, &p121, &p73, &p130, &p17, &p180, &p74, &p184}
	var p18 = sequenceParser{id: 18, commit: 72, name: "root", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{4}}
	var p107 = charParser{id: 107, chars: []rune{114}}
	var p50 = charParser{id: 50, chars: []rune{111}}
	var p122 = charParser{id: 122, chars: []rune{111}}
	var p123 = charParser{id: 123, chars: []rune{116}}
	p18.items = []parser{&p107, &p50, &p122, &p123}
	p4.options = []parser{&p97, &p54, &p129, &p110, &p18}
	p158.items = []parser{&p82, &p4}
	p87.items = []parser{&p88, &p158}
	var p5 = sequenceParser{id: 5, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p159 = charParser{id: 159, chars: []rune{61}}
	p5.items = []parser{&p159}
	var p64 = choiceParser{id: 64, commit: 66, name: "expression"}
	var p104 = choiceParser{id: 104, commit: 66, name: "terminal", generalizations: []int{64, 72, 155}}
	var p103 = sequenceParser{id: 103, commit: 72, name: "any-char", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{104, 64, 72, 155}}
	var p114 = charParser{id: 114, chars: []rune{46}}
	p103.items = []parser{&p114}
	var p51 = sequenceParser{id: 51, commit: 72, name: "char-class", ranges: [][]int{{1, 1}, {0, 1}, {0, -1}, {1, 1}, {1, 1}, {0, 1}, {0, -1}, {1, 1}}, generalizations: []int{104, 64, 72, 155}}
	var p153 = sequenceParser{id: 153, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p115 = charParser{id: 115, chars: []rune{91}}
	p153.items = []parser{&p115}
	var p71 = sequenceParser{id: 71, commit: 72, name: "class-not", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p132 = charParser{id: 132, chars: []rune{94}}
	p71.items = []parser{&p132}
	var p116 = choiceParser{id: 116, commit: 10}
	var p48 = choiceParser{id: 48, commit: 72, name: "class-char", generalizations: []int{116}}
	var p156 = sequenceParser{id: 156, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{48, 116}}
	var p173 = charParser{id: 173, not: true, chars: []rune{92, 91, 93, 94, 45}}
	p156.items = []parser{&p173}
	var p1 = sequenceParser{id: 1, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{48, 116}}
	var p161 = sequenceParser{id: 161, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p55 = charParser{id: 55, chars: []rune{92}}
	p161.items = []parser{&p55}
	var p34 = sequenceParser{id: 34, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p133 = charParser{id: 133, not: true}
	p34.items = []parser{&p133}
	p1.items = []parser{&p161, &p34}
	p48.options = []parser{&p156, &p1}
	var p86 = sequenceParser{id: 86, commit: 72, name: "char-range", ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{116}}
	var p140 = sequenceParser{id: 140, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p39 = charParser{id: 39, chars: []rune{45}}
	p140.items = []parser{&p39}
	p86.items = []parser{&p48, &p140, &p48}
	p116.options = []parser{&p48, &p86}
	var p28 = sequenceParser{id: 28, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p169 = charParser{id: 169, chars: []rune{93}}
	p28.items = []parser{&p169}
	p51.items = []parser{&p153, &p71, &p116, &p28}
	var p36 = sequenceParser{id: 36, commit: 72, name: "char-sequence", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {1, 1}, {0, -1}, {1, 1}}, generalizations: []int{104, 64, 72, 155}}
	var p35 = sequenceParser{id: 35, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p80 = charParser{id: 80, chars: []rune{34}}
	p35.items = []parser{&p80}
	var p141 = choiceParser{id: 141, commit: 72, name: "sequence-char"}
	var p170 = sequenceParser{id: 170, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{141}}
	var p20 = charParser{id: 20, not: true, chars: []rune{92, 34}}
	p170.items = []parser{&p20}
	var p174 = sequenceParser{id: 174, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{141}}
	var p29 = sequenceParser{id: 29, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p79 = charParser{id: 79, chars: []rune{92}}
	p29.items = []parser{&p79}
	var p134 = sequenceParser{id: 134, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p56 = charParser{id: 56, not: true}
	p134.items = []parser{&p56}
	p174.items = []parser{&p29, &p134}
	p141.options = []parser{&p170, &p174}
	var p157 = sequenceParser{id: 157, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p2 = charParser{id: 2, chars: []rune{34}}
	p157.items = []parser{&p2}
	p36.items = []parser{&p35, &p141, &p157}
	p104.options = []parser{&p103, &p51, &p36}
	var p58 = sequenceParser{id: 58, commit: 66, name: "group", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}}, generalizations: []int{64, 72, 155}}
	var p49 = sequenceParser{id: 49, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p105 = charParser{id: 105, chars: []rune{40}}
	p49.items = []parser{&p105}
	var p89 = sequenceParser{id: 89, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p3 = charParser{id: 3, chars: []rune{41}}
	p89.items = []parser{&p3}
	p58.items = []parser{&p49, &p186, &p64, &p186, &p89}
	var p23 = sequenceParser{id: 23, commit: 64, name: "sequence", ranges: [][]int{{1, 1}, {0, -1}}, generalizations: []int{64, 155}}
	var p135 = sequenceParser{id: 135, commit: 72, name: "item", ranges: [][]int{{1, 1}, {0, 1}, {1, 1}, {0, 1}}}
	var p72 = choiceParser{id: 72, commit: 10}
	p72.options = []parser{&p104, &p88, &p58}
	var p60 = choiceParser{id: 60, commit: 66, name: "quantity"}
	var p117 = sequenceParser{id: 117, commit: 64, name: "count-quantifier", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}}, generalizations: []int{60}}
	var p162 = sequenceParser{id: 162, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p81 = charParser{id: 81, chars: []rune{123}}
	p162.items = []parser{&p81}
	var p175 = sequenceParser{id: 175, commit: 64, name: "count", ranges: [][]int{{1, 1}}}
	var p52 = sequenceParser{id: 52, commit: 74, name: "number", ranges: [][]int{{1, -1}, {1, -1}}}
	var p119 = sequenceParser{id: 119, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p126 = charParser{id: 126, ranges: [][]rune{{48, 57}}}
	p119.items = []parser{&p126}
	p52.items = []parser{&p119}
	p175.items = []parser{&p52}
	var p176 = sequenceParser{id: 176, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p142 = charParser{id: 142, chars: []rune{125}}
	p176.items = []parser{&p142}
	p117.items = []parser{&p162, &p186, &p175, &p186, &p176}
	var p154 = sequenceParser{id: 154, commit: 64, name: "range-quantifier", ranges: [][]int{{1, 1}, {0, -1}, {0, 1}, {0, -1}, {1, 1}, {0, -1}, {0, 1}, {0, -1}, {1, 1}}, generalizations: []int{60}}
	var p53 = sequenceParser{id: 53, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p127 = charParser{id: 127, chars: []rune{123}}
	p53.items = []parser{&p127}
	var p183 = sequenceParser{id: 183, commit: 64, name: "range-from", ranges: [][]int{{1, 1}}}
	p183.items = []parser{&p52}
	var p21 = sequenceParser{id: 21, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p25 = charParser{id: 25, chars: []rune{44}}
	p21.items = []parser{&p25}
	var p59 = sequenceParser{id: 59, commit: 64, name: "range-to", ranges: [][]int{{1, 1}}}
	p59.items = []parser{&p52}
	var p145 = sequenceParser{id: 145, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p40 = charParser{id: 40, chars: []rune{125}}
	p145.items = []parser{&p40}
	p154.items = []parser{&p53, &p186, &p183, &p186, &p21, &p186, &p59, &p186, &p145}
	var p66 = sequenceParser{id: 66, commit: 72, name: "one-or-more", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{60}}
	var p177 = charParser{id: 177, chars: []rune{43}}
	p66.items = []parser{&p177}
	var p37 = sequenceParser{id: 37, commit: 72, name: "zero-or-more", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{60}}
	var p96 = charParser{id: 96, chars: []rune{42}}
	p37.items = []parser{&p96}
	var p98 = sequenceParser{id: 98, commit: 72, name: "zero-or-one", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{60}}
	var p90 = charParser{id: 90, chars: []rune{63}}
	p98.items = []parser{&p90}
	p60.options = []parser{&p117, &p154, &p66, &p37, &p98}
	p135.items = []parser{&p72, &p60}
	var p22 = sequenceParser{id: 22, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p22.items = []parser{&p186, &p135}
	p23.items = []parser{&p135, &p22}
	var p137 = sequenceParser{id: 137, commit: 64, name: "choice", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}, generalizations: []int{64}}
	var p155 = choiceParser{id: 155, commit: 66, name: "option"}
	p155.options = []parser{&p104, &p88, &p58, &p23}
	var p120 = sequenceParser{id: 120, commit: 2, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}}}
	var p15 = sequenceParser{id: 15, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p178 = charParser{id: 178, chars: []rune{124}}
	p15.items = []parser{&p178}
	p120.items = []parser{&p15, &p186, &p155}
	var p136 = sequenceParser{id: 136, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p136.items = []parser{&p186, &p120}
	p137.items = []parser{&p155, &p186, &p120, &p136}
	p64.options = []parser{&p104, &p88, &p58, &p23, &p137}
	p42.items = []parser{&p87, &p186, &p5, &p186, &p64}
	var p44 = sequenceParser{id: 44, commit: 2, ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var p109 = sequenceParser{id: 109, commit: 2, ranges: [][]int{{1, 1}, {0, -1}, {0, -1}, {1, 1}}}
	var p91 = sequenceParser{id: 91, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p6 = charParser{id: 6, chars: []rune{59}}
	p91.items = []parser{&p6}
	var p108 = sequenceParser{id: 108, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p108.items = []parser{&p186, &p91}
	p109.items = []parser{&p91, &p108, &p186, &p42}
	var p43 = sequenceParser{id: 43, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p43.items = []parser{&p186, &p109}
	p44.items = []parser{&p186, &p109, &p43}
	p45.items = []parser{&p42, &p44}
	var p10 = sequenceParser{id: 10, commit: 2, ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var p147 = sequenceParser{id: 147, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p164 = charParser{id: 164, chars: []rune{59}}
	p147.items = []parser{&p164}
	var p9 = sequenceParser{id: 9, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p9.items = []parser{&p186, &p147}
	p10.items = []parser{&p186, &p147, &p9}
	p187.items = []parser{&p8, &p186, &p45, &p10}
	p188.items = []parser{&p186, &p187, &p186}
	var b188 = sequenceBuilder{id: 188, commit: 32, name: "syntax", ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var b186 = choiceBuilder{id: 186, commit: 2}
	var b185 = choiceBuilder{id: 185, commit: 70}
	var b124 = choiceBuilder{id: 124, commit: 66}
	var b38 = sequenceBuilder{id: 38, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b160 = charBuilder{}
	b38.items = []builder{&b160}
	var b181 = sequenceBuilder{id: 181, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b32 = charBuilder{}
	b181.items = []builder{&b32}
	var b75 = sequenceBuilder{id: 75, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b92 = charBuilder{}
	b75.items = []builder{&b92}
	var b83 = sequenceBuilder{id: 83, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b61 = charBuilder{}
	b83.items = []builder{&b61}
	var b131 = sequenceBuilder{id: 131, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b139 = charBuilder{}
	b131.items = []builder{&b139}
	var b182 = sequenceBuilder{id: 182, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b12 = charBuilder{}
	b182.items = []builder{&b12}
	var b165 = sequenceBuilder{id: 165, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b118 = charBuilder{}
	b165.items = []builder{&b118}
	b124.options = []builder{&b38, &b181, &b75, &b83, &b131, &b182, &b165}
	var b78 = sequenceBuilder{id: 78, commit: 72, name: "comment", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}}
	var b62 = choiceBuilder{id: 62, commit: 74}
	var b33 = sequenceBuilder{id: 33, commit: 74, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}}
	var b113 = sequenceBuilder{id: 113, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b112 = charBuilder{}
	var b70 = charBuilder{}
	b113.items = []builder{&b112, &b70}
	var b47 = sequenceBuilder{id: 47, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b76 = charBuilder{}
	b47.items = []builder{&b76}
	b33.items = []builder{&b113, &b47}
	var b101 = sequenceBuilder{id: 101, commit: 74, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {1, 1}, {0, -1}, {1, 1}}}
	var b151 = sequenceBuilder{id: 151, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b84 = charBuilder{}
	var b68 = charBuilder{}
	b151.items = []builder{&b84, &b68}
	var b171 = choiceBuilder{id: 171, commit: 10}
	var b100 = sequenceBuilder{id: 100, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b111 = sequenceBuilder{id: 111, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b93 = charBuilder{}
	b111.items = []builder{&b93}
	var b143 = sequenceBuilder{id: 143, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b99 = charBuilder{}
	b143.items = []builder{&b99}
	b100.items = []builder{&b111, &b143}
	var b148 = sequenceBuilder{id: 148, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b166 = charBuilder{}
	b148.items = []builder{&b166}
	b171.options = []builder{&b100, &b148}
	var b46 = sequenceBuilder{id: 46, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b19 = charBuilder{}
	var b69 = charBuilder{}
	b46.items = []builder{&b19, &b69}
	b101.items = []builder{&b151, &b171, &b46}
	b62.options = []builder{&b33, &b101}
	var b13 = sequenceBuilder{id: 13, commit: 10, ranges: [][]int{{0, -1}, {0, 1}, {0, -1}, {1, 1}, {0, -1}, {0, 1}, {0, -1}, {1, 1}}}
	var b125 = choiceBuilder{id: 125, commit: 74}
	var b26 = sequenceBuilder{id: 26, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b77 = charBuilder{}
	b26.items = []builder{&b77}
	var b172 = sequenceBuilder{id: 172, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b94 = charBuilder{}
	b172.items = []builder{&b94}
	var b152 = sequenceBuilder{id: 152, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b102 = charBuilder{}
	b152.items = []builder{&b102}
	var b85 = sequenceBuilder{id: 85, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b63 = charBuilder{}
	b85.items = []builder{&b63}
	var b167 = sequenceBuilder{id: 167, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b149 = charBuilder{}
	b167.items = []builder{&b149}
	var b144 = sequenceBuilder{id: 144, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b168 = charBuilder{}
	b144.items = []builder{&b168}
	b125.options = []builder{&b26, &b172, &b152, &b85, &b167, &b144}
	var b27 = sequenceBuilder{id: 27, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b95 = charBuilder{}
	b27.items = []builder{&b95}
	b13.items = []builder{&b125, &b27, &b125, &b62}
	b78.items = []builder{&b62, &b13}
	b185.options = []builder{&b124, &b78}
	b186.options = []builder{&b185}
	var b187 = sequenceBuilder{id: 187, commit: 66, ranges: [][]int{{0, 1}, {0, -1}, {0, 1}, {0, 1}}}
	var b8 = sequenceBuilder{id: 8, commit: 2, ranges: [][]int{{1, 1}, {0, -1}}}
	var b11 = sequenceBuilder{id: 11, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b31 = charBuilder{}
	b11.items = []builder{&b31}
	var b7 = sequenceBuilder{id: 7, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b7.items = []builder{&b186, &b11}
	b8.items = []builder{&b11, &b7}
	var b45 = sequenceBuilder{id: 45, commit: 66, ranges: [][]int{{1, 1}, {0, 1}}}
	var b42 = sequenceBuilder{id: 42, commit: 64, name: "definition", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}}}
	var b87 = sequenceBuilder{id: 87, commit: 74, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}}
	var b88 = sequenceBuilder{id: 88, commit: 72, name: "symbol", ranges: [][]int{{1, -1}, {1, -1}}}
	var b57 = sequenceBuilder{id: 57, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b14 = charBuilder{}
	b57.items = []builder{&b14}
	b88.items = []builder{&b57}
	var b158 = sequenceBuilder{id: 158, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b82 = sequenceBuilder{id: 82, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b146 = charBuilder{}
	b82.items = []builder{&b146}
	var b4 = choiceBuilder{id: 4, commit: 66}
	var b97 = sequenceBuilder{id: 97, commit: 72, name: "alias", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b128 = charBuilder{}
	var b163 = charBuilder{}
	var b150 = charBuilder{}
	var b106 = charBuilder{}
	var b138 = charBuilder{}
	b97.items = []builder{&b128, &b163, &b150, &b106, &b138}
	var b54 = sequenceBuilder{id: 54, commit: 72, name: "ws", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b67 = charBuilder{}
	var b41 = charBuilder{}
	b54.items = []builder{&b67, &b41}
	var b129 = sequenceBuilder{id: 129, commit: 72, name: "nows", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b30 = charBuilder{}
	var b24 = charBuilder{}
	var b65 = charBuilder{}
	var b16 = charBuilder{}
	b129.items = []builder{&b30, &b24, &b65, &b16}
	var b110 = sequenceBuilder{id: 110, commit: 72, name: "failpass", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b179 = charBuilder{}
	var b121 = charBuilder{}
	var b73 = charBuilder{}
	var b130 = charBuilder{}
	var b17 = charBuilder{}
	var b180 = charBuilder{}
	var b74 = charBuilder{}
	var b184 = charBuilder{}
	b110.items = []builder{&b179, &b121, &b73, &b130, &b17, &b180, &b74, &b184}
	var b18 = sequenceBuilder{id: 18, commit: 72, name: "root", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b107 = charBuilder{}
	var b50 = charBuilder{}
	var b122 = charBuilder{}
	var b123 = charBuilder{}
	b18.items = []builder{&b107, &b50, &b122, &b123}
	b4.options = []builder{&b97, &b54, &b129, &b110, &b18}
	b158.items = []builder{&b82, &b4}
	b87.items = []builder{&b88, &b158}
	var b5 = sequenceBuilder{id: 5, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b159 = charBuilder{}
	b5.items = []builder{&b159}
	var b64 = choiceBuilder{id: 64, commit: 66}
	var b104 = choiceBuilder{id: 104, commit: 66}
	var b103 = sequenceBuilder{id: 103, commit: 72, name: "any-char", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b114 = charBuilder{}
	b103.items = []builder{&b114}
	var b51 = sequenceBuilder{id: 51, commit: 72, name: "char-class", ranges: [][]int{{1, 1}, {0, 1}, {0, -1}, {1, 1}, {1, 1}, {0, 1}, {0, -1}, {1, 1}}}
	var b153 = sequenceBuilder{id: 153, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b115 = charBuilder{}
	b153.items = []builder{&b115}
	var b71 = sequenceBuilder{id: 71, commit: 72, name: "class-not", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b132 = charBuilder{}
	b71.items = []builder{&b132}
	var b116 = choiceBuilder{id: 116, commit: 10}
	var b48 = choiceBuilder{id: 48, commit: 72, name: "class-char"}
	var b156 = sequenceBuilder{id: 156, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b173 = charBuilder{}
	b156.items = []builder{&b173}
	var b1 = sequenceBuilder{id: 1, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b161 = sequenceBuilder{id: 161, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b55 = charBuilder{}
	b161.items = []builder{&b55}
	var b34 = sequenceBuilder{id: 34, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b133 = charBuilder{}
	b34.items = []builder{&b133}
	b1.items = []builder{&b161, &b34}
	b48.options = []builder{&b156, &b1}
	var b86 = sequenceBuilder{id: 86, commit: 72, name: "char-range", ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b140 = sequenceBuilder{id: 140, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b39 = charBuilder{}
	b140.items = []builder{&b39}
	b86.items = []builder{&b48, &b140, &b48}
	b116.options = []builder{&b48, &b86}
	var b28 = sequenceBuilder{id: 28, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b169 = charBuilder{}
	b28.items = []builder{&b169}
	b51.items = []builder{&b153, &b71, &b116, &b28}
	var b36 = sequenceBuilder{id: 36, commit: 72, name: "char-sequence", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {1, 1}, {0, -1}, {1, 1}}}
	var b35 = sequenceBuilder{id: 35, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b80 = charBuilder{}
	b35.items = []builder{&b80}
	var b141 = choiceBuilder{id: 141, commit: 72, name: "sequence-char"}
	var b170 = sequenceBuilder{id: 170, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b20 = charBuilder{}
	b170.items = []builder{&b20}
	var b174 = sequenceBuilder{id: 174, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b29 = sequenceBuilder{id: 29, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b79 = charBuilder{}
	b29.items = []builder{&b79}
	var b134 = sequenceBuilder{id: 134, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b56 = charBuilder{}
	b134.items = []builder{&b56}
	b174.items = []builder{&b29, &b134}
	b141.options = []builder{&b170, &b174}
	var b157 = sequenceBuilder{id: 157, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b2 = charBuilder{}
	b157.items = []builder{&b2}
	b36.items = []builder{&b35, &b141, &b157}
	b104.options = []builder{&b103, &b51, &b36}
	var b58 = sequenceBuilder{id: 58, commit: 66, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}}}
	var b49 = sequenceBuilder{id: 49, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b105 = charBuilder{}
	b49.items = []builder{&b105}
	var b89 = sequenceBuilder{id: 89, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b3 = charBuilder{}
	b89.items = []builder{&b3}
	b58.items = []builder{&b49, &b186, &b64, &b186, &b89}
	var b23 = sequenceBuilder{id: 23, commit: 64, name: "sequence", ranges: [][]int{{1, 1}, {0, -1}}}
	var b135 = sequenceBuilder{id: 135, commit: 72, name: "item", ranges: [][]int{{1, 1}, {0, 1}, {1, 1}, {0, 1}}}
	var b72 = choiceBuilder{id: 72, commit: 10}
	b72.options = []builder{&b104, &b88, &b58}
	var b60 = choiceBuilder{id: 60, commit: 66}
	var b117 = sequenceBuilder{id: 117, commit: 64, name: "count-quantifier", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}}}
	var b162 = sequenceBuilder{id: 162, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b81 = charBuilder{}
	b162.items = []builder{&b81}
	var b175 = sequenceBuilder{id: 175, commit: 64, name: "count", ranges: [][]int{{1, 1}}}
	var b52 = sequenceBuilder{id: 52, commit: 74, ranges: [][]int{{1, -1}, {1, -1}}}
	var b119 = sequenceBuilder{id: 119, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b126 = charBuilder{}
	b119.items = []builder{&b126}
	b52.items = []builder{&b119}
	b175.items = []builder{&b52}
	var b176 = sequenceBuilder{id: 176, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b142 = charBuilder{}
	b176.items = []builder{&b142}
	b117.items = []builder{&b162, &b186, &b175, &b186, &b176}
	var b154 = sequenceBuilder{id: 154, commit: 64, name: "range-quantifier", ranges: [][]int{{1, 1}, {0, -1}, {0, 1}, {0, -1}, {1, 1}, {0, -1}, {0, 1}, {0, -1}, {1, 1}}}
	var b53 = sequenceBuilder{id: 53, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b127 = charBuilder{}
	b53.items = []builder{&b127}
	var b183 = sequenceBuilder{id: 183, commit: 64, name: "range-from", ranges: [][]int{{1, 1}}}
	b183.items = []builder{&b52}
	var b21 = sequenceBuilder{id: 21, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b25 = charBuilder{}
	b21.items = []builder{&b25}
	var b59 = sequenceBuilder{id: 59, commit: 64, name: "range-to", ranges: [][]int{{1, 1}}}
	b59.items = []builder{&b52}
	var b145 = sequenceBuilder{id: 145, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b40 = charBuilder{}
	b145.items = []builder{&b40}
	b154.items = []builder{&b53, &b186, &b183, &b186, &b21, &b186, &b59, &b186, &b145}
	var b66 = sequenceBuilder{id: 66, commit: 72, name: "one-or-more", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b177 = charBuilder{}
	b66.items = []builder{&b177}
	var b37 = sequenceBuilder{id: 37, commit: 72, name: "zero-or-more", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b96 = charBuilder{}
	b37.items = []builder{&b96}
	var b98 = sequenceBuilder{id: 98, commit: 72, name: "zero-or-one", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b90 = charBuilder{}
	b98.items = []builder{&b90}
	b60.options = []builder{&b117, &b154, &b66, &b37, &b98}
	b135.items = []builder{&b72, &b60}
	var b22 = sequenceBuilder{id: 22, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b22.items = []builder{&b186, &b135}
	b23.items = []builder{&b135, &b22}
	var b137 = sequenceBuilder{id: 137, commit: 64, name: "choice", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}}
	var b155 = choiceBuilder{id: 155, commit: 66}
	b155.options = []builder{&b104, &b88, &b58, &b23}
	var b120 = sequenceBuilder{id: 120, commit: 2, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}}}
	var b15 = sequenceBuilder{id: 15, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b178 = charBuilder{}
	b15.items = []builder{&b178}
	b120.items = []builder{&b15, &b186, &b155}
	var b136 = sequenceBuilder{id: 136, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b136.items = []builder{&b186, &b120}
	b137.items = []builder{&b155, &b186, &b120, &b136}
	b64.options = []builder{&b104, &b88, &b58, &b23, &b137}
	b42.items = []builder{&b87, &b186, &b5, &b186, &b64}
	var b44 = sequenceBuilder{id: 44, commit: 2, ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var b109 = sequenceBuilder{id: 109, commit: 2, ranges: [][]int{{1, 1}, {0, -1}, {0, -1}, {1, 1}}}
	var b91 = sequenceBuilder{id: 91, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b6 = charBuilder{}
	b91.items = []builder{&b6}
	var b108 = sequenceBuilder{id: 108, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b108.items = []builder{&b186, &b91}
	b109.items = []builder{&b91, &b108, &b186, &b42}
	var b43 = sequenceBuilder{id: 43, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b43.items = []builder{&b186, &b109}
	b44.items = []builder{&b186, &b109, &b43}
	b45.items = []builder{&b42, &b44}
	var b10 = sequenceBuilder{id: 10, commit: 2, ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var b147 = sequenceBuilder{id: 147, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b164 = charBuilder{}
	b147.items = []builder{&b164}
	var b9 = sequenceBuilder{id: 9, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b9.items = []builder{&b186, &b147}
	b10.items = []builder{&b186, &b147, &b9}
	b187.items = []builder{&b8, &b186, &b45, &b10}
	b188.items = []builder{&b186, &b187, &b186}

	return parse(r, &p188, &b188)
}
