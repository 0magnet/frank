package bun

import (
	gg "github.com/0magnet/frank/gg"
	hotdog "github.com/0magnet/frank/hotdog"
)

func paintBlockElement(ctx *gg.Context, node *hotdog.NodeDOM) {
	// Draw background (padding box)
	bgLeft := node.RenderBox.Left - node.RenderBox.PaddingLeft
	bgTop := node.RenderBox.Top - node.RenderBox.PaddingTop
	bgWidth := node.RenderBox.Width + node.RenderBox.PaddingLeft + node.RenderBox.PaddingRight
	bgHeight := node.RenderBox.Height + node.RenderBox.PaddingTop + node.RenderBox.PaddingBottom

	// Only draw background if it has visible alpha
	if node.Style.BackgroundColor != nil && node.Style.BackgroundColor.A > 0 {
		ctx.DrawRectangle(bgLeft, bgTop, bgWidth, bgHeight)
		ctx.SetRGBA(node.Style.BackgroundColor.R, node.Style.BackgroundColor.G, node.Style.BackgroundColor.B, node.Style.BackgroundColor.A)
		ctx.Fill()
	}

	// Draw borders
	paintBorders(ctx, node, bgLeft, bgTop, bgWidth, bgHeight)

	// Draw text only if there's content
	if len(node.Content) > 0 {
		ctx.SetRGBA(node.Style.Color.R, node.Style.Color.G, node.Style.Color.B, node.Style.Color.A)
		ctx.SetFont(sansSerif[node.Style.FontWeight], node.Style.FontSize)
		lineSpacing := 1.5
		if node.Style.LineHeight > 0 && node.Style.FontSize > 0 {
			lineSpacing = node.Style.LineHeight / node.Style.FontSize
		}
		ctx.DrawStringWrapped(node.Content, node.RenderBox.Left, node.RenderBox.Top+1, 0, 0, node.RenderBox.Width, lineSpacing, gg.AlignLeft)
		ctx.Fill()
	}
}

func paintBorders(ctx *gg.Context, node *hotdog.NodeDOM, bgLeft, bgTop, bgWidth, bgHeight float64) {
	if node.Style == nil {
		return
	}

	// Top border
	if node.Style.BorderTopWidth > 0 && node.Style.BorderTopStyle != "none" {
		borderColor := node.Style.BorderTopColor
		if borderColor == nil {
			borderColor = node.Style.Color
		}
		if borderColor != nil {
			ctx.SetRGBA(borderColor.R, borderColor.G, borderColor.B, borderColor.A)
			ctx.DrawRectangle(bgLeft-node.Style.BorderLeftWidth, bgTop-node.Style.BorderTopWidth,
				bgWidth+node.Style.BorderLeftWidth+node.Style.BorderRightWidth, node.Style.BorderTopWidth)
			ctx.Fill()
		}
	}

	// Bottom border
	if node.Style.BorderBottomWidth > 0 && node.Style.BorderBottomStyle != "none" {
		borderColor := node.Style.BorderBottomColor
		if borderColor == nil {
			borderColor = node.Style.Color
		}
		if borderColor != nil {
			ctx.SetRGBA(borderColor.R, borderColor.G, borderColor.B, borderColor.A)
			ctx.DrawRectangle(bgLeft-node.Style.BorderLeftWidth, bgTop+bgHeight,
				bgWidth+node.Style.BorderLeftWidth+node.Style.BorderRightWidth, node.Style.BorderBottomWidth)
			ctx.Fill()
		}
	}

	// Left border
	if node.Style.BorderLeftWidth > 0 && node.Style.BorderLeftStyle != "none" {
		borderColor := node.Style.BorderLeftColor
		if borderColor == nil {
			borderColor = node.Style.Color
		}
		if borderColor != nil {
			ctx.SetRGBA(borderColor.R, borderColor.G, borderColor.B, borderColor.A)
			ctx.DrawRectangle(bgLeft-node.Style.BorderLeftWidth, bgTop,
				node.Style.BorderLeftWidth, bgHeight)
			ctx.Fill()
		}
	}

	// Right border
	if node.Style.BorderRightWidth > 0 && node.Style.BorderRightStyle != "none" {
		borderColor := node.Style.BorderRightColor
		if borderColor == nil {
			borderColor = node.Style.Color
		}
		if borderColor != nil {
			ctx.SetRGBA(borderColor.R, borderColor.G, borderColor.B, borderColor.A)
			ctx.DrawRectangle(bgLeft+bgWidth, bgTop,
				node.Style.BorderRightWidth, bgHeight)
			ctx.Fill()
		}
	}
}

func calculateBlockLayout(ctx *gg.Context, node *hotdog.NodeDOM, childIdx int) {
	// Apply box model margins/padding to render box
	if node.Style != nil {
		node.RenderBox.MarginTop = node.Style.MarginTop
		node.RenderBox.MarginRight = node.Style.MarginRight
		node.RenderBox.MarginBottom = node.Style.MarginBottom
		node.RenderBox.MarginLeft = node.Style.MarginLeft
		node.RenderBox.PaddingTop = node.Style.PaddingTop
		node.RenderBox.PaddingRight = node.Style.PaddingRight
		node.RenderBox.PaddingBottom = node.Style.PaddingBottom
		node.RenderBox.PaddingLeft = node.Style.PaddingLeft
	}

	// Content width = parent width - margin - padding - border
	if node.Style.Width == 0 {
		if node.Parent != nil {
			node.RenderBox.Width = node.Parent.RenderBox.Width -
				node.RenderBox.MarginLeft - node.RenderBox.MarginRight -
				node.RenderBox.PaddingLeft - node.RenderBox.PaddingRight -
				node.Style.BorderLeftWidth - node.Style.BorderRightWidth
		}
	} else {
		node.RenderBox.Width = node.Style.Width
	}

	if node.RenderBox.Width < 0 {
		node.RenderBox.Width = 0
	}

	// Height
	if node.Style.Height == 0 {
		if len(node.Content) > 0 && node.RenderBox.Width > 0 {
			ctx.SetFont(sansSerif[node.Style.FontWeight], node.Style.FontSize)
			lineSpacing := 1.5
			if node.Style.LineHeight > 0 && node.Style.FontSize > 0 {
				lineSpacing = node.Style.LineHeight / node.Style.FontSize
			}
			node.RenderBox.Height = ctx.MeasureStringWrapped(node.Content, node.RenderBox.Width, lineSpacing) + 2 + ctx.FontHeight()*.5
		}
		// Height 0 for container nodes — will be computed from children
	} else {
		node.RenderBox.Height = node.Style.Height
	}

	// Position
	if childIdx > 0 && node.Parent != nil {
		prev := node.Parent.Children[childIdx-1]
		node.RenderBox.Top = prev.RenderBox.Top + prev.RenderBox.Height +
			prev.RenderBox.PaddingTop + prev.RenderBox.PaddingBottom +
			prev.Style.BorderTopWidth + prev.Style.BorderBottomWidth +
			prev.RenderBox.MarginBottom + node.RenderBox.MarginTop
	} else if node.Parent != nil {
		node.RenderBox.Top = node.Parent.RenderBox.Top +
			node.Parent.RenderBox.PaddingTop + node.Parent.Style.BorderTopWidth +
			node.RenderBox.MarginTop
	}

	// Left position includes parent's content area offset
	if node.Parent != nil {
		node.RenderBox.Left = node.Parent.RenderBox.Left +
			node.Parent.RenderBox.PaddingLeft + node.Parent.Style.BorderLeftWidth +
			node.RenderBox.MarginLeft + node.Style.BorderLeftWidth + node.RenderBox.PaddingLeft
	} else {
		node.RenderBox.Left = node.RenderBox.MarginLeft + node.Style.BorderLeftWidth + node.RenderBox.PaddingLeft
	}

	// Apply positioning offset
	applyPositioning(node)
}
