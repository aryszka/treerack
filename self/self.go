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
	var p38 = choiceParser{id: 38, commit: 66, name: "wschar", generalizations: []int{185, 186}}
	var p25 = sequenceParser{id: 25, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{38, 185, 186}}
	var p6 = charParser{id: 6, chars: []rune{32}}
	p25.items = []parser{&p6}
	var p26 = sequenceParser{id: 26, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{38, 185, 186}}
	var p110 = charParser{id: 110, chars: []rune{9}}
	p26.items = []parser{&p110}
	var p111 = sequenceParser{id: 111, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{38, 185, 186}}
	var p100 = charParser{id: 100, chars: []rune{10}}
	p111.items = []parser{&p100}
	var p136 = sequenceParser{id: 136, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{38, 185, 186}}
	var p142 = charParser{id: 142, chars: []rune{8}}
	p136.items = []parser{&p142}
	var p101 = sequenceParser{id: 101, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{38, 185, 186}}
	var p7 = charParser{id: 7, chars: []rune{12}}
	p101.items = []parser{&p7}
	var p147 = sequenceParser{id: 147, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{38, 185, 186}}
	var p170 = charParser{id: 170, chars: []rune{13}}
	p147.items = []parser{&p170}
	var p63 = sequenceParser{id: 63, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{38, 185, 186}}
	var p44 = charParser{id: 44, chars: []rune{11}}
	p63.items = []parser{&p44}
	p38.options = []parser{&p25, &p26, &p111, &p136, &p101, &p147, &p63}
	var p18 = sequenceParser{id: 18, commit: 72, name: "comment", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}, generalizations: []int{185, 186}}
	var p138 = choiceParser{id: 138, commit: 74, name: "comment-segment"}
	var p20 = sequenceParser{id: 20, commit: 74, name: "line-comment", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}, generalizations: []int{138}}
	var p124 = sequenceParser{id: 124, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var p182 = charParser{id: 182, chars: []rune{47}}
	var p143 = charParser{id: 143, chars: []rune{47}}
	p124.items = []parser{&p182, &p143}
	var p128 = sequenceParser{id: 128, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p77 = charParser{id: 77, not: true, chars: []rune{10}}
	p128.items = []parser{&p77}
	p20.items = []parser{&p124, &p128}
	var p76 = sequenceParser{id: 76, commit: 74, name: "block-comment", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {1, 1}, {0, -1}, {1, 1}}, generalizations: []int{138}}
	var p103 = sequenceParser{id: 103, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var p123 = charParser{id: 123, chars: []rune{47}}
	var p93 = charParser{id: 93, chars: []rune{42}}
	p103.items = []parser{&p123, &p93}
	var p8 = choiceParser{id: 8, commit: 10}
	var p81 = sequenceParser{id: 81, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{8}}
	var p27 = sequenceParser{id: 27, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p112 = charParser{id: 112, chars: []rune{42}}
	p27.items = []parser{&p112}
	var p173 = sequenceParser{id: 173, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p64 = charParser{id: 64, not: true, chars: []rune{47}}
	p173.items = []parser{&p64}
	p81.items = []parser{&p27, &p173}
	var p89 = sequenceParser{id: 89, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{8}}
	var p118 = charParser{id: 118, not: true, chars: []rune{42}}
	p89.items = []parser{&p118}
	p8.options = []parser{&p81, &p89}
	var p47 = sequenceParser{id: 47, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var p137 = charParser{id: 137, chars: []rune{42}}
	var p57 = charParser{id: 57, chars: []rune{47}}
	p47.items = []parser{&p137, &p57}
	p76.items = []parser{&p103, &p8, &p47}
	p138.options = []parser{&p20, &p76}
	var p82 = sequenceParser{id: 82, commit: 10, ranges: [][]int{{0, -1}, {0, 1}, {0, -1}, {1, 1}, {0, -1}, {0, 1}, {0, -1}, {1, 1}}}
	var p106 = choiceParser{id: 106, commit: 74, name: "ws-no-nl"}
	var p130 = sequenceParser{id: 130, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{106}}
	var p48 = charParser{id: 48, chars: []rune{32}}
	p130.items = []parser{&p48}
	var p131 = sequenceParser{id: 131, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{106}}
	var p65 = charParser{id: 65, chars: []rune{9}}
	p131.items = []parser{&p65}
	var p105 = sequenceParser{id: 105, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{106}}
	var p151 = charParser{id: 151, chars: []rune{8}}
	p105.items = []parser{&p151}
	var p17 = sequenceParser{id: 17, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{106}}
	var p90 = charParser{id: 90, chars: []rune{12}}
	p17.items = []parser{&p90}
	var p70 = sequenceParser{id: 70, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{106}}
	var p177 = charParser{id: 177, chars: []rune{13}}
	p70.items = []parser{&p177}
	var p126 = sequenceParser{id: 126, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{106}}
	var p45 = charParser{id: 45, chars: []rune{11}}
	p126.items = []parser{&p45}
	p106.options = []parser{&p130, &p131, &p105, &p17, &p70, &p126}
	var p107 = sequenceParser{id: 107, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p46 = charParser{id: 46, chars: []rune{10}}
	p107.items = []parser{&p46}
	p82.items = []parser{&p106, &p107, &p106, &p138}
	p18.items = []parser{&p138, &p82}
	p185.options = []parser{&p38, &p18}
	p186.options = []parser{&p185}
	var p187 = sequenceParser{id: 187, commit: 66, name: "syntax:wsroot", ranges: [][]int{{0, 1}, {0, -1}, {0, 1}, {0, 1}}}
	var p14 = sequenceParser{id: 14, commit: 2, ranges: [][]int{{1, 1}, {0, -1}}}
	var p117 = sequenceParser{id: 117, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p139 = charParser{id: 139, chars: []rune{59}}
	p117.items = []parser{&p139}
	var p13 = sequenceParser{id: 13, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p13.items = []parser{&p186, &p117}
	p14.items = []parser{&p117, &p13}
	var p159 = sequenceParser{id: 159, commit: 66, name: "definitions", ranges: [][]int{{1, 1}, {0, 1}}}
	var p52 = sequenceParser{id: 52, commit: 64, name: "definition", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}}}
	var p176 = sequenceParser{id: 176, commit: 74, name: "definition-name", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}}
	var p129 = sequenceParser{id: 129, commit: 72, name: "symbol", ranges: [][]int{{1, -1}, {1, -1}}, generalizations: []int{4, 40, 86}}
	var p163 = sequenceParser{id: 163, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p22 = charParser{id: 22, not: true, chars: []rune{92, 32, 10, 9, 8, 12, 13, 11, 47, 46, 91, 93, 34, 123, 125, 94, 43, 42, 63, 124, 40, 41, 58, 61, 59}}
	p163.items = []parser{&p22}
	p129.items = []parser{&p163}
	var p150 = sequenceParser{id: 150, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var p75 = sequenceParser{id: 75, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p30 = charParser{id: 30, chars: []rune{58}}
	p75.items = []parser{&p30}
	var p146 = choiceParser{id: 146, commit: 66, name: "flag"}
	var p116 = sequenceParser{id: 116, commit: 72, name: "alias", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{146}}
	var p69 = charParser{id: 69, chars: []rune{97}}
	var p5 = charParser{id: 5, chars: []rune{108}}
	var p127 = charParser{id: 127, chars: []rune{105}}
	var p62 = charParser{id: 62, chars: []rune{97}}
	var p97 = charParser{id: 97, chars: []rune{115}}
	p116.items = []parser{&p69, &p5, &p127, &p62, &p97}
	var p162 = sequenceParser{id: 162, commit: 72, name: "ws", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{146}}
	var p43 = charParser{id: 43, chars: []rune{119}}
	var p141 = charParser{id: 141, chars: []rune{115}}
	p162.items = []parser{&p43, &p141}
	var p145 = sequenceParser{id: 145, commit: 72, name: "nows", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{146}}
	var p83 = charParser{id: 83, chars: []rune{110}}
	var p98 = charParser{id: 98, chars: []rune{111}}
	var p50 = charParser{id: 50, chars: []rune{119}}
	var p74 = charParser{id: 74, chars: []rune{115}}
	p145.items = []parser{&p83, &p98, &p50, &p74}
	var p165 = sequenceParser{id: 165, commit: 72, name: "failpass", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{146}}
	var p87 = charParser{id: 87, chars: []rune{102}}
	var p29 = charParser{id: 29, chars: []rune{97}}
	var p9 = charParser{id: 9, chars: []rune{105}}
	var p51 = charParser{id: 51, chars: []rune{108}}
	var p37 = charParser{id: 37, chars: []rune{112}}
	var p104 = charParser{id: 104, chars: []rune{97}}
	var p99 = charParser{id: 99, chars: []rune{115}}
	var p154 = charParser{id: 154, chars: []rune{115}}
	p165.items = []parser{&p87, &p29, &p9, &p51, &p37, &p104, &p99, &p154}
	var p88 = sequenceParser{id: 88, commit: 72, name: "root", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{146}}
	var p79 = charParser{id: 79, chars: []rune{114}}
	var p135 = charParser{id: 135, chars: []rune{111}}
	var p114 = charParser{id: 114, chars: []rune{111}}
	var p181 = charParser{id: 181, chars: []rune{116}}
	p88.items = []parser{&p79, &p135, &p114, &p181}
	p146.options = []parser{&p116, &p162, &p145, &p165, &p88}
	p150.items = []parser{&p75, &p146}
	p176.items = []parser{&p129, &p150}
	var p11 = sequenceParser{id: 11, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p80 = charParser{id: 80, chars: []rune{61}}
	p11.items = []parser{&p80}
	var p4 = choiceParser{id: 4, commit: 66, name: "expression"}
	var p39 = choiceParser{id: 39, commit: 66, name: "terminal", generalizations: []int{4, 40, 86}}
	var p183 = sequenceParser{id: 183, commit: 72, name: "any-char", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{39, 4, 40, 86}}
	var p160 = charParser{id: 160, chars: []rune{46}}
	p183.items = []parser{&p160}
	var p184 = sequenceParser{id: 184, commit: 72, name: "char-class", ranges: [][]int{{1, 1}, {0, 1}, {0, -1}, {1, 1}, {1, 1}, {0, 1}, {0, -1}, {1, 1}}, generalizations: []int{39, 4, 40, 86}}
	var p35 = sequenceParser{id: 35, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p153 = charParser{id: 153, chars: []rune{91}}
	p35.items = []parser{&p153}
	var p91 = sequenceParser{id: 91, commit: 72, name: "class-not", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p178 = charParser{id: 178, chars: []rune{94}}
	p91.items = []parser{&p178}
	var p84 = choiceParser{id: 84, commit: 10}
	var p169 = choiceParser{id: 169, commit: 72, name: "class-char", generalizations: []int{84}}
	var p144 = sequenceParser{id: 144, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{169, 84}}
	var p21 = charParser{id: 21, not: true, chars: []rune{92, 91, 93, 94, 45}}
	p144.items = []parser{&p21}
	var p168 = sequenceParser{id: 168, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{169, 84}}
	var p34 = sequenceParser{id: 34, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p167 = charParser{id: 167, chars: []rune{92}}
	p34.items = []parser{&p167}
	var p125 = sequenceParser{id: 125, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p19 = charParser{id: 19, not: true}
	p125.items = []parser{&p19}
	p168.items = []parser{&p34, &p125}
	p169.options = []parser{&p144, &p168}
	var p152 = sequenceParser{id: 152, commit: 72, name: "char-range", ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{84}}
	var p71 = sequenceParser{id: 71, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p41 = charParser{id: 41, chars: []rune{45}}
	p71.items = []parser{&p41}
	p152.items = []parser{&p169, &p71, &p169}
	p84.options = []parser{&p169, &p152}
	var p58 = sequenceParser{id: 58, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p36 = charParser{id: 36, chars: []rune{93}}
	p58.items = []parser{&p36}
	p184.items = []parser{&p35, &p91, &p84, &p58}
	var p10 = sequenceParser{id: 10, commit: 72, name: "char-sequence", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {1, 1}, {0, -1}, {1, 1}}, generalizations: []int{39, 4, 40, 86}}
	var p120 = sequenceParser{id: 120, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p180 = charParser{id: 180, chars: []rune{34}}
	p120.items = []parser{&p180}
	var p171 = choiceParser{id: 171, commit: 72, name: "sequence-char"}
	var p66 = sequenceParser{id: 66, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{171}}
	var p108 = charParser{id: 108, not: true, chars: []rune{92, 34}}
	p66.items = []parser{&p108}
	var p113 = sequenceParser{id: 113, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{171}}
	var p109 = sequenceParser{id: 109, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p119 = charParser{id: 119, chars: []rune{92}}
	p109.items = []parser{&p119}
	var p121 = sequenceParser{id: 121, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p1 = charParser{id: 1, not: true}
	p121.items = []parser{&p1}
	p113.items = []parser{&p109, &p121}
	p171.options = []parser{&p66, &p113}
	var p78 = sequenceParser{id: 78, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p172 = charParser{id: 172, chars: []rune{34}}
	p78.items = []parser{&p172}
	p10.items = []parser{&p120, &p171, &p78}
	p39.options = []parser{&p183, &p184, &p10}
	var p23 = sequenceParser{id: 23, commit: 66, name: "group", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}}, generalizations: []int{4, 40, 86}}
	var p31 = sequenceParser{id: 31, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p60 = charParser{id: 60, chars: []rune{40}}
	p31.items = []parser{&p60}
	var p72 = sequenceParser{id: 72, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p174 = charParser{id: 174, chars: []rune{41}}
	p72.items = []parser{&p174}
	p23.items = []parser{&p31, &p186, &p4, &p186, &p72}
	var p68 = sequenceParser{id: 68, commit: 64, name: "sequence", ranges: [][]int{{1, 1}, {0, -1}}, generalizations: []int{4, 86}}
	var p115 = sequenceParser{id: 115, commit: 72, name: "item", ranges: [][]int{{1, 1}, {0, 1}, {1, 1}, {0, 1}}}
	var p40 = choiceParser{id: 40, commit: 10}
	p40.options = []parser{&p39, &p129, &p23}
	var p56 = choiceParser{id: 56, commit: 66, name: "quantity"}
	var p24 = sequenceParser{id: 24, commit: 64, name: "count-quantifier", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}}, generalizations: []int{56}}
	var p148 = sequenceParser{id: 148, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p55 = charParser{id: 55, chars: []rune{123}}
	p148.items = []parser{&p55}
	var p54 = sequenceParser{id: 54, commit: 64, name: "count", ranges: [][]int{{1, 1}}}
	var p73 = sequenceParser{id: 73, commit: 74, name: "number", ranges: [][]int{{1, -1}, {1, -1}}}
	var p164 = sequenceParser{id: 164, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p2 = charParser{id: 2, ranges: [][]rune{{48, 57}}}
	p164.items = []parser{&p2}
	p73.items = []parser{&p164}
	p54.items = []parser{&p73}
	var p140 = sequenceParser{id: 140, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p94 = charParser{id: 94, chars: []rune{125}}
	p140.items = []parser{&p94}
	p24.items = []parser{&p148, &p186, &p54, &p186, &p140}
	var p42 = sequenceParser{id: 42, commit: 64, name: "range-quantifier", ranges: [][]int{{1, 1}, {0, -1}, {0, 1}, {0, -1}, {1, 1}, {0, -1}, {0, 1}, {0, -1}, {1, 1}}, generalizations: []int{56}}
	var p102 = sequenceParser{id: 102, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p179 = charParser{id: 179, chars: []rune{123}}
	p102.items = []parser{&p179}
	var p59 = sequenceParser{id: 59, commit: 64, name: "range-from", ranges: [][]int{{1, 1}}}
	p59.items = []parser{&p73}
	var p28 = sequenceParser{id: 28, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p32 = charParser{id: 32, chars: []rune{44}}
	p28.items = []parser{&p32}
	var p61 = sequenceParser{id: 61, commit: 64, name: "range-to", ranges: [][]int{{1, 1}}}
	p61.items = []parser{&p73}
	var p33 = sequenceParser{id: 33, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p95 = charParser{id: 95, chars: []rune{125}}
	p33.items = []parser{&p95}
	p42.items = []parser{&p102, &p186, &p59, &p186, &p28, &p186, &p61, &p186, &p33}
	var p132 = sequenceParser{id: 132, commit: 72, name: "one-or-more", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{56}}
	var p122 = charParser{id: 122, chars: []rune{43}}
	p132.items = []parser{&p122}
	var p175 = sequenceParser{id: 175, commit: 72, name: "zero-or-more", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{56}}
	var p92 = charParser{id: 92, chars: []rune{42}}
	p175.items = []parser{&p92}
	var p161 = sequenceParser{id: 161, commit: 72, name: "zero-or-one", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{56}}
	var p85 = charParser{id: 85, chars: []rune{63}}
	p161.items = []parser{&p85}
	p56.options = []parser{&p24, &p42, &p132, &p175, &p161}
	p115.items = []parser{&p40, &p56}
	var p67 = sequenceParser{id: 67, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p67.items = []parser{&p186, &p115}
	p68.items = []parser{&p115, &p67}
	var p134 = sequenceParser{id: 134, commit: 64, name: "choice", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}, generalizations: []int{4}}
	var p86 = choiceParser{id: 86, commit: 66, name: "option"}
	p86.options = []parser{&p39, &p129, &p23, &p68}
	var p3 = sequenceParser{id: 3, commit: 2, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}}}
	var p149 = sequenceParser{id: 149, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p96 = charParser{id: 96, chars: []rune{124}}
	p149.items = []parser{&p96}
	p3.items = []parser{&p149, &p186, &p86}
	var p133 = sequenceParser{id: 133, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p133.items = []parser{&p186, &p3}
	p134.items = []parser{&p86, &p186, &p3, &p133}
	p4.options = []parser{&p39, &p129, &p23, &p68, &p134}
	p52.items = []parser{&p176, &p186, &p11, &p186, &p4}
	var p158 = sequenceParser{id: 158, commit: 2, ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var p156 = sequenceParser{id: 156, commit: 2, ranges: [][]int{{1, 1}, {0, -1}, {0, -1}, {1, 1}}}
	var p53 = sequenceParser{id: 53, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p166 = charParser{id: 166, chars: []rune{59}}
	p53.items = []parser{&p166}
	var p155 = sequenceParser{id: 155, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p155.items = []parser{&p186, &p53}
	p156.items = []parser{&p53, &p155, &p186, &p52}
	var p157 = sequenceParser{id: 157, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p157.items = []parser{&p186, &p156}
	p158.items = []parser{&p186, &p156, &p157}
	p159.items = []parser{&p52, &p158}
	var p16 = sequenceParser{id: 16, commit: 2, ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var p12 = sequenceParser{id: 12, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p49 = charParser{id: 49, chars: []rune{59}}
	p12.items = []parser{&p49}
	var p15 = sequenceParser{id: 15, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p15.items = []parser{&p186, &p12}
	p16.items = []parser{&p186, &p12, &p15}
	p187.items = []parser{&p14, &p186, &p159, &p16}
	p188.items = []parser{&p186, &p187, &p186}
	var b188 = sequenceBuilder{id: 188, commit: 32, name: "syntax", ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var b186 = choiceBuilder{id: 186, commit: 2}
	var b185 = choiceBuilder{id: 185, commit: 70}
	var b38 = choiceBuilder{id: 38, commit: 66}
	var b25 = sequenceBuilder{id: 25, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b6 = charBuilder{}
	b25.items = []builder{&b6}
	var b26 = sequenceBuilder{id: 26, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b110 = charBuilder{}
	b26.items = []builder{&b110}
	var b111 = sequenceBuilder{id: 111, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b100 = charBuilder{}
	b111.items = []builder{&b100}
	var b136 = sequenceBuilder{id: 136, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b142 = charBuilder{}
	b136.items = []builder{&b142}
	var b101 = sequenceBuilder{id: 101, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b7 = charBuilder{}
	b101.items = []builder{&b7}
	var b147 = sequenceBuilder{id: 147, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b170 = charBuilder{}
	b147.items = []builder{&b170}
	var b63 = sequenceBuilder{id: 63, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b44 = charBuilder{}
	b63.items = []builder{&b44}
	b38.options = []builder{&b25, &b26, &b111, &b136, &b101, &b147, &b63}
	var b18 = sequenceBuilder{id: 18, commit: 72, name: "comment", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}}
	var b138 = choiceBuilder{id: 138, commit: 74}
	var b20 = sequenceBuilder{id: 20, commit: 74, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}}
	var b124 = sequenceBuilder{id: 124, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b182 = charBuilder{}
	var b143 = charBuilder{}
	b124.items = []builder{&b182, &b143}
	var b128 = sequenceBuilder{id: 128, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b77 = charBuilder{}
	b128.items = []builder{&b77}
	b20.items = []builder{&b124, &b128}
	var b76 = sequenceBuilder{id: 76, commit: 74, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {1, 1}, {0, -1}, {1, 1}}}
	var b103 = sequenceBuilder{id: 103, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b123 = charBuilder{}
	var b93 = charBuilder{}
	b103.items = []builder{&b123, &b93}
	var b8 = choiceBuilder{id: 8, commit: 10}
	var b81 = sequenceBuilder{id: 81, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b27 = sequenceBuilder{id: 27, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b112 = charBuilder{}
	b27.items = []builder{&b112}
	var b173 = sequenceBuilder{id: 173, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b64 = charBuilder{}
	b173.items = []builder{&b64}
	b81.items = []builder{&b27, &b173}
	var b89 = sequenceBuilder{id: 89, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b118 = charBuilder{}
	b89.items = []builder{&b118}
	b8.options = []builder{&b81, &b89}
	var b47 = sequenceBuilder{id: 47, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b137 = charBuilder{}
	var b57 = charBuilder{}
	b47.items = []builder{&b137, &b57}
	b76.items = []builder{&b103, &b8, &b47}
	b138.options = []builder{&b20, &b76}
	var b82 = sequenceBuilder{id: 82, commit: 10, ranges: [][]int{{0, -1}, {0, 1}, {0, -1}, {1, 1}, {0, -1}, {0, 1}, {0, -1}, {1, 1}}}
	var b106 = choiceBuilder{id: 106, commit: 74}
	var b130 = sequenceBuilder{id: 130, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b48 = charBuilder{}
	b130.items = []builder{&b48}
	var b131 = sequenceBuilder{id: 131, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b65 = charBuilder{}
	b131.items = []builder{&b65}
	var b105 = sequenceBuilder{id: 105, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b151 = charBuilder{}
	b105.items = []builder{&b151}
	var b17 = sequenceBuilder{id: 17, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b90 = charBuilder{}
	b17.items = []builder{&b90}
	var b70 = sequenceBuilder{id: 70, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b177 = charBuilder{}
	b70.items = []builder{&b177}
	var b126 = sequenceBuilder{id: 126, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b45 = charBuilder{}
	b126.items = []builder{&b45}
	b106.options = []builder{&b130, &b131, &b105, &b17, &b70, &b126}
	var b107 = sequenceBuilder{id: 107, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b46 = charBuilder{}
	b107.items = []builder{&b46}
	b82.items = []builder{&b106, &b107, &b106, &b138}
	b18.items = []builder{&b138, &b82}
	b185.options = []builder{&b38, &b18}
	b186.options = []builder{&b185}
	var b187 = sequenceBuilder{id: 187, commit: 66, ranges: [][]int{{0, 1}, {0, -1}, {0, 1}, {0, 1}}}
	var b14 = sequenceBuilder{id: 14, commit: 2, ranges: [][]int{{1, 1}, {0, -1}}}
	var b117 = sequenceBuilder{id: 117, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b139 = charBuilder{}
	b117.items = []builder{&b139}
	var b13 = sequenceBuilder{id: 13, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b13.items = []builder{&b186, &b117}
	b14.items = []builder{&b117, &b13}
	var b159 = sequenceBuilder{id: 159, commit: 66, ranges: [][]int{{1, 1}, {0, 1}}}
	var b52 = sequenceBuilder{id: 52, commit: 64, name: "definition", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}}}
	var b176 = sequenceBuilder{id: 176, commit: 74, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}}
	var b129 = sequenceBuilder{id: 129, commit: 72, name: "symbol", ranges: [][]int{{1, -1}, {1, -1}}}
	var b163 = sequenceBuilder{id: 163, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b22 = charBuilder{}
	b163.items = []builder{&b22}
	b129.items = []builder{&b163}
	var b150 = sequenceBuilder{id: 150, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b75 = sequenceBuilder{id: 75, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b30 = charBuilder{}
	b75.items = []builder{&b30}
	var b146 = choiceBuilder{id: 146, commit: 66}
	var b116 = sequenceBuilder{id: 116, commit: 72, name: "alias", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b69 = charBuilder{}
	var b5 = charBuilder{}
	var b127 = charBuilder{}
	var b62 = charBuilder{}
	var b97 = charBuilder{}
	b116.items = []builder{&b69, &b5, &b127, &b62, &b97}
	var b162 = sequenceBuilder{id: 162, commit: 72, name: "ws", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b43 = charBuilder{}
	var b141 = charBuilder{}
	b162.items = []builder{&b43, &b141}
	var b145 = sequenceBuilder{id: 145, commit: 72, name: "nows", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b83 = charBuilder{}
	var b98 = charBuilder{}
	var b50 = charBuilder{}
	var b74 = charBuilder{}
	b145.items = []builder{&b83, &b98, &b50, &b74}
	var b165 = sequenceBuilder{id: 165, commit: 72, name: "failpass", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b87 = charBuilder{}
	var b29 = charBuilder{}
	var b9 = charBuilder{}
	var b51 = charBuilder{}
	var b37 = charBuilder{}
	var b104 = charBuilder{}
	var b99 = charBuilder{}
	var b154 = charBuilder{}
	b165.items = []builder{&b87, &b29, &b9, &b51, &b37, &b104, &b99, &b154}
	var b88 = sequenceBuilder{id: 88, commit: 72, name: "root", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b79 = charBuilder{}
	var b135 = charBuilder{}
	var b114 = charBuilder{}
	var b181 = charBuilder{}
	b88.items = []builder{&b79, &b135, &b114, &b181}
	b146.options = []builder{&b116, &b162, &b145, &b165, &b88}
	b150.items = []builder{&b75, &b146}
	b176.items = []builder{&b129, &b150}
	var b11 = sequenceBuilder{id: 11, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b80 = charBuilder{}
	b11.items = []builder{&b80}
	var b4 = choiceBuilder{id: 4, commit: 66}
	var b39 = choiceBuilder{id: 39, commit: 66}
	var b183 = sequenceBuilder{id: 183, commit: 72, name: "any-char", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b160 = charBuilder{}
	b183.items = []builder{&b160}
	var b184 = sequenceBuilder{id: 184, commit: 72, name: "char-class", ranges: [][]int{{1, 1}, {0, 1}, {0, -1}, {1, 1}, {1, 1}, {0, 1}, {0, -1}, {1, 1}}}
	var b35 = sequenceBuilder{id: 35, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b153 = charBuilder{}
	b35.items = []builder{&b153}
	var b91 = sequenceBuilder{id: 91, commit: 72, name: "class-not", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b178 = charBuilder{}
	b91.items = []builder{&b178}
	var b84 = choiceBuilder{id: 84, commit: 10}
	var b169 = choiceBuilder{id: 169, commit: 72, name: "class-char"}
	var b144 = sequenceBuilder{id: 144, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b21 = charBuilder{}
	b144.items = []builder{&b21}
	var b168 = sequenceBuilder{id: 168, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b34 = sequenceBuilder{id: 34, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b167 = charBuilder{}
	b34.items = []builder{&b167}
	var b125 = sequenceBuilder{id: 125, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b19 = charBuilder{}
	b125.items = []builder{&b19}
	b168.items = []builder{&b34, &b125}
	b169.options = []builder{&b144, &b168}
	var b152 = sequenceBuilder{id: 152, commit: 72, name: "char-range", ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b71 = sequenceBuilder{id: 71, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b41 = charBuilder{}
	b71.items = []builder{&b41}
	b152.items = []builder{&b169, &b71, &b169}
	b84.options = []builder{&b169, &b152}
	var b58 = sequenceBuilder{id: 58, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b36 = charBuilder{}
	b58.items = []builder{&b36}
	b184.items = []builder{&b35, &b91, &b84, &b58}
	var b10 = sequenceBuilder{id: 10, commit: 72, name: "char-sequence", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {1, 1}, {0, -1}, {1, 1}}}
	var b120 = sequenceBuilder{id: 120, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b180 = charBuilder{}
	b120.items = []builder{&b180}
	var b171 = choiceBuilder{id: 171, commit: 72, name: "sequence-char"}
	var b66 = sequenceBuilder{id: 66, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b108 = charBuilder{}
	b66.items = []builder{&b108}
	var b113 = sequenceBuilder{id: 113, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b109 = sequenceBuilder{id: 109, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b119 = charBuilder{}
	b109.items = []builder{&b119}
	var b121 = sequenceBuilder{id: 121, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b1 = charBuilder{}
	b121.items = []builder{&b1}
	b113.items = []builder{&b109, &b121}
	b171.options = []builder{&b66, &b113}
	var b78 = sequenceBuilder{id: 78, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b172 = charBuilder{}
	b78.items = []builder{&b172}
	b10.items = []builder{&b120, &b171, &b78}
	b39.options = []builder{&b183, &b184, &b10}
	var b23 = sequenceBuilder{id: 23, commit: 66, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}}}
	var b31 = sequenceBuilder{id: 31, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b60 = charBuilder{}
	b31.items = []builder{&b60}
	var b72 = sequenceBuilder{id: 72, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b174 = charBuilder{}
	b72.items = []builder{&b174}
	b23.items = []builder{&b31, &b186, &b4, &b186, &b72}
	var b68 = sequenceBuilder{id: 68, commit: 64, name: "sequence", ranges: [][]int{{1, 1}, {0, -1}}}
	var b115 = sequenceBuilder{id: 115, commit: 72, name: "item", ranges: [][]int{{1, 1}, {0, 1}, {1, 1}, {0, 1}}}
	var b40 = choiceBuilder{id: 40, commit: 10}
	b40.options = []builder{&b39, &b129, &b23}
	var b56 = choiceBuilder{id: 56, commit: 66}
	var b24 = sequenceBuilder{id: 24, commit: 64, name: "count-quantifier", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}}}
	var b148 = sequenceBuilder{id: 148, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b55 = charBuilder{}
	b148.items = []builder{&b55}
	var b54 = sequenceBuilder{id: 54, commit: 64, name: "count", ranges: [][]int{{1, 1}}}
	var b73 = sequenceBuilder{id: 73, commit: 74, ranges: [][]int{{1, -1}, {1, -1}}}
	var b164 = sequenceBuilder{id: 164, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b2 = charBuilder{}
	b164.items = []builder{&b2}
	b73.items = []builder{&b164}
	b54.items = []builder{&b73}
	var b140 = sequenceBuilder{id: 140, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b94 = charBuilder{}
	b140.items = []builder{&b94}
	b24.items = []builder{&b148, &b186, &b54, &b186, &b140}
	var b42 = sequenceBuilder{id: 42, commit: 64, name: "range-quantifier", ranges: [][]int{{1, 1}, {0, -1}, {0, 1}, {0, -1}, {1, 1}, {0, -1}, {0, 1}, {0, -1}, {1, 1}}}
	var b102 = sequenceBuilder{id: 102, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b179 = charBuilder{}
	b102.items = []builder{&b179}
	var b59 = sequenceBuilder{id: 59, commit: 64, name: "range-from", ranges: [][]int{{1, 1}}}
	b59.items = []builder{&b73}
	var b28 = sequenceBuilder{id: 28, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b32 = charBuilder{}
	b28.items = []builder{&b32}
	var b61 = sequenceBuilder{id: 61, commit: 64, name: "range-to", ranges: [][]int{{1, 1}}}
	b61.items = []builder{&b73}
	var b33 = sequenceBuilder{id: 33, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b95 = charBuilder{}
	b33.items = []builder{&b95}
	b42.items = []builder{&b102, &b186, &b59, &b186, &b28, &b186, &b61, &b186, &b33}
	var b132 = sequenceBuilder{id: 132, commit: 72, name: "one-or-more", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b122 = charBuilder{}
	b132.items = []builder{&b122}
	var b175 = sequenceBuilder{id: 175, commit: 72, name: "zero-or-more", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b92 = charBuilder{}
	b175.items = []builder{&b92}
	var b161 = sequenceBuilder{id: 161, commit: 72, name: "zero-or-one", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b85 = charBuilder{}
	b161.items = []builder{&b85}
	b56.options = []builder{&b24, &b42, &b132, &b175, &b161}
	b115.items = []builder{&b40, &b56}
	var b67 = sequenceBuilder{id: 67, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b67.items = []builder{&b186, &b115}
	b68.items = []builder{&b115, &b67}
	var b134 = sequenceBuilder{id: 134, commit: 64, name: "choice", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}}
	var b86 = choiceBuilder{id: 86, commit: 66}
	b86.options = []builder{&b39, &b129, &b23, &b68}
	var b3 = sequenceBuilder{id: 3, commit: 2, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}}}
	var b149 = sequenceBuilder{id: 149, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b96 = charBuilder{}
	b149.items = []builder{&b96}
	b3.items = []builder{&b149, &b186, &b86}
	var b133 = sequenceBuilder{id: 133, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b133.items = []builder{&b186, &b3}
	b134.items = []builder{&b86, &b186, &b3, &b133}
	b4.options = []builder{&b39, &b129, &b23, &b68, &b134}
	b52.items = []builder{&b176, &b186, &b11, &b186, &b4}
	var b158 = sequenceBuilder{id: 158, commit: 2, ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var b156 = sequenceBuilder{id: 156, commit: 2, ranges: [][]int{{1, 1}, {0, -1}, {0, -1}, {1, 1}}}
	var b53 = sequenceBuilder{id: 53, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b166 = charBuilder{}
	b53.items = []builder{&b166}
	var b155 = sequenceBuilder{id: 155, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b155.items = []builder{&b186, &b53}
	b156.items = []builder{&b53, &b155, &b186, &b52}
	var b157 = sequenceBuilder{id: 157, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b157.items = []builder{&b186, &b156}
	b158.items = []builder{&b186, &b156, &b157}
	b159.items = []builder{&b52, &b158}
	var b16 = sequenceBuilder{id: 16, commit: 2, ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var b12 = sequenceBuilder{id: 12, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b49 = charBuilder{}
	b12.items = []builder{&b49}
	var b15 = sequenceBuilder{id: 15, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b15.items = []builder{&b186, &b12}
	b16.items = []builder{&b186, &b12, &b15}
	b187.items = []builder{&b14, &b186, &b159, &b16}
	b188.items = []builder{&b186, &b187, &b186}

	return parseInput(r, &p188, &b188)
}
