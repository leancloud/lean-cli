package wizard

import (
	"fmt"
	"github.com/chzyer/readline"
	"github.com/fatih/color"
	"strconv"
	"strings"
)

// Input ...
type Input struct {
	Hidden bool
	Result *string
}

// Answer ...
type Answer struct {
	Content string
	Handler func()
}

// Question ...
type Question struct {
	Content string
	Answers []Answer
	Input   *Input
}

// Ask ...
func Ask(questions []Question) error {
	rl, err := readline.New(" => ")
	if err != nil {
		return err
	}
	for _, question := range questions {
		printQuestion(question)
		if question.Input != nil {
			if question.Input.Hidden {
				// backup prompt for reset
				prompt := rl.Config.Prompt

				setPasswordCfg := rl.GenPasswordConfig()
				setPasswordCfg.SetListener(func(line []rune, pos int, key rune) (newLine []rune, newPos int, ok bool) {
					display := []byte(" => ")
					for i := 0; i < len(line); i++ {
						display = append(display, '*')
					}
					rl.SetPrompt(string(display))
					rl.Refresh()
					return nil, 0, false
				})

				// restore the origin prompt
				rl.SetPrompt(prompt)

				pswd, err := rl.ReadPasswordWithConfig(setPasswordCfg)
				if err != nil {
					return err
				}
				*question.Input.Result = string(pswd)
			} else {
				line, err := rl.Readline()
				if err != nil {
					return err
				}
				*question.Input.Result = line
			}
			continue
		}
		printAnswers(question)
		handler := scanAnswerNumber(question)
		handler()
	}
	return rl.Close()
}

func printQuestion(qustion Question) {
	green := color.New(color.FgGreen).SprintFunc()
	fmt.Printf("[%s] %s\n", green("?"), qustion.Content)
}

func printAnswers(question Question) {
	blue := color.New(color.FgBlue).SprintFunc()
	for i, answer := range question.Answers {
		fmt.Printf(" %s) %s\n", blue(i+1), answer.Content)
	}
}

func scanAnswerNumber(question Question) func() {
	for true {
		fmt.Print(" => ")
		var input string
		fmt.Scanln(&input)
		strings.TrimSpace(input)
		for i, answer := range question.Answers {
			if strconv.Itoa(i+1) == input {
				return answer.Handler
			}
		}
		fmt.Println("invalid input.")
	}
	// unreachable path
	panic("unreachable path")
}
