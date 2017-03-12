package wizard

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
)

// ForceDisableInquirer acts like it looks like
var ForceDisableInquirer = false

var detectCode = `import os;import inquirer;code=0 if inquirer.__version__ == '2.1.11' else 1;exit(code)`

var inquirerCode = `
import os
import inquirer

print(os.stdin)
`

func useInquirer() bool {
	if ForceDisableInquirer {
		return false
	}

	if _, err := exec.LookPath("python3"); err != nil {
		return false
	}

	cmd := exec.Command("python3", "-c", detectCode)
	return cmd.Run() == nil
}

func inputFile() (*os.File, error) {
	return ioutil.TempFile("", "wizard_input")
}

func inquirer(questions []Question) error {
	dir, err := ioutil.TempDir("", "go_wizard_")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)
	inputPath := filepath.Join(dir, "input.json")
	outputPath := filepath.Join(dir, "output.json")
	scriptPath := filepath.Join(dir, "script.py")

	items := make([]interface{}, len(questions))
	for i, question := range questions {
		if question.Input != nil {
			if question.Input.Hidden {
				items[i] = map[string]interface{}{
					"kind":    "password",
					"name":    i,
					"message": question.Content,
				}
				continue
			}
			items[i] = map[string]interface{}{
				"kind":    "text",
				"name":    i,
				"message": question.Content,
			}
			continue
		}
		choices := make([]string, len(question.Answers))
		for j, answer := range question.Answers {
			choices[j] = answer.Content
		}
		items[i] = map[string]interface{}{
			"kind":    "list",
			"name":    i,
			"message": question.Content,
			"choices": choices,
		}
	}

	input, err := json.Marshal(items)
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(inputPath, input, 0644); err != nil {
		return err
	}

	code := fmt.Sprintf(`
# coding: utf-8

import sys
import json

import inquirer

with open('%s') as f:
    questions = inquirer.load_from_json(f.read())

answers = inquirer.prompt(questions)

if answers is None:
	sys.exit(1)


with open('%s', 'w') as f:
	json.dump(answers, f)
`, inputPath, outputPath)
	if ioutil.WriteFile(scriptPath, []byte(code), 0644); err != nil {
		return err
	}
	cmd := exec.Command("python3", scriptPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err = cmd.Run(); err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			return errors.New("wizard cancelled")
		}
		return err
	}

	resultBytes, err := ioutil.ReadFile(outputPath)
	if err != nil {
		return err
	}
	answers := make(map[int]string)
	if err := json.Unmarshal(resultBytes, &answers); err != nil {
		return err
	}

	for i := 0; i < len(questions); i++ {
		question := questions[i]
		answer := answers[i]
		if question.Input != nil {
			*question.Input.Result = answer
			continue
		}
		for _, a := range question.Answers {
			if a.Content == answer {
				a.Handler()
			}
		}

	}

	return nil
}
