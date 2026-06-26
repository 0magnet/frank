package bun

import (
	"fmt"
	"strings"

	gg "github.com/0magnet/frank/gg"
	hotdog "github.com/0magnet/frank/hotdog"
)

// isWhitespaceOnly returns true if this is a text node with only whitespace
func isWhitespaceOnly(node *hotdog.NodeDOM) bool {
	return node.Element == "html:text" && strings.TrimSpace(node.Content) == ""
}

func RenderDocument(ctx *gg.Context, document *hotdog.Document, experimentalLayout bool) error {
	if !experimentalLayout {
		body, _ := document.DOM.FindChildByName("body")

		canvasWidth := float64(ctx.Width())
		canvasHeight := float64(ctx.Height())

		document.DOM.RenderBox.Width = canvasWidth
		document.DOM.RenderBox.Height = canvasHeight

		// Ensure all ancestors of body also have proper render boxes
		htmlNode, _ := document.DOM.FindChildByName("html")
		if htmlNode != nil && htmlNode.RenderBox != nil {
			htmlNode.RenderBox.Width = canvasWidth
			htmlNode.RenderBox.Height = canvasHeight
		}

		ctx.SetRGB(1, 1, 1)
		ctx.Clear()

		// Pass 1: Layout only (compute positions and sizes)
		layoutOnly(ctx, body, 0)
		// Pass 2: Paint only visible nodes (within canvas bounds)
		paintVisible(ctx, body, canvasHeight)
	} else {
		html, err := document.DOM.FindChildByName("html")
		if err != nil {
			return err
		}

		renderTree := createRenderTree(html)
		renderTree.RenderBox.Width = float64(ctx.Width())
		renderTree.RenderBox.Height = float64(ctx.Height())

		layoutNode(ctx, renderTree)
		paintVisible(ctx, renderTree, float64(ctx.Height()))
	}

	return nil
}

func getNodeContent(NodeDOM *hotdog.NodeDOM) string {
	return NodeDOM.Content
}

func getElementName(NodeDOM *hotdog.NodeDOM) string {
	return NodeDOM.Element
}

func getNodeChildren(NodeDOM *hotdog.NodeDOM) []*hotdog.NodeDOM {
	return NodeDOM.Children
}

func walkDOM(TreeDOM *hotdog.NodeDOM, d string) {
	fmt.Println(d, getElementName(TreeDOM))
	nodeChildren := getNodeChildren(TreeDOM)

	for i := 0; i < len(nodeChildren); i++ {
		walkDOM(nodeChildren[i], d+"-")
	}
}

// layoutOnly computes positions and sizes without painting
func layoutOnly(ctx *gg.Context, node *hotdog.NodeDOM, childIdx int) {
	if node == nil {
		return
	}
	if isWhitespaceOnly(node) {
		node.RenderBox = &hotdog.RenderBox{}
		return
	}
	if node.Style != nil && node.Style.Display == "none" {
		node.RenderBox = &hotdog.RenderBox{}
		return
	}
	nodeChildren := getNodeChildren(node)

	node.RenderBox = &hotdog.RenderBox{}
	calculateNode(ctx, node, childIdx)

	if isFlexContainer(node) {
		for i := 0; i < len(nodeChildren); i++ {
			nodeChildren[i].RenderBox = &hotdog.RenderBox{}
			calculateNode(ctx, nodeChildren[i], i)
			subChildren := getNodeChildren(nodeChildren[i])
			for j := 0; j < len(subChildren); j++ {
				layoutOnly(ctx, subChildren[j], j)
				if subChildren[j].RenderBox != nil {
					nodeChildren[i].RenderBox.Height += subChildren[j].RenderBox.Height +
						subChildren[j].RenderBox.PaddingTop + subChildren[j].RenderBox.PaddingBottom
					if subChildren[j].Style != nil {
						nodeChildren[i].RenderBox.Height += subChildren[j].Style.BorderTopWidth + subChildren[j].Style.BorderBottomWidth
					}
				}
			}
		}
		layoutFlexChildren(ctx, node)
	} else {
		for i := 0; i < len(nodeChildren); i++ {
			layoutOnly(ctx, nodeChildren[i], i)
			if nodeChildren[i].RenderBox != nil {
				childHeight := nodeChildren[i].RenderBox.Height +
					nodeChildren[i].RenderBox.PaddingTop + nodeChildren[i].RenderBox.PaddingBottom +
					nodeChildren[i].RenderBox.MarginTop + nodeChildren[i].RenderBox.MarginBottom
				if nodeChildren[i].Style != nil {
					childHeight += nodeChildren[i].Style.BorderTopWidth + nodeChildren[i].Style.BorderBottomWidth
				}
				node.RenderBox.Height += childHeight
			}
		}
	}
}

// paintVisible only paints nodes that overlap with the visible canvas area [0, canvasHeight]
func paintVisible(ctx *gg.Context, node *hotdog.NodeDOM, canvasHeight float64) {
	if node == nil || node.RenderBox == nil {
		return
	}
	if node.Style != nil && (node.Style.Display == "none" || node.Style.Visibility == "hidden") {
		return
	}
	if isWhitespaceOnly(node) {
		return
	}

	nodeTop := node.RenderBox.Top
	nodeBottom := nodeTop + node.RenderBox.Height +
		node.RenderBox.PaddingTop + node.RenderBox.PaddingBottom

	// If this node's entire subtree is below the canvas, skip it
	if nodeTop > canvasHeight {
		return
	}

	// Paint this node if it overlaps the visible area
	if nodeBottom >= 0 && nodeTop <= canvasHeight {
		paintNode(ctx, node)
	}

	// Recurse into children (they may be visible even if parent extends beyond)
	for _, child := range node.Children {
		paintVisible(ctx, child, canvasHeight)
	}
}

// layoutDOM is kept for compatibility but now just calls layoutOnly
func layoutDOM(ctx *gg.Context, node *hotdog.NodeDOM, childIdx int) {
	layoutOnly(ctx, node, childIdx)
}

func paintNode(ctx *gg.Context, node *hotdog.NodeDOM) {
	if node.Style == nil {
		return
	}
	if node.Style.Visibility == "hidden" || node.Style.Display == "none" {
		return
	}

	switch node.Style.Display {
	case "block", "flex", "grid", "inline-block":
		paintBlockElement(ctx, node)
	case "inline":
		paintInlineElement(ctx, node)
	case "list-item":
		paintListItemElement(ctx, node)
	}
}

func calculateNode(ctx *gg.Context, node *hotdog.NodeDOM, postion int) {
	if node.Style == nil {
		return
	}

	switch node.Style.Display {
	case "block", "flex", "grid", "inline-block":
		calculateBlockLayout(ctx, node, postion)
	case "inline":
		calculateInlineLayout(ctx, node, postion)
	case "list-item":
		calculateListItemLayout(ctx, node, postion)
	case "none":
		// Skip layout entirely
	}
}

func GetPageTitle(TreeDOM *hotdog.NodeDOM) string {
	nodeChildren := getNodeChildren(TreeDOM)
	pageTitle := "Sem Titulo"

	if getElementName(TreeDOM) == "title" {
		return getNodeContent(TreeDOM)
	}

	for i := 0; i < len(nodeChildren); i++ {
		nPageTitle := GetPageTitle(nodeChildren[i])

		if nPageTitle != "Sem Titulo" {
			pageTitle = nPageTitle
		}
	}

	return pageTitle
}
