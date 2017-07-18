package chrysanthemum

import (
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/fatih/color"
	"github.com/mattn/go-isatty"
)

var isTerminal = isatty.IsTerminal(os.Stdout.Fd())

// Frames is spinner's frame. replace it with what you like
var Frames []string

// Success is the flag symbol after you called the Successed
var Success string

// Fail is the flag symbol after you called the Failed
var Fail string

// Chrysanthemum represent a spinner instance
type Chrysanthemum struct {
	stop    chan bool
	stopped bool
	text    string
}

func init() {
	if runtime.GOOS == "windows" {
		Frames = []string{
			"-",
			"\\",
			"|",
			"/",
		}
		Success = " "
		Fail = color.RedString("x")
	} else {
		Frames = []string{
			color.MagentaString("⠋"),
			color.MagentaString("⠙"),
			color.MagentaString("⠹"),
			color.MagentaString("⠸"),
			color.MagentaString("⠼"),
			color.MagentaString("⠴"),
			color.MagentaString("⠦"),
			color.MagentaString("⠧"),
			color.MagentaString("⠇"),
			color.MagentaString("⠏"),
		}
		Success = color.GreenString("✓")
		Fail = color.RedString("✗")
	}
}

// New create a spinner instance
func New(text string) *Chrysanthemum {
	if !isTerminal {
		fmt.Fprint(color.Output, text)
	} else {
		fmt.Fprint(color.Output, text+"  ")
	}
	return &Chrysanthemum{
		stop:    make(chan bool),
		stopped: false,
		text:    text,
	}
}

// Start will let your chrysanthemum spin!
func (c *Chrysanthemum) Start() *Chrysanthemum {
	if !isTerminal {
		return c
	}

	// fmt.Print("\033[?25l") // hide cursor

	i := 0
	go func() {
		for {
			if i == len(Frames) {
				i = 0
			}
			fmt.Fprintf(color.Output, "\r%s %s", c.text, Frames[i])
			i++
			select {
			case <-c.stop:
				return
			default:
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()

	return c
}

func (c *Chrysanthemum) end(flag string) {
	if !isTerminal {
		fmt.Println()
		return
	}

	if c.stopped {
		return
	}
	c.stop <- true
	c.stopped = true
	// fmt.Printf("\033[?25h") // show cursor
	fmt.Fprintf(color.Output, "\r%s %s\n", c.text, flag)
}

func (c *Chrysanthemum) Successed() {
	c.end(Success)
}

func (c *Chrysanthemum) Failed() {
	c.end(Fail)
}

func (c *Chrysanthemum) End() {
	c.Successed()
}

func Successed(args ...interface{}) {
	if isTerminal {
		args = append(args, Success)
	}
	fmt.Fprintln(color.Output, args...)
}

func Failed(args ...interface{}) {
	if isTerminal {
		args = append(args, Fail)
	}
	fmt.Fprintln(color.Output, args...)
}
