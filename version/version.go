package version

import (
	"github.com/leancloud/lean-cli/logger"
)

// Version is lean-cli's version.
const Version = "0.13.0"

func PrintCurrentVersion() {
	logger.Info("当前命令行工具版本：", Version)
}
