package ask

import (
	"fmt"
	"strconv"
)

type Answer struct {
	Content string
	Handler func()
}

type Question struct {
	Content string
	Answers []Answer
}

func Ask(questions []Question) {
	for _, question := range questions {
		printQuestion(question)
		printAnswers(question)
		handler := scanAnswerNumber(question)
		handler()
	}
}

func printQuestion(qustion Question) {
	fmt.Println(qustion.Content)
}

func printAnswers(question Question) {
	for i, answer := range question.Answers {
		fmt.Println(strconv.Itoa(i) + " > " + answer.Content)
	}
}

func scanAnswerNumber(question Question) func() {
	for true {
		fmt.Print(" >> ")
		var input string
		fmt.Scanln(&input)
		for i, answer := range question.Answers{
			if strconv.Itoa(i) == input {
				return answer.Handler
			}
		}
		fmt.Println("invalid input.")
	}
	// unreachable path
	panic("unreachable path")
}
