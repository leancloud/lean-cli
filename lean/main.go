package main

import (
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/getsentry/raven-go"
	"github.com/leancloud/lean-cli/commands"
	"github.com/leancloud/lean-cli/stats"
	"github.com/leancloud/lean-cli/version"
)

func run() {
	if len(os.Args) >= 2 && os.Args[1] == "--_collect-stats" {
		err := stats.Init("Rp8mUcQBVObk8EuyVMDPv39U-gzGzoHsz", "9g3bs563vEsOGdycO2E9ly0y")
		if err != nil {
			raven.CaptureError(err, nil, nil)
		}
		stats.Client.AppVersion = version.Version
		stats.Client.AppChannel = pkgType

		var event string
		if len(os.Args) >= 3 {
			event = os.Args[2]
		}

		stats.Collect([]stats.Event{
			{
				Event: event,
			},
		})
		return
	}

	// disable the log prefix
	log.SetFlags(0)

	go func() {
		_ = checkUpdate()
	}()

	commands.Run(os.Args)
}

func init() {
	err := raven.SetDSN("https://985d436efdb544c49e9389e59724ddce:6a831597d45b4309923f2567bbe7db82@sentry.lean.sh/9")
	if err != nil {
		panic(err)
	}
}

func main() {
	if os.Getenv("LEAN_CLI_DEBUG") == "1" {
		run()
		return
	}

	raven.SetTagsContext(map[string]string{
		"version": version.Version,
		"OS":      runtime.GOOS,
		"arch":    runtime.GOARCH,
	})
	err, id := raven.CapturePanicAndWait(run, nil)
	if err != nil {
		fmt.Printf("panic: %s, 错误 ID: %s\r\n", err, id)
		os.Exit(1)
	}
}
