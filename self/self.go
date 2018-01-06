/*
This file was generated with treerack (https://github.com/aryszka/treerack).

The contents of this file fall under different licenses.

The code between the "// head" and "// eo head" lines falls under the same
license as the source code of treerack (https://github.com/aryszka/treerack),
unless explicitly stated otherwise, if treerack's license allows changing the
license of this source code.

Treerack's license: MIT https://opensource.org/licenses/MIT
where YEAR=2018, COPYRIGHT HOLDER=Arpad Ryszka (arpad.ryszka@gmail.com)

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
	defer func() {
		if err := recover(); err != nil {
			println(len(n.tokens), n.From, n.To)
			panic(err)
		}
	}()
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
	var p114 = choiceParser{id: 114, commit: 66, name: "wschar", generalizations: []int{185, 186}}
	var p1 = sequenceParser{id: 1, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{114, 185, 186}}
	var p29 = charParser{id: 29, chars: []rune{32}}
	p1.items = []parser{&p29}
	var p80 = sequenceParser{id: 80, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{114, 185, 186}}
	var p105 = charParser{id: 105, chars: []rune{9}}
	p80.items = []parser{&p105}
	var p81 = sequenceParser{id: 81, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{114, 185, 186}}
	var p127 = charParser{id: 127, chars: []rune{10}}
	p81.items = []parser{&p127}
	var p2 = sequenceParser{id: 2, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{114, 185, 186}}
	var p61 = charParser{id: 61, chars: []rune{8}}
	p2.items = []parser{&p61}
	var p106 = sequenceParser{id: 106, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{114, 185, 186}}
	var p175 = charParser{id: 175, chars: []rune{12}}
	p106.items = []parser{&p175}
	var p113 = sequenceParser{id: 113, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{114, 185, 186}}
	var p69 = charParser{id: 69, chars: []rune{13}}
	p113.items = []parser{&p69}
	var p20 = sequenceParser{id: 20, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{114, 185, 186}}
	var p7 = charParser{id: 7, chars: []rune{11}}
	p20.items = []parser{&p7}
	p114.options = []parser{&p1, &p80, &p81, &p2, &p106, &p113, &p20}
	var p152 = sequenceParser{id: 152, commit: 72, name: "comment", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}, generalizations: []int{185, 186}}
	var p36 = choiceParser{id: 36, commit: 74, name: "comment-segment"}
	var p35 = sequenceParser{id: 35, commit: 74, name: "line-comment", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}, generalizations: []int{36}}
	var p145 = sequenceParser{id: 145, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var p71 = charParser{id: 71, chars: []rune{47}}
	var p160 = charParser{id: 160, chars: []rune{47}}
	p145.items = []parser{&p71, &p160}
	var p95 = sequenceParser{id: 95, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p21 = charParser{id: 21, not: true, chars: []rune{10}}
	p95.items = []parser{&p21}
	p35.items = []parser{&p145, &p95}
	var p70 = sequenceParser{id: 70, commit: 74, name: "block-comment", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {1, 1}, {0, -1}, {1, 1}}, generalizations: []int{36}}
	var p90 = sequenceParser{id: 90, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var p62 = charParser{id: 62, chars: []rune{47}}
	var p8 = charParser{id: 8, chars: []rune{42}}
	p90.items = []parser{&p62, &p8}
	var p13 = choiceParser{id: 13, commit: 10}
	var p3 = sequenceParser{id: 3, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{13}}
	var p48 = sequenceParser{id: 48, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p47 = charParser{id: 47, chars: []rune{42}}
	p48.items = []parser{&p47}
	var p99 = sequenceParser{id: 99, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p9 = charParser{id: 9, not: true, chars: []rune{47}}
	p99.items = []parser{&p9}
	p3.items = []parser{&p48, &p99}
	var p30 = sequenceParser{id: 30, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{13}}
	var p136 = charParser{id: 136, not: true, chars: []rune{42}}
	p30.items = []parser{&p136}
	p13.options = []parser{&p3, &p30}
	var p151 = sequenceParser{id: 151, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var p54 = charParser{id: 54, chars: []rune{42}}
	var p100 = charParser{id: 100, chars: []rune{47}}
	p151.items = []parser{&p54, &p100}
	p70.items = []parser{&p90, &p13, &p151}
	p36.options = []parser{&p35, &p70}
	var p10 = sequenceParser{id: 10, commit: 10, ranges: [][]int{{0, -1}, {0, 1}, {0, -1}, {1, 1}, {0, -1}, {0, 1}, {0, -1}, {1, 1}}}
	var p146 = choiceParser{id: 146, commit: 74, name: "ws-no-nl"}
	var p63 = sequenceParser{id: 63, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{146}}
	var p161 = charParser{id: 161, chars: []rune{32}}
	p63.items = []parser{&p161}
	var p107 = sequenceParser{id: 107, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{146}}
	var p86 = charParser{id: 86, chars: []rune{9}}
	p107.items = []parser{&p86}
	var p55 = sequenceParser{id: 55, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{146}}
	var p122 = charParser{id: 122, chars: []rune{8}}
	p55.items = []parser{&p122}
	var p108 = sequenceParser{id: 108, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{146}}
	var p96 = charParser{id: 96, chars: []rune{12}}
	p108.items = []parser{&p96}
	var p171 = sequenceParser{id: 171, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{146}}
	var p115 = charParser{id: 115, chars: []rune{13}}
	p171.items = []parser{&p115}
	var p137 = sequenceParser{id: 137, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{146}}
	var p123 = charParser{id: 123, chars: []rune{11}}
	p137.items = []parser{&p123}
	p146.options = []parser{&p63, &p107, &p55, &p108, &p171, &p137}
	var p162 = sequenceParser{id: 162, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p82 = charParser{id: 82, chars: []rune{10}}
	p162.items = []parser{&p82}
	p10.items = []parser{&p146, &p162, &p146, &p36}
	p152.items = []parser{&p36, &p10}
	p185.options = []parser{&p114, &p152}
	p186.options = []parser{&p185}
	var p187 = sequenceParser{id: 187, commit: 66, name: "syntax:wsroot", ranges: [][]int{{0, 1}, {0, -1}, {0, 1}, {0, 1}}}
	var p182 = sequenceParser{id: 182, commit: 2, ranges: [][]int{{1, 1}, {0, -1}}}
	var p135 = sequenceParser{id: 135, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p155 = charParser{id: 155, chars: []rune{59}}
	p135.items = []parser{&p155}
	var p181 = sequenceParser{id: 181, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p181.items = []parser{&p186, &p135}
	p182.items = []parser{&p135, &p181}
	var p121 = sequenceParser{id: 121, commit: 66, name: "definitions", ranges: [][]int{{1, 1}, {0, 1}}}
	var p44 = sequenceParser{id: 44, commit: 64, name: "definition", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}}}
	var p18 = sequenceParser{id: 18, commit: 74, name: "definition-name", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}}
	var p128 = sequenceParser{id: 128, commit: 72, name: "symbol", ranges: [][]int{{1, -1}, {1, -1}}, generalizations: []int{52, 26, 177}}
	var p39 = sequenceParser{id: 39, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p57 = charParser{id: 57, not: true, chars: []rune{92, 32, 10, 9, 8, 12, 13, 11, 47, 46, 91, 93, 34, 123, 125, 94, 43, 42, 63, 124, 40, 41, 58, 61, 59}}
	p39.items = []parser{&p57}
	p128.items = []parser{&p39}
	var p178 = sequenceParser{id: 178, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var p133 = sequenceParser{id: 133, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p43 = charParser{id: 43, chars: []rune{58}}
	p133.items = []parser{&p43}
	var p132 = choiceParser{id: 132, commit: 66, name: "flag"}
	var p93 = sequenceParser{id: 93, commit: 72, name: "alias", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{132}}
	var p163 = charParser{id: 163, chars: []rune{97}}
	var p27 = charParser{id: 27, chars: []rune{108}}
	var p164 = charParser{id: 164, chars: []rune{105}}
	var p158 = charParser{id: 158, chars: []rune{97}}
	var p118 = charParser{id: 118, chars: []rune{115}}
	p93.items = []parser{&p163, &p27, &p164, &p158, &p118}
	var p131 = sequenceParser{id: 131, commit: 72, name: "ws", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{132}}
	var p53 = charParser{id: 53, chars: []rune{119}}
	var p5 = charParser{id: 5, chars: []rune{115}}
	p131.items = []parser{&p53, &p5}
	var p172 = sequenceParser{id: 172, commit: 72, name: "nows", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{132}}
	var p109 = charParser{id: 109, chars: []rune{110}}
	var p110 = charParser{id: 110, chars: []rune{111}}
	var p60 = charParser{id: 60, chars: []rune{119}}
	var p79 = charParser{id: 79, chars: []rune{115}}
	p172.items = []parser{&p109, &p110, &p60, &p79}
	var p25 = sequenceParser{id: 25, commit: 72, name: "failpass", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{132}}
	var p45 = charParser{id: 45, chars: []rune{102}}
	var p94 = charParser{id: 94, chars: []rune{97}}
	var p89 = charParser{id: 89, chars: []rune{105}}
	var p28 = charParser{id: 28, chars: []rune{108}}
	var p42 = charParser{id: 42, chars: []rune{112}}
	var p173 = charParser{id: 173, chars: []rune{97}}
	var p149 = charParser{id: 149, chars: []rune{115}}
	var p150 = charParser{id: 150, chars: []rune{115}}
	p25.items = []parser{&p45, &p94, &p89, &p28, &p42, &p173, &p149, &p150}
	var p144 = sequenceParser{id: 144, commit: 72, name: "root", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{132}}
	var p159 = charParser{id: 159, chars: []rune{114}}
	var p174 = charParser{id: 174, chars: []rune{111}}
	var p76 = charParser{id: 76, chars: []rune{111}}
	var p75 = charParser{id: 75, chars: []rune{116}}
	p144.items = []parser{&p159, &p174, &p76, &p75}
	p132.options = []parser{&p93, &p131, &p172, &p25, &p144}
	p178.items = []parser{&p133, &p132}
	p18.items = []parser{&p128, &p178}
	var p111 = sequenceParser{id: 111, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p19 = charParser{id: 19, chars: []rune{61}}
	p111.items = []parser{&p19}
	var p52 = choiceParser{id: 52, commit: 66, name: "expression"}
	var p11 = choiceParser{id: 11, commit: 66, name: "terminal", generalizations: []int{52, 26, 177}}
	var p37 = sequenceParser{id: 37, commit: 72, name: "any-char", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{11, 52, 26, 177}}
	var p97 = charParser{id: 97, chars: []rune{46}}
	p37.items = []parser{&p97}
	var p49 = sequenceParser{id: 49, commit: 72, name: "char-class", ranges: [][]int{{1, 1}, {0, 1}, {0, -1}, {1, 1}, {1, 1}, {0, 1}, {0, -1}, {1, 1}}, generalizations: []int{11, 52, 26, 177}}
	var p101 = sequenceParser{id: 101, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p67 = charParser{id: 67, chars: []rune{91}}
	p101.items = []parser{&p67}
	var p87 = sequenceParser{id: 87, commit: 72, name: "class-not", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p138 = charParser{id: 138, chars: []rune{94}}
	p87.items = []parser{&p138}
	var p83 = choiceParser{id: 83, commit: 10}
	var p74 = choiceParser{id: 74, commit: 72, name: "class-char", generalizations: []int{83}}
	var p56 = sequenceParser{id: 56, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{74, 83}}
	var p165 = charParser{id: 165, not: true, chars: []rune{92, 91, 93, 94, 45}}
	p56.items = []parser{&p165}
	var p31 = sequenceParser{id: 31, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{74, 83}}
	var p147 = sequenceParser{id: 147, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p91 = charParser{id: 91, chars: []rune{92}}
	p147.items = []parser{&p91}
	var p73 = sequenceParser{id: 73, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p72 = charParser{id: 72, not: true}
	p73.items = []parser{&p72}
	p31.items = []parser{&p147, &p73}
	p74.options = []parser{&p56, &p31}
	var p64 = sequenceParser{id: 64, commit: 72, name: "char-range", ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{83}}
	var p167 = sequenceParser{id: 167, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p77 = charParser{id: 77, chars: []rune{45}}
	p167.items = []parser{&p77}
	p64.items = []parser{&p74, &p167, &p74}
	p83.options = []parser{&p74, &p64}
	var p65 = sequenceParser{id: 65, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p124 = charParser{id: 124, chars: []rune{93}}
	p65.items = []parser{&p124}
	p49.items = []parser{&p101, &p87, &p83, &p65}
	var p140 = sequenceParser{id: 140, commit: 72, name: "char-sequence", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {1, 1}, {0, -1}, {1, 1}}, generalizations: []int{11, 52, 26, 177}}
	var p38 = sequenceParser{id: 38, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p88 = charParser{id: 88, chars: []rune{34}}
	p38.items = []parser{&p88}
	var p22 = choiceParser{id: 22, commit: 72, name: "sequence-char"}
	var p156 = sequenceParser{id: 156, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{22}}
	var p102 = charParser{id: 102, not: true, chars: []rune{92, 34}}
	p156.items = []parser{&p102}
	var p153 = sequenceParser{id: 153, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{22}}
	var p139 = sequenceParser{id: 139, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p50 = charParser{id: 50, chars: []rune{92}}
	p139.items = []parser{&p50}
	var p78 = sequenceParser{id: 78, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p116 = charParser{id: 116, not: true}
	p78.items = []parser{&p116}
	p153.items = []parser{&p139, &p78}
	p22.options = []parser{&p156, &p153}
	var p142 = sequenceParser{id: 142, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p14 = charParser{id: 14, chars: []rune{34}}
	p142.items = []parser{&p14}
	p140.items = []parser{&p38, &p22, &p142}
	p11.options = []parser{&p37, &p49, &p140}
	var p33 = sequenceParser{id: 33, commit: 66, name: "group", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}}, generalizations: []int{52, 26, 177}}
	var p40 = sequenceParser{id: 40, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p166 = charParser{id: 166, chars: []rune{40}}
	p40.items = []parser{&p166}
	var p12 = sequenceParser{id: 12, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p98 = charParser{id: 98, chars: []rune{41}}
	p12.items = []parser{&p98}
	p33.items = []parser{&p40, &p186, &p52, &p186, &p12}
	var p130 = sequenceParser{id: 130, commit: 64, name: "sequence", ranges: [][]int{{1, 1}, {0, -1}}, generalizations: []int{52, 177}}
	var p143 = sequenceParser{id: 143, commit: 72, name: "item", ranges: [][]int{{1, 1}, {0, 1}, {1, 1}, {0, 1}}}
	var p26 = choiceParser{id: 26, commit: 10}
	p26.options = []parser{&p11, &p128, &p33}
	var p176 = choiceParser{id: 176, commit: 66, name: "quantity"}
	var p148 = sequenceParser{id: 148, commit: 64, name: "count-quantifier", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}}, generalizations: []int{176}}
	var p58 = sequenceParser{id: 58, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p6 = charParser{id: 6, chars: []rune{123}}
	p58.items = []parser{&p6}
	var p15 = sequenceParser{id: 15, commit: 64, name: "count", ranges: [][]int{{1, 1}}}
	var p32 = sequenceParser{id: 32, commit: 74, name: "number", ranges: [][]int{{1, -1}, {1, -1}}}
	var p68 = sequenceParser{id: 68, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p34 = charParser{id: 34, ranges: [][]rune{{48, 57}}}
	p68.items = []parser{&p34}
	p32.items = []parser{&p68}
	p15.items = []parser{&p32}
	var p157 = sequenceParser{id: 157, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p103 = charParser{id: 103, chars: []rune{125}}
	p157.items = []parser{&p103}
	p148.items = []parser{&p58, &p186, &p15, &p186, &p157}
	var p141 = sequenceParser{id: 141, commit: 64, name: "range-quantifier", ranges: [][]int{{1, 1}, {0, -1}, {0, 1}, {0, -1}, {1, 1}, {0, -1}, {0, 1}, {0, -1}, {1, 1}}, generalizations: []int{176}}
	var p85 = sequenceParser{id: 85, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p66 = charParser{id: 66, chars: []rune{123}}
	p85.items = []parser{&p66}
	var p125 = sequenceParser{id: 125, commit: 64, name: "range-from", ranges: [][]int{{1, 1}}}
	p125.items = []parser{&p32}
	var p168 = sequenceParser{id: 168, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p41 = charParser{id: 41, chars: []rune{44}}
	p168.items = []parser{&p41}
	var p84 = sequenceParser{id: 84, commit: 64, name: "range-to", ranges: [][]int{{1, 1}}}
	p84.items = []parser{&p32}
	var p16 = sequenceParser{id: 16, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p51 = charParser{id: 51, chars: []rune{125}}
	p16.items = []parser{&p51}
	p141.items = []parser{&p85, &p186, &p125, &p186, &p168, &p186, &p84, &p186, &p16}
	var p4 = sequenceParser{id: 4, commit: 72, name: "one-or-more", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{176}}
	var p23 = charParser{id: 23, chars: []rune{43}}
	p4.items = []parser{&p23}
	var p17 = sequenceParser{id: 17, commit: 72, name: "zero-or-more", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{176}}
	var p134 = charParser{id: 134, chars: []rune{42}}
	p17.items = []parser{&p134}
	var p59 = sequenceParser{id: 59, commit: 72, name: "zero-or-one", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{176}}
	var p117 = charParser{id: 117, chars: []rune{63}}
	p59.items = []parser{&p117}
	p176.options = []parser{&p148, &p141, &p4, &p17, &p59}
	p143.items = []parser{&p26, &p176}
	var p129 = sequenceParser{id: 129, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p129.items = []parser{&p186, &p143}
	p130.items = []parser{&p143, &p129}
	var p170 = sequenceParser{id: 170, commit: 64, name: "choice", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}, generalizations: []int{52}}
	var p177 = choiceParser{id: 177, commit: 66, name: "option"}
	p177.options = []parser{&p11, &p128, &p33, &p130}
	var p92 = sequenceParser{id: 92, commit: 2, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}}}
	var p24 = sequenceParser{id: 24, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p154 = charParser{id: 154, chars: []rune{124}}
	p24.items = []parser{&p154}
	p92.items = []parser{&p24, &p186, &p177}
	var p169 = sequenceParser{id: 169, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p169.items = []parser{&p186, &p92}
	p170.items = []parser{&p177, &p186, &p92, &p169}
	p52.options = []parser{&p11, &p128, &p33, &p130, &p170}
	p44.items = []parser{&p18, &p186, &p111, &p186, &p52}
	var p120 = sequenceParser{id: 120, commit: 2, ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var p180 = sequenceParser{id: 180, commit: 2, ranges: [][]int{{1, 1}, {0, -1}, {0, -1}, {1, 1}}}
	var p112 = sequenceParser{id: 112, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p104 = charParser{id: 104, chars: []rune{59}}
	p112.items = []parser{&p104}
	var p179 = sequenceParser{id: 179, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p179.items = []parser{&p186, &p112}
	p180.items = []parser{&p112, &p179, &p186, &p44}
	var p119 = sequenceParser{id: 119, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p119.items = []parser{&p186, &p180}
	p120.items = []parser{&p186, &p180, &p119}
	p121.items = []parser{&p44, &p120}
	var p184 = sequenceParser{id: 184, commit: 2, ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var p126 = sequenceParser{id: 126, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p46 = charParser{id: 46, chars: []rune{59}}
	p126.items = []parser{&p46}
	var p183 = sequenceParser{id: 183, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p183.items = []parser{&p186, &p126}
	p184.items = []parser{&p186, &p126, &p183}
	p187.items = []parser{&p182, &p186, &p121, &p184}
	p188.items = []parser{&p186, &p187, &p186}
	var b188 = sequenceBuilder{id: 188, commit: 32, name: "syntax", ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var b186 = choiceBuilder{id: 186, commit: 2}
	var b185 = choiceBuilder{id: 185, commit: 70}
	var b114 = choiceBuilder{id: 114, commit: 66}
	var b1 = sequenceBuilder{id: 1, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b29 = charBuilder{}
	b1.items = []builder{&b29}
	var b80 = sequenceBuilder{id: 80, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b105 = charBuilder{}
	b80.items = []builder{&b105}
	var b81 = sequenceBuilder{id: 81, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b127 = charBuilder{}
	b81.items = []builder{&b127}
	var b2 = sequenceBuilder{id: 2, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b61 = charBuilder{}
	b2.items = []builder{&b61}
	var b106 = sequenceBuilder{id: 106, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b175 = charBuilder{}
	b106.items = []builder{&b175}
	var b113 = sequenceBuilder{id: 113, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b69 = charBuilder{}
	b113.items = []builder{&b69}
	var b20 = sequenceBuilder{id: 20, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b7 = charBuilder{}
	b20.items = []builder{&b7}
	b114.options = []builder{&b1, &b80, &b81, &b2, &b106, &b113, &b20}
	var b152 = sequenceBuilder{id: 152, commit: 72, name: "comment", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}}
	var b36 = choiceBuilder{id: 36, commit: 74}
	var b35 = sequenceBuilder{id: 35, commit: 74, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}}
	var b145 = sequenceBuilder{id: 145, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b71 = charBuilder{}
	var b160 = charBuilder{}
	b145.items = []builder{&b71, &b160}
	var b95 = sequenceBuilder{id: 95, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b21 = charBuilder{}
	b95.items = []builder{&b21}
	b35.items = []builder{&b145, &b95}
	var b70 = sequenceBuilder{id: 70, commit: 74, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {1, 1}, {0, -1}, {1, 1}}}
	var b90 = sequenceBuilder{id: 90, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b62 = charBuilder{}
	var b8 = charBuilder{}
	b90.items = []builder{&b62, &b8}
	var b13 = choiceBuilder{id: 13, commit: 10}
	var b3 = sequenceBuilder{id: 3, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b48 = sequenceBuilder{id: 48, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b47 = charBuilder{}
	b48.items = []builder{&b47}
	var b99 = sequenceBuilder{id: 99, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b9 = charBuilder{}
	b99.items = []builder{&b9}
	b3.items = []builder{&b48, &b99}
	var b30 = sequenceBuilder{id: 30, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b136 = charBuilder{}
	b30.items = []builder{&b136}
	b13.options = []builder{&b3, &b30}
	var b151 = sequenceBuilder{id: 151, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b54 = charBuilder{}
	var b100 = charBuilder{}
	b151.items = []builder{&b54, &b100}
	b70.items = []builder{&b90, &b13, &b151}
	b36.options = []builder{&b35, &b70}
	var b10 = sequenceBuilder{id: 10, commit: 10, ranges: [][]int{{0, -1}, {0, 1}, {0, -1}, {1, 1}, {0, -1}, {0, 1}, {0, -1}, {1, 1}}}
	var b146 = choiceBuilder{id: 146, commit: 74}
	var b63 = sequenceBuilder{id: 63, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b161 = charBuilder{}
	b63.items = []builder{&b161}
	var b107 = sequenceBuilder{id: 107, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b86 = charBuilder{}
	b107.items = []builder{&b86}
	var b55 = sequenceBuilder{id: 55, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b122 = charBuilder{}
	b55.items = []builder{&b122}
	var b108 = sequenceBuilder{id: 108, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b96 = charBuilder{}
	b108.items = []builder{&b96}
	var b171 = sequenceBuilder{id: 171, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b115 = charBuilder{}
	b171.items = []builder{&b115}
	var b137 = sequenceBuilder{id: 137, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b123 = charBuilder{}
	b137.items = []builder{&b123}
	b146.options = []builder{&b63, &b107, &b55, &b108, &b171, &b137}
	var b162 = sequenceBuilder{id: 162, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b82 = charBuilder{}
	b162.items = []builder{&b82}
	b10.items = []builder{&b146, &b162, &b146, &b36}
	b152.items = []builder{&b36, &b10}
	b185.options = []builder{&b114, &b152}
	b186.options = []builder{&b185}
	var b187 = sequenceBuilder{id: 187, commit: 66, ranges: [][]int{{0, 1}, {0, -1}, {0, 1}, {0, 1}}}
	var b182 = sequenceBuilder{id: 182, commit: 2, ranges: [][]int{{1, 1}, {0, -1}}}
	var b135 = sequenceBuilder{id: 135, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b155 = charBuilder{}
	b135.items = []builder{&b155}
	var b181 = sequenceBuilder{id: 181, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b181.items = []builder{&b186, &b135}
	b182.items = []builder{&b135, &b181}
	var b121 = sequenceBuilder{id: 121, commit: 66, ranges: [][]int{{1, 1}, {0, 1}}}
	var b44 = sequenceBuilder{id: 44, commit: 64, name: "definition", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}}}
	var b18 = sequenceBuilder{id: 18, commit: 74, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}}
	var b128 = sequenceBuilder{id: 128, commit: 72, name: "symbol", ranges: [][]int{{1, -1}, {1, -1}}}
	var b39 = sequenceBuilder{id: 39, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b57 = charBuilder{}
	b39.items = []builder{&b57}
	b128.items = []builder{&b39}
	var b178 = sequenceBuilder{id: 178, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b133 = sequenceBuilder{id: 133, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b43 = charBuilder{}
	b133.items = []builder{&b43}
	var b132 = choiceBuilder{id: 132, commit: 66}
	var b93 = sequenceBuilder{id: 93, commit: 72, name: "alias", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b163 = charBuilder{}
	var b27 = charBuilder{}
	var b164 = charBuilder{}
	var b158 = charBuilder{}
	var b118 = charBuilder{}
	b93.items = []builder{&b163, &b27, &b164, &b158, &b118}
	var b131 = sequenceBuilder{id: 131, commit: 72, name: "ws", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b53 = charBuilder{}
	var b5 = charBuilder{}
	b131.items = []builder{&b53, &b5}
	var b172 = sequenceBuilder{id: 172, commit: 72, name: "nows", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b109 = charBuilder{}
	var b110 = charBuilder{}
	var b60 = charBuilder{}
	var b79 = charBuilder{}
	b172.items = []builder{&b109, &b110, &b60, &b79}
	var b25 = sequenceBuilder{id: 25, commit: 72, name: "failpass", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b45 = charBuilder{}
	var b94 = charBuilder{}
	var b89 = charBuilder{}
	var b28 = charBuilder{}
	var b42 = charBuilder{}
	var b173 = charBuilder{}
	var b149 = charBuilder{}
	var b150 = charBuilder{}
	b25.items = []builder{&b45, &b94, &b89, &b28, &b42, &b173, &b149, &b150}
	var b144 = sequenceBuilder{id: 144, commit: 72, name: "root", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b159 = charBuilder{}
	var b174 = charBuilder{}
	var b76 = charBuilder{}
	var b75 = charBuilder{}
	b144.items = []builder{&b159, &b174, &b76, &b75}
	b132.options = []builder{&b93, &b131, &b172, &b25, &b144}
	b178.items = []builder{&b133, &b132}
	b18.items = []builder{&b128, &b178}
	var b111 = sequenceBuilder{id: 111, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b19 = charBuilder{}
	b111.items = []builder{&b19}
	var b52 = choiceBuilder{id: 52, commit: 66}
	var b11 = choiceBuilder{id: 11, commit: 66}
	var b37 = sequenceBuilder{id: 37, commit: 72, name: "any-char", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b97 = charBuilder{}
	b37.items = []builder{&b97}
	var b49 = sequenceBuilder{id: 49, commit: 72, name: "char-class", ranges: [][]int{{1, 1}, {0, 1}, {0, -1}, {1, 1}, {1, 1}, {0, 1}, {0, -1}, {1, 1}}}
	var b101 = sequenceBuilder{id: 101, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b67 = charBuilder{}
	b101.items = []builder{&b67}
	var b87 = sequenceBuilder{id: 87, commit: 72, name: "class-not", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b138 = charBuilder{}
	b87.items = []builder{&b138}
	var b83 = choiceBuilder{id: 83, commit: 10}
	var b74 = choiceBuilder{id: 74, commit: 72, name: "class-char"}
	var b56 = sequenceBuilder{id: 56, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b165 = charBuilder{}
	b56.items = []builder{&b165}
	var b31 = sequenceBuilder{id: 31, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b147 = sequenceBuilder{id: 147, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b91 = charBuilder{}
	b147.items = []builder{&b91}
	var b73 = sequenceBuilder{id: 73, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b72 = charBuilder{}
	b73.items = []builder{&b72}
	b31.items = []builder{&b147, &b73}
	b74.options = []builder{&b56, &b31}
	var b64 = sequenceBuilder{id: 64, commit: 72, name: "char-range", ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b167 = sequenceBuilder{id: 167, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b77 = charBuilder{}
	b167.items = []builder{&b77}
	b64.items = []builder{&b74, &b167, &b74}
	b83.options = []builder{&b74, &b64}
	var b65 = sequenceBuilder{id: 65, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b124 = charBuilder{}
	b65.items = []builder{&b124}
	b49.items = []builder{&b101, &b87, &b83, &b65}
	var b140 = sequenceBuilder{id: 140, commit: 72, name: "char-sequence", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {1, 1}, {0, -1}, {1, 1}}}
	var b38 = sequenceBuilder{id: 38, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b88 = charBuilder{}
	b38.items = []builder{&b88}
	var b22 = choiceBuilder{id: 22, commit: 72, name: "sequence-char"}
	var b156 = sequenceBuilder{id: 156, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b102 = charBuilder{}
	b156.items = []builder{&b102}
	var b153 = sequenceBuilder{id: 153, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b139 = sequenceBuilder{id: 139, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b50 = charBuilder{}
	b139.items = []builder{&b50}
	var b78 = sequenceBuilder{id: 78, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b116 = charBuilder{}
	b78.items = []builder{&b116}
	b153.items = []builder{&b139, &b78}
	b22.options = []builder{&b156, &b153}
	var b142 = sequenceBuilder{id: 142, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b14 = charBuilder{}
	b142.items = []builder{&b14}
	b140.items = []builder{&b38, &b22, &b142}
	b11.options = []builder{&b37, &b49, &b140}
	var b33 = sequenceBuilder{id: 33, commit: 66, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}}}
	var b40 = sequenceBuilder{id: 40, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b166 = charBuilder{}
	b40.items = []builder{&b166}
	var b12 = sequenceBuilder{id: 12, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b98 = charBuilder{}
	b12.items = []builder{&b98}
	b33.items = []builder{&b40, &b186, &b52, &b186, &b12}
	var b130 = sequenceBuilder{id: 130, commit: 64, name: "sequence", ranges: [][]int{{1, 1}, {0, -1}}}
	var b143 = sequenceBuilder{id: 143, commit: 72, name: "item", ranges: [][]int{{1, 1}, {0, 1}, {1, 1}, {0, 1}}}
	var b26 = choiceBuilder{id: 26, commit: 10}
	b26.options = []builder{&b11, &b128, &b33}
	var b176 = choiceBuilder{id: 176, commit: 66}
	var b148 = sequenceBuilder{id: 148, commit: 64, name: "count-quantifier", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}}}
	var b58 = sequenceBuilder{id: 58, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b6 = charBuilder{}
	b58.items = []builder{&b6}
	var b15 = sequenceBuilder{id: 15, commit: 64, name: "count", ranges: [][]int{{1, 1}}}
	var b32 = sequenceBuilder{id: 32, commit: 74, ranges: [][]int{{1, -1}, {1, -1}}}
	var b68 = sequenceBuilder{id: 68, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b34 = charBuilder{}
	b68.items = []builder{&b34}
	b32.items = []builder{&b68}
	b15.items = []builder{&b32}
	var b157 = sequenceBuilder{id: 157, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b103 = charBuilder{}
	b157.items = []builder{&b103}
	b148.items = []builder{&b58, &b186, &b15, &b186, &b157}
	var b141 = sequenceBuilder{id: 141, commit: 64, name: "range-quantifier", ranges: [][]int{{1, 1}, {0, -1}, {0, 1}, {0, -1}, {1, 1}, {0, -1}, {0, 1}, {0, -1}, {1, 1}}}
	var b85 = sequenceBuilder{id: 85, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b66 = charBuilder{}
	b85.items = []builder{&b66}
	var b125 = sequenceBuilder{id: 125, commit: 64, name: "range-from", ranges: [][]int{{1, 1}}}
	b125.items = []builder{&b32}
	var b168 = sequenceBuilder{id: 168, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b41 = charBuilder{}
	b168.items = []builder{&b41}
	var b84 = sequenceBuilder{id: 84, commit: 64, name: "range-to", ranges: [][]int{{1, 1}}}
	b84.items = []builder{&b32}
	var b16 = sequenceBuilder{id: 16, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b51 = charBuilder{}
	b16.items = []builder{&b51}
	b141.items = []builder{&b85, &b186, &b125, &b186, &b168, &b186, &b84, &b186, &b16}
	var b4 = sequenceBuilder{id: 4, commit: 72, name: "one-or-more", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b23 = charBuilder{}
	b4.items = []builder{&b23}
	var b17 = sequenceBuilder{id: 17, commit: 72, name: "zero-or-more", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b134 = charBuilder{}
	b17.items = []builder{&b134}
	var b59 = sequenceBuilder{id: 59, commit: 72, name: "zero-or-one", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b117 = charBuilder{}
	b59.items = []builder{&b117}
	b176.options = []builder{&b148, &b141, &b4, &b17, &b59}
	b143.items = []builder{&b26, &b176}
	var b129 = sequenceBuilder{id: 129, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b129.items = []builder{&b186, &b143}
	b130.items = []builder{&b143, &b129}
	var b170 = sequenceBuilder{id: 170, commit: 64, name: "choice", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}}
	var b177 = choiceBuilder{id: 177, commit: 66}
	b177.options = []builder{&b11, &b128, &b33, &b130}
	var b92 = sequenceBuilder{id: 92, commit: 2, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}}}
	var b24 = sequenceBuilder{id: 24, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b154 = charBuilder{}
	b24.items = []builder{&b154}
	b92.items = []builder{&b24, &b186, &b177}
	var b169 = sequenceBuilder{id: 169, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b169.items = []builder{&b186, &b92}
	b170.items = []builder{&b177, &b186, &b92, &b169}
	b52.options = []builder{&b11, &b128, &b33, &b130, &b170}
	b44.items = []builder{&b18, &b186, &b111, &b186, &b52}
	var b120 = sequenceBuilder{id: 120, commit: 2, ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var b180 = sequenceBuilder{id: 180, commit: 2, ranges: [][]int{{1, 1}, {0, -1}, {0, -1}, {1, 1}}}
	var b112 = sequenceBuilder{id: 112, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b104 = charBuilder{}
	b112.items = []builder{&b104}
	var b179 = sequenceBuilder{id: 179, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b179.items = []builder{&b186, &b112}
	b180.items = []builder{&b112, &b179, &b186, &b44}
	var b119 = sequenceBuilder{id: 119, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b119.items = []builder{&b186, &b180}
	b120.items = []builder{&b186, &b180, &b119}
	b121.items = []builder{&b44, &b120}
	var b184 = sequenceBuilder{id: 184, commit: 2, ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var b126 = sequenceBuilder{id: 126, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b46 = charBuilder{}
	b126.items = []builder{&b46}
	var b183 = sequenceBuilder{id: 183, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b183.items = []builder{&b186, &b126}
	b184.items = []builder{&b186, &b126, &b183}
	b187.items = []builder{&b182, &b186, &b121, &b184}
	b188.items = []builder{&b186, &b187, &b186}

	return parse(r, &p188, &b188)
}
