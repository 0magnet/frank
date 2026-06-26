package mustard

import (
	"os"
	"runtime"
	"strconv"
	"testing"

	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

// TestMustard is an interactive GUI demo, not an automated test: it opens a
// GLFW window and spins an infinite event loop that never returns. It is
// skipped by default so `go test ./...` stays headless and terminates. Run it
// on demand with: MUSTARD_GUI=1 go test ./mustard -run TestMustard
func TestMustard(t *testing.T) {
	if os.Getenv("MUSTARD_GUI") == "" {
		t.Skip("interactive GUI demo; set MUSTARD_GUI=1 to run")
	}
	runtime.LockOSThread()
	glfw.Init()
	gl.Init()

	SetGLFWHints()

	app := CreateNewApp("Frank")
	window := CreateNewWindow("Frank", 600, 600, true)
	rootFrame := CreateFrame(HorizontalFrame)

	appBar := CreateFrame(VerticalFrame)

	titleBar := CreateLabelWidget("Frank - nil")
	titleBar.SetFontColor("#fff")

	appBar.SetHeight(28)
	appBar.AttachWidget(titleBar)
	appBar.SetBackgroundColor("#5f6368")

	rootFrame.AttachWidget(appBar)

	viewPort := CreateCanvasWidget(func(canvas *CanvasWidget) {})

	rootFrame.AttachWidget(viewPort)

	statusBar := CreateFrame(HorizontalFrame)
	statusBar.SetBackgroundColor("#babcbe")
	statusBar.SetHeight(20)

	statusLabel := CreateLabelWidget("Processed Events:")
	statusLabel.SetFontSize(16)
	frameEvents := 0

	rootFrame.AttachWidget(statusBar)
	statusBar.AttachWidget(statusLabel)

	window.SetRootFrame(rootFrame)

	app.AddWindow(window)

	window.Show()
	app.Run(func() {
		frameEvents++
		statusLabel.SetContent("Processed Events: " + strconv.Itoa(frameEvents) + "; Resolution: " + strconv.Itoa(window.width) + "X" + strconv.Itoa(window.height))
		//window.RequestRepaint()
	})
}
