# wizard

Command line wizard helper in go.

## Install

```sh
go get -u github.com/aisk/wizard
```

## Example

```go
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
	wizard.Ask(questions)
	fmt.Println("Your username is", username)
	fmt.Println("Your password is", password)
	fmt.Println("Login next time is", loginNextTime)
}
```
