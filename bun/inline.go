package bun

import (
	"bytes"
	"fmt"
	"image"
	"path"
	"strings"

	gg "github.com/0magnet/frank/gg"
	hotdog "github.com/0magnet/frank/hotdog"
	"github.com/0magnet/frank/sauce"
)

func paintInlineElement(ctx *gg.Context, node *hotdog.NodeDOM) {
	// Draw background (padding box) only if visible
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

	if node.Element == "img" {
		im, err := fetchNodeImage(node)
		if err == nil && im != nil {
			ctx.DrawImage(im, int(node.RenderBox.Left), int(node.RenderBox.Top))
		}
		// Silently skip images that can't be decoded
	}

	if len(node.Content) > 0 {
		ctx.SetRGBA(node.Style.Color.R, node.Style.Color.G, node.Style.Color.B, node.Style.Color.A)
		ctx.SetFont(sansSerif[node.Style.FontWeight], node.Style.FontSize)
		ctx.DrawStringWrapped(node.Content, node.RenderBox.Left, node.RenderBox.Top, 0, 0, node.RenderBox.Width, 1, gg.AlignLeft)
		ctx.Fill()
	}
}

// unsupportedImageFormats lists extensions Go's image package can't decode
var unsupportedImageFormats = map[string]bool{
	".svg": true, ".webp": true, ".avif": true, ".bmp": true, ".tiff": true,
}

func isUnsupportedImageFormat(src string) bool {
	ext := strings.ToLower(path.Ext(src))
	return unsupportedImageFormats[ext]
}

func fetchNodeImage(node *hotdog.NodeDOM) (image.Image, error) {
	imgPath := node.Attr("src")

	// Skip unsupported formats entirely
	if isUnsupportedImageFormat(imgPath) {
		return nil, fmt.Errorf("unsupported image format: %s", path.Ext(imgPath))
	}

	imgURL, err := node.Document.URL.Parse(imgPath)
	if err != nil {
		return nil, err
	}

	data, err := sauce.GetImage(imgURL)
	if err != nil {
		return nil, err
	}
	im, _, err := image.Decode(bytes.NewReader(data))

	if err != nil {
		return nil, err
	}
	return im, nil
}

func calculateInlineLayout(ctx *gg.Context, node *hotdog.NodeDOM, childIdx int) {
	ctx.SetFont(sansSerif[node.Style.FontWeight], node.Style.FontSize)

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

	if childIdx > 0 && node.Parent.Children[childIdx-1] != nil {
		prev := node.Parent.Children[childIdx-1]
		if prev.Style.Display == "inline" {
			proposedLeft := prev.RenderBox.Left + prev.RenderBox.Width +
				prev.RenderBox.PaddingRight + prev.Style.BorderRightWidth +
				prev.RenderBox.MarginRight + node.RenderBox.MarginLeft +
				node.Style.BorderLeftWidth + node.RenderBox.PaddingLeft

			// Word wrap: if this inline element would exceed parent width, wrap to next line
			availableWidth := node.Parent.RenderBox.Width
			if node.Parent != nil {
				availableWidth = node.Parent.RenderBox.Left + node.Parent.RenderBox.Width - node.Parent.RenderBox.PaddingRight
			}

			mW, _ := ctx.MeasureString(node.Content)
			if proposedLeft+mW > availableWidth && node.Element != "img" {
				// Wrap to next line
				node.RenderBox.Top = prev.RenderBox.Top + prev.RenderBox.Height +
					prev.RenderBox.PaddingTop + prev.RenderBox.PaddingBottom +
					prev.RenderBox.MarginBottom + node.RenderBox.MarginTop
				node.RenderBox.Left = node.Parent.RenderBox.Left +
					node.Parent.RenderBox.PaddingLeft + node.Parent.Style.BorderLeftWidth +
					node.RenderBox.MarginLeft + node.Style.BorderLeftWidth + node.RenderBox.PaddingLeft
			} else {
				node.RenderBox.Top = prev.RenderBox.Top
				node.RenderBox.Left = proposedLeft
			}
		} else {
			node.RenderBox.Top = prev.RenderBox.Top + prev.RenderBox.Height +
				prev.RenderBox.PaddingTop + prev.RenderBox.PaddingBottom +
				prev.Style.BorderTopWidth + prev.Style.BorderBottomWidth +
				prev.RenderBox.MarginBottom + node.RenderBox.MarginTop
			if node.Parent != nil {
				node.RenderBox.Left = node.Parent.RenderBox.Left +
					node.Parent.RenderBox.PaddingLeft + node.Parent.Style.BorderLeftWidth +
					node.RenderBox.MarginLeft + node.Style.BorderLeftWidth + node.RenderBox.PaddingLeft
			}
		}
	} else {
		if node.Parent != nil {
			node.RenderBox.Top = node.Parent.RenderBox.Top +
				node.Parent.RenderBox.PaddingTop + node.Parent.Style.BorderTopWidth +
				node.RenderBox.MarginTop
			node.RenderBox.Left = node.Parent.RenderBox.Left +
				node.Parent.RenderBox.PaddingLeft + node.Parent.Style.BorderLeftWidth +
				node.RenderBox.MarginLeft + node.Style.BorderLeftWidth + node.RenderBox.PaddingLeft
		}
	}

	if node.Element == "img" {
		im, err := fetchNodeImage(node)
		if err == nil && im != nil {
			imgSize := im.Bounds().Size()
			node.RenderBox.Width = float64(imgSize.X)
			node.RenderBox.Height = float64(imgSize.Y)
		}
		// Undecoded images collapse to zero size
	} else {
		if node.RenderBox.Width == 0 && node.Parent != nil {
			node.RenderBox.Width = node.Parent.RenderBox.Width -
				node.RenderBox.PaddingLeft - node.RenderBox.PaddingRight -
				node.Style.BorderLeftWidth - node.Style.BorderRightWidth
		}

		node.RenderBox.Height = ctx.MeasureStringWrapped(node.Content, node.RenderBox.Width, 1)
		mW, _ := ctx.MeasureString(node.Content)
		if mW < node.RenderBox.Width {
			node.RenderBox.Width = mW
		}
	}

	node.RenderBox.Height++

	// Apply positioning
	applyPositioning(node)
}
