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
	var p16 = choiceParser{id: 16, commit: 66, name: "wschar", generalizations: []int{185, 186}}
	var p96 = sequenceParser{id: 96, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{16, 185, 186}}
	var p179 = charParser{id: 179, chars: []rune{32}}
	p96.items = []parser{&p179}
	var p1 = sequenceParser{id: 1, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{16, 185, 186}}
	var p83 = charParser{id: 83, chars: []rune{9}}
	p1.items = []parser{&p83}
	var p175 = sequenceParser{id: 175, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{16, 185, 186}}
	var p116 = charParser{id: 116, chars: []rune{10}}
	p175.items = []parser{&p116}
	var p85 = sequenceParser{id: 85, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{16, 185, 186}}
	var p84 = charParser{id: 84, chars: []rune{8}}
	p85.items = []parser{&p84}
	var p44 = sequenceParser{id: 44, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{16, 185, 186}}
	var p103 = charParser{id: 103, chars: []rune{12}}
	p44.items = []parser{&p103}
	var p117 = sequenceParser{id: 117, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{16, 185, 186}}
	var p154 = charParser{id: 154, chars: []rune{13}}
	p117.items = []parser{&p154}
	var p90 = sequenceParser{id: 90, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{16, 185, 186}}
	var p55 = charParser{id: 55, chars: []rune{11}}
	p90.items = []parser{&p55}
	p16.options = []parser{&p96, &p1, &p175, &p85, &p44, &p117, &p90}
	var p171 = sequenceParser{id: 171, commit: 72, name: "comment", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}, generalizations: []int{185, 186}}
	var p160 = choiceParser{id: 160, commit: 74, name: "comment-segment"}
	var p124 = sequenceParser{id: 124, commit: 74, name: "line-comment", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}, generalizations: []int{160}}
	var p139 = sequenceParser{id: 139, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var p67 = charParser{id: 67, chars: []rune{47}}
	var p143 = charParser{id: 143, chars: []rune{47}}
	p139.items = []parser{&p67, &p143}
	var p144 = sequenceParser{id: 144, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p29 = charParser{id: 29, not: true, chars: []rune{10}}
	p144.items = []parser{&p29}
	p124.items = []parser{&p139, &p144}
	var p110 = sequenceParser{id: 110, commit: 74, name: "block-comment", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {1, 1}, {0, -1}, {1, 1}}, generalizations: []int{160}}
	var p45 = sequenceParser{id: 45, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var p2 = charParser{id: 2, chars: []rune{47}}
	var p100 = charParser{id: 100, chars: []rune{42}}
	p45.items = []parser{&p2, &p100}
	var p36 = choiceParser{id: 36, commit: 10}
	var p180 = sequenceParser{id: 180, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{36}}
	var p159 = sequenceParser{id: 159, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p3 = charParser{id: 3, chars: []rune{42}}
	p159.items = []parser{&p3}
	var p91 = sequenceParser{id: 91, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p46 = charParser{id: 46, not: true, chars: []rune{47}}
	p91.items = []parser{&p46}
	p180.items = []parser{&p159, &p91}
	var p66 = sequenceParser{id: 66, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{36}}
	var p131 = charParser{id: 131, not: true, chars: []rune{42}}
	p66.items = []parser{&p131}
	p36.options = []parser{&p180, &p66}
	var p97 = sequenceParser{id: 97, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var p47 = charParser{id: 47, chars: []rune{42}}
	var p73 = charParser{id: 73, chars: []rune{47}}
	p97.items = []parser{&p47, &p73}
	p110.items = []parser{&p45, &p36, &p97}
	p160.options = []parser{&p124, &p110}
	var p149 = sequenceParser{id: 149, commit: 10, ranges: [][]int{{0, -1}, {0, 1}, {0, -1}, {1, 1}, {0, -1}, {0, 1}, {0, -1}, {1, 1}}}
	var p111 = choiceParser{id: 111, commit: 74, name: "ws-no-nl"}
	var p37 = sequenceParser{id: 37, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{111}}
	var p74 = charParser{id: 74, chars: []rune{32}}
	p37.items = []parser{&p74}
	var p80 = sequenceParser{id: 80, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{111}}
	var p38 = charParser{id: 38, chars: []rune{9}}
	p80.items = []parser{&p38}
	var p92 = sequenceParser{id: 92, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{111}}
	var p161 = charParser{id: 161, chars: []rune{8}}
	p92.items = []parser{&p161}
	var p148 = sequenceParser{id: 148, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{111}}
	var p56 = charParser{id: 56, chars: []rune{12}}
	p148.items = []parser{&p56}
	var p166 = sequenceParser{id: 166, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{111}}
	var p8 = charParser{id: 8, chars: []rune{13}}
	p166.items = []parser{&p8}
	var p155 = sequenceParser{id: 155, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{111}}
	var p4 = charParser{id: 4, chars: []rune{11}}
	p155.items = []parser{&p4}
	p111.options = []parser{&p37, &p80, &p92, &p148, &p166, &p155}
	var p167 = sequenceParser{id: 167, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p162 = charParser{id: 162, chars: []rune{10}}
	p167.items = []parser{&p162}
	p149.items = []parser{&p111, &p167, &p111, &p160}
	p171.items = []parser{&p160, &p149}
	p185.options = []parser{&p16, &p171}
	p186.options = []parser{&p185}
	var p187 = sequenceParser{id: 187, commit: 66, name: "syntax:wsroot", ranges: [][]int{{0, 1}, {0, -1}, {0, 1}, {0, 1}}}
	var p136 = sequenceParser{id: 136, commit: 2, ranges: [][]int{{1, 1}, {0, -1}}}
	var p7 = sequenceParser{id: 7, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p115 = charParser{id: 115, chars: []rune{59}}
	p7.items = []parser{&p115}
	var p135 = sequenceParser{id: 135, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p135.items = []parser{&p186, &p7}
	p136.items = []parser{&p7, &p135}
	var p35 = sequenceParser{id: 35, commit: 66, name: "definitions", ranges: [][]int{{1, 1}, {0, 1}}}
	var p72 = sequenceParser{id: 72, commit: 64, name: "definition", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}}}
	var p178 = sequenceParser{id: 178, commit: 74, name: "definition-name", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}}
	var p119 = sequenceParser{id: 119, commit: 72, name: "symbol", ranges: [][]int{{1, -1}, {1, -1}}, generalizations: []int{23, 158, 6}}
	var p101 = sequenceParser{id: 101, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p86 = charParser{id: 86, not: true, chars: []rune{92, 32, 10, 9, 8, 12, 13, 11, 47, 46, 91, 93, 34, 123, 125, 94, 43, 42, 63, 124, 40, 41, 58, 61, 59}}
	p101.items = []parser{&p86}
	p119.items = []parser{&p101}
	var p146 = sequenceParser{id: 146, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var p28 = sequenceParser{id: 28, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p82 = charParser{id: 82, chars: []rune{58}}
	p28.items = []parser{&p82}
	var p14 = choiceParser{id: 14, commit: 66, name: "flag"}
	var p123 = sequenceParser{id: 123, commit: 72, name: "alias", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{14}}
	var p60 = charParser{id: 60, chars: []rune{97}}
	var p174 = charParser{id: 174, chars: []rune{108}}
	var p43 = charParser{id: 43, chars: []rune{105}}
	var p51 = charParser{id: 51, chars: []rune{97}}
	var p11 = charParser{id: 11, chars: []rune{115}}
	p123.items = []parser{&p60, &p174, &p43, &p51, &p11}
	var p78 = sequenceParser{id: 78, commit: 72, name: "ws", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{14}}
	var p52 = charParser{id: 52, chars: []rune{119}}
	var p12 = charParser{id: 12, chars: []rune{115}}
	p78.items = []parser{&p52, &p12}
	var p109 = sequenceParser{id: 109, commit: 72, name: "nows", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{14}}
	var p184 = charParser{id: 184, chars: []rune{110}}
	var p79 = charParser{id: 79, chars: []rune{111}}
	var p88 = charParser{id: 88, chars: []rune{119}}
	var p128 = charParser{id: 128, chars: []rune{115}}
	p109.items = []parser{&p184, &p79, &p88, &p128}
	var p71 = sequenceParser{id: 71, commit: 72, name: "failpass", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{14}}
	var p32 = charParser{id: 32, chars: []rune{102}}
	var p134 = charParser{id: 134, chars: []rune{97}}
	var p53 = charParser{id: 53, chars: []rune{105}}
	var p129 = charParser{id: 129, chars: []rune{108}}
	var p65 = charParser{id: 65, chars: []rune{112}}
	var p145 = charParser{id: 145, chars: []rune{97}}
	var p24 = charParser{id: 24, chars: []rune{115}}
	var p13 = charParser{id: 13, chars: []rune{115}}
	p71.items = []parser{&p32, &p134, &p53, &p129, &p65, &p145, &p24, &p13}
	var p54 = sequenceParser{id: 54, commit: 72, name: "root", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{14}}
	var p27 = charParser{id: 27, chars: []rune{114}}
	var p169 = charParser{id: 169, chars: []rune{111}}
	var p142 = charParser{id: 142, chars: []rune{111}}
	var p170 = charParser{id: 170, chars: []rune{116}}
	p54.items = []parser{&p27, &p169, &p142, &p170}
	p14.options = []parser{&p123, &p78, &p109, &p71, &p54}
	p146.items = []parser{&p28, &p14}
	p178.items = []parser{&p119, &p146}
	var p89 = sequenceParser{id: 89, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p151 = charParser{id: 151, chars: []rune{61}}
	p89.items = []parser{&p151}
	var p23 = choiceParser{id: 23, commit: 66, name: "expression"}
	var p181 = choiceParser{id: 181, commit: 66, name: "terminal", generalizations: []int{23, 158, 6}}
	var p39 = sequenceParser{id: 39, commit: 72, name: "any-char", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{181, 23, 158, 6}}
	var p25 = charParser{id: 25, chars: []rune{46}}
	p39.items = []parser{&p25}
	var p58 = sequenceParser{id: 58, commit: 72, name: "char-class", ranges: [][]int{{1, 1}, {0, 1}, {0, -1}, {1, 1}, {1, 1}, {0, 1}, {0, -1}, {1, 1}}, generalizations: []int{181, 23, 158, 6}}
	var p125 = sequenceParser{id: 125, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p18 = charParser{id: 18, chars: []rune{91}}
	p125.items = []parser{&p18}
	var p10 = sequenceParser{id: 10, commit: 72, name: "class-not", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p9 = charParser{id: 9, chars: []rune{94}}
	p10.items = []parser{&p9}
	var p156 = choiceParser{id: 156, commit: 10}
	var p132 = choiceParser{id: 132, commit: 72, name: "class-char", generalizations: []int{156}}
	var p26 = sequenceParser{id: 26, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{132, 156}}
	var p57 = charParser{id: 57, not: true, chars: []rune{92, 91, 93, 94, 45}}
	p26.items = []parser{&p57}
	var p98 = sequenceParser{id: 98, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{132, 156}}
	var p118 = sequenceParser{id: 118, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p5 = charParser{id: 5, chars: []rune{92}}
	p118.items = []parser{&p5}
	var p102 = sequenceParser{id: 102, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p61 = charParser{id: 61, not: true}
	p102.items = []parser{&p61}
	p98.items = []parser{&p118, &p102}
	p132.options = []parser{&p26, &p98}
	var p140 = sequenceParser{id: 140, commit: 72, name: "char-range", ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{156}}
	var p62 = sequenceParser{id: 62, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p17 = charParser{id: 17, chars: []rune{45}}
	p62.items = []parser{&p17}
	p140.items = []parser{&p132, &p62, &p132}
	p156.options = []parser{&p132, &p140}
	var p93 = sequenceParser{id: 93, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p150 = charParser{id: 150, chars: []rune{93}}
	p93.items = []parser{&p150}
	p58.items = []parser{&p125, &p10, &p156, &p93}
	var p20 = sequenceParser{id: 20, commit: 72, name: "char-sequence", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {1, 1}, {0, -1}, {1, 1}}, generalizations: []int{181, 23, 158, 6}}
	var p68 = sequenceParser{id: 68, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p63 = charParser{id: 63, chars: []rune{34}}
	p68.items = []parser{&p63}
	var p31 = choiceParser{id: 31, commit: 72, name: "sequence-char"}
	var p48 = sequenceParser{id: 48, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{31}}
	var p157 = charParser{id: 157, not: true, chars: []rune{92, 34}}
	p48.items = []parser{&p157}
	var p94 = sequenceParser{id: 94, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{31}}
	var p75 = sequenceParser{id: 75, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p19 = charParser{id: 19, chars: []rune{92}}
	p75.items = []parser{&p19}
	var p126 = sequenceParser{id: 126, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p133 = charParser{id: 133, not: true}
	p126.items = []parser{&p133}
	p94.items = []parser{&p75, &p126}
	p31.options = []parser{&p48, &p94}
	var p69 = sequenceParser{id: 69, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p104 = charParser{id: 104, chars: []rune{34}}
	p69.items = []parser{&p104}
	p20.items = []parser{&p68, &p31, &p69}
	p181.options = []parser{&p39, &p58, &p20}
	var p40 = sequenceParser{id: 40, commit: 66, name: "group", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}}, generalizations: []int{23, 158, 6}}
	var p105 = sequenceParser{id: 105, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p95 = charParser{id: 95, chars: []rune{40}}
	p105.items = []parser{&p95}
	var p99 = sequenceParser{id: 99, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p59 = charParser{id: 59, chars: []rune{41}}
	p99.items = []parser{&p59}
	p40.items = []parser{&p105, &p186, &p23, &p186, &p99}
	var p108 = sequenceParser{id: 108, commit: 64, name: "sequence", ranges: [][]int{{1, 1}, {0, -1}}, generalizations: []int{23, 6}}
	var p15 = sequenceParser{id: 15, commit: 72, name: "item", ranges: [][]int{{1, 1}, {0, 1}, {1, 1}, {0, 1}}}
	var p158 = choiceParser{id: 158, commit: 10}
	p158.options = []parser{&p181, &p119, &p40}
	var p50 = choiceParser{id: 50, commit: 66, name: "quantity"}
	var p41 = sequenceParser{id: 41, commit: 64, name: "count-quantifier", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}}, generalizations: []int{50}}
	var p164 = sequenceParser{id: 164, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p49 = charParser{id: 49, chars: []rune{123}}
	p164.items = []parser{&p49}
	var p176 = sequenceParser{id: 176, commit: 64, name: "count", ranges: [][]int{{1, 1}}}
	var p120 = sequenceParser{id: 120, commit: 74, name: "number", ranges: [][]int{{1, -1}, {1, -1}}}
	var p182 = sequenceParser{id: 182, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p163 = charParser{id: 163, ranges: [][]rune{{48, 57}}}
	p182.items = []parser{&p163}
	p120.items = []parser{&p182}
	p176.items = []parser{&p120}
	var p87 = sequenceParser{id: 87, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p141 = charParser{id: 141, chars: []rune{125}}
	p87.items = []parser{&p141}
	p41.items = []parser{&p164, &p186, &p176, &p186, &p87}
	var p70 = sequenceParser{id: 70, commit: 64, name: "range-quantifier", ranges: [][]int{{1, 1}, {0, -1}, {0, 1}, {0, -1}, {1, 1}, {0, -1}, {0, 1}, {0, -1}, {1, 1}}, generalizations: []int{50}}
	var p106 = sequenceParser{id: 106, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p76 = charParser{id: 76, chars: []rune{123}}
	p106.items = []parser{&p76}
	var p42 = sequenceParser{id: 42, commit: 64, name: "range-from", ranges: [][]int{{1, 1}}}
	p42.items = []parser{&p120}
	var p172 = sequenceParser{id: 172, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p112 = charParser{id: 112, chars: []rune{44}}
	p172.items = []parser{&p112}
	var p21 = sequenceParser{id: 21, commit: 64, name: "range-to", ranges: [][]int{{1, 1}}}
	p21.items = []parser{&p120}
	var p165 = sequenceParser{id: 165, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p81 = charParser{id: 81, chars: []rune{125}}
	p165.items = []parser{&p81}
	p70.items = []parser{&p106, &p186, &p42, &p186, &p172, &p186, &p21, &p186, &p165}
	var p127 = sequenceParser{id: 127, commit: 72, name: "one-or-more", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{50}}
	var p183 = charParser{id: 183, chars: []rune{43}}
	p127.items = []parser{&p183}
	var p22 = sequenceParser{id: 22, commit: 72, name: "zero-or-more", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{50}}
	var p177 = charParser{id: 177, chars: []rune{42}}
	p22.items = []parser{&p177}
	var p168 = sequenceParser{id: 168, commit: 72, name: "zero-or-one", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{50}}
	var p30 = charParser{id: 30, chars: []rune{63}}
	p168.items = []parser{&p30}
	p50.options = []parser{&p41, &p70, &p127, &p22, &p168}
	p15.items = []parser{&p158, &p50}
	var p107 = sequenceParser{id: 107, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p107.items = []parser{&p186, &p15}
	p108.items = []parser{&p15, &p107}
	var p122 = sequenceParser{id: 122, commit: 64, name: "choice", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}, generalizations: []int{23}}
	var p6 = choiceParser{id: 6, commit: 66, name: "option"}
	p6.options = []parser{&p181, &p119, &p40, &p108}
	var p77 = sequenceParser{id: 77, commit: 2, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}}}
	var p173 = sequenceParser{id: 173, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p64 = charParser{id: 64, chars: []rune{124}}
	p173.items = []parser{&p64}
	p77.items = []parser{&p173, &p186, &p6}
	var p121 = sequenceParser{id: 121, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p121.items = []parser{&p186, &p77}
	p122.items = []parser{&p6, &p186, &p77, &p121}
	p23.options = []parser{&p181, &p119, &p40, &p108, &p122}
	p72.items = []parser{&p178, &p186, &p89, &p186, &p23}
	var p34 = sequenceParser{id: 34, commit: 2, ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var p114 = sequenceParser{id: 114, commit: 2, ranges: [][]int{{1, 1}, {0, -1}, {0, -1}, {1, 1}}}
	var p147 = sequenceParser{id: 147, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p152 = charParser{id: 152, chars: []rune{59}}
	p147.items = []parser{&p152}
	var p113 = sequenceParser{id: 113, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p113.items = []parser{&p186, &p147}
	p114.items = []parser{&p147, &p113, &p186, &p72}
	var p33 = sequenceParser{id: 33, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p33.items = []parser{&p186, &p114}
	p34.items = []parser{&p186, &p114, &p33}
	p35.items = []parser{&p72, &p34}
	var p138 = sequenceParser{id: 138, commit: 2, ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var p130 = sequenceParser{id: 130, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p153 = charParser{id: 153, chars: []rune{59}}
	p130.items = []parser{&p153}
	var p137 = sequenceParser{id: 137, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p137.items = []parser{&p186, &p130}
	p138.items = []parser{&p186, &p130, &p137}
	p187.items = []parser{&p136, &p186, &p35, &p138}
	p188.items = []parser{&p186, &p187, &p186}
	var b188 = sequenceBuilder{id: 188, commit: 32, name: "syntax", ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var b186 = choiceBuilder{id: 186, commit: 2}
	var b185 = choiceBuilder{id: 185, commit: 70}
	var b16 = choiceBuilder{id: 16, commit: 66}
	var b96 = sequenceBuilder{id: 96, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b179 = charBuilder{}
	b96.items = []builder{&b179}
	var b1 = sequenceBuilder{id: 1, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b83 = charBuilder{}
	b1.items = []builder{&b83}
	var b175 = sequenceBuilder{id: 175, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b116 = charBuilder{}
	b175.items = []builder{&b116}
	var b85 = sequenceBuilder{id: 85, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b84 = charBuilder{}
	b85.items = []builder{&b84}
	var b44 = sequenceBuilder{id: 44, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b103 = charBuilder{}
	b44.items = []builder{&b103}
	var b117 = sequenceBuilder{id: 117, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b154 = charBuilder{}
	b117.items = []builder{&b154}
	var b90 = sequenceBuilder{id: 90, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b55 = charBuilder{}
	b90.items = []builder{&b55}
	b16.options = []builder{&b96, &b1, &b175, &b85, &b44, &b117, &b90}
	var b171 = sequenceBuilder{id: 171, commit: 72, name: "comment", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}}
	var b160 = choiceBuilder{id: 160, commit: 74}
	var b124 = sequenceBuilder{id: 124, commit: 74, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}}
	var b139 = sequenceBuilder{id: 139, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b67 = charBuilder{}
	var b143 = charBuilder{}
	b139.items = []builder{&b67, &b143}
	var b144 = sequenceBuilder{id: 144, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b29 = charBuilder{}
	b144.items = []builder{&b29}
	b124.items = []builder{&b139, &b144}
	var b110 = sequenceBuilder{id: 110, commit: 74, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {1, 1}, {0, -1}, {1, 1}}}
	var b45 = sequenceBuilder{id: 45, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b2 = charBuilder{}
	var b100 = charBuilder{}
	b45.items = []builder{&b2, &b100}
	var b36 = choiceBuilder{id: 36, commit: 10}
	var b180 = sequenceBuilder{id: 180, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b159 = sequenceBuilder{id: 159, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b3 = charBuilder{}
	b159.items = []builder{&b3}
	var b91 = sequenceBuilder{id: 91, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b46 = charBuilder{}
	b91.items = []builder{&b46}
	b180.items = []builder{&b159, &b91}
	var b66 = sequenceBuilder{id: 66, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b131 = charBuilder{}
	b66.items = []builder{&b131}
	b36.options = []builder{&b180, &b66}
	var b97 = sequenceBuilder{id: 97, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b47 = charBuilder{}
	var b73 = charBuilder{}
	b97.items = []builder{&b47, &b73}
	b110.items = []builder{&b45, &b36, &b97}
	b160.options = []builder{&b124, &b110}
	var b149 = sequenceBuilder{id: 149, commit: 10, ranges: [][]int{{0, -1}, {0, 1}, {0, -1}, {1, 1}, {0, -1}, {0, 1}, {0, -1}, {1, 1}}}
	var b111 = choiceBuilder{id: 111, commit: 74}
	var b37 = sequenceBuilder{id: 37, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b74 = charBuilder{}
	b37.items = []builder{&b74}
	var b80 = sequenceBuilder{id: 80, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b38 = charBuilder{}
	b80.items = []builder{&b38}
	var b92 = sequenceBuilder{id: 92, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b161 = charBuilder{}
	b92.items = []builder{&b161}
	var b148 = sequenceBuilder{id: 148, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b56 = charBuilder{}
	b148.items = []builder{&b56}
	var b166 = sequenceBuilder{id: 166, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b8 = charBuilder{}
	b166.items = []builder{&b8}
	var b155 = sequenceBuilder{id: 155, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b4 = charBuilder{}
	b155.items = []builder{&b4}
	b111.options = []builder{&b37, &b80, &b92, &b148, &b166, &b155}
	var b167 = sequenceBuilder{id: 167, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b162 = charBuilder{}
	b167.items = []builder{&b162}
	b149.items = []builder{&b111, &b167, &b111, &b160}
	b171.items = []builder{&b160, &b149}
	b185.options = []builder{&b16, &b171}
	b186.options = []builder{&b185}
	var b187 = sequenceBuilder{id: 187, commit: 66, ranges: [][]int{{0, 1}, {0, -1}, {0, 1}, {0, 1}}}
	var b136 = sequenceBuilder{id: 136, commit: 2, ranges: [][]int{{1, 1}, {0, -1}}}
	var b7 = sequenceBuilder{id: 7, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b115 = charBuilder{}
	b7.items = []builder{&b115}
	var b135 = sequenceBuilder{id: 135, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b135.items = []builder{&b186, &b7}
	b136.items = []builder{&b7, &b135}
	var b35 = sequenceBuilder{id: 35, commit: 66, ranges: [][]int{{1, 1}, {0, 1}}}
	var b72 = sequenceBuilder{id: 72, commit: 64, name: "definition", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}}}
	var b178 = sequenceBuilder{id: 178, commit: 74, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}}
	var b119 = sequenceBuilder{id: 119, commit: 72, name: "symbol", ranges: [][]int{{1, -1}, {1, -1}}}
	var b101 = sequenceBuilder{id: 101, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b86 = charBuilder{}
	b101.items = []builder{&b86}
	b119.items = []builder{&b101}
	var b146 = sequenceBuilder{id: 146, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b28 = sequenceBuilder{id: 28, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b82 = charBuilder{}
	b28.items = []builder{&b82}
	var b14 = choiceBuilder{id: 14, commit: 66}
	var b123 = sequenceBuilder{id: 123, commit: 72, name: "alias", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b60 = charBuilder{}
	var b174 = charBuilder{}
	var b43 = charBuilder{}
	var b51 = charBuilder{}
	var b11 = charBuilder{}
	b123.items = []builder{&b60, &b174, &b43, &b51, &b11}
	var b78 = sequenceBuilder{id: 78, commit: 72, name: "ws", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b52 = charBuilder{}
	var b12 = charBuilder{}
	b78.items = []builder{&b52, &b12}
	var b109 = sequenceBuilder{id: 109, commit: 72, name: "nows", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b184 = charBuilder{}
	var b79 = charBuilder{}
	var b88 = charBuilder{}
	var b128 = charBuilder{}
	b109.items = []builder{&b184, &b79, &b88, &b128}
	var b71 = sequenceBuilder{id: 71, commit: 72, name: "failpass", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b32 = charBuilder{}
	var b134 = charBuilder{}
	var b53 = charBuilder{}
	var b129 = charBuilder{}
	var b65 = charBuilder{}
	var b145 = charBuilder{}
	var b24 = charBuilder{}
	var b13 = charBuilder{}
	b71.items = []builder{&b32, &b134, &b53, &b129, &b65, &b145, &b24, &b13}
	var b54 = sequenceBuilder{id: 54, commit: 72, name: "root", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b27 = charBuilder{}
	var b169 = charBuilder{}
	var b142 = charBuilder{}
	var b170 = charBuilder{}
	b54.items = []builder{&b27, &b169, &b142, &b170}
	b14.options = []builder{&b123, &b78, &b109, &b71, &b54}
	b146.items = []builder{&b28, &b14}
	b178.items = []builder{&b119, &b146}
	var b89 = sequenceBuilder{id: 89, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b151 = charBuilder{}
	b89.items = []builder{&b151}
	var b23 = choiceBuilder{id: 23, commit: 66}
	var b181 = choiceBuilder{id: 181, commit: 66}
	var b39 = sequenceBuilder{id: 39, commit: 72, name: "any-char", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b25 = charBuilder{}
	b39.items = []builder{&b25}
	var b58 = sequenceBuilder{id: 58, commit: 72, name: "char-class", ranges: [][]int{{1, 1}, {0, 1}, {0, -1}, {1, 1}, {1, 1}, {0, 1}, {0, -1}, {1, 1}}}
	var b125 = sequenceBuilder{id: 125, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b18 = charBuilder{}
	b125.items = []builder{&b18}
	var b10 = sequenceBuilder{id: 10, commit: 72, name: "class-not", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b9 = charBuilder{}
	b10.items = []builder{&b9}
	var b156 = choiceBuilder{id: 156, commit: 10}
	var b132 = choiceBuilder{id: 132, commit: 72, name: "class-char"}
	var b26 = sequenceBuilder{id: 26, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b57 = charBuilder{}
	b26.items = []builder{&b57}
	var b98 = sequenceBuilder{id: 98, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b118 = sequenceBuilder{id: 118, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b5 = charBuilder{}
	b118.items = []builder{&b5}
	var b102 = sequenceBuilder{id: 102, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b61 = charBuilder{}
	b102.items = []builder{&b61}
	b98.items = []builder{&b118, &b102}
	b132.options = []builder{&b26, &b98}
	var b140 = sequenceBuilder{id: 140, commit: 72, name: "char-range", ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b62 = sequenceBuilder{id: 62, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b17 = charBuilder{}
	b62.items = []builder{&b17}
	b140.items = []builder{&b132, &b62, &b132}
	b156.options = []builder{&b132, &b140}
	var b93 = sequenceBuilder{id: 93, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b150 = charBuilder{}
	b93.items = []builder{&b150}
	b58.items = []builder{&b125, &b10, &b156, &b93}
	var b20 = sequenceBuilder{id: 20, commit: 72, name: "char-sequence", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {1, 1}, {0, -1}, {1, 1}}}
	var b68 = sequenceBuilder{id: 68, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b63 = charBuilder{}
	b68.items = []builder{&b63}
	var b31 = choiceBuilder{id: 31, commit: 72, name: "sequence-char"}
	var b48 = sequenceBuilder{id: 48, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b157 = charBuilder{}
	b48.items = []builder{&b157}
	var b94 = sequenceBuilder{id: 94, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b75 = sequenceBuilder{id: 75, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b19 = charBuilder{}
	b75.items = []builder{&b19}
	var b126 = sequenceBuilder{id: 126, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b133 = charBuilder{}
	b126.items = []builder{&b133}
	b94.items = []builder{&b75, &b126}
	b31.options = []builder{&b48, &b94}
	var b69 = sequenceBuilder{id: 69, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b104 = charBuilder{}
	b69.items = []builder{&b104}
	b20.items = []builder{&b68, &b31, &b69}
	b181.options = []builder{&b39, &b58, &b20}
	var b40 = sequenceBuilder{id: 40, commit: 66, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}}}
	var b105 = sequenceBuilder{id: 105, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b95 = charBuilder{}
	b105.items = []builder{&b95}
	var b99 = sequenceBuilder{id: 99, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b59 = charBuilder{}
	b99.items = []builder{&b59}
	b40.items = []builder{&b105, &b186, &b23, &b186, &b99}
	var b108 = sequenceBuilder{id: 108, commit: 64, name: "sequence", ranges: [][]int{{1, 1}, {0, -1}}}
	var b15 = sequenceBuilder{id: 15, commit: 72, name: "item", ranges: [][]int{{1, 1}, {0, 1}, {1, 1}, {0, 1}}}
	var b158 = choiceBuilder{id: 158, commit: 10}
	b158.options = []builder{&b181, &b119, &b40}
	var b50 = choiceBuilder{id: 50, commit: 66}
	var b41 = sequenceBuilder{id: 41, commit: 64, name: "count-quantifier", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}}}
	var b164 = sequenceBuilder{id: 164, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b49 = charBuilder{}
	b164.items = []builder{&b49}
	var b176 = sequenceBuilder{id: 176, commit: 64, name: "count", ranges: [][]int{{1, 1}}}
	var b120 = sequenceBuilder{id: 120, commit: 74, ranges: [][]int{{1, -1}, {1, -1}}}
	var b182 = sequenceBuilder{id: 182, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b163 = charBuilder{}
	b182.items = []builder{&b163}
	b120.items = []builder{&b182}
	b176.items = []builder{&b120}
	var b87 = sequenceBuilder{id: 87, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b141 = charBuilder{}
	b87.items = []builder{&b141}
	b41.items = []builder{&b164, &b186, &b176, &b186, &b87}
	var b70 = sequenceBuilder{id: 70, commit: 64, name: "range-quantifier", ranges: [][]int{{1, 1}, {0, -1}, {0, 1}, {0, -1}, {1, 1}, {0, -1}, {0, 1}, {0, -1}, {1, 1}}}
	var b106 = sequenceBuilder{id: 106, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b76 = charBuilder{}
	b106.items = []builder{&b76}
	var b42 = sequenceBuilder{id: 42, commit: 64, name: "range-from", ranges: [][]int{{1, 1}}}
	b42.items = []builder{&b120}
	var b172 = sequenceBuilder{id: 172, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b112 = charBuilder{}
	b172.items = []builder{&b112}
	var b21 = sequenceBuilder{id: 21, commit: 64, name: "range-to", ranges: [][]int{{1, 1}}}
	b21.items = []builder{&b120}
	var b165 = sequenceBuilder{id: 165, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b81 = charBuilder{}
	b165.items = []builder{&b81}
	b70.items = []builder{&b106, &b186, &b42, &b186, &b172, &b186, &b21, &b186, &b165}
	var b127 = sequenceBuilder{id: 127, commit: 72, name: "one-or-more", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b183 = charBuilder{}
	b127.items = []builder{&b183}
	var b22 = sequenceBuilder{id: 22, commit: 72, name: "zero-or-more", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b177 = charBuilder{}
	b22.items = []builder{&b177}
	var b168 = sequenceBuilder{id: 168, commit: 72, name: "zero-or-one", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b30 = charBuilder{}
	b168.items = []builder{&b30}
	b50.options = []builder{&b41, &b70, &b127, &b22, &b168}
	b15.items = []builder{&b158, &b50}
	var b107 = sequenceBuilder{id: 107, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b107.items = []builder{&b186, &b15}
	b108.items = []builder{&b15, &b107}
	var b122 = sequenceBuilder{id: 122, commit: 64, name: "choice", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}}
	var b6 = choiceBuilder{id: 6, commit: 66}
	b6.options = []builder{&b181, &b119, &b40, &b108}
	var b77 = sequenceBuilder{id: 77, commit: 2, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}}}
	var b173 = sequenceBuilder{id: 173, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b64 = charBuilder{}
	b173.items = []builder{&b64}
	b77.items = []builder{&b173, &b186, &b6}
	var b121 = sequenceBuilder{id: 121, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b121.items = []builder{&b186, &b77}
	b122.items = []builder{&b6, &b186, &b77, &b121}
	b23.options = []builder{&b181, &b119, &b40, &b108, &b122}
	b72.items = []builder{&b178, &b186, &b89, &b186, &b23}
	var b34 = sequenceBuilder{id: 34, commit: 2, ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var b114 = sequenceBuilder{id: 114, commit: 2, ranges: [][]int{{1, 1}, {0, -1}, {0, -1}, {1, 1}}}
	var b147 = sequenceBuilder{id: 147, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b152 = charBuilder{}
	b147.items = []builder{&b152}
	var b113 = sequenceBuilder{id: 113, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b113.items = []builder{&b186, &b147}
	b114.items = []builder{&b147, &b113, &b186, &b72}
	var b33 = sequenceBuilder{id: 33, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b33.items = []builder{&b186, &b114}
	b34.items = []builder{&b186, &b114, &b33}
	b35.items = []builder{&b72, &b34}
	var b138 = sequenceBuilder{id: 138, commit: 2, ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var b130 = sequenceBuilder{id: 130, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b153 = charBuilder{}
	b130.items = []builder{&b153}
	var b137 = sequenceBuilder{id: 137, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b137.items = []builder{&b186, &b130}
	b138.items = []builder{&b186, &b130, &b137}
	b187.items = []builder{&b136, &b186, &b35, &b138}
	b188.items = []builder{&b186, &b187, &b186}

	return parseInput(r, &p188, &b188)
}
