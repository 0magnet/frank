package mayo

import (
	"regexp"
	"strconv"
	"strings"

	hotdog "github.com/0magnet/frank/hotdog"
)

func getDefaultElementDisplay(element string) string {
	displayType := "block"

	switch element {
	case "script", "style", "meta", "link", "head", "title":
		displayType = "none"
	case "li":
		displayType = "list-item"
	case "html:text", "a", "abbr", "acronym", "b", "bdo", "big", "br",
		"button", "cite", "code", "dfn", "em", "i", "img", "input", "kbd",
		"label", "map", "object", "output", "q", "samp", "select", "small",
		"span", "strong", "sub", "sup", "textarea", "time", "tt", "var", "font":
		displayType = "inline"
	}

	return displayType
}

func getDefaultElementFontWeight(element string) int {
	switch element {
	case "h1", "h2", "h3", "h4", "h5", "h6":
		return 600
	}

	return 400
}

func mapSizeValue(sizeValue string) float64 {
	sizeValue = strings.TrimSpace(sizeValue)

	// Percentage and auto values can't be resolved to pixels at parse time
	if strings.HasSuffix(sizeValue, "%") || sizeValue == "auto" || sizeValue == "inherit" {
		return 0
	}

	re := regexp.MustCompile("[0-9]+\\.?[0-9]*")
	valueString := re.FindString(sizeValue)
	if valueString == "" {
		return 0
	}

	value, err := strconv.ParseFloat(valueString, 64)
	if err != nil {
		return 0
	}

	// Handle em units (relative to ~14px base; rough approximation)
	if strings.HasSuffix(sizeValue, "em") {
		return value * 14
	}
	// Handle rem (relative to root, assume 16px)
	if strings.HasSuffix(sizeValue, "rem") {
		return value * 16
	}

	return value
}

func mapPropToStylesheet(parsedStyleSheet *hotdog.Stylesheet, propSlice []string) *hotdog.Stylesheet {
	propName := strings.TrimSpace(propSlice[0])
	propValue := strings.TrimSpace(propSlice[1])

	switch propName {
	case "color":
		c := MapCSSColor(propValue)
		if c != nil {
			parsedStyleSheet.Color = c
		}
	case "background-color":
		c := MapCSSColor(propValue)
		if c != nil {
			parsedStyleSheet.BackgroundColor = c
		}
	case "font-size":
		parsedStyleSheet.FontSize = mapSizeValue(propValue)
	case "font-weight":
		parsedStyleSheet.FontWeight = mapFontWeight(propValue)
	case "display":
		parsedStyleSheet.Display = propValue
	case "position":
		parsedStyleSheet.Position = propValue
	case "height":
		parsedStyleSheet.Height = mapSizeValue(propValue)
	case "width":
		parsedStyleSheet.Width = mapSizeValue(propValue)
	case "top":
		parsedStyleSheet.Top = mapSizeValue(propValue)
	case "left":
		parsedStyleSheet.Left = mapSizeValue(propValue)
	case "right":
		parsedStyleSheet.Right = mapSizeValue(propValue)
	case "bottom":
		parsedStyleSheet.Bottom = mapSizeValue(propValue)

	// Margins
	case "margin-top":
		parsedStyleSheet.MarginTop = mapSizeValue(propValue)
	case "margin-right":
		parsedStyleSheet.MarginRight = mapSizeValue(propValue)
	case "margin-bottom":
		parsedStyleSheet.MarginBottom = mapSizeValue(propValue)
	case "margin-left":
		parsedStyleSheet.MarginLeft = mapSizeValue(propValue)

	// Padding
	case "padding-top":
		parsedStyleSheet.PaddingTop = mapSizeValue(propValue)
	case "padding-right":
		parsedStyleSheet.PaddingRight = mapSizeValue(propValue)
	case "padding-bottom":
		parsedStyleSheet.PaddingBottom = mapSizeValue(propValue)
	case "padding-left":
		parsedStyleSheet.PaddingLeft = mapSizeValue(propValue)

	// Borders
	case "border-top-width":
		parsedStyleSheet.BorderTopWidth = mapSizeValue(propValue)
	case "border-right-width":
		parsedStyleSheet.BorderRightWidth = mapSizeValue(propValue)
	case "border-bottom-width":
		parsedStyleSheet.BorderBottomWidth = mapSizeValue(propValue)
	case "border-left-width":
		parsedStyleSheet.BorderLeftWidth = mapSizeValue(propValue)
	case "border-top-color":
		c := MapCSSColor(propValue)
		if c != nil {
			parsedStyleSheet.BorderTopColor = c
		}
	case "border-right-color":
		c := MapCSSColor(propValue)
		if c != nil {
			parsedStyleSheet.BorderRightColor = c
		}
	case "border-bottom-color":
		c := MapCSSColor(propValue)
		if c != nil {
			parsedStyleSheet.BorderBottomColor = c
		}
	case "border-left-color":
		c := MapCSSColor(propValue)
		if c != nil {
			parsedStyleSheet.BorderLeftColor = c
		}
	case "border-top-style":
		parsedStyleSheet.BorderTopStyle = propValue
	case "border-right-style":
		parsedStyleSheet.BorderRightStyle = propValue
	case "border-bottom-style":
		parsedStyleSheet.BorderBottomStyle = propValue
	case "border-left-style":
		parsedStyleSheet.BorderLeftStyle = propValue

	// Text
	case "text-align":
		parsedStyleSheet.TextAlign = propValue
	case "text-decoration":
		parsedStyleSheet.TextDecoration = propValue
	case "line-height":
		parsedStyleSheet.LineHeight = mapSizeValue(propValue)
	case "font-family":
		parsedStyleSheet.FontFamily = propValue
	case "font-style":
		parsedStyleSheet.FontStyle = propValue
	case "white-space":
		parsedStyleSheet.WhiteSpace = propValue
	case "text-transform":
		parsedStyleSheet.TextTransform = propValue

	// Layout
	case "min-width":
		parsedStyleSheet.MinWidth = mapSizeValue(propValue)
	case "max-width":
		parsedStyleSheet.MaxWidth = mapSizeValue(propValue)
	case "min-height":
		parsedStyleSheet.MinHeight = mapSizeValue(propValue)
	case "max-height":
		parsedStyleSheet.MaxHeight = mapSizeValue(propValue)
	case "overflow":
		parsedStyleSheet.Overflow = propValue
	case "float":
		parsedStyleSheet.Float = propValue
	case "clear":
		parsedStyleSheet.Clear = propValue

	// Visual
	case "opacity":
		parsedStyleSheet.Opacity = mapSizeValue(propValue)
	case "visibility":
		parsedStyleSheet.Visibility = propValue
	case "list-style-type":
		parsedStyleSheet.ListStyleType = propValue

	// Z-index
	case "z-index":
		parsedStyleSheet.ZIndex = int(mapSizeValue(propValue))

	// Flexbox
	case "flex-direction":
		parsedStyleSheet.FlexDirection = propValue
	case "justify-content":
		parsedStyleSheet.JustifyContent = propValue
	case "align-items":
		parsedStyleSheet.AlignItems = propValue
	case "flex-grow":
		parsedStyleSheet.FlexGrow = mapSizeValue(propValue)
	case "flex-shrink":
		parsedStyleSheet.FlexShrink = mapSizeValue(propValue)
	case "flex-basis":
		parsedStyleSheet.FlexBasis = mapSizeValue(propValue)
	}

	return parsedStyleSheet
}

func mapFontWeight(value string) int {
	switch value {
	case "normal":
		return 400
	case "bold":
		return 700
	case "lighter":
		return 300
	case "bolder":
		return 800
	default:
		v := mapSizeValue(value)
		if v > 0 {
			return int(v)
		}
		return 400
	}
}

func parseInlineStylesheet(attributes []*hotdog.Attribute, elementStylesheet *hotdog.Stylesheet) *hotdog.Stylesheet {
	for i := 0; i < len(attributes); i++ {
		attributeName := attributes[i].Name
		if attributeName == "style" {

			styleString := attributes[i].Value
			styleProps := strings.Split(strings.Replace(styleString, " ", "", -1), ";")

			for i := 0; i < len(styleProps); i++ {
				styledProperty := strings.Split(styleProps[i], ":")
				if len(styledProperty) >= 2 {
					elementStylesheet = mapPropToStylesheet(elementStylesheet, styledProperty)
				}
			}
		}
	}

	return elementStylesheet
}

func hasInlineStyle(attributes []*hotdog.Attribute) bool {
	inlineStyle := false

	for i := 0; i < len(attributes); i++ {
		attributeName := attributes[i].Name
		if attributeName == "style" {
			inlineStyle = true
		}
	}

	return inlineStyle
}

func GetElementStylesheet(elementName string, attributes []*hotdog.Attribute) *hotdog.Stylesheet {
	elementStylesheet := &hotdog.Stylesheet{
		BackgroundColor: &hotdog.ColorRGBA{R: 1, G: 1, B: 1, A: 0},
		FontSize:        0,
		Display:         "",
		Position:        "static",
		Opacity:         1,
	}

	if hasInlineStyle(attributes) {
		elementStylesheet = parseInlineStylesheet(attributes, elementStylesheet)
	}

	if elementStylesheet.FontSize == float64(0) {
		fontSize := elementFontTable[elementName]

		if fontSize != float64(0) {
			elementStylesheet.FontSize = fontSize
		} else {
			elementStylesheet.FontSize = float64(14)
		}
	}

	if elementStylesheet.Color == nil {
		color := elementColorTable[elementName]
		if color != nil {
			elementStylesheet.Color = color
		} else {
			elementStylesheet.Color = &hotdog.ColorRGBA{R: 0, G: 0, B: 0, A: 1}
		}
	}

	if elementStylesheet.FontWeight == 0 {
		elementStylesheet.FontWeight = getDefaultElementFontWeight(elementName)
	}

	if elementStylesheet.Display == "" {
		elementStylesheet.Display = getDefaultElementDisplay(elementName)
	}

	return elementStylesheet
}
