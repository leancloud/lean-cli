package main

import (
	"fmt"
	"github.com/aisk/wizard"
)

func main() {
	username := ""
	password := ""
	loginNextTime := false

	questions := []wizard.Question{
		{
			Content: "PLease input your user name:",
			Input: &wizard.Input{
				Result: &username,
			},
		},
		{
			Content: "Please input your password:",
			Input: &wizard.Input{
				Result: &password,
				Hidden: true,
			},
		},
		{
			Content: "Login next time?",
			Answers: []wizard.Answer{
				{
					Content: "yes",
					Handler: func() {
						loginNextTime = true
					},
				},
				{
					Content: "no",
					Handler: func() {
						loginNextTime = false
					},
				},
			},
		},
	}

	err := wizard.Ask(questions)
	if err != nil {
		panic(err)
	}

	fmt.Println("Your username is", username)
	fmt.Println("Your password is", password)
	fmt.Println("Login next time is", loginNextTime)
}
