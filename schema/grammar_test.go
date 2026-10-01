// Copyright (c) 2026 Michael D Henderson. All rights reserved.

// Package schema holds the wxx reference grammars for Worldographer files. It
// has no Go API: this test keeps 1.06.rnc honest by checking every tracked 2.08
// fixture against it, fixture -> grammar. A fixture element or attribute that
// the grammar does not allow at that position fails the test, and so does
// non-whitespace text inside an element whose pattern has no `text`. Grammar
// rules no fixture exercises do not.
//
// The test parses a small subset of RELAX NG compact syntax, documented in the
// header of 1.06.rnc, and fails loudly on anything outside it rather than
// skipping it.
package schema

import (
	"bytes"
	"compress/gzip"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

const (
	// grammarPath is the grammar under test.
	grammarPath = "1.06.rnc"
	// fixtureGlob selects the fixtures checked against it. The grammar is
	// checked against 2.08 saves only.
	fixtureGlob = "../testdata/2025-2.08-*.wxx"
	// minFixtures is the number of 2.08 fixtures tracked when the grammar was
	// written. Fewer matches means the glob or the tree is wrong.
	minFixtures = 6
	// minElements and minAttributes guard against a vacuous pass: the
	// smallest tracked fixture visits 97 elements and 686 attributes.
	minElements   = 50
	minAttributes = 300
)

// TestGrammarCoversFixtures walks every 2.08 fixture and reports each element
// or attribute that the grammar does not allow where the fixture has it.
func TestGrammarCoversFixtures(t *testing.T) {
	g := loadGrammar(t, grammarPath)
	checkFixtures(t, g, fixtureGlob)
}

func checkFixtures(t *testing.T, g *grammar, glob string) {
	t.Helper()
	paths, err := filepath.Glob(glob)
	if err != nil {
		t.Fatalf("glob %q: %v", glob, err)
	}
	if len(paths) < minFixtures {
		t.Fatalf("glob %q matched %d fixtures, want at least %d: the check would pass vacuously", glob, len(paths), minFixtures)
	}
	for _, path := range paths {
		name := filepath.Base(path)
		data, err := readWXX(path)
		if err != nil {
			t.Errorf("%s: %v", name, err)
			continue
		}
		res, err := g.check(data)
		if err != nil {
			t.Errorf("%s: %v", name, err)
			continue
		}
		t.Logf("%s: %d elements, %d attributes checked", name, res.elements, res.attributes)
		if res.elements < minElements || res.attributes < minAttributes {
			t.Errorf("%s: visited %d elements and %d attributes, want at least %d and %d: the check would pass vacuously",
				name, res.elements, res.attributes, minElements, minAttributes)
		}
		for _, v := range res.violations {
			t.Errorf("%s: grammar does not allow %s in pattern %s (first at %s; %d occurrence(s))",
				name, v.item, v.patternPath, v.xmlPath, v.count)
		}
	}
}

// readWXX gunzips a .wxx, decodes its UTF-16 to UTF-8 and strips the XML
// declaration, which names version 1.1 and an encoding the bytes no longer
// have. Nothing else is removed.
func readWXX(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return nil, fmt.Errorf("gunzip: %w", err)
	}
	utf16 := unicode.UTF16(unicode.BigEndian, unicode.UseBOM).NewDecoder()
	data, err := io.ReadAll(transform.NewReader(gz, utf16))
	if err != nil {
		return nil, fmt.Errorf("utf-16: %w", err)
	}
	data = bytes.TrimPrefix(data, []byte("\xef\xbb\xbf"))
	if bytes.HasPrefix(data, []byte("<?xml")) {
		end := bytes.Index(data, []byte("?>"))
		if end < 0 {
			return nil, fmt.Errorf("unterminated xml declaration")
		}
		data = data[end+2:]
	}
	return data, nil
}

// ---------------------------------------------------------------------------
// The grammar graph.

// elemNode is one `element` in the grammar: the attributes it allows, the
// child elements its content allows, by name, and whether its content allows
// text.
type elemNode struct {
	name     string
	label    string // the named pattern it is the body of, for messages
	attrs    map[string]bool
	children map[string]*elemNode
	text     bool // the content holds a `text` pattern
	content  *pnode
}

type grammar struct {
	defs  map[string]*pnode
	start *pnode
	roots map[string]*elemNode // the elements start allows, by name
	elems []*elemNode
}

type violation struct {
	patternPath, item, xmlPath string
	count                      int
}

type result struct {
	elements, attributes int
	violations           []*violation
}

// check walks one document from its root, tracking the grammar element each
// XML element matched.
func (g *grammar) check(data []byte) (*result, error) {
	type frame struct {
		node        *elemNode // nil below an element the grammar does not allow
		patternPath string
		xmlPath     string
		seen        map[string]int
		textSeen    bool // non-whitespace text already reported or allowed
	}
	res := &result{}
	byKey := map[string]*violation{}
	report := func(patternPath, item, xmlPath string) {
		key := patternPath + "\x00" + item
		if v, ok := byKey[key]; ok {
			v.count++
			return
		}
		v := &violation{patternPath: patternPath, item: item, xmlPath: xmlPath, count: 1}
		byKey[key] = v
		res.violations = append(res.violations, v)
	}
	var stack []*frame
	dec := xml.NewDecoder(bytes.NewReader(data))
	dec.Strict = true
	sawRoot := false
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("xml: %w", err)
		}
		switch tok := tok.(type) {
		case xml.StartElement:
			res.elements++
			name := qname(tok.Name)
			var node *elemNode
			var patternPath, xmlPath string
			if len(stack) == 0 {
				if sawRoot {
					return nil, fmt.Errorf("more than one root element")
				}
				sawRoot = true
				xmlPath = "/" + name
				node = g.roots[name]
				if node == nil {
					report("start", "root element "+name, xmlPath)
				} else {
					patternPath = node.label
				}
			} else {
				parent := stack[len(stack)-1]
				parent.seen[name]++
				xmlPath = fmt.Sprintf("%s/%s[%d]", parent.xmlPath, name, parent.seen[name])
				if parent.node != nil {
					node = parent.node.children[name]
					if node == nil {
						report(parent.patternPath, "element "+name, xmlPath)
					} else {
						patternPath = parent.patternPath + "/" + node.label
					}
				}
			}
			for _, a := range tok.Attr {
				res.attributes++
				if node != nil && !node.attrs[qname(a.Name)] {
					report(patternPath, "attribute "+qname(a.Name), xmlPath)
				}
			}
			stack = append(stack, &frame{node: node, patternPath: patternPath, xmlPath: xmlPath, seen: map[string]int{}})
		case xml.EndElement:
			stack = stack[:len(stack)-1]
		case xml.CharData:
			// Text, CDATA sections included. Whitespace-only text is
			// allowed anywhere, as in RELAX NG; anything else needs a
			// `text` in the element's pattern. Reported once per element.
			if len(stack) == 0 || isXMLSpace(tok) {
				break
			}
			top := stack[len(stack)-1]
			if top.node != nil && !top.textSeen && !top.node.text {
				report(top.patternPath, "text", top.xmlPath)
			}
			top.textSeen = true
		}
		// Comment, ProcInst and Directive carry no element, attribute or
		// text, so they are not checked.
	}
	if !sawRoot {
		return nil, fmt.Errorf("no root element")
	}
	return res, nil
}

// isXMLSpace reports whether b is empty or holds only XML whitespace (space,
// tab, CR, LF).
func isXMLSpace(b []byte) bool {
	return len(bytes.Trim(b, " \t\r\n")) == 0
}

// qname spells a name with its namespace, so a namespaced element or
// attribute never matches a grammar name that lacks one.
func qname(n xml.Name) string {
	if n.Space != "" {
		return n.Space + ":" + n.Local
	}
	return n.Local
}

// ---------------------------------------------------------------------------
// The pattern tree.

type pkind int

const (
	pElement pkind = iota
	pAttribute
	pText
	pEmpty
	pRef
	pGroup  // a, b
	pChoice // a | b
	pOpt    // a?
	pStar   // a*
	pPlus   // a+
)

type pnode struct {
	kind pkind
	name string   // element/attribute name, or the referenced pattern
	kids []*pnode // group/choice members, quantifier operand, element content
	elem *elemNode
	bare bool // a group written without parentheses: may not sit beside "|"
}

// loadGrammar parses the grammar file and builds the element graph.
func loadGrammar(t *testing.T, path string) *grammar {
	t.Helper()
	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("%s: %v", path, err)
	}
	g, err := parseGrammar(string(src))
	if err != nil {
		t.Fatalf("%s: %v", path, err)
	}
	return g
}

// parseGrammar parses the subset and builds the graph. Any construct outside
// the subset is an error.
func parseGrammar(src string) (*grammar, error) {
	toks, err := lex(src)
	if err != nil {
		return nil, err
	}
	p := &parser{toks: toks}
	g := &grammar{defs: map[string]*pnode{}}
	var order []string
	for !p.atEOF() {
		name, err := p.ident()
		if err != nil {
			return nil, err
		}
		if eq := p.next(); eq.text != "=" {
			return nil, p.errAt(eq, "want \"=\" after %q (combination with |= or &= is not supported)", name)
		}
		pat, err := p.pattern()
		if err != nil {
			return nil, err
		}
		if name == "start" {
			if g.start != nil {
				return nil, fmt.Errorf("start defined twice")
			}
			g.start = pat
			continue
		}
		if _, dup := g.defs[name]; dup {
			return nil, fmt.Errorf("pattern %s defined twice", name)
		}
		g.defs[name] = pat
		order = append(order, name)
		if pat.kind == pElement {
			pat.elem.label = name
		}
	}
	if g.start == nil {
		return nil, fmt.Errorf("no start pattern")
	}

	// Every reference resolves, and every named pattern is reachable.
	reached := map[string]bool{}
	var reach func(n *pnode) error
	reach = func(n *pnode) error {
		if n.kind == pRef {
			if _, ok := g.defs[n.name]; !ok {
				return fmt.Errorf("reference to undefined pattern %s", n.name)
			}
			if reached[n.name] {
				return nil
			}
			reached[n.name] = true
			return reach(g.defs[n.name])
		}
		for _, k := range n.kids {
			if err := reach(k); err != nil {
				return err
			}
		}
		return nil
	}
	if err := reach(g.start); err != nil {
		return nil, err
	}
	for _, name := range order {
		if !reached[name] {
			return nil, fmt.Errorf("pattern %s is never referenced from start", name)
		}
	}

	// Collect every element node and resolve its attributes and children.
	var collect func(n *pnode)
	collect = func(n *pnode) {
		if n.kind == pElement {
			g.elems = append(g.elems, n.elem)
		}
		for _, k := range n.kids {
			collect(k)
		}
	}
	collect(g.start)
	for _, name := range order {
		collect(g.defs[name])
	}
	for _, e := range g.elems {
		if e.label == "" {
			e.label = "element " + e.name
		}
		e.attrs = map[string]bool{}
		e.children = map[string]*elemNode{}
		for _, k := range e.content.kids {
			if err := g.content(e, k, map[string]bool{}); err != nil {
				return nil, fmt.Errorf("pattern %s: %w", e.label, err)
			}
		}
	}

	// start must allow elements only.
	root := &elemNode{label: "start", attrs: map[string]bool{}, children: map[string]*elemNode{}}
	if err := g.content(root, g.start, map[string]bool{}); err != nil {
		return nil, fmt.Errorf("start: %w", err)
	}
	if len(root.attrs) > 0 {
		return nil, fmt.Errorf("start allows an attribute")
	}
	if len(root.children) == 0 {
		return nil, fmt.Errorf("start allows no element")
	}
	g.roots = root.children

	// The drop check: every element and attribute keyword in the source
	// became a node. A parser that silently skipped a rule fails here.
	var nElem, nAttr int
	var count func(n *pnode)
	seen := map[*pnode]bool{}
	count = func(n *pnode) {
		if seen[n] {
			return
		}
		seen[n] = true
		switch n.kind {
		case pElement:
			nElem++
		case pAttribute:
			nAttr++
		}
		for _, k := range n.kids {
			count(k)
		}
	}
	count(g.start)
	for _, name := range order {
		count(g.defs[name])
	}
	if want := keywordCount(toks, "element"); nElem != want {
		return nil, fmt.Errorf("parsed %d element patterns, source has %d", nElem, want)
	}
	if want := keywordCount(toks, "attribute"); nAttr != want {
		return nil, fmt.Errorf("parsed %d attribute patterns, source has %d", nAttr, want)
	}
	return g, nil
}

// content adds what pattern n allows inside element e: attributes to e.attrs
// and child elements to e.children. It follows references but stops at a
// nested element, whose own content is its own.
func (g *grammar) content(e *elemNode, n *pnode, refs map[string]bool) error {
	switch n.kind {
	case pAttribute:
		e.attrs[n.name] = true
	case pElement:
		if prev, ok := e.children[n.name]; ok && prev != n.elem {
			return fmt.Errorf("two different patterns (%s, %s) for child element %s: the check cannot tell them apart", prev.label, n.elem.label, n.name)
		}
		e.children[n.name] = n.elem
	case pText:
		e.text = true
	case pEmpty:
	case pRef:
		if refs[n.name] {
			return fmt.Errorf("pattern %s refers to itself without an element in between", n.name)
		}
		refs[n.name] = true
		defer delete(refs, n.name)
		return g.content(e, g.defs[n.name], refs)
	case pGroup, pChoice, pOpt, pStar, pPlus:
		for _, k := range n.kids {
			if err := g.content(e, k, refs); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("unknown pattern kind %d", n.kind)
	}
	return nil
}

// ---------------------------------------------------------------------------
// The lexer.

type tokKind int

const (
	tIdent   tokKind = iota // a name or keyword; escaped is set for \name
	tCName                  // prefix:local, e.g. xsd:string
	tLiteral                // "..." or '...'
	tPunct                  // = { } ( ) , | ? * +
	tEOF
)

type token struct {
	kind    tokKind
	text    string
	escaped bool
	line    int
}

// keywords are RELAX NG compact keywords. The subset uses only element,
// attribute, text, empty and start; any other unescaped keyword is an error.
var keywords = map[string]bool{
	"attribute": true, "default": true, "datatypes": true, "div": true, "element": true,
	"empty": true, "external": true, "grammar": true, "include": true, "inherit": true,
	"list": true, "mixed": true, "namespace": true, "notAllowed": true, "parent": true,
	"start": true, "string": true, "text": true, "token": true,
}

// xsdTypes are the datatypes the subset accepts.
var xsdTypes = map[string]bool{
	"string": true, "token": true, "boolean": true, "integer": true, "int": true,
	"long": true, "decimal": true, "double": true, "float": true, "NCName": true,
	"NMTOKEN": true,
}

var nameRE = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_.\-]*`)

func lex(src string) ([]token, error) {
	var toks []token
	line := 1
	for i := 0; i < len(src); {
		c := src[i]
		switch {
		case c == '\n':
			line++
			i++
		case c == ' ' || c == '\t' || c == '\r':
			i++
		case c == '#':
			if strings.HasPrefix(src[i:], "##") {
				return nil, fmt.Errorf("line %d: \"##\" documentation comments are not supported", line)
			}
			for i < len(src) && src[i] != '\n' {
				i++
			}
		case strings.ContainsRune("={}(),|?*+", rune(c)):
			if c == '|' && i+1 < len(src) && src[i+1] == '=' {
				return nil, fmt.Errorf("line %d: \"|=\" is not supported", line)
			}
			toks = append(toks, token{kind: tPunct, text: string(c), line: line})
			i++
		case c == '"' || c == '\'':
			end := strings.IndexByte(src[i+1:], c)
			if end < 0 || strings.ContainsRune(src[i+1:i+1+end], '\n') {
				return nil, fmt.Errorf("line %d: unterminated string literal", line)
			}
			toks = append(toks, token{kind: tLiteral, text: src[i+1 : i+1+end], line: line})
			i += end + 2
		case c == '\\' || nameRE.MatchString(src[i:]):
			escaped := c == '\\'
			if escaped {
				i++
			}
			name := nameRE.FindString(src[i:])
			if name == "" {
				return nil, fmt.Errorf("line %d: \"\\\" not followed by a name", line)
			}
			i += len(name)
			if i < len(src) && src[i] == ':' {
				if escaped {
					return nil, fmt.Errorf("line %d: escaped prefixed name", line)
				}
				local := nameRE.FindString(src[i+1:])
				if local == "" {
					return nil, fmt.Errorf("line %d: %s: namespace wildcards are not supported", line, name)
				}
				i += 1 + len(local)
				toks = append(toks, token{kind: tCName, text: name + ":" + local, line: line})
				continue
			}
			if !escaped && keywords[name] && !map[string]bool{"element": true, "attribute": true, "text": true, "empty": true, "start": true}[name] {
				return nil, fmt.Errorf("line %d: keyword %q is not supported (escape it as \\%s if it is a name)", line, name, name)
			}
			toks = append(toks, token{kind: tIdent, text: name, escaped: escaped, line: line})
		default:
			return nil, fmt.Errorf("line %d: unsupported character %q", line, c)
		}
	}
	return append(toks, token{kind: tEOF, line: line}), nil
}

// keywordCount counts unescaped occurrences of a keyword token.
func keywordCount(toks []token, kw string) int {
	n := 0
	for _, t := range toks {
		if t.kind == tIdent && !t.escaped && t.text == kw {
			n++
		}
	}
	return n
}

// ---------------------------------------------------------------------------
// The parser.

type parser struct {
	toks []token
	pos  int
}

func (p *parser) peek() token { return p.toks[p.pos] }
func (p *parser) atEOF() bool { return p.peek().kind == tEOF }
func (p *parser) next() token {
	t := p.toks[p.pos]
	if t.kind != tEOF {
		p.pos++
	}
	return t
}

func (p *parser) errAt(t token, format string, args ...any) error {
	return fmt.Errorf("line %d: %s", t.line, fmt.Sprintf(format, args...))
}

func (p *parser) expect(text string) error {
	if t := p.next(); t.kind != tPunct || t.text != text {
		return p.errAt(t, "want %q, got %q", text, t.text)
	}
	return nil
}

// ident reads a pattern name (a definition's left-hand side or a reference).
func (p *parser) ident() (string, error) {
	t := p.next()
	if t.kind != tIdent {
		return "", p.errAt(t, "want a pattern name, got %q", t.text)
	}
	if !t.escaped && keywords[t.text] && t.text != "start" {
		return "", p.errAt(t, "keyword %q is not a pattern name", t.text)
	}
	return t.text, nil
}

// pattern := seq ("|" seq)*   with "," and "|" never mixed at one level.
func (p *parser) pattern() (*pnode, error) {
	first, err := p.seq()
	if err != nil {
		return nil, err
	}
	if t := p.peek(); !(t.kind == tPunct && t.text == "|") {
		return first, nil
	}
	alts := []*pnode{first}
	for p.peek().kind == tPunct && p.peek().text == "|" {
		bar := p.next()
		alt, err := p.seq()
		if err != nil {
			return nil, err
		}
		alts = append(alts, alt)
		for _, a := range alts {
			if a.kind == pGroup && a.bare {
				return nil, p.errAt(bar, "\",\" and \"|\" mixed without parentheses")
			}
		}
	}
	return &pnode{kind: pChoice, kids: alts}, nil
}

func (p *parser) seq() (*pnode, error) {
	first, err := p.unary()
	if err != nil {
		return nil, err
	}
	if t := p.peek(); !(t.kind == tPunct && t.text == ",") {
		return first, nil
	}
	items := []*pnode{first}
	for p.peek().kind == tPunct && p.peek().text == "," {
		p.next()
		item, err := p.unary()
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if t := p.peek(); t.kind == tPunct && t.text == "|" {
		return nil, p.errAt(t, "\",\" and \"|\" mixed without parentheses")
	}
	return &pnode{kind: pGroup, kids: items, bare: true}, nil
}

func (p *parser) unary() (*pnode, error) {
	n, err := p.primary()
	if err != nil {
		return nil, err
	}
	if t := p.peek(); t.kind == tPunct {
		switch t.text {
		case "?":
			p.next()
			return &pnode{kind: pOpt, kids: []*pnode{n}}, nil
		case "*":
			p.next()
			return &pnode{kind: pStar, kids: []*pnode{n}}, nil
		case "+":
			p.next()
			return &pnode{kind: pPlus, kids: []*pnode{n}}, nil
		}
	}
	return n, nil
}

func (p *parser) primary() (*pnode, error) {
	t := p.next()
	switch {
	case t.kind == tPunct && t.text == "(":
		n, err := p.pattern()
		if err != nil {
			return nil, err
		}
		if err := p.expect(")"); err != nil {
			return nil, err
		}
		n.bare = false // parenthesised: may sit beside "|"
		return n, nil
	case t.kind == tIdent && !t.escaped && t.text == "element":
		name, err := p.nodeName("element")
		if err != nil {
			return nil, err
		}
		if err := p.expect("{"); err != nil {
			return nil, err
		}
		body, err := p.pattern()
		if err != nil {
			return nil, err
		}
		if err := p.expect("}"); err != nil {
			return nil, err
		}
		n := &pnode{kind: pElement, name: name, kids: []*pnode{body}}
		n.elem = &elemNode{name: name, content: n}
		return n, nil
	case t.kind == tIdent && !t.escaped && t.text == "attribute":
		name, err := p.nodeName("attribute")
		if err != nil {
			return nil, err
		}
		if err := p.expect("{"); err != nil {
			return nil, err
		}
		if err := p.value(); err != nil {
			return nil, err
		}
		if err := p.expect("}"); err != nil {
			return nil, err
		}
		return &pnode{kind: pAttribute, name: name}, nil
	case t.kind == tIdent && !t.escaped && t.text == "text":
		return &pnode{kind: pText}, nil
	case t.kind == tIdent && !t.escaped && t.text == "empty":
		return &pnode{kind: pEmpty}, nil
	case t.kind == tIdent && (t.escaped || !keywords[t.text]):
		return &pnode{kind: pRef, name: t.text}, nil
	}
	return nil, p.errAt(t, "unsupported construct %q", t.text)
}

// nodeName reads an element or attribute name: a plain or escaped name, no
// namespace prefix, no wildcard.
func (p *parser) nodeName(what string) (string, error) {
	t := p.next()
	if t.kind != tIdent {
		return "", p.errAt(t, "%s name: want a plain name, got %q (namespaces and wildcards are not supported)", what, t.text)
	}
	if !t.escaped && keywords[t.text] {
		return "", p.errAt(t, "%s name %q is a keyword: escape it as \\%s", what, t.text, t.text)
	}
	return t.text, nil
}

// value := atom ("|" atom)*   atom := xsd:type | "literal"
func (p *parser) value() error {
	for {
		t := p.next()
		switch t.kind {
		case tCName:
			prefix, local, _ := strings.Cut(t.text, ":")
			if prefix != "xsd" {
				return p.errAt(t, "datatype %q: only the predeclared xsd prefix is supported", t.text)
			}
			if !xsdTypes[local] {
				return p.errAt(t, "datatype %q is not in the subset", t.text)
			}
		case tLiteral:
		default:
			return p.errAt(t, "attribute value: want xsd:type or a literal, got %q", t.text)
		}
		if n := p.peek(); !(n.kind == tPunct && n.text == "|") {
			return nil
		}
		p.next()
	}
}

// ---------------------------------------------------------------------------
// The parser's own tests: it must fail loudly on constructs outside the
// subset, and the check must report what it should.

func TestParserRejectsUnsupported(t *testing.T) {
	cases := map[string]string{
		"interleave":       `start = A A = element a { attribute x { xsd:string } & attribute y { xsd:string } }`,
		"mixed":            `start = A A = element a { mixed { empty } }`,
		"list":             `start = A A = element a { attribute x { list { xsd:string } } }`,
		"doc comment":      "## doc\nstart = A A = element a { empty }",
		"mixed operators":  `start = A A = element a { attribute x { xsd:string }, attribute y { xsd:string } | text }`,
		"mixed operators2": `start = A A = element a { text | attribute x { xsd:string }, attribute y { xsd:string } }`,
		"undefined ref":    `start = A A = element a { B }`,
		"unreferenced":     `start = A A = element a { empty } B = element b { empty }`,
		"facet":            `start = A A = element a { attribute x { xsd:string { length = "1" } } }`,
		"unknown type":     `start = A A = element a { attribute x { xsd:dateTime } }`,
		"other prefix":     `start = A A = element a { attribute x { foo:bar } }`,
		"wildcard":         `start = A A = element * { empty }`,
		"namespaced":       `start = A A = element ns:a { empty }`,
		"ns wildcard":      `start = A A = element ns:* { empty }`,
		"include":          `include "x.rnc" start = A A = element a { empty }`,
		"combine":          `start = A A = element a { empty } A |= element b { empty }`,
		"bare keyword":     `start = A A = element a { attribute parent { xsd:string } }`,
		"no start":         `A = element a { empty }`,
		"attribute start":  `start = attribute a { xsd:string }`,
		"ambiguous child":  `start = A A = element a { B, C } B = element b { empty } C = element b { text }`,
		"self reference":   `start = A A = element a { B } B = B?`,
	}
	for name, src := range cases {
		if _, err := parseGrammar(src); err == nil {
			t.Errorf("%s: parsed without error, want a loud failure", name)
		} else {
			t.Logf("%s: %v", name, err)
		}
	}
}

func TestCheckReportsMissing(t *testing.T) {
	g, err := parseGrammar(`start = A
A = element a { attribute x { xsd:string }?, B*, E? }
B = element b { (attribute y { xsd:string } | attribute z { xsd:string }), B* }
E = element e { text }`)
	if err != nil {
		t.Fatal(err)
	}
	// Whitespace between elements is allowed everywhere; the text in the
	// outer b (plain) and the inner b (CDATA) is not, and is reported once
	// per element; e's text is allowed.
	res, err := g.check([]byte("<a x=\"1\" w=\"2\">\n <b y=\"1\">t<b z=\"2\"><c/><![CDATA[u]]></b>v</b><d/><e>ok</e>\n</a>"))
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, v := range res.violations {
		got = append(got, v.patternPath+" "+v.item+" "+v.xmlPath)
	}
	sort.Strings(got)
	want := []string{
		"A attribute w /a",
		"A element d /a/d[1]",
		"A/B text /a/b[1]",
		"A/B/B element c /a/b[1]/b[1]/c[1]",
		"A/B/B text /a/b[1]/b[1]",
	}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Errorf("violations:\n%s\nwant:\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
	if res.elements != 6 || res.attributes != 4 {
		t.Errorf("visited %d elements, %d attributes; want 6, 4", res.elements, res.attributes)
	}
}
