package version

import (
	"github.com/aisk/logp"
)

// Version is lean-cli's version.
const Version = "0.24.4"

var Distribution = "lean"

func PrintCurrentVersion() {
	logp.Info("Current CLI tool version: ", Version)
}
