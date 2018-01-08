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
	var p148 = choiceParser{id: 148, commit: 66, name: "wschar", generalizations: []int{185, 186}}
	var p109 = sequenceParser{id: 109, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{148, 185, 186}}
	var p52 = charParser{id: 52, chars: []rune{32}}
	p109.items = []parser{&p52}
	var p24 = sequenceParser{id: 24, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{148, 185, 186}}
	var p147 = charParser{id: 147, chars: []rune{9}}
	p24.items = []parser{&p147}
	var p15 = sequenceParser{id: 15, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{148, 185, 186}}
	var p77 = charParser{id: 77, chars: []rune{10}}
	p15.items = []parser{&p77}
	var p33 = sequenceParser{id: 33, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{148, 185, 186}}
	var p37 = charParser{id: 37, chars: []rune{8}}
	p33.items = []parser{&p37}
	var p45 = sequenceParser{id: 45, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{148, 185, 186}}
	var p129 = charParser{id: 129, chars: []rune{12}}
	p45.items = []parser{&p129}
	var p5 = sequenceParser{id: 5, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{148, 185, 186}}
	var p84 = charParser{id: 84, chars: []rune{13}}
	p5.items = []parser{&p84}
	var p85 = sequenceParser{id: 85, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{148, 185, 186}}
	var p25 = charParser{id: 25, chars: []rune{11}}
	p85.items = []parser{&p25}
	p148.options = []parser{&p109, &p24, &p15, &p33, &p45, &p5, &p85}
	var p103 = sequenceParser{id: 103, commit: 72, name: "comment", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}, generalizations: []int{185, 186}}
	var p110 = choiceParser{id: 110, commit: 74, name: "comment-segment"}
	var p34 = sequenceParser{id: 34, commit: 74, name: "line-comment", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}, generalizations: []int{110}}
	var p53 = sequenceParser{id: 53, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var p17 = charParser{id: 17, chars: []rune{47}}
	var p47 = charParser{id: 47, chars: []rune{47}}
	p53.items = []parser{&p17, &p47}
	var p119 = sequenceParser{id: 119, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p150 = charParser{id: 150, not: true, chars: []rune{10}}
	p119.items = []parser{&p150}
	p34.items = []parser{&p53, &p119}
	var p73 = sequenceParser{id: 73, commit: 74, name: "block-comment", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {1, 1}, {0, -1}, {1, 1}}, generalizations: []int{110}}
	var p46 = sequenceParser{id: 46, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var p130 = charParser{id: 130, chars: []rune{47}}
	var p101 = charParser{id: 101, chars: []rune{42}}
	p46.items = []parser{&p130, &p101}
	var p131 = choiceParser{id: 131, commit: 10}
	var p161 = sequenceParser{id: 161, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{131}}
	var p62 = sequenceParser{id: 62, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p61 = charParser{id: 61, chars: []rune{42}}
	p62.items = []parser{&p61}
	var p118 = sequenceParser{id: 118, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p72 = charParser{id: 72, not: true, chars: []rune{47}}
	p118.items = []parser{&p72}
	p161.items = []parser{&p62, &p118}
	var p78 = sequenceParser{id: 78, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{131}}
	var p95 = charParser{id: 95, not: true, chars: []rune{42}}
	p78.items = []parser{&p95}
	p131.options = []parser{&p161, &p78}
	var p140 = sequenceParser{id: 140, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var p96 = charParser{id: 96, chars: []rune{42}}
	var p16 = charParser{id: 16, chars: []rune{47}}
	p140.items = []parser{&p96, &p16}
	p73.items = []parser{&p46, &p131, &p140}
	p110.options = []parser{&p34, &p73}
	var p31 = sequenceParser{id: 31, commit: 10, ranges: [][]int{{0, -1}, {0, 1}, {0, -1}, {1, 1}, {0, -1}, {0, 1}, {0, -1}, {1, 1}}}
	var p111 = choiceParser{id: 111, commit: 74, name: "ws-no-nl"}
	var p26 = sequenceParser{id: 26, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{111}}
	var p182 = charParser{id: 182, chars: []rune{32}}
	p26.items = []parser{&p182}
	var p124 = sequenceParser{id: 124, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{111}}
	var p10 = charParser{id: 10, chars: []rune{9}}
	p124.items = []parser{&p10}
	var p30 = sequenceParser{id: 30, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{111}}
	var p29 = charParser{id: 29, chars: []rune{8}}
	p30.items = []parser{&p29}
	var p133 = sequenceParser{id: 133, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{111}}
	var p40 = charParser{id: 40, chars: []rune{12}}
	p133.items = []parser{&p40}
	var p102 = sequenceParser{id: 102, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{111}}
	var p41 = charParser{id: 41, chars: []rune{13}}
	p102.items = []parser{&p41}
	var p149 = sequenceParser{id: 149, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{111}}
	var p35 = charParser{id: 35, chars: []rune{11}}
	p149.items = []parser{&p35}
	p111.options = []parser{&p26, &p124, &p30, &p133, &p102, &p149}
	var p141 = sequenceParser{id: 141, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p162 = charParser{id: 162, chars: []rune{10}}
	p141.items = []parser{&p162}
	p31.items = []parser{&p111, &p141, &p111, &p110}
	p103.items = []parser{&p110, &p31}
	p185.options = []parser{&p148, &p103}
	p186.options = []parser{&p185}
	var p187 = sequenceParser{id: 187, commit: 66, name: "syntax:wsroot", ranges: [][]int{{0, 1}, {0, -1}, {0, 1}, {0, 1}}}
	var p164 = sequenceParser{id: 164, commit: 2, ranges: [][]int{{1, 1}, {0, -1}}}
	var p181 = sequenceParser{id: 181, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p154 = charParser{id: 154, chars: []rune{59}}
	p181.items = []parser{&p154}
	var p163 = sequenceParser{id: 163, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p163.items = []parser{&p186, &p181}
	p164.items = []parser{&p181, &p163}
	var p117 = sequenceParser{id: 117, commit: 66, name: "definitions", ranges: [][]int{{1, 1}, {0, 1}}}
	var p18 = sequenceParser{id: 18, commit: 64, name: "definition", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}}}
	var p114 = sequenceParser{id: 114, commit: 74, name: "definition-name", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}}
	var p93 = sequenceParser{id: 93, commit: 72, name: "symbol", ranges: [][]int{{1, -1}, {1, -1}}, generalizations: []int{59, 88, 39}}
	var p80 = sequenceParser{id: 80, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p107 = charParser{id: 107, not: true, chars: []rune{92, 32, 10, 9, 8, 12, 13, 11, 47, 46, 91, 93, 34, 123, 125, 94, 43, 42, 63, 124, 40, 41, 58, 61, 59}}
	p80.items = []parser{&p107}
	p93.items = []parser{&p80}
	var p83 = sequenceParser{id: 83, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var p71 = sequenceParser{id: 71, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p70 = charParser{id: 70, chars: []rune{58}}
	p71.items = []parser{&p70}
	var p4 = choiceParser{id: 4, commit: 66, name: "flag"}
	var p56 = sequenceParser{id: 56, commit: 72, name: "alias", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{4}}
	var p134 = charParser{id: 134, chars: []rune{97}}
	var p64 = charParser{id: 64, chars: []rune{108}}
	var p170 = charParser{id: 170, chars: []rune{105}}
	var p99 = charParser{id: 99, chars: []rune{97}}
	var p123 = charParser{id: 123, chars: []rune{115}}
	p56.items = []parser{&p134, &p64, &p170, &p99, &p123}
	var p22 = sequenceParser{id: 22, commit: 72, name: "ws", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{4}}
	var p184 = charParser{id: 184, chars: []rune{119}}
	var p174 = charParser{id: 174, chars: []rune{115}}
	p22.items = []parser{&p184, &p174}
	var p57 = sequenceParser{id: 57, commit: 72, name: "nows", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{4}}
	var p23 = charParser{id: 23, chars: []rune{110}}
	var p135 = charParser{id: 135, chars: []rune{111}}
	var p159 = charParser{id: 159, chars: []rune{119}}
	var p54 = charParser{id: 54, chars: []rune{115}}
	p57.items = []parser{&p23, &p135, &p159, &p54}
	var p51 = sequenceParser{id: 51, commit: 72, name: "failpass", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{4}}
	var p50 = charParser{id: 50, chars: []rune{102}}
	var p60 = charParser{id: 60, chars: []rune{97}}
	var p178 = charParser{id: 178, chars: []rune{105}}
	var p89 = charParser{id: 89, chars: []rune{108}}
	var p69 = charParser{id: 69, chars: []rune{112}}
	var p175 = charParser{id: 175, chars: []rune{97}}
	var p138 = charParser{id: 138, chars: []rune{115}}
	var p113 = charParser{id: 113, chars: []rune{115}}
	p51.items = []parser{&p50, &p60, &p178, &p89, &p69, &p175, &p138, &p113}
	var p28 = sequenceParser{id: 28, commit: 72, name: "root", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{4}}
	var p9 = charParser{id: 9, chars: []rune{114}}
	var p100 = charParser{id: 100, chars: []rune{111}}
	var p160 = charParser{id: 160, chars: []rune{111}}
	var p76 = charParser{id: 76, chars: []rune{116}}
	p28.items = []parser{&p9, &p100, &p160, &p76}
	p4.options = []parser{&p56, &p22, &p57, &p51, &p28}
	p83.items = []parser{&p71, &p4}
	p114.items = []parser{&p93, &p83}
	var p49 = sequenceParser{id: 49, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p132 = charParser{id: 132, chars: []rune{61}}
	p49.items = []parser{&p132}
	var p59 = choiceParser{id: 59, commit: 66, name: "expression"}
	var p136 = choiceParser{id: 136, commit: 66, name: "terminal", generalizations: []int{59, 88, 39}}
	var p6 = sequenceParser{id: 6, commit: 72, name: "any-char", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{136, 59, 88, 39}}
	var p104 = charParser{id: 104, chars: []rune{46}}
	p6.items = []parser{&p104}
	var p75 = sequenceParser{id: 75, commit: 72, name: "char-class", ranges: [][]int{{1, 1}, {0, 1}, {0, -1}, {1, 1}, {1, 1}, {0, 1}, {0, -1}, {1, 1}}, generalizations: []int{136, 59, 88, 39}}
	var p86 = sequenceParser{id: 86, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p105 = charParser{id: 105, chars: []rune{91}}
	p86.items = []parser{&p105}
	var p79 = sequenceParser{id: 79, commit: 72, name: "class-not", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p171 = charParser{id: 171, chars: []rune{94}}
	p79.items = []parser{&p171}
	var p125 = choiceParser{id: 125, commit: 10}
	var p157 = choiceParser{id: 157, commit: 72, name: "class-char", generalizations: []int{125}}
	var p7 = sequenceParser{id: 7, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{157, 125}}
	var p90 = charParser{id: 90, not: true, chars: []rune{92, 91, 93, 94, 45}}
	p7.items = []parser{&p90}
	var p173 = sequenceParser{id: 173, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{157, 125}}
	var p172 = sequenceParser{id: 172, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p155 = charParser{id: 155, chars: []rune{92}}
	p172.items = []parser{&p155}
	var p183 = sequenceParser{id: 183, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p156 = charParser{id: 156, not: true}
	p183.items = []parser{&p156}
	p173.items = []parser{&p172, &p183}
	p157.options = []parser{&p7, &p173}
	var p91 = sequenceParser{id: 91, commit: 72, name: "char-range", ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{125}}
	var p120 = sequenceParser{id: 120, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p74 = charParser{id: 74, chars: []rune{45}}
	p120.items = []parser{&p74}
	p91.items = []parser{&p157, &p120, &p157}
	p125.options = []parser{&p157, &p91}
	var p1 = sequenceParser{id: 1, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p112 = charParser{id: 112, chars: []rune{93}}
	p1.items = []parser{&p112}
	p75.items = []parser{&p86, &p79, &p125, &p1}
	var p2 = sequenceParser{id: 2, commit: 72, name: "char-sequence", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {1, 1}, {0, -1}, {1, 1}}, generalizations: []int{136, 59, 88, 39}}
	var p127 = sequenceParser{id: 127, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p97 = charParser{id: 97, chars: []rune{34}}
	p127.items = []parser{&p97}
	var p43 = choiceParser{id: 43, commit: 72, name: "sequence-char"}
	var p142 = sequenceParser{id: 142, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{43}}
	var p92 = charParser{id: 92, not: true, chars: []rune{92, 34}}
	p142.items = []parser{&p92}
	var p19 = sequenceParser{id: 19, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{43}}
	var p42 = sequenceParser{id: 42, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p38 = charParser{id: 38, chars: []rune{92}}
	p42.items = []parser{&p38}
	var p87 = sequenceParser{id: 87, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p106 = charParser{id: 106, not: true}
	p87.items = []parser{&p106}
	p19.items = []parser{&p42, &p87}
	p43.options = []parser{&p142, &p19}
	var p11 = sequenceParser{id: 11, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p20 = charParser{id: 20, chars: []rune{34}}
	p11.items = []parser{&p20}
	p2.items = []parser{&p127, &p43, &p11}
	p136.options = []parser{&p6, &p75, &p2}
	var p167 = sequenceParser{id: 167, commit: 66, name: "group", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}}, generalizations: []int{59, 88, 39}}
	var p126 = sequenceParser{id: 126, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p151 = charParser{id: 151, chars: []rune{40}}
	p126.items = []parser{&p151}
	var p12 = sequenceParser{id: 12, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p176 = charParser{id: 176, chars: []rune{41}}
	p12.items = []parser{&p176}
	p167.items = []parser{&p126, &p186, &p59, &p186, &p12}
	var p146 = sequenceParser{id: 146, commit: 64, name: "sequence", ranges: [][]int{{1, 1}, {0, -1}}, generalizations: []int{59, 39}}
	var p68 = sequenceParser{id: 68, commit: 72, name: "item", ranges: [][]int{{1, 1}, {0, 1}, {1, 1}, {0, 1}}}
	var p88 = choiceParser{id: 88, commit: 10}
	p88.options = []parser{&p136, &p93, &p167}
	var p122 = choiceParser{id: 122, commit: 66, name: "quantity"}
	var p108 = sequenceParser{id: 108, commit: 64, name: "count-quantifier", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}}, generalizations: []int{122}}
	var p44 = sequenceParser{id: 44, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p128 = charParser{id: 128, chars: []rune{123}}
	p44.items = []parser{&p128}
	var p27 = sequenceParser{id: 27, commit: 64, name: "count", ranges: [][]int{{1, 1}}}
	var p98 = sequenceParser{id: 98, commit: 74, name: "number", ranges: [][]int{{1, -1}, {1, -1}}}
	var p121 = sequenceParser{id: 121, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p143 = charParser{id: 143, ranges: [][]rune{{48, 57}}}
	p121.items = []parser{&p143}
	p98.items = []parser{&p121}
	p27.items = []parser{&p98}
	var p63 = sequenceParser{id: 63, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p55 = charParser{id: 55, chars: []rune{125}}
	p63.items = []parser{&p55}
	p108.items = []parser{&p44, &p186, &p27, &p186, &p63}
	var p81 = sequenceParser{id: 81, commit: 64, name: "range-quantifier", ranges: [][]int{{1, 1}, {0, -1}, {0, 1}, {0, -1}, {1, 1}, {0, -1}, {0, 1}, {0, -1}, {1, 1}}, generalizations: []int{122}}
	var p21 = sequenceParser{id: 21, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p8 = charParser{id: 8, chars: []rune{123}}
	p21.items = []parser{&p8}
	var p65 = sequenceParser{id: 65, commit: 64, name: "range-from", ranges: [][]int{{1, 1}}}
	p65.items = []parser{&p98}
	var p58 = sequenceParser{id: 58, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p153 = charParser{id: 153, chars: []rune{44}}
	p58.items = []parser{&p153}
	var p152 = sequenceParser{id: 152, commit: 64, name: "range-to", ranges: [][]int{{1, 1}}}
	p152.items = []parser{&p98}
	var p48 = sequenceParser{id: 48, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p158 = charParser{id: 158, chars: []rune{125}}
	p48.items = []parser{&p158}
	p81.items = []parser{&p21, &p186, &p65, &p186, &p58, &p186, &p152, &p186, &p48}
	var p82 = sequenceParser{id: 82, commit: 72, name: "one-or-more", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{122}}
	var p144 = charParser{id: 144, chars: []rune{43}}
	p82.items = []parser{&p144}
	var p13 = sequenceParser{id: 13, commit: 72, name: "zero-or-more", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{122}}
	var p137 = charParser{id: 137, chars: []rune{42}}
	p13.items = []parser{&p137}
	var p3 = sequenceParser{id: 3, commit: 72, name: "zero-or-one", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{122}}
	var p14 = charParser{id: 14, chars: []rune{63}}
	p3.items = []parser{&p14}
	p122.options = []parser{&p108, &p81, &p82, &p13, &p3}
	p68.items = []parser{&p88, &p122}
	var p145 = sequenceParser{id: 145, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p145.items = []parser{&p186, &p68}
	p146.items = []parser{&p68, &p145}
	var p169 = sequenceParser{id: 169, commit: 64, name: "choice", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}, generalizations: []int{59}}
	var p39 = choiceParser{id: 39, commit: 66, name: "option"}
	p39.options = []parser{&p136, &p93, &p167, &p146}
	var p177 = sequenceParser{id: 177, commit: 2, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}}}
	var p67 = sequenceParser{id: 67, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p66 = charParser{id: 66, chars: []rune{124}}
	p67.items = []parser{&p66}
	p177.items = []parser{&p67, &p186, &p39}
	var p168 = sequenceParser{id: 168, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p168.items = []parser{&p186, &p177}
	p169.items = []parser{&p39, &p186, &p177, &p168}
	p59.options = []parser{&p136, &p93, &p167, &p146, &p169}
	p18.items = []parser{&p114, &p186, &p49, &p186, &p59}
	var p116 = sequenceParser{id: 116, commit: 2, ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var p180 = sequenceParser{id: 180, commit: 2, ranges: [][]int{{1, 1}, {0, -1}, {0, -1}, {1, 1}}}
	var p36 = sequenceParser{id: 36, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p32 = charParser{id: 32, chars: []rune{59}}
	p36.items = []parser{&p32}
	var p179 = sequenceParser{id: 179, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p179.items = []parser{&p186, &p36}
	p180.items = []parser{&p36, &p179, &p186, &p18}
	var p115 = sequenceParser{id: 115, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p115.items = []parser{&p186, &p180}
	p116.items = []parser{&p186, &p180, &p115}
	p117.items = []parser{&p18, &p116}
	var p166 = sequenceParser{id: 166, commit: 2, ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var p139 = sequenceParser{id: 139, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p94 = charParser{id: 94, chars: []rune{59}}
	p139.items = []parser{&p94}
	var p165 = sequenceParser{id: 165, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p165.items = []parser{&p186, &p139}
	p166.items = []parser{&p186, &p139, &p165}
	p187.items = []parser{&p164, &p186, &p117, &p166}
	p188.items = []parser{&p186, &p187, &p186}
	var b188 = sequenceBuilder{id: 188, commit: 32, name: "syntax", ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var b186 = choiceBuilder{id: 186, commit: 2}
	var b185 = choiceBuilder{id: 185, commit: 70}
	var b148 = choiceBuilder{id: 148, commit: 66}
	var b109 = sequenceBuilder{id: 109, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b52 = charBuilder{}
	b109.items = []builder{&b52}
	var b24 = sequenceBuilder{id: 24, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b147 = charBuilder{}
	b24.items = []builder{&b147}
	var b15 = sequenceBuilder{id: 15, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b77 = charBuilder{}
	b15.items = []builder{&b77}
	var b33 = sequenceBuilder{id: 33, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b37 = charBuilder{}
	b33.items = []builder{&b37}
	var b45 = sequenceBuilder{id: 45, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b129 = charBuilder{}
	b45.items = []builder{&b129}
	var b5 = sequenceBuilder{id: 5, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b84 = charBuilder{}
	b5.items = []builder{&b84}
	var b85 = sequenceBuilder{id: 85, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b25 = charBuilder{}
	b85.items = []builder{&b25}
	b148.options = []builder{&b109, &b24, &b15, &b33, &b45, &b5, &b85}
	var b103 = sequenceBuilder{id: 103, commit: 72, name: "comment", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}}
	var b110 = choiceBuilder{id: 110, commit: 74}
	var b34 = sequenceBuilder{id: 34, commit: 74, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}}
	var b53 = sequenceBuilder{id: 53, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b17 = charBuilder{}
	var b47 = charBuilder{}
	b53.items = []builder{&b17, &b47}
	var b119 = sequenceBuilder{id: 119, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b150 = charBuilder{}
	b119.items = []builder{&b150}
	b34.items = []builder{&b53, &b119}
	var b73 = sequenceBuilder{id: 73, commit: 74, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {1, 1}, {0, -1}, {1, 1}}}
	var b46 = sequenceBuilder{id: 46, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b130 = charBuilder{}
	var b101 = charBuilder{}
	b46.items = []builder{&b130, &b101}
	var b131 = choiceBuilder{id: 131, commit: 10}
	var b161 = sequenceBuilder{id: 161, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b62 = sequenceBuilder{id: 62, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b61 = charBuilder{}
	b62.items = []builder{&b61}
	var b118 = sequenceBuilder{id: 118, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b72 = charBuilder{}
	b118.items = []builder{&b72}
	b161.items = []builder{&b62, &b118}
	var b78 = sequenceBuilder{id: 78, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b95 = charBuilder{}
	b78.items = []builder{&b95}
	b131.options = []builder{&b161, &b78}
	var b140 = sequenceBuilder{id: 140, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b96 = charBuilder{}
	var b16 = charBuilder{}
	b140.items = []builder{&b96, &b16}
	b73.items = []builder{&b46, &b131, &b140}
	b110.options = []builder{&b34, &b73}
	var b31 = sequenceBuilder{id: 31, commit: 10, ranges: [][]int{{0, -1}, {0, 1}, {0, -1}, {1, 1}, {0, -1}, {0, 1}, {0, -1}, {1, 1}}}
	var b111 = choiceBuilder{id: 111, commit: 74}
	var b26 = sequenceBuilder{id: 26, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b182 = charBuilder{}
	b26.items = []builder{&b182}
	var b124 = sequenceBuilder{id: 124, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b10 = charBuilder{}
	b124.items = []builder{&b10}
	var b30 = sequenceBuilder{id: 30, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b29 = charBuilder{}
	b30.items = []builder{&b29}
	var b133 = sequenceBuilder{id: 133, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b40 = charBuilder{}
	b133.items = []builder{&b40}
	var b102 = sequenceBuilder{id: 102, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b41 = charBuilder{}
	b102.items = []builder{&b41}
	var b149 = sequenceBuilder{id: 149, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b35 = charBuilder{}
	b149.items = []builder{&b35}
	b111.options = []builder{&b26, &b124, &b30, &b133, &b102, &b149}
	var b141 = sequenceBuilder{id: 141, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b162 = charBuilder{}
	b141.items = []builder{&b162}
	b31.items = []builder{&b111, &b141, &b111, &b110}
	b103.items = []builder{&b110, &b31}
	b185.options = []builder{&b148, &b103}
	b186.options = []builder{&b185}
	var b187 = sequenceBuilder{id: 187, commit: 66, ranges: [][]int{{0, 1}, {0, -1}, {0, 1}, {0, 1}}}
	var b164 = sequenceBuilder{id: 164, commit: 2, ranges: [][]int{{1, 1}, {0, -1}}}
	var b181 = sequenceBuilder{id: 181, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b154 = charBuilder{}
	b181.items = []builder{&b154}
	var b163 = sequenceBuilder{id: 163, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b163.items = []builder{&b186, &b181}
	b164.items = []builder{&b181, &b163}
	var b117 = sequenceBuilder{id: 117, commit: 66, ranges: [][]int{{1, 1}, {0, 1}}}
	var b18 = sequenceBuilder{id: 18, commit: 64, name: "definition", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}}}
	var b114 = sequenceBuilder{id: 114, commit: 74, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}}
	var b93 = sequenceBuilder{id: 93, commit: 72, name: "symbol", ranges: [][]int{{1, -1}, {1, -1}}}
	var b80 = sequenceBuilder{id: 80, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b107 = charBuilder{}
	b80.items = []builder{&b107}
	b93.items = []builder{&b80}
	var b83 = sequenceBuilder{id: 83, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b71 = sequenceBuilder{id: 71, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b70 = charBuilder{}
	b71.items = []builder{&b70}
	var b4 = choiceBuilder{id: 4, commit: 66}
	var b56 = sequenceBuilder{id: 56, commit: 72, name: "alias", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b134 = charBuilder{}
	var b64 = charBuilder{}
	var b170 = charBuilder{}
	var b99 = charBuilder{}
	var b123 = charBuilder{}
	b56.items = []builder{&b134, &b64, &b170, &b99, &b123}
	var b22 = sequenceBuilder{id: 22, commit: 72, name: "ws", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b184 = charBuilder{}
	var b174 = charBuilder{}
	b22.items = []builder{&b184, &b174}
	var b57 = sequenceBuilder{id: 57, commit: 72, name: "nows", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b23 = charBuilder{}
	var b135 = charBuilder{}
	var b159 = charBuilder{}
	var b54 = charBuilder{}
	b57.items = []builder{&b23, &b135, &b159, &b54}
	var b51 = sequenceBuilder{id: 51, commit: 72, name: "failpass", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b50 = charBuilder{}
	var b60 = charBuilder{}
	var b178 = charBuilder{}
	var b89 = charBuilder{}
	var b69 = charBuilder{}
	var b175 = charBuilder{}
	var b138 = charBuilder{}
	var b113 = charBuilder{}
	b51.items = []builder{&b50, &b60, &b178, &b89, &b69, &b175, &b138, &b113}
	var b28 = sequenceBuilder{id: 28, commit: 72, name: "root", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b9 = charBuilder{}
	var b100 = charBuilder{}
	var b160 = charBuilder{}
	var b76 = charBuilder{}
	b28.items = []builder{&b9, &b100, &b160, &b76}
	b4.options = []builder{&b56, &b22, &b57, &b51, &b28}
	b83.items = []builder{&b71, &b4}
	b114.items = []builder{&b93, &b83}
	var b49 = sequenceBuilder{id: 49, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b132 = charBuilder{}
	b49.items = []builder{&b132}
	var b59 = choiceBuilder{id: 59, commit: 66}
	var b136 = choiceBuilder{id: 136, commit: 66}
	var b6 = sequenceBuilder{id: 6, commit: 72, name: "any-char", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b104 = charBuilder{}
	b6.items = []builder{&b104}
	var b75 = sequenceBuilder{id: 75, commit: 72, name: "char-class", ranges: [][]int{{1, 1}, {0, 1}, {0, -1}, {1, 1}, {1, 1}, {0, 1}, {0, -1}, {1, 1}}}
	var b86 = sequenceBuilder{id: 86, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b105 = charBuilder{}
	b86.items = []builder{&b105}
	var b79 = sequenceBuilder{id: 79, commit: 72, name: "class-not", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b171 = charBuilder{}
	b79.items = []builder{&b171}
	var b125 = choiceBuilder{id: 125, commit: 10}
	var b157 = choiceBuilder{id: 157, commit: 72, name: "class-char"}
	var b7 = sequenceBuilder{id: 7, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b90 = charBuilder{}
	b7.items = []builder{&b90}
	var b173 = sequenceBuilder{id: 173, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b172 = sequenceBuilder{id: 172, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b155 = charBuilder{}
	b172.items = []builder{&b155}
	var b183 = sequenceBuilder{id: 183, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b156 = charBuilder{}
	b183.items = []builder{&b156}
	b173.items = []builder{&b172, &b183}
	b157.options = []builder{&b7, &b173}
	var b91 = sequenceBuilder{id: 91, commit: 72, name: "char-range", ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b120 = sequenceBuilder{id: 120, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b74 = charBuilder{}
	b120.items = []builder{&b74}
	b91.items = []builder{&b157, &b120, &b157}
	b125.options = []builder{&b157, &b91}
	var b1 = sequenceBuilder{id: 1, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b112 = charBuilder{}
	b1.items = []builder{&b112}
	b75.items = []builder{&b86, &b79, &b125, &b1}
	var b2 = sequenceBuilder{id: 2, commit: 72, name: "char-sequence", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {1, 1}, {0, -1}, {1, 1}}}
	var b127 = sequenceBuilder{id: 127, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b97 = charBuilder{}
	b127.items = []builder{&b97}
	var b43 = choiceBuilder{id: 43, commit: 72, name: "sequence-char"}
	var b142 = sequenceBuilder{id: 142, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b92 = charBuilder{}
	b142.items = []builder{&b92}
	var b19 = sequenceBuilder{id: 19, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b42 = sequenceBuilder{id: 42, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b38 = charBuilder{}
	b42.items = []builder{&b38}
	var b87 = sequenceBuilder{id: 87, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b106 = charBuilder{}
	b87.items = []builder{&b106}
	b19.items = []builder{&b42, &b87}
	b43.options = []builder{&b142, &b19}
	var b11 = sequenceBuilder{id: 11, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b20 = charBuilder{}
	b11.items = []builder{&b20}
	b2.items = []builder{&b127, &b43, &b11}
	b136.options = []builder{&b6, &b75, &b2}
	var b167 = sequenceBuilder{id: 167, commit: 66, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}}}
	var b126 = sequenceBuilder{id: 126, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b151 = charBuilder{}
	b126.items = []builder{&b151}
	var b12 = sequenceBuilder{id: 12, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b176 = charBuilder{}
	b12.items = []builder{&b176}
	b167.items = []builder{&b126, &b186, &b59, &b186, &b12}
	var b146 = sequenceBuilder{id: 146, commit: 64, name: "sequence", ranges: [][]int{{1, 1}, {0, -1}}}
	var b68 = sequenceBuilder{id: 68, commit: 72, name: "item", ranges: [][]int{{1, 1}, {0, 1}, {1, 1}, {0, 1}}}
	var b88 = choiceBuilder{id: 88, commit: 10}
	b88.options = []builder{&b136, &b93, &b167}
	var b122 = choiceBuilder{id: 122, commit: 66}
	var b108 = sequenceBuilder{id: 108, commit: 64, name: "count-quantifier", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}}}
	var b44 = sequenceBuilder{id: 44, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b128 = charBuilder{}
	b44.items = []builder{&b128}
	var b27 = sequenceBuilder{id: 27, commit: 64, name: "count", ranges: [][]int{{1, 1}}}
	var b98 = sequenceBuilder{id: 98, commit: 74, ranges: [][]int{{1, -1}, {1, -1}}}
	var b121 = sequenceBuilder{id: 121, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b143 = charBuilder{}
	b121.items = []builder{&b143}
	b98.items = []builder{&b121}
	b27.items = []builder{&b98}
	var b63 = sequenceBuilder{id: 63, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b55 = charBuilder{}
	b63.items = []builder{&b55}
	b108.items = []builder{&b44, &b186, &b27, &b186, &b63}
	var b81 = sequenceBuilder{id: 81, commit: 64, name: "range-quantifier", ranges: [][]int{{1, 1}, {0, -1}, {0, 1}, {0, -1}, {1, 1}, {0, -1}, {0, 1}, {0, -1}, {1, 1}}}
	var b21 = sequenceBuilder{id: 21, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b8 = charBuilder{}
	b21.items = []builder{&b8}
	var b65 = sequenceBuilder{id: 65, commit: 64, name: "range-from", ranges: [][]int{{1, 1}}}
	b65.items = []builder{&b98}
	var b58 = sequenceBuilder{id: 58, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b153 = charBuilder{}
	b58.items = []builder{&b153}
	var b152 = sequenceBuilder{id: 152, commit: 64, name: "range-to", ranges: [][]int{{1, 1}}}
	b152.items = []builder{&b98}
	var b48 = sequenceBuilder{id: 48, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b158 = charBuilder{}
	b48.items = []builder{&b158}
	b81.items = []builder{&b21, &b186, &b65, &b186, &b58, &b186, &b152, &b186, &b48}
	var b82 = sequenceBuilder{id: 82, commit: 72, name: "one-or-more", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b144 = charBuilder{}
	b82.items = []builder{&b144}
	var b13 = sequenceBuilder{id: 13, commit: 72, name: "zero-or-more", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b137 = charBuilder{}
	b13.items = []builder{&b137}
	var b3 = sequenceBuilder{id: 3, commit: 72, name: "zero-or-one", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b14 = charBuilder{}
	b3.items = []builder{&b14}
	b122.options = []builder{&b108, &b81, &b82, &b13, &b3}
	b68.items = []builder{&b88, &b122}
	var b145 = sequenceBuilder{id: 145, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b145.items = []builder{&b186, &b68}
	b146.items = []builder{&b68, &b145}
	var b169 = sequenceBuilder{id: 169, commit: 64, name: "choice", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}}
	var b39 = choiceBuilder{id: 39, commit: 66}
	b39.options = []builder{&b136, &b93, &b167, &b146}
	var b177 = sequenceBuilder{id: 177, commit: 2, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}}}
	var b67 = sequenceBuilder{id: 67, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b66 = charBuilder{}
	b67.items = []builder{&b66}
	b177.items = []builder{&b67, &b186, &b39}
	var b168 = sequenceBuilder{id: 168, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b168.items = []builder{&b186, &b177}
	b169.items = []builder{&b39, &b186, &b177, &b168}
	b59.options = []builder{&b136, &b93, &b167, &b146, &b169}
	b18.items = []builder{&b114, &b186, &b49, &b186, &b59}
	var b116 = sequenceBuilder{id: 116, commit: 2, ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var b180 = sequenceBuilder{id: 180, commit: 2, ranges: [][]int{{1, 1}, {0, -1}, {0, -1}, {1, 1}}}
	var b36 = sequenceBuilder{id: 36, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b32 = charBuilder{}
	b36.items = []builder{&b32}
	var b179 = sequenceBuilder{id: 179, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b179.items = []builder{&b186, &b36}
	b180.items = []builder{&b36, &b179, &b186, &b18}
	var b115 = sequenceBuilder{id: 115, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b115.items = []builder{&b186, &b180}
	b116.items = []builder{&b186, &b180, &b115}
	b117.items = []builder{&b18, &b116}
	var b166 = sequenceBuilder{id: 166, commit: 2, ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var b139 = sequenceBuilder{id: 139, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b94 = charBuilder{}
	b139.items = []builder{&b94}
	var b165 = sequenceBuilder{id: 165, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b165.items = []builder{&b186, &b139}
	b166.items = []builder{&b186, &b139, &b165}
	b187.items = []builder{&b164, &b186, &b117, &b166}
	b188.items = []builder{&b186, &b187, &b186}

	return parseInput(r, &p188, &b188)
}
