package hotdog

import (
	"fmt"
	"github.com/0magnet/frank/mustard"
	profiler "github.com/0magnet/frank/profiler"
	"net/url"
)

type WebBrowser struct {
	ActiveDocument *Document
	Documents      []*Document

	Viewport    *mustard.CanvasWidget
	StatusLabel *mustard.LabelWidget
	History     *History
	Window      *mustard.Window
	Profiler    *profiler.Profiler
	BuildInfo   *BuildInfo
	Settings    *Settings
}

type Document struct {
	Title       string
	ContentType string
	URL         *url.URL

	RawDocument string
	DOM         *NodeDOM

	DebugFlag       bool
	DebugWindow     *mustard.Window
	DebugTree       *mustard.TreeWidget
	SelectedElement *NodeDOM

	OffsetY int
}
type BuildInfo struct {
	GitRevision string
	GitBranch   string

	HostInfo  string
	BuildTime string
}

type History struct {
	previousPages []*url.URL
	nextPages     []*url.URL
}

func (history *History) NextPages() []*url.URL {
	return history.nextPages
}

func (history *History) AllPages() []*url.URL {
	return history.previousPages
}

func (history *History) PageCount() int {
	return len(history.previousPages)
}

func (history *History) Push(URL *url.URL) {
	history.nextPages = nil
	history.previousPages = append(history.previousPages, URL)
}

func (history *History) Last() *url.URL {
	return history.previousPages[len(history.previousPages)-1]
}

func (history *History) PopNext() {
	if len(history.nextPages) > 0 {
		history.previousPages = append(history.previousPages, history.nextPages[len(history.nextPages)-1])
		history.nextPages = nil
	}
}

func (history *History) Pop() {
	if len(history.previousPages) > 0 {
		history.nextPages = append(history.nextPages, history.previousPages[len(history.previousPages)-1])
		history.previousPages = history.previousPages[:len(history.previousPages)-1]
	}
}

type RenderBox struct {
	Node *NodeDOM

	Top  float64
	Left float64

	Width  float64
	Height float64

	MarginTop    float64
	MarginLeft   float64
	MarginRight  float64
	MarginBottom float64

	PaddingTop    float64
	PaddingLeft   float64
	PaddingRight  float64
	PaddingBottom float64
}

func (box *RenderBox) GetRect() (float64, float64, float64, float64) {
	return box.Top, box.Left, box.Width, box.Height
}

type NoSuchElementError string

func (e NoSuchElementError) Error() string {
	return fmt.Sprintf("no such element: %q", string(e))
}

//Resource "HTTP resource struct definition"
type Resource struct {
	Body        string
	ContentType string
	Code        int
	URL         *url.URL
	Key         string
}

//Attribute "Generic key:value attribute definition"
type Attribute struct {
	Name  string
	Value string
}

//Stylesheet "Stylesheet definition for DOM Nodes"
type Stylesheet struct {
	Color           *ColorRGBA
	BackgroundColor *ColorRGBA

	FontSize   float64
	FontWeight int

	Display  string
	Position string

	Width  float64
	Height float64
	Top    float64
	Left   float64

	// Box model
	MarginTop    float64
	MarginRight  float64
	MarginBottom float64
	MarginLeft   float64

	PaddingTop    float64
	PaddingRight  float64
	PaddingBottom float64
	PaddingLeft   float64

	BorderTopWidth    float64
	BorderRightWidth  float64
	BorderBottomWidth float64
	BorderLeftWidth   float64

	BorderTopColor    *ColorRGBA
	BorderRightColor  *ColorRGBA
	BorderBottomColor *ColorRGBA
	BorderLeftColor   *ColorRGBA

	BorderTopStyle    string
	BorderRightStyle  string
	BorderBottomStyle string
	BorderLeftStyle   string

	// Text properties
	TextAlign      string
	TextDecoration string
	LineHeight     float64
	FontFamily     string
	FontStyle      string
	WhiteSpace     string
	TextTransform  string

	// Layout properties
	MinWidth  float64
	MaxWidth  float64
	MinHeight float64
	MaxHeight float64
	Overflow  string
	Float     string
	Clear     string

	// Visual properties
	Opacity       float64
	Visibility    string
	ListStyleType string

	// Positioning
	Right  float64
	Bottom float64
	ZIndex int

	// Flexbox
	FlexDirection  string
	JustifyContent string
	AlignItems     string
	FlexGrow       float64
	FlexShrink     float64
	FlexBasis      float64
}

// StyleRule represents a parsed CSS rule with selector and properties
type StyleRule struct {
	Selector   string
	Properties map[string]string
	Specificity [4]int
}

//StyleElement "hmtl <style> element"
type StyleElement struct {
	Selector string
	Style    *Stylesheet
}

//ColorRGBA "RGBA color model"
type ColorRGBA struct {
	R float64
	G float64
	B float64
	A float64
}

type ResourceCache struct {
	cachedResources []*Resource
}

func (cache *ResourceCache) AddResource(resource *Resource) {
	cache.cachedResources = append(cache.cachedResources, resource)
}

func (cache *ResourceCache) GetResource(resourceKey string) *Resource {
	for _, resource := range cache.cachedResources {
		if resource.Key == resourceKey {
			return resource
		}
	}

	return nil
}

type CachedImage struct {
	Key   string
	Image []byte
}
type ImgCache struct {
	cachedImages []*CachedImage
}

func (cache *ImgCache) AddImage(key string, value []byte) {
	cache.cachedImages = append(cache.cachedImages,
		&CachedImage{
			Key:   key,
			Image: value,
		},
	)
}

func (cache *ImgCache) GetImage(imageKey string) *CachedImage {
	for _, image := range cache.cachedImages {
		if image.Key == imageKey {
			return image
		}
	}

	return nil
}

func Log(component, msg string) {
	str := "(" + "\033[95m" + component + "\033[0m" + ")"
	fmt.Println(str, msg)
}
