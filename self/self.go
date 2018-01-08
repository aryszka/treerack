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
	var p155 = choiceParser{id: 155, commit: 66, name: "wschar", generalizations: []int{185, 186}}
	var p97 = sequenceParser{id: 97, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{155, 185, 186}}
	var p61 = charParser{id: 61, chars: []rune{32}}
	p97.items = []parser{&p61}
	var p179 = sequenceParser{id: 179, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{155, 185, 186}}
	var p1 = charParser{id: 1, chars: []rune{9}}
	p179.items = []parser{&p1}
	var p2 = sequenceParser{id: 2, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{155, 185, 186}}
	var p119 = charParser{id: 119, chars: []rune{10}}
	p2.items = []parser{&p119}
	var p120 = sequenceParser{id: 120, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{155, 185, 186}}
	var p98 = charParser{id: 98, chars: []rune{8}}
	p120.items = []parser{&p98}
	var p105 = sequenceParser{id: 105, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{155, 185, 186}}
	var p71 = charParser{id: 71, chars: []rune{12}}
	p105.items = []parser{&p71}
	var p149 = sequenceParser{id: 149, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{155, 185, 186}}
	var p139 = charParser{id: 139, chars: []rune{13}}
	p149.items = []parser{&p139}
	var p99 = sequenceParser{id: 99, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{155, 185, 186}}
	var p20 = charParser{id: 20, chars: []rune{11}}
	p99.items = []parser{&p20}
	p155.options = []parser{&p97, &p179, &p2, &p120, &p105, &p149, &p99}
	var p63 = sequenceParser{id: 63, commit: 72, name: "comment", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}, generalizations: []int{185, 186}}
	var p33 = choiceParser{id: 33, commit: 74, name: "comment-segment"}
	var p101 = sequenceParser{id: 101, commit: 74, name: "line-comment", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}, generalizations: []int{33}}
	var p9 = sequenceParser{id: 9, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var p88 = charParser{id: 88, chars: []rune{47}}
	var p164 = charParser{id: 164, chars: []rune{47}}
	p9.items = []parser{&p88, &p164}
	var p89 = sequenceParser{id: 89, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p62 = charParser{id: 62, not: true, chars: []rune{10}}
	p89.items = []parser{&p62}
	p101.items = []parser{&p9, &p89}
	var p72 = sequenceParser{id: 72, commit: 74, name: "block-comment", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {1, 1}, {0, -1}, {1, 1}}, generalizations: []int{33}}
	var p21 = sequenceParser{id: 21, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var p172 = charParser{id: 172, chars: []rune{47}}
	var p140 = charParser{id: 140, chars: []rune{42}}
	p21.items = []parser{&p172, &p140}
	var p27 = choiceParser{id: 27, commit: 10}
	var p8 = sequenceParser{id: 8, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{27}}
	var p7 = sequenceParser{id: 7, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p141 = charParser{id: 141, chars: []rune{42}}
	p7.items = []parser{&p141}
	var p22 = sequenceParser{id: 22, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p3 = charParser{id: 3, not: true, chars: []rune{47}}
	p22.items = []parser{&p3}
	p8.items = []parser{&p7, &p22}
	var p100 = sequenceParser{id: 100, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{27}}
	var p76 = charParser{id: 76, not: true, chars: []rune{42}}
	p100.items = []parser{&p76}
	p27.options = []parser{&p8, &p100}
	var p49 = sequenceParser{id: 49, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var p127 = charParser{id: 127, chars: []rune{42}}
	var p150 = charParser{id: 150, chars: []rune{47}}
	p49.items = []parser{&p127, &p150}
	p72.items = []parser{&p21, &p27, &p49}
	p33.options = []parser{&p101, &p72}
	var p133 = sequenceParser{id: 133, commit: 10, ranges: [][]int{{0, -1}, {0, 1}, {0, -1}, {1, 1}, {0, -1}, {0, 1}, {0, -1}, {1, 1}}}
	var p53 = choiceParser{id: 53, commit: 74, name: "ws-no-nl"}
	var p10 = sequenceParser{id: 10, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{53}}
	var p34 = charParser{id: 34, chars: []rune{32}}
	p10.items = []parser{&p34}
	var p151 = sequenceParser{id: 151, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{53}}
	var p110 = charParser{id: 110, chars: []rune{9}}
	p151.items = []parser{&p110}
	var p142 = sequenceParser{id: 142, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{53}}
	var p102 = charParser{id: 102, chars: []rune{8}}
	p142.items = []parser{&p102}
	var p152 = sequenceParser{id: 152, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{53}}
	var p35 = charParser{id: 35, chars: []rune{12}}
	p152.items = []parser{&p35}
	var p28 = sequenceParser{id: 28, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{53}}
	var p50 = charParser{id: 50, chars: []rune{13}}
	p28.items = []parser{&p50}
	var p159 = sequenceParser{id: 159, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{53}}
	var p23 = charParser{id: 23, chars: []rune{11}}
	p159.items = []parser{&p23}
	p53.options = []parser{&p10, &p151, &p142, &p152, &p28, &p159}
	var p43 = sequenceParser{id: 43, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p95 = charParser{id: 95, chars: []rune{10}}
	p43.items = []parser{&p95}
	p133.items = []parser{&p53, &p43, &p53, &p33}
	p63.items = []parser{&p33, &p133}
	p185.options = []parser{&p155, &p63}
	p186.options = []parser{&p185}
	var p187 = sequenceParser{id: 187, commit: 66, name: "syntax:wsroot", ranges: [][]int{{0, 1}, {0, -1}, {0, 1}, {0, 1}}}
	var p85 = sequenceParser{id: 85, commit: 2, ranges: [][]int{{1, 1}, {0, -1}}}
	var p145 = sequenceParser{id: 145, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p58 = charParser{id: 58, chars: []rune{59}}
	p145.items = []parser{&p58}
	var p84 = sequenceParser{id: 84, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p84.items = []parser{&p186, &p145}
	p85.items = []parser{&p145, &p84}
	var p42 = sequenceParser{id: 42, commit: 66, name: "definitions", ranges: [][]int{{1, 1}, {0, 1}}}
	var p4 = sequenceParser{id: 4, commit: 64, name: "definition", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}}}
	var p138 = sequenceParser{id: 138, commit: 74, name: "definition-name", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}}
	var p11 = sequenceParser{id: 11, commit: 72, name: "symbol", ranges: [][]int{{1, -1}, {1, -1}}, generalizations: []int{154, 160, 158}}
	var p47 = sequenceParser{id: 47, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p46 = charParser{id: 46, not: true, chars: []rune{92, 32, 10, 9, 8, 12, 13, 11, 47, 46, 91, 93, 34, 123, 125, 94, 43, 42, 63, 124, 40, 41, 58, 61, 59}}
	p47.items = []parser{&p46}
	p11.items = []parser{&p47}
	var p103 = sequenceParser{id: 103, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var p26 = sequenceParser{id: 26, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p19 = charParser{id: 19, chars: []rune{58}}
	p26.items = []parser{&p19}
	var p39 = choiceParser{id: 39, commit: 66, name: "flag"}
	var p96 = sequenceParser{id: 96, commit: 72, name: "alias", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{39}}
	var p69 = charParser{id: 69, chars: []rune{97}}
	var p114 = charParser{id: 114, chars: []rune{108}}
	var p162 = charParser{id: 162, chars: []rune{105}}
	var p161 = charParser{id: 161, chars: []rune{97}}
	var p137 = charParser{id: 137, chars: []rune{115}}
	p96.items = []parser{&p69, &p114, &p162, &p161, &p137}
	var p31 = sequenceParser{id: 31, commit: 72, name: "ws", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{39}}
	var p184 = charParser{id: 184, chars: []rune{119}}
	var p115 = charParser{id: 115, chars: []rune{115}}
	p31.items = []parser{&p184, &p115}
	var p83 = sequenceParser{id: 83, commit: 72, name: "nows", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{39}}
	var p148 = charParser{id: 148, chars: []rune{110}}
	var p38 = charParser{id: 38, chars: []rune{111}}
	var p109 = charParser{id: 109, chars: []rune{119}}
	var p178 = charParser{id: 178, chars: []rune{115}}
	p83.items = []parser{&p148, &p38, &p109, &p178}
	var p163 = sequenceParser{id: 163, commit: 72, name: "failpass", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{39}}
	var p32 = charParser{id: 32, chars: []rune{102}}
	var p74 = charParser{id: 74, chars: []rune{97}}
	var p126 = charParser{id: 126, chars: []rune{105}}
	var p52 = charParser{id: 52, chars: []rune{108}}
	var p5 = charParser{id: 5, chars: []rune{112}}
	var p147 = charParser{id: 147, chars: []rune{97}}
	var p75 = charParser{id: 75, chars: []rune{115}}
	var p6 = charParser{id: 6, chars: []rune{115}}
	p163.items = []parser{&p32, &p74, &p126, &p52, &p5, &p147, &p75, &p6}
	var p70 = sequenceParser{id: 70, commit: 72, name: "root", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{39}}
	var p177 = charParser{id: 177, chars: []rune{114}}
	var p16 = charParser{id: 16, chars: []rune{111}}
	var p170 = charParser{id: 170, chars: []rune{111}}
	var p79 = charParser{id: 79, chars: []rune{116}}
	p70.items = []parser{&p177, &p16, &p170, &p79}
	p39.options = []parser{&p96, &p31, &p83, &p163, &p70}
	p103.items = []parser{&p26, &p39}
	p138.items = []parser{&p11, &p103}
	var p171 = sequenceParser{id: 171, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p132 = charParser{id: 132, chars: []rune{61}}
	p171.items = []parser{&p132}
	var p154 = choiceParser{id: 154, commit: 66, name: "expression"}
	var p29 = choiceParser{id: 29, commit: 66, name: "terminal", generalizations: []int{154, 160, 158}}
	var p44 = sequenceParser{id: 44, commit: 72, name: "any-char", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{29, 154, 160, 158}}
	var p13 = charParser{id: 13, chars: []rune{46}}
	p44.items = []parser{&p13}
	var p153 = sequenceParser{id: 153, commit: 72, name: "char-class", ranges: [][]int{{1, 1}, {0, 1}, {0, -1}, {1, 1}, {1, 1}, {0, 1}, {0, -1}, {1, 1}}, generalizations: []int{29, 154, 160, 158}}
	var p45 = sequenceParser{id: 45, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p112 = charParser{id: 112, chars: []rune{91}}
	p45.items = []parser{&p112}
	var p107 = sequenceParser{id: 107, commit: 72, name: "class-not", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p106 = charParser{id: 106, chars: []rune{94}}
	p107.items = []parser{&p106}
	var p77 = choiceParser{id: 77, commit: 10}
	var p165 = choiceParser{id: 165, commit: 72, name: "class-char", generalizations: []int{77}}
	var p173 = sequenceParser{id: 173, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{165, 77}}
	var p24 = charParser{id: 24, not: true, chars: []rune{92, 91, 93, 94, 45}}
	p173.items = []parser{&p24}
	var p143 = sequenceParser{id: 143, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{165, 77}}
	var p111 = sequenceParser{id: 111, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p180 = charParser{id: 180, chars: []rune{92}}
	p111.items = []parser{&p180}
	var p25 = sequenceParser{id: 25, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p146 = charParser{id: 146, not: true}
	p25.items = []parser{&p146}
	p143.items = []parser{&p111, &p25}
	p165.options = []parser{&p173, &p143}
	var p181 = sequenceParser{id: 181, commit: 72, name: "char-range", ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{77}}
	var p80 = sequenceParser{id: 80, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p156 = charParser{id: 156, chars: []rune{45}}
	p80.items = []parser{&p156}
	p181.items = []parser{&p165, &p80, &p165}
	p77.options = []parser{&p165, &p181}
	var p174 = sequenceParser{id: 174, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p14 = charParser{id: 14, chars: []rune{93}}
	p174.items = []parser{&p14}
	p153.items = []parser{&p45, &p107, &p77, &p174}
	var p59 = sequenceParser{id: 59, commit: 72, name: "char-sequence", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {1, 1}, {0, -1}, {1, 1}}, generalizations: []int{29, 154, 160, 158}}
	var p65 = sequenceParser{id: 65, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p91 = charParser{id: 91, chars: []rune{34}}
	p65.items = []parser{&p91}
	var p90 = choiceParser{id: 90, commit: 72, name: "sequence-char"}
	var p166 = sequenceParser{id: 166, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{90}}
	var p51 = charParser{id: 51, not: true, chars: []rune{92, 34}}
	p166.items = []parser{&p51}
	var p54 = sequenceParser{id: 54, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{90}}
	var p182 = sequenceParser{id: 182, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p64 = charParser{id: 64, chars: []rune{92}}
	p182.items = []parser{&p64}
	var p124 = sequenceParser{id: 124, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p121 = charParser{id: 121, not: true}
	p124.items = []parser{&p121}
	p54.items = []parser{&p182, &p124}
	p90.options = []parser{&p166, &p54}
	var p175 = sequenceParser{id: 175, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p66 = charParser{id: 66, chars: []rune{34}}
	p175.items = []parser{&p66}
	p59.items = []parser{&p65, &p90, &p175}
	p29.options = []parser{&p44, &p153, &p59}
	var p55 = sequenceParser{id: 55, commit: 66, name: "group", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}}, generalizations: []int{154, 160, 158}}
	var p92 = sequenceParser{id: 92, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p167 = charParser{id: 167, chars: []rune{40}}
	p92.items = []parser{&p167}
	var p12 = sequenceParser{id: 12, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p168 = charParser{id: 168, chars: []rune{41}}
	p12.items = []parser{&p168}
	p55.items = []parser{&p92, &p186, &p154, &p186, &p12}
	var p131 = sequenceParser{id: 131, commit: 64, name: "sequence", ranges: [][]int{{1, 1}, {0, -1}}, generalizations: []int{154, 158}}
	var p176 = sequenceParser{id: 176, commit: 72, name: "item", ranges: [][]int{{1, 1}, {0, 1}, {1, 1}, {0, 1}}}
	var p160 = choiceParser{id: 160, commit: 10}
	p160.options = []parser{&p29, &p11, &p55}
	var p48 = choiceParser{id: 48, commit: 66, name: "quantity"}
	var p36 = sequenceParser{id: 36, commit: 64, name: "count-quantifier", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}}, generalizations: []int{48}}
	var p81 = sequenceParser{id: 81, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p169 = charParser{id: 169, chars: []rune{123}}
	p81.items = []parser{&p169}
	var p129 = sequenceParser{id: 129, commit: 64, name: "count", ranges: [][]int{{1, 1}}}
	var p60 = sequenceParser{id: 60, commit: 74, name: "number", ranges: [][]int{{1, -1}, {1, -1}}}
	var p113 = sequenceParser{id: 113, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p128 = charParser{id: 128, ranges: [][]rune{{48, 57}}}
	p113.items = []parser{&p128}
	p60.items = []parser{&p113}
	p129.items = []parser{&p60}
	var p15 = sequenceParser{id: 15, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p144 = charParser{id: 144, chars: []rune{125}}
	p15.items = []parser{&p144}
	p36.items = []parser{&p81, &p186, &p129, &p186, &p15}
	var p56 = sequenceParser{id: 56, commit: 64, name: "range-quantifier", ranges: [][]int{{1, 1}, {0, -1}, {0, 1}, {0, -1}, {1, 1}, {0, -1}, {0, 1}, {0, -1}, {1, 1}}, generalizations: []int{48}}
	var p122 = sequenceParser{id: 122, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p82 = charParser{id: 82, chars: []rune{123}}
	p122.items = []parser{&p82}
	var p67 = sequenceParser{id: 67, commit: 64, name: "range-from", ranges: [][]int{{1, 1}}}
	p67.items = []parser{&p60}
	var p68 = sequenceParser{id: 68, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p125 = charParser{id: 125, chars: []rune{44}}
	p68.items = []parser{&p125}
	var p157 = sequenceParser{id: 157, commit: 64, name: "range-to", ranges: [][]int{{1, 1}}}
	p157.items = []parser{&p60}
	var p135 = sequenceParser{id: 135, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p134 = charParser{id: 134, chars: []rune{125}}
	p135.items = []parser{&p134}
	p56.items = []parser{&p122, &p186, &p67, &p186, &p68, &p186, &p157, &p186, &p135}
	var p78 = sequenceParser{id: 78, commit: 72, name: "one-or-more", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{48}}
	var p73 = charParser{id: 73, chars: []rune{43}}
	p78.items = []parser{&p73}
	var p57 = sequenceParser{id: 57, commit: 72, name: "zero-or-more", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{48}}
	var p30 = charParser{id: 30, chars: []rune{42}}
	p57.items = []parser{&p30}
	var p136 = sequenceParser{id: 136, commit: 72, name: "zero-or-one", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{48}}
	var p108 = charParser{id: 108, chars: []rune{63}}
	p136.items = []parser{&p108}
	p48.options = []parser{&p36, &p56, &p78, &p57, &p136}
	p176.items = []parser{&p160, &p48}
	var p130 = sequenceParser{id: 130, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p130.items = []parser{&p186, &p176}
	p131.items = []parser{&p176, &p130}
	var p94 = sequenceParser{id: 94, commit: 64, name: "choice", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}, generalizations: []int{154}}
	var p158 = choiceParser{id: 158, commit: 66, name: "option"}
	p158.options = []parser{&p29, &p11, &p55, &p131}
	var p183 = sequenceParser{id: 183, commit: 2, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}}}
	var p18 = sequenceParser{id: 18, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p37 = charParser{id: 37, chars: []rune{124}}
	p18.items = []parser{&p37}
	p183.items = []parser{&p18, &p186, &p158}
	var p93 = sequenceParser{id: 93, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p93.items = []parser{&p186, &p183}
	p94.items = []parser{&p158, &p186, &p183, &p93}
	p154.options = []parser{&p29, &p11, &p55, &p131, &p94}
	p4.items = []parser{&p138, &p186, &p171, &p186, &p154}
	var p41 = sequenceParser{id: 41, commit: 2, ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var p118 = sequenceParser{id: 118, commit: 2, ranges: [][]int{{1, 1}, {0, -1}, {0, -1}, {1, 1}}}
	var p17 = sequenceParser{id: 17, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p116 = charParser{id: 116, chars: []rune{59}}
	p17.items = []parser{&p116}
	var p117 = sequenceParser{id: 117, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p117.items = []parser{&p186, &p17}
	p118.items = []parser{&p17, &p117, &p186, &p4}
	var p40 = sequenceParser{id: 40, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p40.items = []parser{&p186, &p118}
	p41.items = []parser{&p186, &p118, &p40}
	p42.items = []parser{&p4, &p41}
	var p87 = sequenceParser{id: 87, commit: 2, ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var p104 = sequenceParser{id: 104, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p123 = charParser{id: 123, chars: []rune{59}}
	p104.items = []parser{&p123}
	var p86 = sequenceParser{id: 86, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p86.items = []parser{&p186, &p104}
	p87.items = []parser{&p186, &p104, &p86}
	p187.items = []parser{&p85, &p186, &p42, &p87}
	p188.items = []parser{&p186, &p187, &p186}
	var b188 = sequenceBuilder{id: 188, commit: 32, name: "syntax", ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var b186 = choiceBuilder{id: 186, commit: 2}
	var b185 = choiceBuilder{id: 185, commit: 70}
	var b155 = choiceBuilder{id: 155, commit: 66}
	var b97 = sequenceBuilder{id: 97, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b61 = charBuilder{}
	b97.items = []builder{&b61}
	var b179 = sequenceBuilder{id: 179, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b1 = charBuilder{}
	b179.items = []builder{&b1}
	var b2 = sequenceBuilder{id: 2, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b119 = charBuilder{}
	b2.items = []builder{&b119}
	var b120 = sequenceBuilder{id: 120, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b98 = charBuilder{}
	b120.items = []builder{&b98}
	var b105 = sequenceBuilder{id: 105, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b71 = charBuilder{}
	b105.items = []builder{&b71}
	var b149 = sequenceBuilder{id: 149, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b139 = charBuilder{}
	b149.items = []builder{&b139}
	var b99 = sequenceBuilder{id: 99, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b20 = charBuilder{}
	b99.items = []builder{&b20}
	b155.options = []builder{&b97, &b179, &b2, &b120, &b105, &b149, &b99}
	var b63 = sequenceBuilder{id: 63, commit: 72, name: "comment", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}}
	var b33 = choiceBuilder{id: 33, commit: 74}
	var b101 = sequenceBuilder{id: 101, commit: 74, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}}
	var b9 = sequenceBuilder{id: 9, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b88 = charBuilder{}
	var b164 = charBuilder{}
	b9.items = []builder{&b88, &b164}
	var b89 = sequenceBuilder{id: 89, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b62 = charBuilder{}
	b89.items = []builder{&b62}
	b101.items = []builder{&b9, &b89}
	var b72 = sequenceBuilder{id: 72, commit: 74, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {1, 1}, {0, -1}, {1, 1}}}
	var b21 = sequenceBuilder{id: 21, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b172 = charBuilder{}
	var b140 = charBuilder{}
	b21.items = []builder{&b172, &b140}
	var b27 = choiceBuilder{id: 27, commit: 10}
	var b8 = sequenceBuilder{id: 8, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b7 = sequenceBuilder{id: 7, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b141 = charBuilder{}
	b7.items = []builder{&b141}
	var b22 = sequenceBuilder{id: 22, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b3 = charBuilder{}
	b22.items = []builder{&b3}
	b8.items = []builder{&b7, &b22}
	var b100 = sequenceBuilder{id: 100, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b76 = charBuilder{}
	b100.items = []builder{&b76}
	b27.options = []builder{&b8, &b100}
	var b49 = sequenceBuilder{id: 49, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b127 = charBuilder{}
	var b150 = charBuilder{}
	b49.items = []builder{&b127, &b150}
	b72.items = []builder{&b21, &b27, &b49}
	b33.options = []builder{&b101, &b72}
	var b133 = sequenceBuilder{id: 133, commit: 10, ranges: [][]int{{0, -1}, {0, 1}, {0, -1}, {1, 1}, {0, -1}, {0, 1}, {0, -1}, {1, 1}}}
	var b53 = choiceBuilder{id: 53, commit: 74}
	var b10 = sequenceBuilder{id: 10, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b34 = charBuilder{}
	b10.items = []builder{&b34}
	var b151 = sequenceBuilder{id: 151, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b110 = charBuilder{}
	b151.items = []builder{&b110}
	var b142 = sequenceBuilder{id: 142, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b102 = charBuilder{}
	b142.items = []builder{&b102}
	var b152 = sequenceBuilder{id: 152, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b35 = charBuilder{}
	b152.items = []builder{&b35}
	var b28 = sequenceBuilder{id: 28, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b50 = charBuilder{}
	b28.items = []builder{&b50}
	var b159 = sequenceBuilder{id: 159, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b23 = charBuilder{}
	b159.items = []builder{&b23}
	b53.options = []builder{&b10, &b151, &b142, &b152, &b28, &b159}
	var b43 = sequenceBuilder{id: 43, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b95 = charBuilder{}
	b43.items = []builder{&b95}
	b133.items = []builder{&b53, &b43, &b53, &b33}
	b63.items = []builder{&b33, &b133}
	b185.options = []builder{&b155, &b63}
	b186.options = []builder{&b185}
	var b187 = sequenceBuilder{id: 187, commit: 66, ranges: [][]int{{0, 1}, {0, -1}, {0, 1}, {0, 1}}}
	var b85 = sequenceBuilder{id: 85, commit: 2, ranges: [][]int{{1, 1}, {0, -1}}}
	var b145 = sequenceBuilder{id: 145, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b58 = charBuilder{}
	b145.items = []builder{&b58}
	var b84 = sequenceBuilder{id: 84, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b84.items = []builder{&b186, &b145}
	b85.items = []builder{&b145, &b84}
	var b42 = sequenceBuilder{id: 42, commit: 66, ranges: [][]int{{1, 1}, {0, 1}}}
	var b4 = sequenceBuilder{id: 4, commit: 64, name: "definition", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}}}
	var b138 = sequenceBuilder{id: 138, commit: 74, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}}
	var b11 = sequenceBuilder{id: 11, commit: 72, name: "symbol", ranges: [][]int{{1, -1}, {1, -1}}}
	var b47 = sequenceBuilder{id: 47, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b46 = charBuilder{}
	b47.items = []builder{&b46}
	b11.items = []builder{&b47}
	var b103 = sequenceBuilder{id: 103, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b26 = sequenceBuilder{id: 26, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b19 = charBuilder{}
	b26.items = []builder{&b19}
	var b39 = choiceBuilder{id: 39, commit: 66}
	var b96 = sequenceBuilder{id: 96, commit: 72, name: "alias", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b69 = charBuilder{}
	var b114 = charBuilder{}
	var b162 = charBuilder{}
	var b161 = charBuilder{}
	var b137 = charBuilder{}
	b96.items = []builder{&b69, &b114, &b162, &b161, &b137}
	var b31 = sequenceBuilder{id: 31, commit: 72, name: "ws", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b184 = charBuilder{}
	var b115 = charBuilder{}
	b31.items = []builder{&b184, &b115}
	var b83 = sequenceBuilder{id: 83, commit: 72, name: "nows", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b148 = charBuilder{}
	var b38 = charBuilder{}
	var b109 = charBuilder{}
	var b178 = charBuilder{}
	b83.items = []builder{&b148, &b38, &b109, &b178}
	var b163 = sequenceBuilder{id: 163, commit: 72, name: "failpass", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b32 = charBuilder{}
	var b74 = charBuilder{}
	var b126 = charBuilder{}
	var b52 = charBuilder{}
	var b5 = charBuilder{}
	var b147 = charBuilder{}
	var b75 = charBuilder{}
	var b6 = charBuilder{}
	b163.items = []builder{&b32, &b74, &b126, &b52, &b5, &b147, &b75, &b6}
	var b70 = sequenceBuilder{id: 70, commit: 72, name: "root", allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b177 = charBuilder{}
	var b16 = charBuilder{}
	var b170 = charBuilder{}
	var b79 = charBuilder{}
	b70.items = []builder{&b177, &b16, &b170, &b79}
	b39.options = []builder{&b96, &b31, &b83, &b163, &b70}
	b103.items = []builder{&b26, &b39}
	b138.items = []builder{&b11, &b103}
	var b171 = sequenceBuilder{id: 171, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b132 = charBuilder{}
	b171.items = []builder{&b132}
	var b154 = choiceBuilder{id: 154, commit: 66}
	var b29 = choiceBuilder{id: 29, commit: 66}
	var b44 = sequenceBuilder{id: 44, commit: 72, name: "any-char", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b13 = charBuilder{}
	b44.items = []builder{&b13}
	var b153 = sequenceBuilder{id: 153, commit: 72, name: "char-class", ranges: [][]int{{1, 1}, {0, 1}, {0, -1}, {1, 1}, {1, 1}, {0, 1}, {0, -1}, {1, 1}}}
	var b45 = sequenceBuilder{id: 45, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b112 = charBuilder{}
	b45.items = []builder{&b112}
	var b107 = sequenceBuilder{id: 107, commit: 72, name: "class-not", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b106 = charBuilder{}
	b107.items = []builder{&b106}
	var b77 = choiceBuilder{id: 77, commit: 10}
	var b165 = choiceBuilder{id: 165, commit: 72, name: "class-char"}
	var b173 = sequenceBuilder{id: 173, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b24 = charBuilder{}
	b173.items = []builder{&b24}
	var b143 = sequenceBuilder{id: 143, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b111 = sequenceBuilder{id: 111, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b180 = charBuilder{}
	b111.items = []builder{&b180}
	var b25 = sequenceBuilder{id: 25, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b146 = charBuilder{}
	b25.items = []builder{&b146}
	b143.items = []builder{&b111, &b25}
	b165.options = []builder{&b173, &b143}
	var b181 = sequenceBuilder{id: 181, commit: 72, name: "char-range", ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b80 = sequenceBuilder{id: 80, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b156 = charBuilder{}
	b80.items = []builder{&b156}
	b181.items = []builder{&b165, &b80, &b165}
	b77.options = []builder{&b165, &b181}
	var b174 = sequenceBuilder{id: 174, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b14 = charBuilder{}
	b174.items = []builder{&b14}
	b153.items = []builder{&b45, &b107, &b77, &b174}
	var b59 = sequenceBuilder{id: 59, commit: 72, name: "char-sequence", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {1, 1}, {0, -1}, {1, 1}}}
	var b65 = sequenceBuilder{id: 65, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b91 = charBuilder{}
	b65.items = []builder{&b91}
	var b90 = choiceBuilder{id: 90, commit: 72, name: "sequence-char"}
	var b166 = sequenceBuilder{id: 166, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b51 = charBuilder{}
	b166.items = []builder{&b51}
	var b54 = sequenceBuilder{id: 54, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b182 = sequenceBuilder{id: 182, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b64 = charBuilder{}
	b182.items = []builder{&b64}
	var b124 = sequenceBuilder{id: 124, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b121 = charBuilder{}
	b124.items = []builder{&b121}
	b54.items = []builder{&b182, &b124}
	b90.options = []builder{&b166, &b54}
	var b175 = sequenceBuilder{id: 175, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b66 = charBuilder{}
	b175.items = []builder{&b66}
	b59.items = []builder{&b65, &b90, &b175}
	b29.options = []builder{&b44, &b153, &b59}
	var b55 = sequenceBuilder{id: 55, commit: 66, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}}}
	var b92 = sequenceBuilder{id: 92, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b167 = charBuilder{}
	b92.items = []builder{&b167}
	var b12 = sequenceBuilder{id: 12, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b168 = charBuilder{}
	b12.items = []builder{&b168}
	b55.items = []builder{&b92, &b186, &b154, &b186, &b12}
	var b131 = sequenceBuilder{id: 131, commit: 64, name: "sequence", ranges: [][]int{{1, 1}, {0, -1}}}
	var b176 = sequenceBuilder{id: 176, commit: 72, name: "item", ranges: [][]int{{1, 1}, {0, 1}, {1, 1}, {0, 1}}}
	var b160 = choiceBuilder{id: 160, commit: 10}
	b160.options = []builder{&b29, &b11, &b55}
	var b48 = choiceBuilder{id: 48, commit: 66}
	var b36 = sequenceBuilder{id: 36, commit: 64, name: "count-quantifier", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}}}
	var b81 = sequenceBuilder{id: 81, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b169 = charBuilder{}
	b81.items = []builder{&b169}
	var b129 = sequenceBuilder{id: 129, commit: 64, name: "count", ranges: [][]int{{1, 1}}}
	var b60 = sequenceBuilder{id: 60, commit: 74, ranges: [][]int{{1, -1}, {1, -1}}}
	var b113 = sequenceBuilder{id: 113, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b128 = charBuilder{}
	b113.items = []builder{&b128}
	b60.items = []builder{&b113}
	b129.items = []builder{&b60}
	var b15 = sequenceBuilder{id: 15, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b144 = charBuilder{}
	b15.items = []builder{&b144}
	b36.items = []builder{&b81, &b186, &b129, &b186, &b15}
	var b56 = sequenceBuilder{id: 56, commit: 64, name: "range-quantifier", ranges: [][]int{{1, 1}, {0, -1}, {0, 1}, {0, -1}, {1, 1}, {0, -1}, {0, 1}, {0, -1}, {1, 1}}}
	var b122 = sequenceBuilder{id: 122, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b82 = charBuilder{}
	b122.items = []builder{&b82}
	var b67 = sequenceBuilder{id: 67, commit: 64, name: "range-from", ranges: [][]int{{1, 1}}}
	b67.items = []builder{&b60}
	var b68 = sequenceBuilder{id: 68, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b125 = charBuilder{}
	b68.items = []builder{&b125}
	var b157 = sequenceBuilder{id: 157, commit: 64, name: "range-to", ranges: [][]int{{1, 1}}}
	b157.items = []builder{&b60}
	var b135 = sequenceBuilder{id: 135, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b134 = charBuilder{}
	b135.items = []builder{&b134}
	b56.items = []builder{&b122, &b186, &b67, &b186, &b68, &b186, &b157, &b186, &b135}
	var b78 = sequenceBuilder{id: 78, commit: 72, name: "one-or-more", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b73 = charBuilder{}
	b78.items = []builder{&b73}
	var b57 = sequenceBuilder{id: 57, commit: 72, name: "zero-or-more", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b30 = charBuilder{}
	b57.items = []builder{&b30}
	var b136 = sequenceBuilder{id: 136, commit: 72, name: "zero-or-one", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b108 = charBuilder{}
	b136.items = []builder{&b108}
	b48.options = []builder{&b36, &b56, &b78, &b57, &b136}
	b176.items = []builder{&b160, &b48}
	var b130 = sequenceBuilder{id: 130, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b130.items = []builder{&b186, &b176}
	b131.items = []builder{&b176, &b130}
	var b94 = sequenceBuilder{id: 94, commit: 64, name: "choice", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}}
	var b158 = choiceBuilder{id: 158, commit: 66}
	b158.options = []builder{&b29, &b11, &b55, &b131}
	var b183 = sequenceBuilder{id: 183, commit: 2, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}}}
	var b18 = sequenceBuilder{id: 18, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b37 = charBuilder{}
	b18.items = []builder{&b37}
	b183.items = []builder{&b18, &b186, &b158}
	var b93 = sequenceBuilder{id: 93, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b93.items = []builder{&b186, &b183}
	b94.items = []builder{&b158, &b186, &b183, &b93}
	b154.options = []builder{&b29, &b11, &b55, &b131, &b94}
	b4.items = []builder{&b138, &b186, &b171, &b186, &b154}
	var b41 = sequenceBuilder{id: 41, commit: 2, ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var b118 = sequenceBuilder{id: 118, commit: 2, ranges: [][]int{{1, 1}, {0, -1}, {0, -1}, {1, 1}}}
	var b17 = sequenceBuilder{id: 17, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b116 = charBuilder{}
	b17.items = []builder{&b116}
	var b117 = sequenceBuilder{id: 117, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b117.items = []builder{&b186, &b17}
	b118.items = []builder{&b17, &b117, &b186, &b4}
	var b40 = sequenceBuilder{id: 40, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b40.items = []builder{&b186, &b118}
	b41.items = []builder{&b186, &b118, &b40}
	b42.items = []builder{&b4, &b41}
	var b87 = sequenceBuilder{id: 87, commit: 2, ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var b104 = sequenceBuilder{id: 104, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b123 = charBuilder{}
	b104.items = []builder{&b123}
	var b86 = sequenceBuilder{id: 86, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b86.items = []builder{&b186, &b104}
	b87.items = []builder{&b186, &b104, &b86}
	b187.items = []builder{&b85, &b186, &b42, &b87}
	b188.items = []builder{&b186, &b187, &b186}

	return parseInput(r, &p188, &b188)
}
