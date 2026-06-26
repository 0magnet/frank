package bun

import (
	gg "github.com/0magnet/frank/gg"
	hotdog "github.com/0magnet/frank/hotdog"
)

func isFlexContainer(node *hotdog.NodeDOM) bool {
	return node.Style != nil && node.Style.Display == "flex"
}

func calculateFlexLayout(ctx *gg.Context, node *hotdog.NodeDOM, childIdx int) {
	if node.Style.Width == 0 && node.Parent != nil {
		node.RenderBox.Width = node.Parent.RenderBox.Width -
			node.Style.MarginLeft - node.Style.MarginRight -
			node.Style.PaddingLeft - node.Style.PaddingRight -
			node.Style.BorderLeftWidth - node.Style.BorderRightWidth
	} else if node.Style.Width > 0 {
		node.RenderBox.Width = node.Style.Width
	}

	// Position within parent
	if childIdx > 0 && node.Parent != nil {
		prev := node.Parent.Children[childIdx-1]
		node.RenderBox.Top = prev.RenderBox.Top + prev.RenderBox.Height +
			prev.RenderBox.MarginBottom + node.RenderBox.MarginTop
	} else if node.Parent != nil {
		node.RenderBox.Top = node.Parent.RenderBox.Top + node.Parent.RenderBox.PaddingTop + node.Parent.Style.BorderTopWidth
	}

	node.RenderBox.Left = node.Style.MarginLeft + node.Style.BorderLeftWidth + node.Style.PaddingLeft
	if node.Parent != nil {
		node.RenderBox.Left += node.Parent.RenderBox.Left + node.Parent.RenderBox.PaddingLeft + node.Parent.Style.BorderLeftWidth
	}
}

func layoutFlexChildren(ctx *gg.Context, node *hotdog.NodeDOM) {
	if node.Style == nil || !isFlexContainer(node) {
		return
	}

	direction := node.Style.FlexDirection
	if direction == "" {
		direction = "row"
	}

	children := node.Children
	if len(children) == 0 {
		return
	}

	contentLeft := node.RenderBox.Left + node.RenderBox.PaddingLeft + node.Style.BorderLeftWidth
	contentTop := node.RenderBox.Top + node.RenderBox.PaddingTop + node.Style.BorderTopWidth
	contentWidth := node.RenderBox.Width - node.RenderBox.PaddingLeft - node.RenderBox.PaddingRight -
		node.Style.BorderLeftWidth - node.Style.BorderRightWidth
	contentHeight := node.RenderBox.Height - node.RenderBox.PaddingTop - node.RenderBox.PaddingBottom -
		node.Style.BorderTopWidth - node.Style.BorderBottomWidth

	// Calculate total flex-grow
	totalGrow := 0.0
	totalFixedSize := 0.0
	for _, child := range children {
		if child.Style == nil || child.Style.Display == "none" {
			continue
		}
		if child.Style.FlexGrow > 0 {
			totalGrow += child.Style.FlexGrow
		} else {
			if direction == "row" {
				totalFixedSize += child.RenderBox.Width + child.RenderBox.MarginLeft + child.RenderBox.MarginRight
			} else {
				totalFixedSize += child.RenderBox.Height + child.RenderBox.MarginTop + child.RenderBox.MarginBottom
			}
		}
	}

	var availableSpace float64
	if direction == "row" {
		availableSpace = contentWidth - totalFixedSize
	} else {
		availableSpace = contentHeight - totalFixedSize
	}
	if availableSpace < 0 {
		availableSpace = 0
	}

	offset := 0.0

	// Calculate spacing for justify-content
	numChildren := 0
	for _, child := range children {
		if child.Style != nil && child.Style.Display != "none" {
			numChildren++
		}
	}

	gap := 0.0
	startOffset := 0.0
	if totalGrow == 0 && numChildren > 0 {
		switch node.Style.JustifyContent {
		case "center":
			startOffset = availableSpace / 2
		case "flex-end":
			startOffset = availableSpace
		case "space-between":
			if numChildren > 1 {
				gap = availableSpace / float64(numChildren-1)
			}
		case "space-around":
			gap = availableSpace / float64(numChildren)
			startOffset = gap / 2
		case "space-evenly":
			gap = availableSpace / float64(numChildren+1)
			startOffset = gap
		}
	}
	offset = startOffset

	childIdx := 0
	for _, child := range children {
		if child.Style == nil || child.Style.Display == "none" {
			continue
		}

		// Distribute flex-grow space
		if totalGrow > 0 && child.Style.FlexGrow > 0 {
			extraSize := (child.Style.FlexGrow / totalGrow) * availableSpace
			if direction == "row" {
				child.RenderBox.Width += extraSize
			} else {
				child.RenderBox.Height += extraSize
			}
		}

		if direction == "row" {
			child.RenderBox.Left = contentLeft + offset + child.RenderBox.MarginLeft
			child.RenderBox.Top = contentTop + child.RenderBox.MarginTop

			// align-items
			switch node.Style.AlignItems {
			case "center":
				childH := child.RenderBox.Height + child.RenderBox.MarginTop + child.RenderBox.MarginBottom
				child.RenderBox.Top = contentTop + (contentHeight-childH)/2 + child.RenderBox.MarginTop
			case "flex-end":
				child.RenderBox.Top = contentTop + contentHeight - child.RenderBox.Height - child.RenderBox.MarginBottom
			case "stretch":
				if child.Style.Height == 0 {
					child.RenderBox.Height = contentHeight - child.RenderBox.MarginTop - child.RenderBox.MarginBottom
				}
			}

			offset += child.RenderBox.Width + child.RenderBox.MarginLeft + child.RenderBox.MarginRight + gap
		} else {
			child.RenderBox.Top = contentTop + offset + child.RenderBox.MarginTop
			child.RenderBox.Left = contentLeft + child.RenderBox.MarginLeft

			// align-items (cross-axis in column is horizontal)
			switch node.Style.AlignItems {
			case "center":
				childW := child.RenderBox.Width + child.RenderBox.MarginLeft + child.RenderBox.MarginRight
				child.RenderBox.Left = contentLeft + (contentWidth-childW)/2 + child.RenderBox.MarginLeft
			case "flex-end":
				child.RenderBox.Left = contentLeft + contentWidth - child.RenderBox.Width - child.RenderBox.MarginRight
			case "stretch":
				if child.Style.Width == 0 {
					child.RenderBox.Width = contentWidth - child.RenderBox.MarginLeft - child.RenderBox.MarginRight
				}
			}

			offset += child.RenderBox.Height + child.RenderBox.MarginTop + child.RenderBox.MarginBottom + gap
		}
		childIdx++
	}

	// Update container height if auto
	if node.Style.Height == 0 {
		if direction == "column" {
			node.RenderBox.Height = offset + node.RenderBox.PaddingTop + node.RenderBox.PaddingBottom +
				node.Style.BorderTopWidth + node.Style.BorderBottomWidth
		} else {
			maxChildH := 0.0
			for _, child := range children {
				if child.Style != nil && child.Style.Display != "none" {
					childBottom := child.RenderBox.Top + child.RenderBox.Height + child.RenderBox.MarginBottom - contentTop
					if childBottom > maxChildH {
						maxChildH = childBottom
					}
				}
			}
			node.RenderBox.Height = maxChildH + node.RenderBox.PaddingTop + node.RenderBox.PaddingBottom +
				node.Style.BorderTopWidth + node.Style.BorderBottomWidth
		}
	}
}
