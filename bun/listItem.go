package bun

import (
	gg "github.com/0magnet/frank/gg"
	hotdog "github.com/0magnet/frank/hotdog"
)

func paintListItemElement(ctx *gg.Context, node *hotdog.NodeDOM) {
	// Draw background (padding box)
	bgLeft := node.RenderBox.Left - node.RenderBox.PaddingLeft
	bgTop := node.RenderBox.Top - node.RenderBox.PaddingTop
	bgWidth := node.RenderBox.Width + node.RenderBox.PaddingLeft + node.RenderBox.PaddingRight
	bgHeight := node.RenderBox.Height + node.RenderBox.PaddingTop + node.RenderBox.PaddingBottom

	if node.Style.BackgroundColor != nil && node.Style.BackgroundColor.A > 0 {
		ctx.DrawRectangle(bgLeft, bgTop, bgWidth, bgHeight)
		ctx.SetRGBA(node.Style.BackgroundColor.R, node.Style.BackgroundColor.G, node.Style.BackgroundColor.B, node.Style.BackgroundColor.A)
		ctx.Fill()
	}

	// Draw borders
	paintBorders(ctx, node, bgLeft, bgTop, bgWidth, bgHeight)

	// Draw bullet and text
	ctx.DrawCircle(node.RenderBox.Left-15, node.RenderBox.Top+node.Style.FontSize/2, 3)
	ctx.SetRGBA(node.Style.Color.R, node.Style.Color.G, node.Style.Color.B, node.Style.Color.A)
	ctx.Fill()

	if len(node.Content) > 0 {
		ctx.SetFont(sansSerif[node.Style.FontWeight], node.Style.FontSize)
		lineSpacing := 1.5
		if node.Style.LineHeight > 0 && node.Style.FontSize > 0 {
			lineSpacing = node.Style.LineHeight / node.Style.FontSize
		}
		ctx.DrawStringWrapped(node.Content, node.RenderBox.Left, node.RenderBox.Top+1, 0, 0, node.RenderBox.Width, lineSpacing, gg.AlignLeft)
		ctx.Fill()
	}
}

func calculateListItemLayout(ctx *gg.Context, node *hotdog.NodeDOM, childIdx int) {
	// Apply box model
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

	if node.Style.Width == 0 && node.Parent != nil {
		node.RenderBox.Width = node.Parent.RenderBox.Width - 30 -
			node.RenderBox.PaddingLeft - node.RenderBox.PaddingRight -
			node.Style.BorderLeftWidth - node.Style.BorderRightWidth
	} else if node.Style.Width > 0 {
		node.RenderBox.Width = node.Style.Width
	}

	if node.Style.Height == 0 && len(node.Content) > 0 {
		ctx.SetFont(sansSerif[node.Style.FontWeight], node.Style.FontSize)
		lineSpacing := 1.5
		if node.Style.LineHeight > 0 {
			lineSpacing = node.Style.LineHeight / node.Style.FontSize
		}
		node.RenderBox.Height = ctx.MeasureStringWrapped(node.Content, node.RenderBox.Width, lineSpacing) + 2 + ctx.FontHeight()*.5
	} else if node.Style.Height > 0 {
		node.RenderBox.Height = node.Style.Height
	}

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

	node.RenderBox.Left = 30 + node.RenderBox.MarginLeft + node.Style.BorderLeftWidth + node.RenderBox.PaddingLeft

	// Apply positioning
	applyPositioning(node)
}
