package version

import (
	"github.com/leancloud/lean-cli/logger"
)

// Version is lean-cli's version.
const Version = "0.12.0"

func PrintCurrentVersion() {
	logger.Info("当前版本：", Version)
}
