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
	var p21 = choiceParser{id: 21, commit: 66, name: "wschar", generalizations: []int{185, 186}}
	var p135 = sequenceParser{id: 135, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{21, 185, 186}}
	var p36 = charParser{id: 36, chars: []rune{32}}
	p135.items = []parser{&p36}
	var p170 = sequenceParser{id: 170, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{21, 185, 186}}
	var p98 = charParser{id: 98, chars: []rune{9}}
	p170.items = []parser{&p98}
	var p117 = sequenceParser{id: 117, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{21, 185, 186}}
	var p111 = charParser{id: 111, chars: []rune{10}}
	p117.items = []parser{&p111}
	var p149 = sequenceParser{id: 149, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{21, 185, 186}}
	var p118 = charParser{id: 118, chars: []rune{8}}
	p149.items = []parser{&p118}
	var p71 = sequenceParser{id: 71, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{21, 185, 186}}
	var p130 = charParser{id: 130, chars: []rune{12}}
	p71.items = []parser{&p130}
	var p64 = sequenceParser{id: 64, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{21, 185, 186}}
	var p99 = charParser{id: 99, chars: []rune{13}}
	p64.items = []parser{&p99}
	var p112 = sequenceParser{id: 112, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{21, 185, 186}}
	var p37 = charParser{id: 37, chars: []rune{11}}
	p112.items = []parser{&p37}
	p21.options = []parser{&p135, &p170, &p117, &p149, &p71, &p64, &p112}
	var p35 = sequenceParser{id: 35, commit: 72, name: "comment", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}, generalizations: []int{185, 186}}
	var p113 = choiceParser{id: 113, commit: 74, name: "comment-segment"}
	var p136 = sequenceParser{id: 136, commit: 74, name: "line-comment", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}, generalizations: []int{113}}
	var p22 = sequenceParser{id: 22, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var p159 = charParser{id: 159, chars: []rune{47}}
	var p144 = charParser{id: 144, chars: []rune{47}}
	p22.items = []parser{&p159, &p144}
	var p40 = sequenceParser{id: 40, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p39 = charParser{id: 39, not: true, chars: []rune{10}}
	p40.items = []parser{&p39}
	p136.items = []parser{&p22, &p40}
	var p10 = sequenceParser{id: 10, commit: 74, name: "block-comment", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {1, 1}, {0, -1}, {1, 1}}, generalizations: []int{113}}
	var p119 = sequenceParser{id: 119, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var p13 = charParser{id: 13, chars: []rune{47}}
	var p91 = charParser{id: 91, chars: []rune{42}}
	p119.items = []parser{&p13, &p91}
	var p163 = choiceParser{id: 163, commit: 10}
	var p53 = sequenceParser{id: 53, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{163}}
	var p143 = sequenceParser{id: 143, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p65 = charParser{id: 65, chars: []rune{42}}
	p143.items = []parser{&p65}
	var p8 = sequenceParser{id: 8, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p92 = charParser{id: 92, not: true, chars: []rune{47}}
	p8.items = []parser{&p92}
	p53.items = []parser{&p143, &p8}
	var p9 = sequenceParser{id: 9, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{163}}
	var p38 = charParser{id: 38, not: true, chars: []rune{42}}
	p9.items = []parser{&p38}
	p163.options = []parser{&p53, &p9}
	var p2 = sequenceParser{id: 2, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var p150 = charParser{id: 150, chars: []rune{42}}
	var p181 = charParser{id: 181, chars: []rune{47}}
	p2.items = []parser{&p150, &p181}
	p10.items = []parser{&p119, &p163, &p2}
	p113.options = []parser{&p136, &p10}
	var p79 = sequenceParser{id: 79, commit: 10, ranges: [][]int{{0, -1}, {0, 1}, {0, -1}, {1, 1}, {0, -1}, {0, 1}, {0, -1}, {1, 1}}}
	var p125 = choiceParser{id: 125, commit: 74, name: "ws-no-nl"}
	var p123 = sequenceParser{id: 123, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{125}}
	var p23 = charParser{id: 23, chars: []rune{32}}
	p123.items = []parser{&p23}
	var p48 = sequenceParser{id: 48, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{125}}
	var p47 = charParser{id: 47, chars: []rune{9}}
	p48.items = []parser{&p47}
	var p41 = sequenceParser{id: 41, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{125}}
	var p93 = charParser{id: 93, chars: []rune{8}}
	p41.items = []parser{&p93}
	var p124 = sequenceParser{id: 124, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{125}}
	var p114 = charParser{id: 114, chars: []rune{12}}
	p124.items = []parser{&p114}
	var p151 = sequenceParser{id: 151, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{125}}
	var p54 = charParser{id: 54, chars: []rune{13}}
	p151.items = []parser{&p54}
	var p104 = sequenceParser{id: 104, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{125}}
	var p84 = charParser{id: 84, chars: []rune{11}}
	p104.items = []parser{&p84}
	p125.options = []parser{&p123, &p48, &p41, &p124, &p151, &p104}
	var p34 = sequenceParser{id: 34, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p61 = charParser{id: 61, chars: []rune{10}}
	p34.items = []parser{&p61}
	p79.items = []parser{&p125, &p34, &p125, &p113}
	p35.items = []parser{&p113, &p79}
	p185.options = []parser{&p21, &p35}
	p186.options = []parser{&p185}
	var p187 = sequenceParser{id: 187, commit: 66, name: "syntax:wsroot", ranges: [][]int{{0, 1}, {0, -1}, {0, 1}, {0, 1}}}
	var p176 = sequenceParser{id: 176, commit: 2, ranges: [][]int{{1, 1}, {0, -1}}}
	var p108 = sequenceParser{id: 108, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p60 = charParser{id: 60, chars: []rune{59}}
	p108.items = []parser{&p60}
	var p175 = sequenceParser{id: 175, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p175.items = []parser{&p186, &p108}
	p176.items = []parser{&p108, &p175}
	var p169 = sequenceParser{id: 169, commit: 66, name: "definitions", ranges: [][]int{{1, 1}, {0, 1}}}
	var p174 = sequenceParser{id: 174, commit: 64, name: "definition", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}}}
	var p83 = sequenceParser{id: 83, commit: 74, name: "definition-name", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}}
	var p156 = sequenceParser{id: 156, commit: 72, name: "symbol", ranges: [][]int{{1, -1}, {1, -1}}, generalizations: []int{6, 57, 102}}
	var p100 = sequenceParser{id: 100, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p183 = charParser{id: 183, not: true, chars: []rune{92, 32, 10, 9, 8, 12, 13, 11, 47, 46, 91, 93, 34, 123, 125, 94, 43, 42, 63, 124, 40, 41, 58, 61, 59}}
	p100.items = []parser{&p183}
	p156.items = []parser{&p100}
	var p20 = sequenceParser{id: 20, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var p162 = sequenceParser{id: 162, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p154 = charParser{id: 154, chars: []rune{58}}
	p162.items = []parser{&p154}
	var p107 = choiceParser{id: 107, commit: 66, name: "flag"}
	var p158 = sequenceParser{id: 158, commit: 72, name: "alias", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{107}}
	var p58 = charParser{id: 58, chars: []rune{97}}
	var p68 = charParser{id: 68, chars: []rune{108}}
	var p45 = charParser{id: 45, chars: []rune{105}}
	var p160 = charParser{id: 160, chars: []rune{97}}
	var p51 = charParser{id: 51, chars: []rune{115}}
	p158.items = []parser{&p58, &p68, &p45, &p160, &p51}
	var p11 = sequenceParser{id: 11, commit: 72, name: "ws", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{107}}
	var p116 = charParser{id: 116, chars: []rune{119}}
	var p161 = charParser{id: 161, chars: []rune{115}}
	p11.items = []parser{&p116, &p161}
	var p141 = sequenceParser{id: 141, commit: 72, name: "nows", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{107}}
	var p106 = charParser{id: 106, chars: []rune{110}}
	var p77 = charParser{id: 77, chars: []rune{111}}
	var p173 = charParser{id: 173, chars: []rune{119}}
	var p129 = charParser{id: 129, chars: []rune{115}}
	p141.items = []parser{&p106, &p77, &p173, &p129}
	var p142 = sequenceParser{id: 142, commit: 72, name: "failpass", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{107}}
	var p7 = charParser{id: 7, chars: []rune{102}}
	var p32 = charParser{id: 32, chars: []rune{97}}
	var p52 = charParser{id: 52, chars: []rune{105}}
	var p134 = charParser{id: 134, chars: []rune{108}}
	var p78 = charParser{id: 78, chars: []rune{112}}
	var p46 = charParser{id: 46, chars: []rune{97}}
	var p82 = charParser{id: 82, chars: []rune{115}}
	var p110 = charParser{id: 110, chars: []rune{115}}
	p142.items = []parser{&p7, &p32, &p52, &p134, &p78, &p46, &p82, &p110}
	var p103 = sequenceParser{id: 103, commit: 72, name: "root", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{107}}
	var p1 = charParser{id: 1, chars: []rune{114}}
	var p18 = charParser{id: 18, chars: []rune{111}}
	var p33 = charParser{id: 33, chars: []rune{111}}
	var p50 = charParser{id: 50, chars: []rune{116}}
	p103.items = []parser{&p1, &p18, &p33, &p50}
	p107.options = []parser{&p158, &p11, &p141, &p142, &p103}
	p20.items = []parser{&p162, &p107}
	p83.items = []parser{&p156, &p20}
	var p59 = sequenceParser{id: 59, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p12 = charParser{id: 12, chars: []rune{61}}
	p59.items = []parser{&p12}
	var p6 = choiceParser{id: 6, commit: 66, name: "expression"}
	var p42 = choiceParser{id: 42, commit: 66, name: "terminal", generalizations: []int{6, 57, 102}}
	var p73 = sequenceParser{id: 73, commit: 72, name: "any-char", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{42, 6, 57, 102}}
	var p145 = charParser{id: 145, chars: []rune{46}}
	p73.items = []parser{&p145}
	var p16 = sequenceParser{id: 16, commit: 72, name: "char-class", ranges: [][]int{{1, 1}, {0, 1}, {0, -1}, {1, 1}, {1, 1}, {0, 1}, {0, -1}, {1, 1}}, generalizations: []int{42, 6, 57, 102}}
	var p165 = sequenceParser{id: 165, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p74 = charParser{id: 74, chars: []rune{91}}
	p165.items = []parser{&p74}
	var p137 = sequenceParser{id: 137, commit: 72, name: "class-not", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p164 = charParser{id: 164, chars: []rune{94}}
	p137.items = []parser{&p164}
	var p179 = choiceParser{id: 179, commit: 10}
	var p24 = choiceParser{id: 24, commit: 72, name: "class-char", generalizations: []int{179}}
	var p146 = sequenceParser{id: 146, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{24, 179}}
	var p109 = charParser{id: 109, not: true, chars: []rune{92, 91, 93, 94, 45}}
	p146.items = []parser{&p109}
	var p14 = sequenceParser{id: 14, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{24, 179}}
	var p66 = sequenceParser{id: 66, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p94 = charParser{id: 94, chars: []rune{92}}
	p66.items = []parser{&p94}
	var p155 = sequenceParser{id: 155, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p85 = charParser{id: 85, not: true}
	p155.items = []parser{&p85}
	p14.items = []parser{&p66, &p155}
	p24.options = []parser{&p146, &p14}
	var p86 = sequenceParser{id: 86, commit: 72, name: "char-range", ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{179}}
	var p131 = sequenceParser{id: 131, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p95 = charParser{id: 95, chars: []rune{45}}
	p131.items = []parser{&p95}
	p86.items = []parser{&p24, &p131, &p24}
	p179.options = []parser{&p24, &p86}
	var p15 = sequenceParser{id: 15, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p27 = charParser{id: 27, chars: []rune{93}}
	p15.items = []parser{&p27}
	p16.items = []parser{&p165, &p137, &p179, &p15}
	var p180 = sequenceParser{id: 180, commit: 72, name: "char-sequence", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {1, 1}, {0, -1}, {1, 1}}, generalizations: []int{42, 6, 57, 102}}
	var p49 = sequenceParser{id: 49, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p26 = charParser{id: 26, chars: []rune{34}}
	p49.items = []parser{&p26}
	var p25 = choiceParser{id: 25, commit: 72, name: "sequence-char"}
	var p96 = sequenceParser{id: 96, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{25}}
	var p182 = charParser{id: 182, not: true, chars: []rune{92, 34}}
	p96.items = []parser{&p182}
	var p147 = sequenceParser{id: 147, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{25}}
	var p87 = sequenceParser{id: 87, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p62 = charParser{id: 62, chars: []rune{92}}
	p87.items = []parser{&p62}
	var p90 = sequenceParser{id: 90, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p75 = charParser{id: 75, not: true}
	p90.items = []parser{&p75}
	p147.items = []parser{&p87, &p90}
	p25.options = []parser{&p96, &p147}
	var p88 = sequenceParser{id: 88, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p19 = charParser{id: 19, chars: []rune{34}}
	p88.items = []parser{&p19}
	p180.items = []parser{&p49, &p25, &p88}
	p42.options = []parser{&p73, &p16, &p180}
	var p43 = sequenceParser{id: 43, commit: 66, name: "group", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}}, generalizations: []int{6, 57, 102}}
	var p70 = sequenceParser{id: 70, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p28 = charParser{id: 28, chars: []rune{40}}
	p70.items = []parser{&p28}
	var p138 = sequenceParser{id: 138, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p171 = charParser{id: 171, chars: []rune{41}}
	p138.items = []parser{&p171}
	p43.items = []parser{&p70, &p186, &p6, &p186, &p138}
	var p30 = sequenceParser{id: 30, commit: 64, name: "sequence", ranges: [][]int{{1, 1}, {0, -1}}, generalizations: []int{6, 102}}
	var p132 = sequenceParser{id: 132, commit: 72, name: "item", ranges: [][]int{{1, 1}, {0, 1}, {1, 1}, {0, 1}}}
	var p57 = choiceParser{id: 57, commit: 10}
	p57.options = []parser{&p42, &p156, &p43}
	var p126 = choiceParser{id: 126, commit: 66, name: "quantity"}
	var p56 = sequenceParser{id: 56, commit: 64, name: "count-quantifier", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}}, generalizations: []int{126}}
	var p17 = sequenceParser{id: 17, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p120 = charParser{id: 120, chars: []rune{123}}
	p17.items = []parser{&p120}
	var p4 = sequenceParser{id: 4, commit: 64, name: "count", ranges: [][]int{{1, 1}}}
	var p3 = sequenceParser{id: 3, commit: 74, name: "number", ranges: [][]int{{1, -1}, {1, -1}}}
	var p76 = sequenceParser{id: 76, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p157 = charParser{id: 157, ranges: [][]rune{{48, 57}}}
	p76.items = []parser{&p157}
	p3.items = []parser{&p76}
	p4.items = []parser{&p3}
	var p67 = sequenceParser{id: 67, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p55 = charParser{id: 55, chars: []rune{125}}
	p67.items = []parser{&p55}
	p56.items = []parser{&p17, &p186, &p4, &p186, &p67}
	var p44 = sequenceParser{id: 44, commit: 64, name: "range-quantifier", ranges: [][]int{{1, 1}, {0, -1}, {0, 1}, {0, -1}, {1, 1}, {0, -1}, {0, 1}, {0, -1}, {1, 1}}, generalizations: []int{126}}
	var p139 = sequenceParser{id: 139, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p166 = charParser{id: 166, chars: []rune{123}}
	p139.items = []parser{&p166}
	var p148 = sequenceParser{id: 148, commit: 64, name: "range-from", ranges: [][]int{{1, 1}}}
	p148.items = []parser{&p3}
	var p101 = sequenceParser{id: 101, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p80 = charParser{id: 80, chars: []rune{44}}
	p101.items = []parser{&p80}
	var p105 = sequenceParser{id: 105, commit: 64, name: "range-to", ranges: [][]int{{1, 1}}}
	p105.items = []parser{&p3}
	var p140 = sequenceParser{id: 140, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p5 = charParser{id: 5, chars: []rune{125}}
	p140.items = []parser{&p5}
	p44.items = []parser{&p139, &p186, &p148, &p186, &p101, &p186, &p105, &p186, &p140}
	var p89 = sequenceParser{id: 89, commit: 72, name: "one-or-more", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{126}}
	var p115 = charParser{id: 115, chars: []rune{43}}
	p89.items = []parser{&p115}
	var p152 = sequenceParser{id: 152, commit: 72, name: "zero-or-more", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{126}}
	var p81 = charParser{id: 81, chars: []rune{42}}
	p152.items = []parser{&p81}
	var p153 = sequenceParser{id: 153, commit: 72, name: "zero-or-one", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{126}}
	var p63 = charParser{id: 63, chars: []rune{63}}
	p153.items = []parser{&p63}
	p126.options = []parser{&p56, &p44, &p89, &p152, &p153}
	p132.items = []parser{&p57, &p126}
	var p29 = sequenceParser{id: 29, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p29.items = []parser{&p186, &p132}
	p30.items = []parser{&p132, &p29}
	var p128 = sequenceParser{id: 128, commit: 64, name: "choice", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}, generalizations: []int{6}}
	var p102 = choiceParser{id: 102, commit: 66, name: "option"}
	p102.options = []parser{&p42, &p156, &p43, &p30}
	var p184 = sequenceParser{id: 184, commit: 2, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}}}
	var p31 = sequenceParser{id: 31, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p172 = charParser{id: 172, chars: []rune{124}}
	p31.items = []parser{&p172}
	p184.items = []parser{&p31, &p186, &p102}
	var p127 = sequenceParser{id: 127, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p127.items = []parser{&p186, &p184}
	p128.items = []parser{&p102, &p186, &p184, &p127}
	p6.options = []parser{&p42, &p156, &p43, &p30, &p128}
	p174.items = []parser{&p83, &p186, &p59, &p186, &p6}
	var p168 = sequenceParser{id: 168, commit: 2, ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var p122 = sequenceParser{id: 122, commit: 2, ranges: [][]int{{1, 1}, {0, -1}, {0, -1}, {1, 1}}}
	var p133 = sequenceParser{id: 133, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p97 = charParser{id: 97, chars: []rune{59}}
	p133.items = []parser{&p97}
	var p121 = sequenceParser{id: 121, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p121.items = []parser{&p186, &p133}
	p122.items = []parser{&p133, &p121, &p186, &p174}
	var p167 = sequenceParser{id: 167, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p167.items = []parser{&p186, &p122}
	p168.items = []parser{&p186, &p122, &p167}
	p169.items = []parser{&p174, &p168}
	var p178 = sequenceParser{id: 178, commit: 2, ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var p69 = sequenceParser{id: 69, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p72 = charParser{id: 72, chars: []rune{59}}
	p69.items = []parser{&p72}
	var p177 = sequenceParser{id: 177, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p177.items = []parser{&p186, &p69}
	p178.items = []parser{&p186, &p69, &p177}
	p187.items = []parser{&p176, &p186, &p169, &p178}
	p188.items = []parser{&p186, &p187, &p186}
	var b188 = sequenceBuilder{id: 188, commit: 32, name: "syntax", ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var b186 = choiceBuilder{id: 186, commit: 2}
	var b185 = choiceBuilder{id: 185, commit: 70}
	var b21 = choiceBuilder{id: 21, commit: 66}
	var b135 = sequenceBuilder{id: 135, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b36 = charBuilder{}
	b135.items = []builder{&b36}
	var b170 = sequenceBuilder{id: 170, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b98 = charBuilder{}
	b170.items = []builder{&b98}
	var b117 = sequenceBuilder{id: 117, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b111 = charBuilder{}
	b117.items = []builder{&b111}
	var b149 = sequenceBuilder{id: 149, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b118 = charBuilder{}
	b149.items = []builder{&b118}
	var b71 = sequenceBuilder{id: 71, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b130 = charBuilder{}
	b71.items = []builder{&b130}
	var b64 = sequenceBuilder{id: 64, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b99 = charBuilder{}
	b64.items = []builder{&b99}
	var b112 = sequenceBuilder{id: 112, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b37 = charBuilder{}
	b112.items = []builder{&b37}
	b21.options = []builder{&b135, &b170, &b117, &b149, &b71, &b64, &b112}
	var b35 = sequenceBuilder{id: 35, commit: 72, name: "comment", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}}
	var b113 = choiceBuilder{id: 113, commit: 74}
	var b136 = sequenceBuilder{id: 136, commit: 74, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}}
	var b22 = sequenceBuilder{id: 22, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b159 = charBuilder{}
	var b144 = charBuilder{}
	b22.items = []builder{&b159, &b144}
	var b40 = sequenceBuilder{id: 40, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b39 = charBuilder{}
	b40.items = []builder{&b39}
	b136.items = []builder{&b22, &b40}
	var b10 = sequenceBuilder{id: 10, commit: 74, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {1, 1}, {0, -1}, {1, 1}}}
	var b119 = sequenceBuilder{id: 119, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b13 = charBuilder{}
	var b91 = charBuilder{}
	b119.items = []builder{&b13, &b91}
	var b163 = choiceBuilder{id: 163, commit: 10}
	var b53 = sequenceBuilder{id: 53, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b143 = sequenceBuilder{id: 143, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b65 = charBuilder{}
	b143.items = []builder{&b65}
	var b8 = sequenceBuilder{id: 8, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b92 = charBuilder{}
	b8.items = []builder{&b92}
	b53.items = []builder{&b143, &b8}
	var b9 = sequenceBuilder{id: 9, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b38 = charBuilder{}
	b9.items = []builder{&b38}
	b163.options = []builder{&b53, &b9}
	var b2 = sequenceBuilder{id: 2, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b150 = charBuilder{}
	var b181 = charBuilder{}
	b2.items = []builder{&b150, &b181}
	b10.items = []builder{&b119, &b163, &b2}
	b113.options = []builder{&b136, &b10}
	var b79 = sequenceBuilder{id: 79, commit: 10, ranges: [][]int{{0, -1}, {0, 1}, {0, -1}, {1, 1}, {0, -1}, {0, 1}, {0, -1}, {1, 1}}}
	var b125 = choiceBuilder{id: 125, commit: 74}
	var b123 = sequenceBuilder{id: 123, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b23 = charBuilder{}
	b123.items = []builder{&b23}
	var b48 = sequenceBuilder{id: 48, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b47 = charBuilder{}
	b48.items = []builder{&b47}
	var b41 = sequenceBuilder{id: 41, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b93 = charBuilder{}
	b41.items = []builder{&b93}
	var b124 = sequenceBuilder{id: 124, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b114 = charBuilder{}
	b124.items = []builder{&b114}
	var b151 = sequenceBuilder{id: 151, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b54 = charBuilder{}
	b151.items = []builder{&b54}
	var b104 = sequenceBuilder{id: 104, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b84 = charBuilder{}
	b104.items = []builder{&b84}
	b125.options = []builder{&b123, &b48, &b41, &b124, &b151, &b104}
	var b34 = sequenceBuilder{id: 34, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b61 = charBuilder{}
	b34.items = []builder{&b61}
	b79.items = []builder{&b125, &b34, &b125, &b113}
	b35.items = []builder{&b113, &b79}
	b185.options = []builder{&b21, &b35}
	b186.options = []builder{&b185}
	var b187 = sequenceBuilder{id: 187, commit: 66, ranges: [][]int{{0, 1}, {0, -1}, {0, 1}, {0, 1}}}
	var b176 = sequenceBuilder{id: 176, commit: 2, ranges: [][]int{{1, 1}, {0, -1}}}
	var b108 = sequenceBuilder{id: 108, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b60 = charBuilder{}
	b108.items = []builder{&b60}
	var b175 = sequenceBuilder{id: 175, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b175.items = []builder{&b186, &b108}
	b176.items = []builder{&b108, &b175}
	var b169 = sequenceBuilder{id: 169, commit: 66, ranges: [][]int{{1, 1}, {0, 1}}}
	var b174 = sequenceBuilder{id: 174, commit: 64, name: "definition", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}}}
	var b83 = sequenceBuilder{id: 83, commit: 74, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}}
	var b156 = sequenceBuilder{id: 156, commit: 72, name: "symbol", ranges: [][]int{{1, -1}, {1, -1}}}
	var b100 = sequenceBuilder{id: 100, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b183 = charBuilder{}
	b100.items = []builder{&b183}
	b156.items = []builder{&b100}
	var b20 = sequenceBuilder{id: 20, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b162 = sequenceBuilder{id: 162, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b154 = charBuilder{}
	b162.items = []builder{&b154}
	var b107 = choiceBuilder{id: 107, commit: 66}
	var b158 = sequenceBuilder{id: 158, commit: 72, name: "alias", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b58 = charBuilder{}
	var b68 = charBuilder{}
	var b45 = charBuilder{}
	var b160 = charBuilder{}
	var b51 = charBuilder{}
	b158.items = []builder{&b58, &b68, &b45, &b160, &b51}
	var b11 = sequenceBuilder{id: 11, commit: 72, name: "ws", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b116 = charBuilder{}
	var b161 = charBuilder{}
	b11.items = []builder{&b116, &b161}
	var b141 = sequenceBuilder{id: 141, commit: 72, name: "nows", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b106 = charBuilder{}
	var b77 = charBuilder{}
	var b173 = charBuilder{}
	var b129 = charBuilder{}
	b141.items = []builder{&b106, &b77, &b173, &b129}
	var b142 = sequenceBuilder{id: 142, commit: 72, name: "failpass", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b7 = charBuilder{}
	var b32 = charBuilder{}
	var b52 = charBuilder{}
	var b134 = charBuilder{}
	var b78 = charBuilder{}
	var b46 = charBuilder{}
	var b82 = charBuilder{}
	var b110 = charBuilder{}
	b142.items = []builder{&b7, &b32, &b52, &b134, &b78, &b46, &b82, &b110}
	var b103 = sequenceBuilder{id: 103, commit: 72, name: "root", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b1 = charBuilder{}
	var b18 = charBuilder{}
	var b33 = charBuilder{}
	var b50 = charBuilder{}
	b103.items = []builder{&b1, &b18, &b33, &b50}
	b107.options = []builder{&b158, &b11, &b141, &b142, &b103}
	b20.items = []builder{&b162, &b107}
	b83.items = []builder{&b156, &b20}
	var b59 = sequenceBuilder{id: 59, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b12 = charBuilder{}
	b59.items = []builder{&b12}
	var b6 = choiceBuilder{id: 6, commit: 66}
	var b42 = choiceBuilder{id: 42, commit: 66}
	var b73 = sequenceBuilder{id: 73, commit: 72, name: "any-char", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b145 = charBuilder{}
	b73.items = []builder{&b145}
	var b16 = sequenceBuilder{id: 16, commit: 72, name: "char-class", ranges: [][]int{{1, 1}, {0, 1}, {0, -1}, {1, 1}, {1, 1}, {0, 1}, {0, -1}, {1, 1}}}
	var b165 = sequenceBuilder{id: 165, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b74 = charBuilder{}
	b165.items = []builder{&b74}
	var b137 = sequenceBuilder{id: 137, commit: 72, name: "class-not", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b164 = charBuilder{}
	b137.items = []builder{&b164}
	var b179 = choiceBuilder{id: 179, commit: 10}
	var b24 = choiceBuilder{id: 24, commit: 72, name: "class-char"}
	var b146 = sequenceBuilder{id: 146, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b109 = charBuilder{}
	b146.items = []builder{&b109}
	var b14 = sequenceBuilder{id: 14, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b66 = sequenceBuilder{id: 66, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b94 = charBuilder{}
	b66.items = []builder{&b94}
	var b155 = sequenceBuilder{id: 155, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b85 = charBuilder{}
	b155.items = []builder{&b85}
	b14.items = []builder{&b66, &b155}
	b24.options = []builder{&b146, &b14}
	var b86 = sequenceBuilder{id: 86, commit: 72, name: "char-range", ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b131 = sequenceBuilder{id: 131, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b95 = charBuilder{}
	b131.items = []builder{&b95}
	b86.items = []builder{&b24, &b131, &b24}
	b179.options = []builder{&b24, &b86}
	var b15 = sequenceBuilder{id: 15, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b27 = charBuilder{}
	b15.items = []builder{&b27}
	b16.items = []builder{&b165, &b137, &b179, &b15}
	var b180 = sequenceBuilder{id: 180, commit: 72, name: "char-sequence", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {1, 1}, {0, -1}, {1, 1}}}
	var b49 = sequenceBuilder{id: 49, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b26 = charBuilder{}
	b49.items = []builder{&b26}
	var b25 = choiceBuilder{id: 25, commit: 72, name: "sequence-char"}
	var b96 = sequenceBuilder{id: 96, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b182 = charBuilder{}
	b96.items = []builder{&b182}
	var b147 = sequenceBuilder{id: 147, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b87 = sequenceBuilder{id: 87, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b62 = charBuilder{}
	b87.items = []builder{&b62}
	var b90 = sequenceBuilder{id: 90, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b75 = charBuilder{}
	b90.items = []builder{&b75}
	b147.items = []builder{&b87, &b90}
	b25.options = []builder{&b96, &b147}
	var b88 = sequenceBuilder{id: 88, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b19 = charBuilder{}
	b88.items = []builder{&b19}
	b180.items = []builder{&b49, &b25, &b88}
	b42.options = []builder{&b73, &b16, &b180}
	var b43 = sequenceBuilder{id: 43, commit: 66, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}}}
	var b70 = sequenceBuilder{id: 70, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b28 = charBuilder{}
	b70.items = []builder{&b28}
	var b138 = sequenceBuilder{id: 138, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b171 = charBuilder{}
	b138.items = []builder{&b171}
	b43.items = []builder{&b70, &b186, &b6, &b186, &b138}
	var b30 = sequenceBuilder{id: 30, commit: 64, name: "sequence", ranges: [][]int{{1, 1}, {0, -1}}}
	var b132 = sequenceBuilder{id: 132, commit: 72, name: "item", ranges: [][]int{{1, 1}, {0, 1}, {1, 1}, {0, 1}}}
	var b57 = choiceBuilder{id: 57, commit: 10}
	b57.options = []builder{&b42, &b156, &b43}
	var b126 = choiceBuilder{id: 126, commit: 66}
	var b56 = sequenceBuilder{id: 56, commit: 64, name: "count-quantifier", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}}}
	var b17 = sequenceBuilder{id: 17, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b120 = charBuilder{}
	b17.items = []builder{&b120}
	var b4 = sequenceBuilder{id: 4, commit: 64, name: "count", ranges: [][]int{{1, 1}}}
	var b3 = sequenceBuilder{id: 3, commit: 74, ranges: [][]int{{1, -1}, {1, -1}}}
	var b76 = sequenceBuilder{id: 76, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b157 = charBuilder{}
	b76.items = []builder{&b157}
	b3.items = []builder{&b76}
	b4.items = []builder{&b3}
	var b67 = sequenceBuilder{id: 67, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b55 = charBuilder{}
	b67.items = []builder{&b55}
	b56.items = []builder{&b17, &b186, &b4, &b186, &b67}
	var b44 = sequenceBuilder{id: 44, commit: 64, name: "range-quantifier", ranges: [][]int{{1, 1}, {0, -1}, {0, 1}, {0, -1}, {1, 1}, {0, -1}, {0, 1}, {0, -1}, {1, 1}}}
	var b139 = sequenceBuilder{id: 139, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b166 = charBuilder{}
	b139.items = []builder{&b166}
	var b148 = sequenceBuilder{id: 148, commit: 64, name: "range-from", ranges: [][]int{{1, 1}}}
	b148.items = []builder{&b3}
	var b101 = sequenceBuilder{id: 101, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b80 = charBuilder{}
	b101.items = []builder{&b80}
	var b105 = sequenceBuilder{id: 105, commit: 64, name: "range-to", ranges: [][]int{{1, 1}}}
	b105.items = []builder{&b3}
	var b140 = sequenceBuilder{id: 140, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b5 = charBuilder{}
	b140.items = []builder{&b5}
	b44.items = []builder{&b139, &b186, &b148, &b186, &b101, &b186, &b105, &b186, &b140}
	var b89 = sequenceBuilder{id: 89, commit: 72, name: "one-or-more", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b115 = charBuilder{}
	b89.items = []builder{&b115}
	var b152 = sequenceBuilder{id: 152, commit: 72, name: "zero-or-more", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b81 = charBuilder{}
	b152.items = []builder{&b81}
	var b153 = sequenceBuilder{id: 153, commit: 72, name: "zero-or-one", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b63 = charBuilder{}
	b153.items = []builder{&b63}
	b126.options = []builder{&b56, &b44, &b89, &b152, &b153}
	b132.items = []builder{&b57, &b126}
	var b29 = sequenceBuilder{id: 29, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b29.items = []builder{&b186, &b132}
	b30.items = []builder{&b132, &b29}
	var b128 = sequenceBuilder{id: 128, commit: 64, name: "choice", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}}
	var b102 = choiceBuilder{id: 102, commit: 66}
	b102.options = []builder{&b42, &b156, &b43, &b30}
	var b184 = sequenceBuilder{id: 184, commit: 2, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}}}
	var b31 = sequenceBuilder{id: 31, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b172 = charBuilder{}
	b31.items = []builder{&b172}
	b184.items = []builder{&b31, &b186, &b102}
	var b127 = sequenceBuilder{id: 127, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b127.items = []builder{&b186, &b184}
	b128.items = []builder{&b102, &b186, &b184, &b127}
	b6.options = []builder{&b42, &b156, &b43, &b30, &b128}
	b174.items = []builder{&b83, &b186, &b59, &b186, &b6}
	var b168 = sequenceBuilder{id: 168, commit: 2, ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var b122 = sequenceBuilder{id: 122, commit: 2, ranges: [][]int{{1, 1}, {0, -1}, {0, -1}, {1, 1}}}
	var b133 = sequenceBuilder{id: 133, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b97 = charBuilder{}
	b133.items = []builder{&b97}
	var b121 = sequenceBuilder{id: 121, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b121.items = []builder{&b186, &b133}
	b122.items = []builder{&b133, &b121, &b186, &b174}
	var b167 = sequenceBuilder{id: 167, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b167.items = []builder{&b186, &b122}
	b168.items = []builder{&b186, &b122, &b167}
	b169.items = []builder{&b174, &b168}
	var b178 = sequenceBuilder{id: 178, commit: 2, ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var b69 = sequenceBuilder{id: 69, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b72 = charBuilder{}
	b69.items = []builder{&b72}
	var b177 = sequenceBuilder{id: 177, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b177.items = []builder{&b186, &b69}
	b178.items = []builder{&b186, &b69, &b177}
	b187.items = []builder{&b176, &b186, &b169, &b178}
	b188.items = []builder{&b186, &b187, &b186}

	return parseInput(r, &p188, &b188)
}
