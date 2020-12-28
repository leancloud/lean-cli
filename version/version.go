package version

import (
	"github.com/aisk/logp"
)

// Version is lean-cli's version.
const Version = "0.24.2"

func PrintCurrentVersion() {
	logp.Info("Current CLI tool version: ", Version)
}
