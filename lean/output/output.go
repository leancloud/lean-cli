package output

import (
	"fmt"
	"io"

	"github.com/fatih/color"
)

// Output ...
type Output struct {
	ended  bool
	writer io.Writer
}

// NewOutput ...
func NewOutput(writer io.Writer) *Output {
	return &Output{
		writer: writer,
		ended:  true,
	}
}

// Write ...
func (op *Output) Write(line string) {
	if !op.ended {
		op.Successed()
	}
	fmt.Fprintf(op.writer, "> %s ...", line)
	op.ended = false
}

// Successed ...
func (op *Output) Successed() {
	fmt.Fprintf(op.writer, "\b\b\b%s\r\n", color.GreenString("[SUCCESS]"))
	op.ended = true
}

// Failed ...
func (op *Output) Failed() {
	fmt.Fprintf(op.writer, "\b\b\b%s\r\n", color.RedString("[FAIL]"))
	op.ended = true
}
