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
	var p56 = choiceParser{id: 56, commit: 66, name: "wschar", generalizations: []int{185, 186}}
	var p162 = sequenceParser{id: 162, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{56, 185, 186}}
	var p138 = charParser{id: 138, chars: []rune{32}}
	p162.items = []parser{&p138}
	var p140 = sequenceParser{id: 140, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{56, 185, 186}}
	var p139 = charParser{id: 139, chars: []rune{9}}
	p140.items = []parser{&p139}
	var p112 = sequenceParser{id: 112, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{56, 185, 186}}
	var p141 = charParser{id: 141, chars: []rune{10}}
	p112.items = []parser{&p141}
	var p54 = sequenceParser{id: 54, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{56, 185, 186}}
	var p133 = charParser{id: 133, chars: []rune{8}}
	p54.items = []parser{&p133}
	var p99 = sequenceParser{id: 99, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{56, 185, 186}}
	var p98 = charParser{id: 98, chars: []rune{12}}
	p99.items = []parser{&p98}
	var p121 = sequenceParser{id: 121, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{56, 185, 186}}
	var p21 = charParser{id: 21, chars: []rune{13}}
	p121.items = []parser{&p21}
	var p168 = sequenceParser{id: 168, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{56, 185, 186}}
	var p55 = charParser{id: 55, chars: []rune{11}}
	p168.items = []parser{&p55}
	p56.options = []parser{&p162, &p140, &p112, &p54, &p99, &p121, &p168}
	var p101 = sequenceParser{id: 101, commit: 72, name: "comment", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}, generalizations: []int{185, 186}}
	var p175 = choiceParser{id: 175, commit: 74, name: "comment-segment"}
	var p77 = sequenceParser{id: 77, commit: 74, name: "line-comment", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}, generalizations: []int{175}}
	var p76 = sequenceParser{id: 76, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var p18 = charParser{id: 18, chars: []rune{47}}
	var p57 = charParser{id: 57, chars: []rune{47}}
	p76.items = []parser{&p18, &p57}
	var p127 = sequenceParser{id: 127, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p100 = charParser{id: 100, not: true, chars: []rune{10}}
	p127.items = []parser{&p100}
	p77.items = []parser{&p76, &p127}
	var p27 = sequenceParser{id: 27, commit: 74, name: "block-comment", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {1, 1}, {0, -1}, {1, 1}}, generalizations: []int{175}}
	var p12 = sequenceParser{id: 12, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var p148 = charParser{id: 148, chars: []rune{47}}
	var p169 = charParser{id: 169, chars: []rune{42}}
	p12.items = []parser{&p148, &p169}
	var p34 = choiceParser{id: 34, commit: 10}
	var p1 = sequenceParser{id: 1, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{34}}
	var p170 = sequenceParser{id: 170, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p134 = charParser{id: 134, chars: []rune{42}}
	p170.items = []parser{&p134}
	var p179 = sequenceParser{id: 179, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p42 = charParser{id: 42, not: true, chars: []rune{47}}
	p179.items = []parser{&p42}
	p1.items = []parser{&p170, &p179}
	var p174 = sequenceParser{id: 174, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{34}}
	var p155 = charParser{id: 155, not: true, chars: []rune{42}}
	p174.items = []parser{&p155}
	p34.options = []parser{&p1, &p174}
	var p117 = sequenceParser{id: 117, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var p163 = charParser{id: 163, chars: []rune{42}}
	var p149 = charParser{id: 149, chars: []rune{47}}
	p117.items = []parser{&p163, &p149}
	p27.items = []parser{&p12, &p34, &p117}
	p175.options = []parser{&p77, &p27}
	var p176 = sequenceParser{id: 176, commit: 10, ranges: [][]int{{0, -1}, {0, 1}, {0, -1}, {1, 1}, {0, -1}, {0, 1}, {0, -1}, {1, 1}}}
	var p47 = choiceParser{id: 47, commit: 74, name: "ws-no-nl"}
	var p22 = sequenceParser{id: 22, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{47}}
	var p19 = charParser{id: 19, chars: []rune{32}}
	p22.items = []parser{&p19}
	var p28 = sequenceParser{id: 28, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{47}}
	var p81 = charParser{id: 81, chars: []rune{9}}
	p28.items = []parser{&p81}
	var p23 = sequenceParser{id: 23, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{47}}
	var p90 = charParser{id: 90, chars: []rune{8}}
	p23.items = []parser{&p90}
	var p5 = sequenceParser{id: 5, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{47}}
	var p68 = charParser{id: 68, chars: []rune{12}}
	p5.items = []parser{&p68}
	var p35 = sequenceParser{id: 35, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{47}}
	var p6 = charParser{id: 6, chars: []rune{13}}
	p35.items = []parser{&p6}
	var p69 = sequenceParser{id: 69, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{47}}
	var p61 = charParser{id: 61, chars: []rune{11}}
	p69.items = []parser{&p61}
	p47.options = []parser{&p22, &p28, &p23, &p5, &p35, &p69}
	var p118 = sequenceParser{id: 118, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p58 = charParser{id: 58, chars: []rune{10}}
	p118.items = []parser{&p58}
	p176.items = []parser{&p47, &p118, &p47, &p175}
	p101.items = []parser{&p175, &p176}
	p185.options = []parser{&p56, &p101}
	p186.options = []parser{&p185}
	var p187 = sequenceParser{id: 187, commit: 66, name: "syntax:wsroot", ranges: [][]int{{0, 1}, {0, -1}, {0, 1}, {0, 1}}}
	var p87 = sequenceParser{id: 87, commit: 2, ranges: [][]int{{1, 1}, {0, -1}}}
	var p142 = sequenceParser{id: 142, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p126 = charParser{id: 126, chars: []rune{59}}
	p142.items = []parser{&p126}
	var p86 = sequenceParser{id: 86, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p86.items = []parser{&p186, &p142}
	p87.items = []parser{&p142, &p86}
	var p66 = sequenceParser{id: 66, commit: 66, name: "definitions", ranges: [][]int{{1, 1}, {0, 1}}}
	var p73 = sequenceParser{id: 73, commit: 64, name: "definition", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}}}
	var p154 = sequenceParser{id: 154, commit: 74, name: "definition-name", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}}
	var p49 = sequenceParser{id: 49, commit: 72, name: "symbol", ranges: [][]int{{1, -1}, {1, -1}}, generalizations: []int{26, 103, 63}}
	var p2 = sequenceParser{id: 2, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p144 = charParser{id: 144, not: true, chars: []rune{92, 32, 10, 9, 8, 12, 13, 11, 47, 46, 91, 93, 34, 123, 125, 94, 43, 42, 63, 124, 40, 41, 58, 61, 59}}
	p2.items = []parser{&p144}
	p49.items = []parser{&p2}
	var p46 = sequenceParser{id: 46, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var p172 = sequenceParser{id: 172, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p125 = charParser{id: 125, chars: []rune{58}}
	p172.items = []parser{&p125}
	var p4 = choiceParser{id: 4, commit: 66, name: "flag"}
	var p44 = sequenceParser{id: 44, commit: 72, name: "alias", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{4}}
	var p33 = charParser{id: 33, chars: []rune{97}}
	var p75 = charParser{id: 75, chars: []rune{108}}
	var p16 = charParser{id: 16, chars: []rune{105}}
	var p51 = charParser{id: 51, chars: []rune{97}}
	var p40 = charParser{id: 40, chars: []rune{115}}
	p44.items = []parser{&p33, &p75, &p16, &p51, &p40}
	var p45 = sequenceParser{id: 45, commit: 72, name: "ws", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{4}}
	var p31 = charParser{id: 31, chars: []rune{119}}
	var p120 = charParser{id: 120, chars: []rune{115}}
	p45.items = []parser{&p31, &p120}
	var p17 = sequenceParser{id: 17, commit: 72, name: "nows", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{4}}
	var p97 = charParser{id: 97, chars: []rune{110}}
	var p183 = charParser{id: 183, chars: []rune{111}}
	var p93 = charParser{id: 93, chars: []rune{119}}
	var p146 = charParser{id: 146, chars: []rune{115}}
	p17.items = []parser{&p97, &p183, &p93, &p146}
	var p160 = sequenceParser{id: 160, commit: 72, name: "failpass", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{4}}
	var p79 = charParser{id: 79, chars: []rune{102}}
	var p80 = charParser{id: 80, chars: []rune{97}}
	var p111 = charParser{id: 111, chars: []rune{105}}
	var p94 = charParser{id: 94, chars: []rune{108}}
	var p132 = charParser{id: 132, chars: []rune{112}}
	var p53 = charParser{id: 53, chars: []rune{97}}
	var p72 = charParser{id: 72, chars: []rune{115}}
	var p105 = charParser{id: 105, chars: []rune{115}}
	p160.items = []parser{&p79, &p80, &p111, &p94, &p132, &p53, &p72, &p105}
	var p20 = sequenceParser{id: 20, commit: 72, name: "root", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{4}}
	var p147 = charParser{id: 147, chars: []rune{114}}
	var p184 = charParser{id: 184, chars: []rune{111}}
	var p153 = charParser{id: 153, chars: []rune{111}}
	var p137 = charParser{id: 137, chars: []rune{116}}
	p20.items = []parser{&p147, &p184, &p153, &p137}
	p4.options = []parser{&p44, &p45, &p17, &p160, &p20}
	p46.items = []parser{&p172, &p4}
	p154.items = []parser{&p49, &p46}
	var p152 = sequenceParser{id: 152, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p173 = charParser{id: 173, chars: []rune{61}}
	p152.items = []parser{&p173}
	var p26 = choiceParser{id: 26, commit: 66, name: "expression"}
	var p129 = choiceParser{id: 129, commit: 66, name: "terminal", generalizations: []int{26, 103, 63}}
	var p70 = sequenceParser{id: 70, commit: 72, name: "any-char", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{129, 26, 103, 63}}
	var p91 = charParser{id: 91, chars: []rune{46}}
	p70.items = []parser{&p91}
	var p62 = sequenceParser{id: 62, commit: 72, name: "char-class", ranges: [][]int{{1, 1}, {0, 1}, {0, -1}, {1, 1}, {1, 1}, {0, 1}, {0, -1}, {1, 1}}, generalizations: []int{129, 26, 103, 63}}
	var p71 = sequenceParser{id: 71, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p82 = charParser{id: 82, chars: []rune{91}}
	p71.items = []parser{&p82}
	var p180 = sequenceParser{id: 180, commit: 72, name: "class-not", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p29 = charParser{id: 29, chars: []rune{94}}
	p180.items = []parser{&p29}
	var p83 = choiceParser{id: 83, commit: 10}
	var p171 = choiceParser{id: 171, commit: 72, name: "class-char", generalizations: []int{83}}
	var p150 = sequenceParser{id: 150, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{171, 83}}
	var p113 = charParser{id: 113, not: true, chars: []rune{92, 91, 93, 94, 45}}
	p150.items = []parser{&p113}
	var p164 = sequenceParser{id: 164, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{171, 83}}
	var p59 = sequenceParser{id: 59, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p122 = charParser{id: 122, chars: []rune{92}}
	p59.items = []parser{&p122}
	var p106 = sequenceParser{id: 106, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p7 = charParser{id: 7, not: true}
	p106.items = []parser{&p7}
	p164.items = []parser{&p59, &p106}
	p171.options = []parser{&p150, &p164}
	var p177 = sequenceParser{id: 177, commit: 72, name: "char-range", ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{83}}
	var p78 = sequenceParser{id: 78, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p24 = charParser{id: 24, chars: []rune{45}}
	p78.items = []parser{&p24}
	p177.items = []parser{&p171, &p78, &p171}
	p83.options = []parser{&p171, &p177}
	var p84 = sequenceParser{id: 84, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p43 = charParser{id: 43, chars: []rune{93}}
	p84.items = []parser{&p43}
	p62.items = []parser{&p71, &p180, &p83, &p84}
	var p36 = sequenceParser{id: 36, commit: 72, name: "char-sequence", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {1, 1}, {0, -1}, {1, 1}}, generalizations: []int{129, 26, 103, 63}}
	var p128 = sequenceParser{id: 128, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p157 = charParser{id: 157, chars: []rune{34}}
	p128.items = []parser{&p157}
	var p114 = choiceParser{id: 114, commit: 72, name: "sequence-char"}
	var p143 = sequenceParser{id: 143, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{114}}
	var p107 = charParser{id: 107, not: true, chars: []rune{92, 34}}
	p143.items = []parser{&p107}
	var p25 = sequenceParser{id: 25, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{114}}
	var p48 = sequenceParser{id: 48, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p156 = charParser{id: 156, chars: []rune{92}}
	p48.items = []parser{&p156}
	var p8 = sequenceParser{id: 8, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p13 = charParser{id: 13, not: true}
	p8.items = []parser{&p13}
	p25.items = []parser{&p48, &p8}
	p114.options = []parser{&p143, &p25}
	var p115 = sequenceParser{id: 115, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p151 = charParser{id: 151, chars: []rune{34}}
	p115.items = []parser{&p151}
	p36.items = []parser{&p128, &p114, &p115}
	p129.options = []parser{&p70, &p62, &p36}
	var p165 = sequenceParser{id: 165, commit: 66, name: "group", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}}, generalizations: []int{26, 103, 63}}
	var p104 = sequenceParser{id: 104, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p178 = charParser{id: 178, chars: []rune{40}}
	p104.items = []parser{&p178}
	var p135 = sequenceParser{id: 135, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p37 = charParser{id: 37, chars: []rune{41}}
	p135.items = []parser{&p37}
	p165.items = []parser{&p104, &p186, &p26, &p186, &p135}
	var p11 = sequenceParser{id: 11, commit: 64, name: "sequence", ranges: [][]int{{1, 1}, {0, -1}}, generalizations: []int{26, 63}}
	var p167 = sequenceParser{id: 167, commit: 72, name: "item", ranges: [][]int{{1, 1}, {0, 1}, {1, 1}, {0, 1}}}
	var p103 = choiceParser{id: 103, commit: 10}
	p103.options = []parser{&p129, &p49, &p165}
	var p123 = choiceParser{id: 123, commit: 66, name: "quantity"}
	var p116 = sequenceParser{id: 116, commit: 64, name: "count-quantifier", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}}, generalizations: []int{123}}
	var p136 = sequenceParser{id: 136, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p130 = charParser{id: 130, chars: []rune{123}}
	p136.items = []parser{&p130}
	var p14 = sequenceParser{id: 14, commit: 64, name: "count", ranges: [][]int{{1, 1}}}
	var p109 = sequenceParser{id: 109, commit: 74, name: "number", ranges: [][]int{{1, -1}, {1, -1}}}
	var p158 = sequenceParser{id: 158, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p108 = charParser{id: 108, ranges: [][]rune{{48, 57}}}
	p158.items = []parser{&p108}
	p109.items = []parser{&p158}
	p14.items = []parser{&p109}
	var p30 = sequenceParser{id: 30, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p38 = charParser{id: 38, chars: []rune{125}}
	p30.items = []parser{&p38}
	p116.items = []parser{&p136, &p186, &p14, &p186, &p30}
	var p3 = sequenceParser{id: 3, commit: 64, name: "range-quantifier", ranges: [][]int{{1, 1}, {0, -1}, {0, 1}, {0, -1}, {1, 1}, {0, -1}, {0, 1}, {0, -1}, {1, 1}}, generalizations: []int{123}}
	var p32 = sequenceParser{id: 32, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p39 = charParser{id: 39, chars: []rune{123}}
	p32.items = []parser{&p39}
	var p159 = sequenceParser{id: 159, commit: 64, name: "range-from", ranges: [][]int{{1, 1}}}
	p159.items = []parser{&p109}
	var p92 = sequenceParser{id: 92, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p102 = charParser{id: 102, chars: []rune{44}}
	p92.items = []parser{&p102}
	var p9 = sequenceParser{id: 9, commit: 64, name: "range-to", ranges: [][]int{{1, 1}}}
	p9.items = []parser{&p109}
	var p119 = sequenceParser{id: 119, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p15 = charParser{id: 15, chars: []rune{125}}
	p119.items = []parser{&p15}
	p3.items = []parser{&p32, &p186, &p159, &p186, &p92, &p186, &p9, &p186, &p119}
	var p50 = sequenceParser{id: 50, commit: 72, name: "one-or-more", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{123}}
	var p166 = charParser{id: 166, chars: []rune{43}}
	p50.items = []parser{&p166}
	var p145 = sequenceParser{id: 145, commit: 72, name: "zero-or-more", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{123}}
	var p131 = charParser{id: 131, chars: []rune{42}}
	p145.items = []parser{&p131}
	var p85 = sequenceParser{id: 85, commit: 72, name: "zero-or-one", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{123}}
	var p74 = charParser{id: 74, chars: []rune{63}}
	p85.items = []parser{&p74}
	p123.options = []parser{&p116, &p3, &p50, &p145, &p85}
	p167.items = []parser{&p103, &p123}
	var p10 = sequenceParser{id: 10, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p10.items = []parser{&p186, &p167}
	p11.items = []parser{&p167, &p10}
	var p182 = sequenceParser{id: 182, commit: 64, name: "choice", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}, generalizations: []int{26}}
	var p63 = choiceParser{id: 63, commit: 66, name: "option"}
	p63.options = []parser{&p129, &p49, &p165, &p11}
	var p52 = sequenceParser{id: 52, commit: 2, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}}}
	var p110 = sequenceParser{id: 110, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p124 = charParser{id: 124, chars: []rune{124}}
	p110.items = []parser{&p124}
	p52.items = []parser{&p110, &p186, &p63}
	var p181 = sequenceParser{id: 181, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p181.items = []parser{&p186, &p52}
	p182.items = []parser{&p63, &p186, &p52, &p181}
	p26.options = []parser{&p129, &p49, &p165, &p11, &p182}
	p73.items = []parser{&p154, &p186, &p152, &p186, &p26}
	var p65 = sequenceParser{id: 65, commit: 2, ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var p96 = sequenceParser{id: 96, commit: 2, ranges: [][]int{{1, 1}, {0, -1}, {0, -1}, {1, 1}}}
	var p67 = sequenceParser{id: 67, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p60 = charParser{id: 60, chars: []rune{59}}
	p67.items = []parser{&p60}
	var p95 = sequenceParser{id: 95, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p95.items = []parser{&p186, &p67}
	p96.items = []parser{&p67, &p95, &p186, &p73}
	var p64 = sequenceParser{id: 64, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p64.items = []parser{&p186, &p96}
	p65.items = []parser{&p186, &p96, &p64}
	p66.items = []parser{&p73, &p65}
	var p89 = sequenceParser{id: 89, commit: 2, ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var p161 = sequenceParser{id: 161, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p41 = charParser{id: 41, chars: []rune{59}}
	p161.items = []parser{&p41}
	var p88 = sequenceParser{id: 88, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p88.items = []parser{&p186, &p161}
	p89.items = []parser{&p186, &p161, &p88}
	p187.items = []parser{&p87, &p186, &p66, &p89}
	p188.items = []parser{&p186, &p187, &p186}
	var b188 = sequenceBuilder{id: 188, commit: 32, name: "syntax", ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var b186 = choiceBuilder{id: 186, commit: 2}
	var b185 = choiceBuilder{id: 185, commit: 70}
	var b56 = choiceBuilder{id: 56, commit: 66}
	var b162 = sequenceBuilder{id: 162, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b138 = charBuilder{}
	b162.items = []builder{&b138}
	var b140 = sequenceBuilder{id: 140, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b139 = charBuilder{}
	b140.items = []builder{&b139}
	var b112 = sequenceBuilder{id: 112, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b141 = charBuilder{}
	b112.items = []builder{&b141}
	var b54 = sequenceBuilder{id: 54, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b133 = charBuilder{}
	b54.items = []builder{&b133}
	var b99 = sequenceBuilder{id: 99, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b98 = charBuilder{}
	b99.items = []builder{&b98}
	var b121 = sequenceBuilder{id: 121, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b21 = charBuilder{}
	b121.items = []builder{&b21}
	var b168 = sequenceBuilder{id: 168, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b55 = charBuilder{}
	b168.items = []builder{&b55}
	b56.options = []builder{&b162, &b140, &b112, &b54, &b99, &b121, &b168}
	var b101 = sequenceBuilder{id: 101, commit: 72, name: "comment", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}}
	var b175 = choiceBuilder{id: 175, commit: 74}
	var b77 = sequenceBuilder{id: 77, commit: 74, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}}
	var b76 = sequenceBuilder{id: 76, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b18 = charBuilder{}
	var b57 = charBuilder{}
	b76.items = []builder{&b18, &b57}
	var b127 = sequenceBuilder{id: 127, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b100 = charBuilder{}
	b127.items = []builder{&b100}
	b77.items = []builder{&b76, &b127}
	var b27 = sequenceBuilder{id: 27, commit: 74, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {1, 1}, {0, -1}, {1, 1}}}
	var b12 = sequenceBuilder{id: 12, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b148 = charBuilder{}
	var b169 = charBuilder{}
	b12.items = []builder{&b148, &b169}
	var b34 = choiceBuilder{id: 34, commit: 10}
	var b1 = sequenceBuilder{id: 1, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b170 = sequenceBuilder{id: 170, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b134 = charBuilder{}
	b170.items = []builder{&b134}
	var b179 = sequenceBuilder{id: 179, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b42 = charBuilder{}
	b179.items = []builder{&b42}
	b1.items = []builder{&b170, &b179}
	var b174 = sequenceBuilder{id: 174, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b155 = charBuilder{}
	b174.items = []builder{&b155}
	b34.options = []builder{&b1, &b174}
	var b117 = sequenceBuilder{id: 117, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b163 = charBuilder{}
	var b149 = charBuilder{}
	b117.items = []builder{&b163, &b149}
	b27.items = []builder{&b12, &b34, &b117}
	b175.options = []builder{&b77, &b27}
	var b176 = sequenceBuilder{id: 176, commit: 10, ranges: [][]int{{0, -1}, {0, 1}, {0, -1}, {1, 1}, {0, -1}, {0, 1}, {0, -1}, {1, 1}}}
	var b47 = choiceBuilder{id: 47, commit: 74}
	var b22 = sequenceBuilder{id: 22, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b19 = charBuilder{}
	b22.items = []builder{&b19}
	var b28 = sequenceBuilder{id: 28, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b81 = charBuilder{}
	b28.items = []builder{&b81}
	var b23 = sequenceBuilder{id: 23, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b90 = charBuilder{}
	b23.items = []builder{&b90}
	var b5 = sequenceBuilder{id: 5, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b68 = charBuilder{}
	b5.items = []builder{&b68}
	var b35 = sequenceBuilder{id: 35, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b6 = charBuilder{}
	b35.items = []builder{&b6}
	var b69 = sequenceBuilder{id: 69, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b61 = charBuilder{}
	b69.items = []builder{&b61}
	b47.options = []builder{&b22, &b28, &b23, &b5, &b35, &b69}
	var b118 = sequenceBuilder{id: 118, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b58 = charBuilder{}
	b118.items = []builder{&b58}
	b176.items = []builder{&b47, &b118, &b47, &b175}
	b101.items = []builder{&b175, &b176}
	b185.options = []builder{&b56, &b101}
	b186.options = []builder{&b185}
	var b187 = sequenceBuilder{id: 187, commit: 66, ranges: [][]int{{0, 1}, {0, -1}, {0, 1}, {0, 1}}}
	var b87 = sequenceBuilder{id: 87, commit: 2, ranges: [][]int{{1, 1}, {0, -1}}}
	var b142 = sequenceBuilder{id: 142, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b126 = charBuilder{}
	b142.items = []builder{&b126}
	var b86 = sequenceBuilder{id: 86, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b86.items = []builder{&b186, &b142}
	b87.items = []builder{&b142, &b86}
	var b66 = sequenceBuilder{id: 66, commit: 66, ranges: [][]int{{1, 1}, {0, 1}}}
	var b73 = sequenceBuilder{id: 73, commit: 64, name: "definition", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}}}
	var b154 = sequenceBuilder{id: 154, commit: 74, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}}
	var b49 = sequenceBuilder{id: 49, commit: 72, name: "symbol", ranges: [][]int{{1, -1}, {1, -1}}}
	var b2 = sequenceBuilder{id: 2, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b144 = charBuilder{}
	b2.items = []builder{&b144}
	b49.items = []builder{&b2}
	var b46 = sequenceBuilder{id: 46, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b172 = sequenceBuilder{id: 172, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b125 = charBuilder{}
	b172.items = []builder{&b125}
	var b4 = choiceBuilder{id: 4, commit: 66}
	var b44 = sequenceBuilder{id: 44, commit: 72, name: "alias", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b33 = charBuilder{}
	var b75 = charBuilder{}
	var b16 = charBuilder{}
	var b51 = charBuilder{}
	var b40 = charBuilder{}
	b44.items = []builder{&b33, &b75, &b16, &b51, &b40}
	var b45 = sequenceBuilder{id: 45, commit: 72, name: "ws", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b31 = charBuilder{}
	var b120 = charBuilder{}
	b45.items = []builder{&b31, &b120}
	var b17 = sequenceBuilder{id: 17, commit: 72, name: "nows", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b97 = charBuilder{}
	var b183 = charBuilder{}
	var b93 = charBuilder{}
	var b146 = charBuilder{}
	b17.items = []builder{&b97, &b183, &b93, &b146}
	var b160 = sequenceBuilder{id: 160, commit: 72, name: "failpass", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b79 = charBuilder{}
	var b80 = charBuilder{}
	var b111 = charBuilder{}
	var b94 = charBuilder{}
	var b132 = charBuilder{}
	var b53 = charBuilder{}
	var b72 = charBuilder{}
	var b105 = charBuilder{}
	b160.items = []builder{&b79, &b80, &b111, &b94, &b132, &b53, &b72, &b105}
	var b20 = sequenceBuilder{id: 20, commit: 72, name: "root", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b147 = charBuilder{}
	var b184 = charBuilder{}
	var b153 = charBuilder{}
	var b137 = charBuilder{}
	b20.items = []builder{&b147, &b184, &b153, &b137}
	b4.options = []builder{&b44, &b45, &b17, &b160, &b20}
	b46.items = []builder{&b172, &b4}
	b154.items = []builder{&b49, &b46}
	var b152 = sequenceBuilder{id: 152, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b173 = charBuilder{}
	b152.items = []builder{&b173}
	var b26 = choiceBuilder{id: 26, commit: 66}
	var b129 = choiceBuilder{id: 129, commit: 66}
	var b70 = sequenceBuilder{id: 70, commit: 72, name: "any-char", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b91 = charBuilder{}
	b70.items = []builder{&b91}
	var b62 = sequenceBuilder{id: 62, commit: 72, name: "char-class", ranges: [][]int{{1, 1}, {0, 1}, {0, -1}, {1, 1}, {1, 1}, {0, 1}, {0, -1}, {1, 1}}}
	var b71 = sequenceBuilder{id: 71, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b82 = charBuilder{}
	b71.items = []builder{&b82}
	var b180 = sequenceBuilder{id: 180, commit: 72, name: "class-not", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b29 = charBuilder{}
	b180.items = []builder{&b29}
	var b83 = choiceBuilder{id: 83, commit: 10}
	var b171 = choiceBuilder{id: 171, commit: 72, name: "class-char"}
	var b150 = sequenceBuilder{id: 150, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b113 = charBuilder{}
	b150.items = []builder{&b113}
	var b164 = sequenceBuilder{id: 164, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b59 = sequenceBuilder{id: 59, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b122 = charBuilder{}
	b59.items = []builder{&b122}
	var b106 = sequenceBuilder{id: 106, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b7 = charBuilder{}
	b106.items = []builder{&b7}
	b164.items = []builder{&b59, &b106}
	b171.options = []builder{&b150, &b164}
	var b177 = sequenceBuilder{id: 177, commit: 72, name: "char-range", ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b78 = sequenceBuilder{id: 78, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b24 = charBuilder{}
	b78.items = []builder{&b24}
	b177.items = []builder{&b171, &b78, &b171}
	b83.options = []builder{&b171, &b177}
	var b84 = sequenceBuilder{id: 84, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b43 = charBuilder{}
	b84.items = []builder{&b43}
	b62.items = []builder{&b71, &b180, &b83, &b84}
	var b36 = sequenceBuilder{id: 36, commit: 72, name: "char-sequence", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {1, 1}, {0, -1}, {1, 1}}}
	var b128 = sequenceBuilder{id: 128, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b157 = charBuilder{}
	b128.items = []builder{&b157}
	var b114 = choiceBuilder{id: 114, commit: 72, name: "sequence-char"}
	var b143 = sequenceBuilder{id: 143, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b107 = charBuilder{}
	b143.items = []builder{&b107}
	var b25 = sequenceBuilder{id: 25, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b48 = sequenceBuilder{id: 48, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b156 = charBuilder{}
	b48.items = []builder{&b156}
	var b8 = sequenceBuilder{id: 8, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b13 = charBuilder{}
	b8.items = []builder{&b13}
	b25.items = []builder{&b48, &b8}
	b114.options = []builder{&b143, &b25}
	var b115 = sequenceBuilder{id: 115, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b151 = charBuilder{}
	b115.items = []builder{&b151}
	b36.items = []builder{&b128, &b114, &b115}
	b129.options = []builder{&b70, &b62, &b36}
	var b165 = sequenceBuilder{id: 165, commit: 66, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}}}
	var b104 = sequenceBuilder{id: 104, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b178 = charBuilder{}
	b104.items = []builder{&b178}
	var b135 = sequenceBuilder{id: 135, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b37 = charBuilder{}
	b135.items = []builder{&b37}
	b165.items = []builder{&b104, &b186, &b26, &b186, &b135}
	var b11 = sequenceBuilder{id: 11, commit: 64, name: "sequence", ranges: [][]int{{1, 1}, {0, -1}}}
	var b167 = sequenceBuilder{id: 167, commit: 72, name: "item", ranges: [][]int{{1, 1}, {0, 1}, {1, 1}, {0, 1}}}
	var b103 = choiceBuilder{id: 103, commit: 10}
	b103.options = []builder{&b129, &b49, &b165}
	var b123 = choiceBuilder{id: 123, commit: 66}
	var b116 = sequenceBuilder{id: 116, commit: 64, name: "count-quantifier", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}}}
	var b136 = sequenceBuilder{id: 136, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b130 = charBuilder{}
	b136.items = []builder{&b130}
	var b14 = sequenceBuilder{id: 14, commit: 64, name: "count", ranges: [][]int{{1, 1}}}
	var b109 = sequenceBuilder{id: 109, commit: 74, ranges: [][]int{{1, -1}, {1, -1}}}
	var b158 = sequenceBuilder{id: 158, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b108 = charBuilder{}
	b158.items = []builder{&b108}
	b109.items = []builder{&b158}
	b14.items = []builder{&b109}
	var b30 = sequenceBuilder{id: 30, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b38 = charBuilder{}
	b30.items = []builder{&b38}
	b116.items = []builder{&b136, &b186, &b14, &b186, &b30}
	var b3 = sequenceBuilder{id: 3, commit: 64, name: "range-quantifier", ranges: [][]int{{1, 1}, {0, -1}, {0, 1}, {0, -1}, {1, 1}, {0, -1}, {0, 1}, {0, -1}, {1, 1}}}
	var b32 = sequenceBuilder{id: 32, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b39 = charBuilder{}
	b32.items = []builder{&b39}
	var b159 = sequenceBuilder{id: 159, commit: 64, name: "range-from", ranges: [][]int{{1, 1}}}
	b159.items = []builder{&b109}
	var b92 = sequenceBuilder{id: 92, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b102 = charBuilder{}
	b92.items = []builder{&b102}
	var b9 = sequenceBuilder{id: 9, commit: 64, name: "range-to", ranges: [][]int{{1, 1}}}
	b9.items = []builder{&b109}
	var b119 = sequenceBuilder{id: 119, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b15 = charBuilder{}
	b119.items = []builder{&b15}
	b3.items = []builder{&b32, &b186, &b159, &b186, &b92, &b186, &b9, &b186, &b119}
	var b50 = sequenceBuilder{id: 50, commit: 72, name: "one-or-more", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b166 = charBuilder{}
	b50.items = []builder{&b166}
	var b145 = sequenceBuilder{id: 145, commit: 72, name: "zero-or-more", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b131 = charBuilder{}
	b145.items = []builder{&b131}
	var b85 = sequenceBuilder{id: 85, commit: 72, name: "zero-or-one", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b74 = charBuilder{}
	b85.items = []builder{&b74}
	b123.options = []builder{&b116, &b3, &b50, &b145, &b85}
	b167.items = []builder{&b103, &b123}
	var b10 = sequenceBuilder{id: 10, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b10.items = []builder{&b186, &b167}
	b11.items = []builder{&b167, &b10}
	var b182 = sequenceBuilder{id: 182, commit: 64, name: "choice", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}}
	var b63 = choiceBuilder{id: 63, commit: 66}
	b63.options = []builder{&b129, &b49, &b165, &b11}
	var b52 = sequenceBuilder{id: 52, commit: 2, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}}}
	var b110 = sequenceBuilder{id: 110, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b124 = charBuilder{}
	b110.items = []builder{&b124}
	b52.items = []builder{&b110, &b186, &b63}
	var b181 = sequenceBuilder{id: 181, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b181.items = []builder{&b186, &b52}
	b182.items = []builder{&b63, &b186, &b52, &b181}
	b26.options = []builder{&b129, &b49, &b165, &b11, &b182}
	b73.items = []builder{&b154, &b186, &b152, &b186, &b26}
	var b65 = sequenceBuilder{id: 65, commit: 2, ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var b96 = sequenceBuilder{id: 96, commit: 2, ranges: [][]int{{1, 1}, {0, -1}, {0, -1}, {1, 1}}}
	var b67 = sequenceBuilder{id: 67, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b60 = charBuilder{}
	b67.items = []builder{&b60}
	var b95 = sequenceBuilder{id: 95, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b95.items = []builder{&b186, &b67}
	b96.items = []builder{&b67, &b95, &b186, &b73}
	var b64 = sequenceBuilder{id: 64, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b64.items = []builder{&b186, &b96}
	b65.items = []builder{&b186, &b96, &b64}
	b66.items = []builder{&b73, &b65}
	var b89 = sequenceBuilder{id: 89, commit: 2, ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var b161 = sequenceBuilder{id: 161, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b41 = charBuilder{}
	b161.items = []builder{&b41}
	var b88 = sequenceBuilder{id: 88, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b88.items = []builder{&b186, &b161}
	b89.items = []builder{&b186, &b161, &b88}
	b187.items = []builder{&b87, &b186, &b66, &b89}
	b188.items = []builder{&b186, &b187, &b186}

	return parseInput(r, &p188, &b188)
}
