package prompt

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
)

var nonInteractive bool

// SetNonInteractive configures prompt behavior.
func SetNonInteractive(value bool) {
	nonInteractive = value
}

// AskString prompts for a string input.
func AskString(question string, defaultValue string, required bool) (string, error) {
	if nonInteractive {
		if required && defaultValue == "" {
			return "", errors.New("missing required input in non-interactive mode")
		}
		return defaultValue, nil
	}
	reader := bufio.NewReader(os.Stdin)
	if defaultValue != "" {
		fmt.Printf("%s [%s]: ", question, defaultValue)
	} else {
		fmt.Printf("%s: ", question)
	}
	text, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	text = strings.TrimSpace(text)
	if text == "" {
		if required && defaultValue == "" {
			return "", errors.New("missing required input")
		}
		return defaultValue, nil
	}
	return text, nil
}

// AskChoice prompts for a choice from a list.
func AskChoice(question string, options []string, defaultValue string, required bool) (string, error) {
	if len(options) == 0 {
		return "", errors.New("no options provided")
	}
	optionList := strings.Join(options, "/")
	prompt := fmt.Sprintf("%s (%s)", question, optionList)
	return AskString(prompt, defaultValue, required)
}

// AskBool prompts for a yes/no input.
func AskBool(question string, defaultValue bool) (bool, error) {
	defaultText := "n"
	if defaultValue {
		defaultText = "y"
	}
	answer, err := AskString(fmt.Sprintf("%s (y/n)", question), defaultText, false)
	if err != nil {
		return false, err
	}
	answer = strings.ToLower(strings.TrimSpace(answer))
	return answer == "y" || answer == "yes", nil
}
