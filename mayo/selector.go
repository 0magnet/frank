package mayo

import (
	"strings"

	hotdog "github.com/0magnet/frank/hotdog"
)

// selectorPart represents one simple selector component
type selectorPart struct {
	Element    string // e.g., "div", "*", ""
	ID         string // e.g., "main"
	Classes    []string
	Attributes []attrSelector
	Pseudos    []string
}

type attrSelector struct {
	Name  string
	Op    string // "", "=", "~=", "|=", "^=", "$=", "*="
	Value string
}

type combinatorType int

const (
	combDescendant combinatorType = iota
	combChild
	combAdjacentSibling
	combGeneralSibling
)

type selectorSegment struct {
	Part       selectorPart
	Combinator combinatorType
}

// MatchSelector checks if a node matches a CSS selector string
func MatchSelector(node *hotdog.NodeDOM, selector string) bool {
	segments := parseSelector(selector)
	if len(segments) == 0 {
		return false
	}
	return matchSegments(node, segments, len(segments)-1)
}

func matchSegments(node *hotdog.NodeDOM, segments []selectorSegment, idx int) bool {
	if idx < 0 {
		return true
	}
	if node == nil {
		return false
	}

	seg := segments[idx]
	if !matchPart(node, seg.Part) {
		return false
	}

	if idx == 0 {
		return true
	}

	prevSeg := segments[idx-1]
	switch seg.Combinator {
	case combDescendant:
		// Match any ancestor
		parent := node.Parent
		for parent != nil {
			if matchSegments(parent, segments, idx-1) {
				return true
			}
			parent = parent.Parent
		}
		return false

	case combChild:
		// Match direct parent
		if node.Parent == nil {
			return false
		}
		return matchSegments(node.Parent, segments, idx-1)

	case combAdjacentSibling:
		prev := getPreviousSibling(node)
		if prev == nil {
			return false
		}
		return matchSegments(prev, segments, idx-1)

	case combGeneralSibling:
		if node.Parent == nil {
			return false
		}
		for _, sibling := range node.Parent.Children {
			if sibling == node {
				break
			}
			if matchSegments(sibling, segments, idx-1) {
				return true
			}
		}
		return false

	default:
		_ = prevSeg
		return false
	}
}

func matchPart(node *hotdog.NodeDOM, part selectorPart) bool {
	// Element
	if part.Element != "" && part.Element != "*" {
		if node.Element != part.Element {
			return false
		}
	}

	// ID
	if part.ID != "" {
		if node.ID() != part.ID {
			return false
		}
	}

	// Classes
	for _, cls := range part.Classes {
		if !node.HasClass(cls) {
			return false
		}
	}

	// Attributes
	for _, attr := range part.Attributes {
		nodeVal := node.Attr(attr.Name)
		switch attr.Op {
		case "":
			if nodeVal == "" {
				return false
			}
		case "=":
			if nodeVal != attr.Value {
				return false
			}
		case "~=":
			found := false
			for _, v := range strings.Fields(nodeVal) {
				if v == attr.Value {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		case "|=":
			if nodeVal != attr.Value && !strings.HasPrefix(nodeVal, attr.Value+"-") {
				return false
			}
		case "^=":
			if !strings.HasPrefix(nodeVal, attr.Value) {
				return false
			}
		case "$=":
			if !strings.HasSuffix(nodeVal, attr.Value) {
				return false
			}
		case "*=":
			if !strings.Contains(nodeVal, attr.Value) {
				return false
			}
		}
	}

	// Pseudo-classes
	for _, pseudo := range part.Pseudos {
		switch pseudo {
		case "first-child":
			if !isFirstChild(node) {
				return false
			}
		case "last-child":
			if !isLastChild(node) {
				return false
			}
		}
	}

	return true
}

func isFirstChild(node *hotdog.NodeDOM) bool {
	if node.Parent == nil {
		return true
	}
	for _, child := range node.Parent.Children {
		if child.Style != nil && child.Style.Display == "none" {
			continue
		}
		return child == node
	}
	return false
}

func isLastChild(node *hotdog.NodeDOM) bool {
	if node.Parent == nil {
		return true
	}
	children := node.Parent.Children
	for i := len(children) - 1; i >= 0; i-- {
		child := children[i]
		if child.Style != nil && child.Style.Display == "none" {
			continue
		}
		return child == node
	}
	return false
}

func getPreviousSibling(node *hotdog.NodeDOM) *hotdog.NodeDOM {
	if node.Parent == nil {
		return nil
	}
	var prev *hotdog.NodeDOM
	for _, child := range node.Parent.Children {
		if child == node {
			return prev
		}
		prev = child
	}
	return nil
}

// parseSelector parses a selector string into segments with combinators
func parseSelector(selector string) []selectorSegment {
	selector = strings.TrimSpace(selector)
	if selector == "" {
		return nil
	}

	var segments []selectorSegment
	tokens := tokenizeSelector(selector)
	if len(tokens) == 0 {
		return nil
	}

	var currentTokens []string
	var currentCombinator combinatorType = combDescendant

	for _, tok := range tokens {
		switch tok {
		case ">":
			if len(currentTokens) > 0 {
				part := parseSelectorPart(strings.Join(currentTokens, ""))
				segments = append(segments, selectorSegment{Part: part, Combinator: currentCombinator})
				currentTokens = nil
			}
			currentCombinator = combChild
		case "+":
			if len(currentTokens) > 0 {
				part := parseSelectorPart(strings.Join(currentTokens, ""))
				segments = append(segments, selectorSegment{Part: part, Combinator: currentCombinator})
				currentTokens = nil
			}
			currentCombinator = combAdjacentSibling
		case "~":
			if len(currentTokens) > 0 {
				part := parseSelectorPart(strings.Join(currentTokens, ""))
				segments = append(segments, selectorSegment{Part: part, Combinator: currentCombinator})
				currentTokens = nil
			}
			currentCombinator = combGeneralSibling
		case " ":
			if len(currentTokens) > 0 {
				part := parseSelectorPart(strings.Join(currentTokens, ""))
				segments = append(segments, selectorSegment{Part: part, Combinator: currentCombinator})
				currentTokens = nil
				currentCombinator = combDescendant
			}
		default:
			currentTokens = append(currentTokens, tok)
		}
	}

	if len(currentTokens) > 0 {
		part := parseSelectorPart(strings.Join(currentTokens, ""))
		segments = append(segments, selectorSegment{Part: part, Combinator: currentCombinator})
	}

	return segments
}

// tokenizeSelector splits a selector into atomic tokens
func tokenizeSelector(selector string) []string {
	var tokens []string
	i := 0
	runes := []rune(selector)

	for i < len(runes) {
		ch := runes[i]

		// Whitespace
		if ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' {
			for i < len(runes) && (runes[i] == ' ' || runes[i] == '\t' || runes[i] == '\n' || runes[i] == '\r') {
				i++
			}
			// Check if next char is a combinator
			if i < len(runes) && (runes[i] == '>' || runes[i] == '+' || runes[i] == '~') {
				continue // let the combinator be handled next iteration
			}
			tokens = append(tokens, " ")
			continue
		}

		// Combinators
		if ch == '>' || ch == '+' || ch == '~' {
			tokens = append(tokens, string(ch))
			i++
			// Skip trailing whitespace
			for i < len(runes) && (runes[i] == ' ' || runes[i] == '\t') {
				i++
			}
			continue
		}

		// Attribute selector [...]
		if ch == '[' {
			j := i
			depth := 0
			for j < len(runes) {
				if runes[j] == '[' {
					depth++
				} else if runes[j] == ']' {
					depth--
					if depth == 0 {
						j++
						break
					}
				}
				j++
			}
			tokens = append(tokens, string(runes[i:j]))
			i = j
			continue
		}

		// Everything else accumulates as a token part
		j := i
		for j < len(runes) {
			c := runes[j]
			if c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == '>' || c == '+' || c == '~' || c == '[' {
				break
			}
			j++
		}
		if j > i {
			tokens = append(tokens, string(runes[i:j]))
		}
		i = j
	}

	return tokens
}

// parseSelectorPart parses a simple selector like "div.class#id:pseudo"
func parseSelectorPart(s string) selectorPart {
	var part selectorPart
	runes := []rune(s)
	i := 0

	// Handle attribute selectors that may be mixed in
	if strings.HasPrefix(s, "[") {
		attr := parseAttrSelector(s)
		part.Attributes = append(part.Attributes, attr)
		return part
	}

	for i < len(runes) {
		ch := runes[i]

		switch ch {
		case '#':
			i++
			start := i
			for i < len(runes) && runes[i] != '.' && runes[i] != '#' && runes[i] != ':' && runes[i] != '[' {
				i++
			}
			part.ID = string(runes[start:i])

		case '.':
			i++
			start := i
			for i < len(runes) && runes[i] != '.' && runes[i] != '#' && runes[i] != ':' && runes[i] != '[' {
				i++
			}
			part.Classes = append(part.Classes, string(runes[start:i]))

		case ':':
			i++
			// Skip second ':' for pseudo-elements
			if i < len(runes) && runes[i] == ':' {
				i++
			}
			start := i
			for i < len(runes) && runes[i] != '.' && runes[i] != '#' && runes[i] != ':' && runes[i] != '[' && runes[i] != '(' {
				i++
			}
			pseudo := string(runes[start:i])
			// Skip function arguments like :nth-child(2)
			if i < len(runes) && runes[i] == '(' {
				depth := 0
				for i < len(runes) {
					if runes[i] == '(' {
						depth++
					} else if runes[i] == ')' {
						depth--
						if depth == 0 {
							i++
							break
						}
					}
					i++
				}
			}
			part.Pseudos = append(part.Pseudos, pseudo)

		case '[':
			j := i
			depth := 0
			for j < len(runes) {
				if runes[j] == '[' {
					depth++
				} else if runes[j] == ']' {
					depth--
					if depth == 0 {
						j++
						break
					}
				}
				j++
			}
			attr := parseAttrSelector(string(runes[i:j]))
			part.Attributes = append(part.Attributes, attr)
			i = j

		case '*':
			part.Element = "*"
			i++

		default:
			// Element name
			start := i
			for i < len(runes) && runes[i] != '.' && runes[i] != '#' && runes[i] != ':' && runes[i] != '[' && runes[i] != '*' {
				i++
			}
			part.Element = string(runes[start:i])
		}
	}

	return part
}

func parseAttrSelector(s string) attrSelector {
	s = strings.Trim(s, "[]")
	var attr attrSelector

	for _, op := range []string{"~=", "|=", "^=", "$=", "*=", "="} {
		idx := strings.Index(s, op)
		if idx >= 0 {
			attr.Name = strings.TrimSpace(s[:idx])
			attr.Op = op
			attr.Value = strings.Trim(strings.TrimSpace(s[idx+len(op):]), "\"'")
			return attr
		}
	}

	attr.Name = strings.TrimSpace(s)
	return attr
}

// calculateSpecificity computes [inline, ids, classes+attrs+pseudos, elements]
func calculateSpecificity(selector string) [4]int {
	var spec [4]int
	segments := parseSelector(selector)
	for _, seg := range segments {
		part := seg.Part
		if part.ID != "" {
			spec[1]++
		}
		spec[2] += len(part.Classes) + len(part.Attributes) + len(part.Pseudos)
		if part.Element != "" && part.Element != "*" {
			spec[3]++
		}
	}
	return spec
}
