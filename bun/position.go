package bun

import (
	hotdog "github.com/0magnet/frank/hotdog"
)

// applyPositioning adjusts a node's position based on CSS position property
func applyPositioning(node *hotdog.NodeDOM) {
	if node.Style == nil || node.RenderBox == nil {
		return
	}

	switch node.Style.Position {
	case "relative":
		node.RenderBox.Top += node.Style.Top
		node.RenderBox.Left += node.Style.Left

	case "absolute":
		// Position relative to nearest positioned ancestor
		ancestor := findPositionedAncestor(node)
		if ancestor != nil && ancestor.RenderBox != nil {
			node.RenderBox.Top = ancestor.RenderBox.Top + node.Style.Top
			node.RenderBox.Left = ancestor.RenderBox.Left + node.Style.Left
		} else {
			node.RenderBox.Top = node.Style.Top
			node.RenderBox.Left = node.Style.Left
		}
		if node.Style.Width > 0 {
			node.RenderBox.Width = node.Style.Width
		}
		if node.Style.Height > 0 {
			node.RenderBox.Height = node.Style.Height
		}

	case "fixed":
		// Position relative to viewport (same as absolute for now since we don't have viewport tracking)
		node.RenderBox.Top = node.Style.Top
		node.RenderBox.Left = node.Style.Left
		if node.Style.Width > 0 {
			node.RenderBox.Width = node.Style.Width
		}
		if node.Style.Height > 0 {
			node.RenderBox.Height = node.Style.Height
		}
	}
}

func findPositionedAncestor(node *hotdog.NodeDOM) *hotdog.NodeDOM {
	parent := node.Parent
	for parent != nil {
		if parent.Style != nil {
			pos := parent.Style.Position
			if pos == "relative" || pos == "absolute" || pos == "fixed" {
				return parent
			}
		}
		parent = parent.Parent
	}
	return nil
}
