package mayo

import (
	"sort"
	"strings"

	hotdog "github.com/0magnet/frank/hotdog"
)

// inheritableProperties lists CSS properties that inherit from parent
var inheritableProperties = map[string]bool{
	"color":           true,
	"font-size":       true,
	"font-family":     true,
	"font-weight":     true,
	"font-style":      true,
	"line-height":     true,
	"text-align":      true,
	"text-decoration": true,
	"text-transform":  true,
	"white-space":     true,
	"visibility":      true,
	"list-style-type": true,
}

type matchedRule struct {
	rule  hotdog.StyleRule
	order int
}

// ruleIndex groups rules by the element name in their rightmost selector segment
// for fast rejection. Rules with no element constraint go in "*".
type ruleIndex struct {
	byElement map[string][]indexedRule
	universal []indexedRule
}

type indexedRule struct {
	rule  hotdog.StyleRule
	order int
}

func buildRuleIndex(rules []hotdog.StyleRule) *ruleIndex {
	idx := &ruleIndex{byElement: make(map[string][]indexedRule)}
	for i, rule := range rules {
		ir := indexedRule{rule: rule, order: i}
		// Extract the rightmost simple selector's element name
		elem := extractRightmostElement(rule.Selector)
		if elem == "" || elem == "*" {
			idx.universal = append(idx.universal, ir)
		} else {
			idx.byElement[elem] = append(idx.byElement[elem], ir)
		}
	}
	return idx
}

func extractRightmostElement(selector string) string {
	// Find the last segment (after last space, >, +, ~)
	selector = strings.TrimSpace(selector)
	lastSpace := -1
	for i := len(selector) - 1; i >= 0; i-- {
		ch := selector[i]
		if ch == ' ' || ch == '>' || ch == '+' || ch == '~' {
			lastSpace = i
			break
		}
	}
	last := selector
	if lastSpace >= 0 {
		last = strings.TrimSpace(selector[lastSpace+1:])
	}
	// Extract element name (before any . # : [)
	for i, ch := range last {
		if ch == '.' || ch == '#' || ch == ':' || ch == '[' {
			return last[:i]
		}
	}
	return last
}

func (idx *ruleIndex) candidateRules(elementName string) []indexedRule {
	candidates := idx.universal
	if specific, ok := idx.byElement[elementName]; ok {
		candidates = append(candidates, specific...)
	}
	return candidates
}

// ComputeStyles walks the DOM tree and applies cascaded styles to every node
func ComputeStyles(root *hotdog.NodeDOM, rules []hotdog.StyleRule) {
	idx := buildRuleIndex(rules)
	computeNodeStyles(root, nil, idx)
}

func computeNodeStyles(node *hotdog.NodeDOM, parent *hotdog.NodeDOM, idx *ruleIndex) {
	// Start with default element stylesheet
	node.Style = GetElementStylesheet(node.Element, node.Attributes)

	// Apply inherited properties from parent
	if parent != nil && parent.Style != nil {
		inheritStyles(node.Style, parent.Style)
	}

	// Skip selector matching for text nodes and nodes with display:none
	if node.Element != "html:text" && node.Element != "html:doctype" && node.Element != "html:raw" {
		// Collect matching rules using the index for fast rejection
		candidates := idx.candidateRules(node.Element)
		var matched []matchedRule
		for _, ir := range candidates {
			if MatchSelector(node, ir.rule.Selector) {
				matched = append(matched, matchedRule{rule: ir.rule, order: ir.order})
			}
		}

		// Sort by specificity then source order
		if len(matched) > 1 {
			sort.SliceStable(matched, func(i, j int) bool {
				si := matched[i].rule.Specificity
				sj := matched[j].rule.Specificity
				for k := 0; k < 4; k++ {
					if si[k] != sj[k] {
						return si[k] < sj[k]
					}
				}
				return matched[i].order < matched[j].order
			})
		}

		// Apply matched rules in order (lowest specificity first)
		for _, m := range matched {
			for prop, val := range m.rule.Properties {
				applyPropertyToStylesheet(node.Style, prop, val)
			}
		}
	}

	// Recurse into children
	for _, child := range node.Children {
		computeNodeStyles(child, node, idx)
	}
}

func inheritStyles(child *hotdog.Stylesheet, parent *hotdog.Stylesheet) {
	if child.Color == nil && parent.Color != nil {
		child.Color = parent.Color
	}
	if child.FontFamily == "" && parent.FontFamily != "" {
		child.FontFamily = parent.FontFamily
	}
	if child.FontStyle == "" && parent.FontStyle != "" {
		child.FontStyle = parent.FontStyle
	}
	if child.TextAlign == "" && parent.TextAlign != "" {
		child.TextAlign = parent.TextAlign
	}
	if child.TextDecoration == "" && parent.TextDecoration != "" {
		child.TextDecoration = parent.TextDecoration
	}
	if child.TextTransform == "" && parent.TextTransform != "" {
		child.TextTransform = parent.TextTransform
	}
	if child.WhiteSpace == "" && parent.WhiteSpace != "" {
		child.WhiteSpace = parent.WhiteSpace
	}
	if child.LineHeight == 0 && parent.LineHeight != 0 {
		child.LineHeight = parent.LineHeight
	}
	if child.Visibility == "" && parent.Visibility != "" {
		child.Visibility = parent.Visibility
	}
	if child.ListStyleType == "" && parent.ListStyleType != "" {
		child.ListStyleType = parent.ListStyleType
	}
}

// applyPropertyToStylesheet applies a CSS property value to a stylesheet struct
func applyPropertyToStylesheet(s *hotdog.Stylesheet, prop, value string) {
	mapPropToStylesheet(s, []string{prop, value})
}
