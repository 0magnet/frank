package bun

import (
	"github.com/0magnet/frank/gg"
	hotdog "github.com/0magnet/frank/hotdog"
)

func createRenderTree(root *hotdog.NodeDOM) *hotdog.NodeDOM {
	if root.Style != nil && root.Style.Display == "none" {
		return nil
	}

	node := &hotdog.NodeDOM{
		Style:      root.Style,
		Element:    root.Element,
		Content:    root.Content,
		Attributes: root.Attributes,
	}

	node.RenderBox = &hotdog.RenderBox{}
	for _, child := range root.Children {
		r := createRenderTree(child)
		if r != nil {
			r.Parent = node
			node.Children = append(node.Children, r)
		}
	}

	return node
}

func layoutNode(ctx *gg.Context, node *hotdog.NodeDOM) {
	if node == nil || node.Style == nil {
		return
	}

	// Calculate this node's layout
	for i, child := range node.Children {
		child.RenderBox = &hotdog.RenderBox{}
		calculateNode(ctx, child, i)
		layoutNode(ctx, child)
		// Accumulate height
		if child.Style != nil && child.Style.Display != "inline" {
			node.RenderBox.Height += child.RenderBox.Height +
				child.RenderBox.PaddingTop + child.RenderBox.PaddingBottom +
				child.RenderBox.MarginTop + child.RenderBox.MarginBottom
		}
	}
}

func paintText(ctx *gg.Context, node *hotdog.NodeDOM) {
	if node == nil || node.Style == nil {
		return
	}

	paintNode(ctx, node)
	for _, child := range node.Children {
		paintText(ctx, child)
	}
}
