package mayo

import (
	"strings"

	hotdog "github.com/0magnet/frank/hotdog"
)

// ParseStylesheet parses a CSS string into a list of style rules
func ParseStylesheet(css string) []hotdog.StyleRule {
	tokens := newTokenizer(css).tokenize()
	parser := &cssParser{tokens: tokens, pos: 0}
	return parser.parseRules()
}

// parseStylesheet is the old function signature kept for compatibility
func parseStylesheet(css string) []*hotdog.StyleElement {
	rules := ParseStylesheet(css)
	var elements []*hotdog.StyleElement
	for _, rule := range rules {
		stylesheet := &hotdog.Stylesheet{
			BackgroundColor: &hotdog.ColorRGBA{R: 1, G: 1, B: 1, A: 0},
			FontSize:        0,
			Display:         "",
			Position:        "static",
			Opacity:         1,
		}
		for prop, val := range rule.Properties {
			applyProperty(stylesheet, prop, val)
		}
		elements = append(elements, &hotdog.StyleElement{
			Selector: rule.Selector,
			Style:    stylesheet,
		})
	}
	return elements
}

type cssParser struct {
	tokens []cssToken
	pos    int
}

func (p *cssParser) peek() cssToken {
	if p.pos >= len(p.tokens) {
		return cssToken{tokenEOF, ""}
	}
	return p.tokens[p.pos]
}

func (p *cssParser) advance() cssToken {
	tok := p.peek()
	if p.pos < len(p.tokens) {
		p.pos++
	}
	return tok
}

func (p *cssParser) skipWhitespace() {
	for p.pos < len(p.tokens) && p.tokens[p.pos].Type == tokenWhitespace {
		p.pos++
	}
}

func (p *cssParser) parseRules() []hotdog.StyleRule {
	var rules []hotdog.StyleRule

	for p.peek().Type != tokenEOF {
		p.skipWhitespace()
		if p.peek().Type == tokenEOF {
			break
		}

		// Handle @import (basic)
		if p.peek().Type == tokenAtKeyword {
			p.skipAtRule()
			continue
		}

		// Parse a ruleset
		ruleList := p.parseRuleset()
		rules = append(rules, ruleList...)
	}

	return rules
}

func (p *cssParser) skipAtRule() {
	p.advance() // skip @keyword
	// Skip until semicolon or block
	depth := 0
	for p.peek().Type != tokenEOF {
		tok := p.advance()
		if tok.Type == tokenOpenBrace {
			depth++
		} else if tok.Type == tokenCloseBrace {
			depth--
			if depth <= 0 {
				return
			}
		} else if tok.Type == tokenSemicolon && depth == 0 {
			return
		}
	}
}

func (p *cssParser) parseRuleset() []hotdog.StyleRule {
	// Read selector(s) until '{'
	var selectorParts []string
	var currentSelector strings.Builder

	for p.peek().Type != tokenEOF && p.peek().Type != tokenOpenBrace {
		tok := p.advance()
		if tok.Type == tokenComma {
			sel := strings.TrimSpace(currentSelector.String())
			if sel != "" {
				selectorParts = append(selectorParts, sel)
			}
			currentSelector.Reset()
		} else {
			currentSelector.WriteString(tok.Value)
		}
	}
	sel := strings.TrimSpace(currentSelector.String())
	if sel != "" {
		selectorParts = append(selectorParts, sel)
	}

	if p.peek().Type != tokenOpenBrace {
		return nil
	}
	p.advance() // consume '{'

	// Parse declarations
	properties := p.parseDeclarations()

	// Expand shorthands
	expanded := expandShorthands(properties)

	var rules []hotdog.StyleRule
	for _, selector := range selectorParts {
		spec := calculateSpecificity(selector)
		rules = append(rules, hotdog.StyleRule{
			Selector:    selector,
			Properties:  copyMap(expanded),
			Specificity: spec,
		})
	}

	return rules
}

func (p *cssParser) parseDeclarations() map[string]string {
	props := make(map[string]string)

	for p.peek().Type != tokenEOF && p.peek().Type != tokenCloseBrace {
		p.skipWhitespace()
		if p.peek().Type == tokenCloseBrace || p.peek().Type == tokenEOF {
			break
		}

		// Read property name
		var propName strings.Builder
		for p.peek().Type != tokenEOF && p.peek().Type != tokenColon && p.peek().Type != tokenCloseBrace {
			tok := p.advance()
			if tok.Type != tokenWhitespace {
				propName.WriteString(tok.Value)
			}
		}

		if p.peek().Type == tokenColon {
			p.advance() // consume ':'
		} else {
			continue
		}

		p.skipWhitespace()

		// Read property value
		var propValue strings.Builder
		for p.peek().Type != tokenEOF && p.peek().Type != tokenSemicolon && p.peek().Type != tokenCloseBrace {
			tok := p.advance()
			propValue.WriteString(tok.Value)
		}

		if p.peek().Type == tokenSemicolon {
			p.advance()
		}

		name := strings.TrimSpace(propName.String())
		value := strings.TrimSpace(propValue.String())
		if name != "" && value != "" {
			props[name] = value
		}
	}

	if p.peek().Type == tokenCloseBrace {
		p.advance()
	}

	return props
}

func expandShorthands(props map[string]string) map[string]string {
	result := make(map[string]string)

	for k, v := range props {
		switch k {
		case "margin":
			parts := splitCSSValues(v)
			t, r, b, l := expandBoxShorthand(parts)
			result["margin-top"] = t
			result["margin-right"] = r
			result["margin-bottom"] = b
			result["margin-left"] = l

		case "padding":
			parts := splitCSSValues(v)
			t, r, b, l := expandBoxShorthand(parts)
			result["padding-top"] = t
			result["padding-right"] = r
			result["padding-bottom"] = b
			result["padding-left"] = l

		case "border":
			parts := splitCSSValues(v)
			width, style, color := "medium", "none", ""
			for _, part := range parts {
				if isBorderStyle(part) {
					style = part
				} else if isColorValue(part) {
					color = part
				} else {
					width = part
				}
			}
			for _, side := range []string{"top", "right", "bottom", "left"} {
				result["border-"+side+"-width"] = width
				result["border-"+side+"-style"] = style
				if color != "" {
					result["border-"+side+"-color"] = color
				}
			}

		case "background":
			// Simple: just treat as background-color for now
			result["background-color"] = v

		case "font":
			// Very simplified font shorthand
			result["font"] = v

		default:
			result[k] = v
		}
	}

	return result
}

func splitCSSValues(value string) []string {
	var parts []string
	for _, p := range strings.Fields(value) {
		p = strings.TrimSpace(p)
		if p != "" {
			parts = append(parts, p)
		}
	}
	return parts
}

func expandBoxShorthand(parts []string) (top, right, bottom, left string) {
	switch len(parts) {
	case 1:
		return parts[0], parts[0], parts[0], parts[0]
	case 2:
		return parts[0], parts[1], parts[0], parts[1]
	case 3:
		return parts[0], parts[1], parts[2], parts[1]
	case 4:
		return parts[0], parts[1], parts[2], parts[3]
	default:
		return "0", "0", "0", "0"
	}
}

func isBorderStyle(s string) bool {
	switch s {
	case "none", "hidden", "dotted", "dashed", "solid", "double", "groove", "ridge", "inset", "outset":
		return true
	}
	return false
}

func isColorValue(s string) bool {
	if strings.HasPrefix(s, "#") || strings.HasPrefix(s, "rgb") || strings.HasPrefix(s, "rgba") {
		return true
	}
	if _, ok := colorTable[s]; ok {
		return true
	}
	return false
}

func copyMap(m map[string]string) map[string]string {
	result := make(map[string]string, len(m))
	for k, v := range m {
		result[k] = v
	}
	return result
}

// applyProperty applies a single CSS property to a stylesheet
func applyProperty(s *hotdog.Stylesheet, prop, value string) {
	propSlice := []string{prop, value}
	mapPropToStylesheet(s, propSlice)
}
