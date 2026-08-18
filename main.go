// Command frank is a native web browser built on an embedded WebKitGTK-6.0
// (GTK4) engine, so it renders modern JS/WebAssembly/WebGL/CSS at full fidelity.
//
// This is the WebKit-GTK4 line of Frank. The original pure-Go (GLFW/OpenGL)
// renderer is preserved on the `pure-go` branch; its engine packages
// (mustard/bun/mayo/ketchup/gg) remain in this module but are no longer wired to
// the entry point.
package main

/*
#cgo pkg-config: gtk4 webkitgtk-6.0
#include <stdlib.h>
#include "browser.h"
*/
import "C"

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"unsafe"

	"github.com/0magnet/calvin"
	cc "github.com/ivanpirog/coloredcobra"
	"github.com/spf13/cobra"
)

func setIfUnset(k, v string) {
	if os.Getenv(k) == "" {
		os.Setenv(k, v)
	}
}

var skynet bool

func init() {
	RootCmd.Flags().BoolVarP(&skynet, "skynet", "s", false, "start and manage a resolving proxy for the browser's lifetime")
	var helpflag bool
	RootCmd.SetUsageTemplate(help)
	RootCmd.PersistentFlags().BoolVarP(&helpflag, "help", "h", false, "help for "+RootCmd.Use)
	RootCmd.SetHelpCommand(&cobra.Command{Hidden: true})
	RootCmd.PersistentFlags().MarkHidden("help") //nolint
}

// RootCmd is the root command
var RootCmd = &cobra.Command{
	Use:                   "frank [url]",
	Short:                 "a native browser on embedded WebKitGTK",
	Long:                  calvin.AsciiFont("frank") + "\na native browser on embedded WebKitGTK",
	Args:                  cobra.MaximumNArgs(1),
	SilenceErrors:         true,
	SilenceUsage:          true,
	DisableSuggestions:    true,
	DisableFlagsInUseLine: true,
	Run: func(_ *cobra.Command, args []string) {
		url := "" // empty -> the configured homepage (or built-in welcome), chosen C-side
		if len(args) > 0 {
			url = args[0]
		}
		run(url)
	},
}

func main() {
	// GTK/WebKit must run on the main thread. Cobra runs the command on this
	// same goroutine, so locking here still covers the C call.
	runtime.LockOSThread()

	cc.Init(&cc.Config{
		RootCmd:         RootCmd,
		Headings:        cc.HiBlue + cc.Bold,
		Commands:        cc.HiBlue + cc.Bold,
		CmdShortDescr:   cc.HiBlue,
		Example:         cc.HiBlue + cc.Italic,
		ExecName:        cc.HiBlue + cc.Bold,
		Flags:           cc.HiBlue + cc.Bold,
		FlagsDescr:      cc.HiBlue,
		NoExtraNewlines: true,
		NoBottomNewline: true,
	})
	if err := RootCmd.Execute(); err != nil {
		log.Fatal("Failed to execute command: ", err)
	}
}

const help = "{{if .HasAvailableSubCommands}}{{end}} {{if gt (len .Aliases) 0}}\r\n\r\n" +
	"{{.NameAndAliases}}{{end}}{{if .HasAvailableSubCommands}}" +
	"Available Commands:{{range .Commands}}  {{if and (ne .Name \"completion\") .IsAvailableCommand}}\r\n  " +
	"{{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}\r\n\r\n" +
	"Flags:\r\n" +
	"{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}\r\n\r\n" +
	"Global Flags:\r\n" +
	"{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}\r\n\r\n"

func run(url string) {
	// Old-GPU compatibility profile. WebKit's fast dmabuf zero-copy compositing
	// path can't be presented by some older drivers (e.g. Gen7 Intel / GLES 3.0),
	// where it renders one frame then freezes. Default is the fast path (correct
	// on modern GPUs); set FRANK_GL_COMPAT=1 to force the stable readback path
	// (slower but freeze-free) on affected hardware.
	if os.Getenv("FRANK_GL_COMPAT") == "1" {
		setIfUnset("MESA_EXTENSION_OVERRIDE", "+GL_KHR_robustness")
		setIfUnset("WEBKIT_DISABLE_DMABUF_RENDERER", "1")
	}

	// Two ways to reach the Skywire network, both wired through WebKit's proxy:
	//   FRANK_PROXY set -> bring-your-own: a visor that spawns Frank (or an
	//                      already-running one) passes its resolving-proxy URI and
	//                      owns its own lifecycle (it survives Frank closing).
	//   --skynet        -> Frank starts and manages a resolving proxy on an
	//                      ephemeral loopback port for the browser's lifetime
	//                      (stand-in for the in-process routable visor to come).
	proxy := os.Getenv("FRANK_PROXY")
	if proxy == "" && skynet {
		p, cleanup, err := startManagedProxy()
		if err != nil {
			fmt.Fprintln(os.Stderr, "frank: skynet:", err)
			os.Exit(1)
		}
		defer cleanup()
		proxy = p
	}

	// Run the real wasm-visor under the hood: start `hv serve` and let the anchor
	// boot it. Skip with FRANK_NO_VISOR=1 (anchor uses the SharedWorker stub).
	if os.Getenv("FRANK_NO_VISOR") == "" {
		if vurl, cleanup, err := startVisorServer(); err == nil {
			os.Setenv("FRANK_VISOR_URL", vurl)
			defer cleanup()
		} else {
			fmt.Fprintln(os.Stderr, "frank: wasm-visor:", err, "(anchor uses stub)")
		}
	}

	curl := C.CString(url)
	cproxy := C.CString(proxy)
	defer C.free(unsafe.Pointer(curl))
	defer C.free(unsafe.Pointer(cproxy))
	C.frank_run(curl, cproxy)
}
