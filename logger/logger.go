package logger

import (
	"fmt"
	"os"

	"github.com/fatih/color"
)

func Info(msgs ...interface{}) {
	msgs = append([]interface{}{color.GreenString("[INFO]")}, msgs...)
	fmt.Fprintln(os.Stderr, msgs...)
}

func Warn(msgs ...interface{}) {
	msgs = append([]interface{}{color.RedString("[WARNING]")}, msgs...)
	fmt.Fprintln(os.Stderr, msgs...)
}
