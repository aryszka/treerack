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
	var p142 = choiceParser{id: 142, commit: 66, name: "wschar", generalizations: []int{185, 186}}
	var p180 = sequenceParser{id: 180, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{142, 185, 186}}
	var p13 = charParser{id: 13, chars: []rune{32}}
	p180.items = []parser{&p13}
	var p3 = sequenceParser{id: 3, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{142, 185, 186}}
	var p48 = charParser{id: 48, chars: []rune{9}}
	p3.items = []parser{&p48}
	var p112 = sequenceParser{id: 112, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{142, 185, 186}}
	var p64 = charParser{id: 64, chars: []rune{10}}
	p112.items = []parser{&p64}
	var p170 = sequenceParser{id: 170, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{142, 185, 186}}
	var p169 = charParser{id: 169, chars: []rune{8}}
	p170.items = []parser{&p169}
	var p76 = sequenceParser{id: 76, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{142, 185, 186}}
	var p104 = charParser{id: 104, chars: []rune{12}}
	p76.items = []parser{&p104}
	var p33 = sequenceParser{id: 33, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{142, 185, 186}}
	var p91 = charParser{id: 91, chars: []rune{13}}
	p33.items = []parser{&p91}
	var p41 = sequenceParser{id: 41, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{142, 185, 186}}
	var p77 = charParser{id: 77, chars: []rune{11}}
	p41.items = []parser{&p77}
	p142.options = []parser{&p180, &p3, &p112, &p170, &p76, &p33, &p41}
	var p143 = sequenceParser{id: 143, commit: 72, name: "comment", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}, generalizations: []int{185, 186}}
	var p54 = choiceParser{id: 54, commit: 74, name: "comment-segment"}
	var p159 = sequenceParser{id: 159, commit: 74, name: "line-comment", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}, generalizations: []int{54}}
	var p8 = sequenceParser{id: 8, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var p35 = charParser{id: 35, chars: []rune{47}}
	var p53 = charParser{id: 53, chars: []rune{47}}
	p8.items = []parser{&p35, &p53}
	var p9 = sequenceParser{id: 9, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p43 = charParser{id: 43, not: true, chars: []rune{10}}
	p9.items = []parser{&p43}
	p159.items = []parser{&p8, &p9}
	var p93 = sequenceParser{id: 93, commit: 74, name: "block-comment", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {1, 1}, {0, -1}, {1, 1}}, generalizations: []int{54}}
	var p131 = sequenceParser{id: 131, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var p92 = charParser{id: 92, chars: []rune{47}}
	var p99 = charParser{id: 99, chars: []rune{42}}
	p131.items = []parser{&p92, &p99}
	var p34 = choiceParser{id: 34, commit: 10}
	var p164 = sequenceParser{id: 164, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{34}}
	var p171 = sequenceParser{id: 171, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p59 = charParser{id: 59, chars: []rune{42}}
	p171.items = []parser{&p59}
	var p120 = sequenceParser{id: 120, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p116 = charParser{id: 116, not: true, chars: []rune{47}}
	p120.items = []parser{&p116}
	p164.items = []parser{&p171, &p120}
	var p28 = sequenceParser{id: 28, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{34}}
	var p155 = charParser{id: 155, not: true, chars: []rune{42}}
	p28.items = []parser{&p155}
	p34.options = []parser{&p164, &p28}
	var p42 = sequenceParser{id: 42, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var p60 = charParser{id: 60, chars: []rune{42}}
	var p83 = charParser{id: 83, chars: []rune{47}}
	p42.items = []parser{&p60, &p83}
	p93.items = []parser{&p131, &p34, &p42}
	p54.options = []parser{&p159, &p93}
	var p7 = sequenceParser{id: 7, commit: 10, ranges: [][]int{{0, -1}, {0, 1}, {0, -1}, {1, 1}, {0, -1}, {0, 1}, {0, -1}, {1, 1}}}
	var p44 = choiceParser{id: 44, commit: 74, name: "ws-no-nl"}
	var p65 = sequenceParser{id: 65, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{44}}
	var p49 = charParser{id: 49, chars: []rune{32}}
	p65.items = []parser{&p49}
	var p136 = sequenceParser{id: 136, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{44}}
	var p78 = charParser{id: 78, chars: []rune{9}}
	p136.items = []parser{&p78}
	var p23 = sequenceParser{id: 23, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{44}}
	var p4 = charParser{id: 4, chars: []rune{8}}
	p23.items = []parser{&p4}
	var p137 = sequenceParser{id: 137, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{44}}
	var p69 = charParser{id: 69, chars: []rune{12}}
	p137.items = []parser{&p69}
	var p5 = sequenceParser{id: 5, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{44}}
	var p24 = charParser{id: 24, chars: []rune{13}}
	p5.items = []parser{&p24}
	var p29 = sequenceParser{id: 29, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{44}}
	var p79 = charParser{id: 79, chars: []rune{11}}
	p29.items = []parser{&p79}
	p44.options = []parser{&p65, &p136, &p23, &p137, &p5, &p29}
	var p45 = sequenceParser{id: 45, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p6 = charParser{id: 6, chars: []rune{10}}
	p45.items = []parser{&p6}
	p7.items = []parser{&p44, &p45, &p44, &p54}
	p143.items = []parser{&p54, &p7}
	p185.options = []parser{&p142, &p143}
	p186.options = []parser{&p185}
	var p187 = sequenceParser{id: 187, commit: 66, name: "syntax:wsroot", ranges: [][]int{{0, 1}, {0, -1}, {0, 1}, {0, 1}}}
	var p177 = sequenceParser{id: 177, commit: 2, ranges: [][]int{{1, 1}, {0, -1}}}
	var p58 = sequenceParser{id: 58, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p174 = charParser{id: 174, chars: []rune{59}}
	p58.items = []parser{&p174}
	var p176 = sequenceParser{id: 176, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p176.items = []parser{&p186, &p58}
	p177.items = []parser{&p58, &p176}
	var p126 = sequenceParser{id: 126, commit: 66, name: "definitions", ranges: [][]int{{1, 1}, {0, 1}}}
	var p96 = sequenceParser{id: 96, commit: 64, name: "definition", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}}}
	var p158 = sequenceParser{id: 158, commit: 74, name: "definition-name", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}}
	var p10 = sequenceParser{id: 10, commit: 72, name: "symbol", ranges: [][]int{{1, -1}, {1, -1}}, generalizations: []int{88, 56, 57}}
	var p86 = sequenceParser{id: 86, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p134 = charParser{id: 134, not: true, chars: []rune{92, 32, 10, 9, 8, 12, 13, 11, 47, 46, 91, 93, 34, 123, 125, 94, 43, 42, 63, 124, 40, 41, 58, 61, 59}}
	p86.items = []parser{&p134}
	p10.items = []parser{&p86}
	var p138 = sequenceParser{id: 138, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var p82 = sequenceParser{id: 82, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p17 = charParser{id: 17, chars: []rune{58}}
	p82.items = []parser{&p17}
	var p184 = choiceParser{id: 184, commit: 66, name: "flag"}
	var p81 = sequenceParser{id: 81, commit: 72, name: "alias", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{184}}
	var p162 = charParser{id: 162, chars: []rune{97}}
	var p51 = charParser{id: 51, chars: []rune{108}}
	var p73 = charParser{id: 73, chars: []rune{105}}
	var p182 = charParser{id: 182, chars: []rune{97}}
	var p12 = charParser{id: 12, chars: []rune{115}}
	p81.items = []parser{&p162, &p51, &p73, &p182, &p12}
	var p30 = sequenceParser{id: 30, commit: 72, name: "ws", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{184}}
	var p89 = charParser{id: 89, chars: []rune{119}}
	var p154 = charParser{id: 154, chars: []rune{115}}
	p30.items = []parser{&p89, &p154}
	var p123 = sequenceParser{id: 123, commit: 72, name: "nows", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{184}}
	var p118 = charParser{id: 118, chars: []rune{110}}
	var p111 = charParser{id: 111, chars: []rune{111}}
	var p21 = charParser{id: 21, chars: []rune{119}}
	var p146 = charParser{id: 146, chars: []rune{115}}
	p123.items = []parser{&p118, &p111, &p21, &p146}
	var p63 = sequenceParser{id: 63, commit: 72, name: "failpass", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{184}}
	var p141 = charParser{id: 141, chars: []rune{102}}
	var p47 = charParser{id: 47, chars: []rune{97}}
	var p103 = charParser{id: 103, chars: []rune{105}}
	var p90 = charParser{id: 90, chars: []rune{108}}
	var p68 = charParser{id: 68, chars: []rune{112}}
	var p173 = charParser{id: 173, chars: []rune{97}}
	var p107 = charParser{id: 107, chars: []rune{115}}
	var p108 = charParser{id: 108, chars: []rune{115}}
	p63.items = []parser{&p141, &p47, &p103, &p90, &p68, &p173, &p107, &p108}
	var p183 = sequenceParser{id: 183, commit: 72, name: "root", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{184}}
	var p119 = charParser{id: 119, chars: []rune{114}}
	var p52 = charParser{id: 52, chars: []rune{111}}
	var p97 = charParser{id: 97, chars: []rune{111}}
	var p50 = charParser{id: 50, chars: []rune{116}}
	p183.items = []parser{&p119, &p52, &p97, &p50}
	p184.options = []parser{&p81, &p30, &p123, &p63, &p183}
	p138.items = []parser{&p82, &p184}
	p158.items = []parser{&p10, &p138}
	var p98 = sequenceParser{id: 98, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p31 = charParser{id: 31, chars: []rune{61}}
	p98.items = []parser{&p31}
	var p88 = choiceParser{id: 88, commit: 66, name: "expression"}
	var p160 = choiceParser{id: 160, commit: 66, name: "terminal", generalizations: []int{88, 56, 57}}
	var p165 = sequenceParser{id: 165, commit: 72, name: "any-char", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{160, 88, 56, 57}}
	var p36 = charParser{id: 36, chars: []rune{46}}
	p165.items = []parser{&p36}
	var p156 = sequenceParser{id: 156, commit: 72, name: "char-class", ranges: [][]int{{1, 1}, {0, 1}, {0, -1}, {1, 1}, {1, 1}, {0, 1}, {0, -1}, {1, 1}}, generalizations: []int{160, 88, 56, 57}}
	var p18 = sequenceParser{id: 18, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p144 = charParser{id: 144, chars: []rune{91}}
	p18.items = []parser{&p144}
	var p38 = sequenceParser{id: 38, commit: 72, name: "class-not", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p14 = charParser{id: 14, chars: []rune{94}}
	p38.items = []parser{&p14}
	var p132 = choiceParser{id: 132, commit: 10}
	var p172 = choiceParser{id: 172, commit: 72, name: "class-char", generalizations: []int{132}}
	var p84 = sequenceParser{id: 84, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{172, 132}}
	var p147 = charParser{id: 147, not: true, chars: []rune{92, 91, 93, 94, 45}}
	p84.items = []parser{&p147}
	var p148 = sequenceParser{id: 148, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{172, 132}}
	var p85 = sequenceParser{id: 85, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p105 = charParser{id: 105, chars: []rune{92}}
	p85.items = []parser{&p105}
	var p39 = sequenceParser{id: 39, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p94 = charParser{id: 94, not: true}
	p39.items = []parser{&p94}
	p148.items = []parser{&p85, &p39}
	p172.options = []parser{&p84, &p148}
	var p113 = sequenceParser{id: 113, commit: 72, name: "char-range", ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{132}}
	var p166 = sequenceParser{id: 166, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p70 = charParser{id: 70, chars: []rune{45}}
	p166.items = []parser{&p70}
	p113.items = []parser{&p172, &p166, &p172}
	p132.options = []parser{&p172, &p113}
	var p80 = sequenceParser{id: 80, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p100 = charParser{id: 100, chars: []rune{93}}
	p80.items = []parser{&p100}
	p156.items = []parser{&p18, &p38, &p132, &p80}
	var p72 = sequenceParser{id: 72, commit: 72, name: "char-sequence", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {1, 1}, {0, -1}, {1, 1}}, generalizations: []int{160, 88, 56, 57}}
	var p26 = sequenceParser{id: 26, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p19 = charParser{id: 19, chars: []rune{34}}
	p26.items = []parser{&p19}
	var p114 = choiceParser{id: 114, commit: 72, name: "sequence-char"}
	var p25 = sequenceParser{id: 25, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{114}}
	var p71 = charParser{id: 71, not: true, chars: []rune{92, 34}}
	p25.items = []parser{&p71}
	var p109 = sequenceParser{id: 109, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{114}}
	var p181 = sequenceParser{id: 181, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p101 = charParser{id: 101, chars: []rune{92}}
	p181.items = []parser{&p101}
	var p66 = sequenceParser{id: 66, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p133 = charParser{id: 133, not: true}
	p66.items = []parser{&p133}
	p109.items = []parser{&p181, &p66}
	p114.options = []parser{&p25, &p109}
	var p40 = sequenceParser{id: 40, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p55 = charParser{id: 55, chars: []rune{34}}
	p40.items = []parser{&p55}
	p72.items = []parser{&p26, &p114, &p40}
	p160.options = []parser{&p165, &p156, &p72}
	var p27 = sequenceParser{id: 27, commit: 66, name: "group", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}}, generalizations: []int{88, 56, 57}}
	var p20 = sequenceParser{id: 20, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p61 = charParser{id: 61, chars: []rune{40}}
	p20.items = []parser{&p61}
	var p62 = sequenceParser{id: 62, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p145 = charParser{id: 145, chars: []rune{41}}
	p62.items = []parser{&p145}
	p27.items = []parser{&p20, &p186, &p88, &p186, &p62}
	var p153 = sequenceParser{id: 153, commit: 64, name: "sequence", ranges: [][]int{{1, 1}, {0, -1}}, generalizations: []int{88, 57}}
	var p106 = sequenceParser{id: 106, commit: 72, name: "item", ranges: [][]int{{1, 1}, {0, 1}, {1, 1}, {0, 1}}}
	var p56 = choiceParser{id: 56, commit: 10}
	p56.options = []parser{&p160, &p10, &p27}
	var p163 = choiceParser{id: 163, commit: 66, name: "quantity"}
	var p149 = sequenceParser{id: 149, commit: 64, name: "count-quantifier", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}}, generalizations: []int{163}}
	var p121 = sequenceParser{id: 121, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p127 = charParser{id: 127, chars: []rune{123}}
	p121.items = []parser{&p127}
	var p135 = sequenceParser{id: 135, commit: 64, name: "count", ranges: [][]int{{1, 1}}}
	var p15 = sequenceParser{id: 15, commit: 74, name: "number", ranges: [][]int{{1, -1}, {1, -1}}}
	var p110 = sequenceParser{id: 110, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p117 = charParser{id: 117, ranges: [][]rune{{48, 57}}}
	p110.items = []parser{&p117}
	p15.items = []parser{&p110}
	p135.items = []parser{&p15}
	var p102 = sequenceParser{id: 102, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p157 = charParser{id: 157, chars: []rune{125}}
	p102.items = []parser{&p157}
	p149.items = []parser{&p121, &p186, &p135, &p186, &p102}
	var p67 = sequenceParser{id: 67, commit: 64, name: "range-quantifier", ranges: [][]int{{1, 1}, {0, -1}, {0, 1}, {0, -1}, {1, 1}, {0, -1}, {0, 1}, {0, -1}, {1, 1}}, generalizations: []int{163}}
	var p46 = sequenceParser{id: 46, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p168 = charParser{id: 168, chars: []rune{123}}
	p46.items = []parser{&p168}
	var p115 = sequenceParser{id: 115, commit: 64, name: "range-from", ranges: [][]int{{1, 1}}}
	p115.items = []parser{&p15}
	var p1 = sequenceParser{id: 1, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p11 = charParser{id: 11, chars: []rune{44}}
	p1.items = []parser{&p11}
	var p167 = sequenceParser{id: 167, commit: 64, name: "range-to", ranges: [][]int{{1, 1}}}
	p167.items = []parser{&p15}
	var p37 = sequenceParser{id: 37, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p161 = charParser{id: 161, chars: []rune{125}}
	p37.items = []parser{&p161}
	p67.items = []parser{&p46, &p186, &p115, &p186, &p1, &p186, &p167, &p186, &p37}
	var p150 = sequenceParser{id: 150, commit: 72, name: "one-or-more", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{163}}
	var p122 = charParser{id: 122, chars: []rune{43}}
	p150.items = []parser{&p122}
	var p87 = sequenceParser{id: 87, commit: 72, name: "zero-or-more", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{163}}
	var p74 = charParser{id: 74, chars: []rune{42}}
	p87.items = []parser{&p74}
	var p151 = sequenceParser{id: 151, commit: 72, name: "zero-or-one", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{163}}
	var p128 = charParser{id: 128, chars: []rune{63}}
	p151.items = []parser{&p128}
	p163.options = []parser{&p149, &p67, &p150, &p87, &p151}
	p106.items = []parser{&p56, &p163}
	var p152 = sequenceParser{id: 152, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p152.items = []parser{&p186, &p106}
	p153.items = []parser{&p106, &p152}
	var p130 = sequenceParser{id: 130, commit: 64, name: "choice", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}, generalizations: []int{88}}
	var p57 = choiceParser{id: 57, commit: 66, name: "option"}
	p57.options = []parser{&p160, &p10, &p27, &p153}
	var p95 = sequenceParser{id: 95, commit: 2, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}}}
	var p16 = sequenceParser{id: 16, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p75 = charParser{id: 75, chars: []rune{124}}
	p16.items = []parser{&p75}
	p95.items = []parser{&p16, &p186, &p57}
	var p129 = sequenceParser{id: 129, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p129.items = []parser{&p186, &p95}
	p130.items = []parser{&p57, &p186, &p95, &p129}
	p88.options = []parser{&p160, &p10, &p27, &p153, &p130}
	p96.items = []parser{&p158, &p186, &p98, &p186, &p88}
	var p125 = sequenceParser{id: 125, commit: 2, ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var p140 = sequenceParser{id: 140, commit: 2, ranges: [][]int{{1, 1}, {0, -1}, {0, -1}, {1, 1}}}
	var p22 = sequenceParser{id: 22, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p2 = charParser{id: 2, chars: []rune{59}}
	p22.items = []parser{&p2}
	var p139 = sequenceParser{id: 139, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p139.items = []parser{&p186, &p22}
	p140.items = []parser{&p22, &p139, &p186, &p96}
	var p124 = sequenceParser{id: 124, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p124.items = []parser{&p186, &p140}
	p125.items = []parser{&p186, &p140, &p124}
	p126.items = []parser{&p96, &p125}
	var p179 = sequenceParser{id: 179, commit: 2, ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var p32 = sequenceParser{id: 32, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p175 = charParser{id: 175, chars: []rune{59}}
	p32.items = []parser{&p175}
	var p178 = sequenceParser{id: 178, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p178.items = []parser{&p186, &p32}
	p179.items = []parser{&p186, &p32, &p178}
	p187.items = []parser{&p177, &p186, &p126, &p179}
	p188.items = []parser{&p186, &p187, &p186}
	var b188 = sequenceBuilder{id: 188, commit: 32, name: "syntax", ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var b186 = choiceBuilder{id: 186, commit: 2}
	var b185 = choiceBuilder{id: 185, commit: 70}
	var b142 = choiceBuilder{id: 142, commit: 66}
	var b180 = sequenceBuilder{id: 180, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b13 = charBuilder{}
	b180.items = []builder{&b13}
	var b3 = sequenceBuilder{id: 3, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b48 = charBuilder{}
	b3.items = []builder{&b48}
	var b112 = sequenceBuilder{id: 112, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b64 = charBuilder{}
	b112.items = []builder{&b64}
	var b170 = sequenceBuilder{id: 170, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b169 = charBuilder{}
	b170.items = []builder{&b169}
	var b76 = sequenceBuilder{id: 76, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b104 = charBuilder{}
	b76.items = []builder{&b104}
	var b33 = sequenceBuilder{id: 33, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b91 = charBuilder{}
	b33.items = []builder{&b91}
	var b41 = sequenceBuilder{id: 41, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b77 = charBuilder{}
	b41.items = []builder{&b77}
	b142.options = []builder{&b180, &b3, &b112, &b170, &b76, &b33, &b41}
	var b143 = sequenceBuilder{id: 143, commit: 72, name: "comment", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}}
	var b54 = choiceBuilder{id: 54, commit: 74}
	var b159 = sequenceBuilder{id: 159, commit: 74, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}}
	var b8 = sequenceBuilder{id: 8, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b35 = charBuilder{}
	var b53 = charBuilder{}
	b8.items = []builder{&b35, &b53}
	var b9 = sequenceBuilder{id: 9, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b43 = charBuilder{}
	b9.items = []builder{&b43}
	b159.items = []builder{&b8, &b9}
	var b93 = sequenceBuilder{id: 93, commit: 74, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {1, 1}, {0, -1}, {1, 1}}}
	var b131 = sequenceBuilder{id: 131, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b92 = charBuilder{}
	var b99 = charBuilder{}
	b131.items = []builder{&b92, &b99}
	var b34 = choiceBuilder{id: 34, commit: 10}
	var b164 = sequenceBuilder{id: 164, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b171 = sequenceBuilder{id: 171, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b59 = charBuilder{}
	b171.items = []builder{&b59}
	var b120 = sequenceBuilder{id: 120, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b116 = charBuilder{}
	b120.items = []builder{&b116}
	b164.items = []builder{&b171, &b120}
	var b28 = sequenceBuilder{id: 28, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b155 = charBuilder{}
	b28.items = []builder{&b155}
	b34.options = []builder{&b164, &b28}
	var b42 = sequenceBuilder{id: 42, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b60 = charBuilder{}
	var b83 = charBuilder{}
	b42.items = []builder{&b60, &b83}
	b93.items = []builder{&b131, &b34, &b42}
	b54.options = []builder{&b159, &b93}
	var b7 = sequenceBuilder{id: 7, commit: 10, ranges: [][]int{{0, -1}, {0, 1}, {0, -1}, {1, 1}, {0, -1}, {0, 1}, {0, -1}, {1, 1}}}
	var b44 = choiceBuilder{id: 44, commit: 74}
	var b65 = sequenceBuilder{id: 65, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b49 = charBuilder{}
	b65.items = []builder{&b49}
	var b136 = sequenceBuilder{id: 136, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b78 = charBuilder{}
	b136.items = []builder{&b78}
	var b23 = sequenceBuilder{id: 23, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b4 = charBuilder{}
	b23.items = []builder{&b4}
	var b137 = sequenceBuilder{id: 137, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b69 = charBuilder{}
	b137.items = []builder{&b69}
	var b5 = sequenceBuilder{id: 5, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b24 = charBuilder{}
	b5.items = []builder{&b24}
	var b29 = sequenceBuilder{id: 29, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b79 = charBuilder{}
	b29.items = []builder{&b79}
	b44.options = []builder{&b65, &b136, &b23, &b137, &b5, &b29}
	var b45 = sequenceBuilder{id: 45, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b6 = charBuilder{}
	b45.items = []builder{&b6}
	b7.items = []builder{&b44, &b45, &b44, &b54}
	b143.items = []builder{&b54, &b7}
	b185.options = []builder{&b142, &b143}
	b186.options = []builder{&b185}
	var b187 = sequenceBuilder{id: 187, commit: 66, ranges: [][]int{{0, 1}, {0, -1}, {0, 1}, {0, 1}}}
	var b177 = sequenceBuilder{id: 177, commit: 2, ranges: [][]int{{1, 1}, {0, -1}}}
	var b58 = sequenceBuilder{id: 58, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b174 = charBuilder{}
	b58.items = []builder{&b174}
	var b176 = sequenceBuilder{id: 176, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b176.items = []builder{&b186, &b58}
	b177.items = []builder{&b58, &b176}
	var b126 = sequenceBuilder{id: 126, commit: 66, ranges: [][]int{{1, 1}, {0, 1}}}
	var b96 = sequenceBuilder{id: 96, commit: 64, name: "definition", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}}}
	var b158 = sequenceBuilder{id: 158, commit: 74, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}}
	var b10 = sequenceBuilder{id: 10, commit: 72, name: "symbol", ranges: [][]int{{1, -1}, {1, -1}}}
	var b86 = sequenceBuilder{id: 86, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b134 = charBuilder{}
	b86.items = []builder{&b134}
	b10.items = []builder{&b86}
	var b138 = sequenceBuilder{id: 138, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b82 = sequenceBuilder{id: 82, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b17 = charBuilder{}
	b82.items = []builder{&b17}
	var b184 = choiceBuilder{id: 184, commit: 66}
	var b81 = sequenceBuilder{id: 81, commit: 72, name: "alias", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b162 = charBuilder{}
	var b51 = charBuilder{}
	var b73 = charBuilder{}
	var b182 = charBuilder{}
	var b12 = charBuilder{}
	b81.items = []builder{&b162, &b51, &b73, &b182, &b12}
	var b30 = sequenceBuilder{id: 30, commit: 72, name: "ws", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b89 = charBuilder{}
	var b154 = charBuilder{}
	b30.items = []builder{&b89, &b154}
	var b123 = sequenceBuilder{id: 123, commit: 72, name: "nows", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b118 = charBuilder{}
	var b111 = charBuilder{}
	var b21 = charBuilder{}
	var b146 = charBuilder{}
	b123.items = []builder{&b118, &b111, &b21, &b146}
	var b63 = sequenceBuilder{id: 63, commit: 72, name: "failpass", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b141 = charBuilder{}
	var b47 = charBuilder{}
	var b103 = charBuilder{}
	var b90 = charBuilder{}
	var b68 = charBuilder{}
	var b173 = charBuilder{}
	var b107 = charBuilder{}
	var b108 = charBuilder{}
	b63.items = []builder{&b141, &b47, &b103, &b90, &b68, &b173, &b107, &b108}
	var b183 = sequenceBuilder{id: 183, commit: 72, name: "root", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b119 = charBuilder{}
	var b52 = charBuilder{}
	var b97 = charBuilder{}
	var b50 = charBuilder{}
	b183.items = []builder{&b119, &b52, &b97, &b50}
	b184.options = []builder{&b81, &b30, &b123, &b63, &b183}
	b138.items = []builder{&b82, &b184}
	b158.items = []builder{&b10, &b138}
	var b98 = sequenceBuilder{id: 98, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b31 = charBuilder{}
	b98.items = []builder{&b31}
	var b88 = choiceBuilder{id: 88, commit: 66}
	var b160 = choiceBuilder{id: 160, commit: 66}
	var b165 = sequenceBuilder{id: 165, commit: 72, name: "any-char", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b36 = charBuilder{}
	b165.items = []builder{&b36}
	var b156 = sequenceBuilder{id: 156, commit: 72, name: "char-class", ranges: [][]int{{1, 1}, {0, 1}, {0, -1}, {1, 1}, {1, 1}, {0, 1}, {0, -1}, {1, 1}}}
	var b18 = sequenceBuilder{id: 18, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b144 = charBuilder{}
	b18.items = []builder{&b144}
	var b38 = sequenceBuilder{id: 38, commit: 72, name: "class-not", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b14 = charBuilder{}
	b38.items = []builder{&b14}
	var b132 = choiceBuilder{id: 132, commit: 10}
	var b172 = choiceBuilder{id: 172, commit: 72, name: "class-char"}
	var b84 = sequenceBuilder{id: 84, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b147 = charBuilder{}
	b84.items = []builder{&b147}
	var b148 = sequenceBuilder{id: 148, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b85 = sequenceBuilder{id: 85, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b105 = charBuilder{}
	b85.items = []builder{&b105}
	var b39 = sequenceBuilder{id: 39, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b94 = charBuilder{}
	b39.items = []builder{&b94}
	b148.items = []builder{&b85, &b39}
	b172.options = []builder{&b84, &b148}
	var b113 = sequenceBuilder{id: 113, commit: 72, name: "char-range", ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b166 = sequenceBuilder{id: 166, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b70 = charBuilder{}
	b166.items = []builder{&b70}
	b113.items = []builder{&b172, &b166, &b172}
	b132.options = []builder{&b172, &b113}
	var b80 = sequenceBuilder{id: 80, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b100 = charBuilder{}
	b80.items = []builder{&b100}
	b156.items = []builder{&b18, &b38, &b132, &b80}
	var b72 = sequenceBuilder{id: 72, commit: 72, name: "char-sequence", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {1, 1}, {0, -1}, {1, 1}}}
	var b26 = sequenceBuilder{id: 26, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b19 = charBuilder{}
	b26.items = []builder{&b19}
	var b114 = choiceBuilder{id: 114, commit: 72, name: "sequence-char"}
	var b25 = sequenceBuilder{id: 25, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b71 = charBuilder{}
	b25.items = []builder{&b71}
	var b109 = sequenceBuilder{id: 109, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b181 = sequenceBuilder{id: 181, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b101 = charBuilder{}
	b181.items = []builder{&b101}
	var b66 = sequenceBuilder{id: 66, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b133 = charBuilder{}
	b66.items = []builder{&b133}
	b109.items = []builder{&b181, &b66}
	b114.options = []builder{&b25, &b109}
	var b40 = sequenceBuilder{id: 40, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b55 = charBuilder{}
	b40.items = []builder{&b55}
	b72.items = []builder{&b26, &b114, &b40}
	b160.options = []builder{&b165, &b156, &b72}
	var b27 = sequenceBuilder{id: 27, commit: 66, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}}}
	var b20 = sequenceBuilder{id: 20, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b61 = charBuilder{}
	b20.items = []builder{&b61}
	var b62 = sequenceBuilder{id: 62, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b145 = charBuilder{}
	b62.items = []builder{&b145}
	b27.items = []builder{&b20, &b186, &b88, &b186, &b62}
	var b153 = sequenceBuilder{id: 153, commit: 64, name: "sequence", ranges: [][]int{{1, 1}, {0, -1}}}
	var b106 = sequenceBuilder{id: 106, commit: 72, name: "item", ranges: [][]int{{1, 1}, {0, 1}, {1, 1}, {0, 1}}}
	var b56 = choiceBuilder{id: 56, commit: 10}
	b56.options = []builder{&b160, &b10, &b27}
	var b163 = choiceBuilder{id: 163, commit: 66}
	var b149 = sequenceBuilder{id: 149, commit: 64, name: "count-quantifier", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}}}
	var b121 = sequenceBuilder{id: 121, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b127 = charBuilder{}
	b121.items = []builder{&b127}
	var b135 = sequenceBuilder{id: 135, commit: 64, name: "count", ranges: [][]int{{1, 1}}}
	var b15 = sequenceBuilder{id: 15, commit: 74, ranges: [][]int{{1, -1}, {1, -1}}}
	var b110 = sequenceBuilder{id: 110, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b117 = charBuilder{}
	b110.items = []builder{&b117}
	b15.items = []builder{&b110}
	b135.items = []builder{&b15}
	var b102 = sequenceBuilder{id: 102, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b157 = charBuilder{}
	b102.items = []builder{&b157}
	b149.items = []builder{&b121, &b186, &b135, &b186, &b102}
	var b67 = sequenceBuilder{id: 67, commit: 64, name: "range-quantifier", ranges: [][]int{{1, 1}, {0, -1}, {0, 1}, {0, -1}, {1, 1}, {0, -1}, {0, 1}, {0, -1}, {1, 1}}}
	var b46 = sequenceBuilder{id: 46, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b168 = charBuilder{}
	b46.items = []builder{&b168}
	var b115 = sequenceBuilder{id: 115, commit: 64, name: "range-from", ranges: [][]int{{1, 1}}}
	b115.items = []builder{&b15}
	var b1 = sequenceBuilder{id: 1, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b11 = charBuilder{}
	b1.items = []builder{&b11}
	var b167 = sequenceBuilder{id: 167, commit: 64, name: "range-to", ranges: [][]int{{1, 1}}}
	b167.items = []builder{&b15}
	var b37 = sequenceBuilder{id: 37, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b161 = charBuilder{}
	b37.items = []builder{&b161}
	b67.items = []builder{&b46, &b186, &b115, &b186, &b1, &b186, &b167, &b186, &b37}
	var b150 = sequenceBuilder{id: 150, commit: 72, name: "one-or-more", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b122 = charBuilder{}
	b150.items = []builder{&b122}
	var b87 = sequenceBuilder{id: 87, commit: 72, name: "zero-or-more", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b74 = charBuilder{}
	b87.items = []builder{&b74}
	var b151 = sequenceBuilder{id: 151, commit: 72, name: "zero-or-one", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b128 = charBuilder{}
	b151.items = []builder{&b128}
	b163.options = []builder{&b149, &b67, &b150, &b87, &b151}
	b106.items = []builder{&b56, &b163}
	var b152 = sequenceBuilder{id: 152, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b152.items = []builder{&b186, &b106}
	b153.items = []builder{&b106, &b152}
	var b130 = sequenceBuilder{id: 130, commit: 64, name: "choice", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}}
	var b57 = choiceBuilder{id: 57, commit: 66}
	b57.options = []builder{&b160, &b10, &b27, &b153}
	var b95 = sequenceBuilder{id: 95, commit: 2, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}}}
	var b16 = sequenceBuilder{id: 16, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b75 = charBuilder{}
	b16.items = []builder{&b75}
	b95.items = []builder{&b16, &b186, &b57}
	var b129 = sequenceBuilder{id: 129, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b129.items = []builder{&b186, &b95}
	b130.items = []builder{&b57, &b186, &b95, &b129}
	b88.options = []builder{&b160, &b10, &b27, &b153, &b130}
	b96.items = []builder{&b158, &b186, &b98, &b186, &b88}
	var b125 = sequenceBuilder{id: 125, commit: 2, ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var b140 = sequenceBuilder{id: 140, commit: 2, ranges: [][]int{{1, 1}, {0, -1}, {0, -1}, {1, 1}}}
	var b22 = sequenceBuilder{id: 22, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b2 = charBuilder{}
	b22.items = []builder{&b2}
	var b139 = sequenceBuilder{id: 139, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b139.items = []builder{&b186, &b22}
	b140.items = []builder{&b22, &b139, &b186, &b96}
	var b124 = sequenceBuilder{id: 124, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b124.items = []builder{&b186, &b140}
	b125.items = []builder{&b186, &b140, &b124}
	b126.items = []builder{&b96, &b125}
	var b179 = sequenceBuilder{id: 179, commit: 2, ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var b32 = sequenceBuilder{id: 32, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b175 = charBuilder{}
	b32.items = []builder{&b175}
	var b178 = sequenceBuilder{id: 178, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b178.items = []builder{&b186, &b32}
	b179.items = []builder{&b186, &b32, &b178}
	b187.items = []builder{&b177, &b186, &b126, &b179}
	b188.items = []builder{&b186, &b187, &b186}

	return parseInput(r, &p188, &b188)
}
