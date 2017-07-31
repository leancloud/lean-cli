package logp

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/mattn/go-colorable"
)

var (
	output      = colorable.NewColorableStderr()
	InfoPrefix  = color.GreenString("[INFO]")
	WarnPrefix  = color.RedString("[WARNING]")
	ErrorPrefix = color.RedString("[ERROR]")
)

func loggerPrintln(prefix string, msgs ...interface{}) {
	msgs = append([]interface{}{prefix}, msgs...)
	fmt.Fprintln(output, msgs...)
}

func loggerPrintf(prefix string, format string, msgs ...interface{}) {
	format = prefix + " " + format
	fmt.Fprintf(output, format, msgs...)
}

func Info(msgs ...interface{}) {
	loggerPrintln(InfoPrefix, msgs...)
}

func Infof(format string, msgs ...interface{}) {
	loggerPrintf(InfoPrefix, format, msgs...)
}

func Warn(msgs ...interface{}) {
	loggerPrintln(WarnPrefix, msgs...)
}

func Warnf(format string, msgs ...interface{}) {
	loggerPrintf(WarnPrefix, format, msgs...)
}

func Error(msgs ...interface{}) {
	loggerPrintln(ErrorPrefix, msgs...)
}

func Errorf(format string, msgs ...interface{}) {
	loggerPrintf(ErrorPrefix, format, msgs...)
}
