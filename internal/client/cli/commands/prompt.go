package commands

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"golang.org/x/term"
)

var (
	promptLabel = color.New(color.Bold)
	stdin       = bufio.NewReader(os.Stdin)
)

func prompt(label string) string {
	fmt.Print(promptLabel.Sprint(label) + ": ")
	input, _ := stdin.ReadString('\n')
	return strings.TrimSpace(input)
}

func promptPassphrase(label string) (string, error) {
	first, err := promptHidden(label)
	if err != nil {
		return "", fmt.Errorf("failed to read passphrase: %w", err)
	}
	if first == "" {
		return "", errors.New(color.RedString("Passphrase cannot be empty"))
	}

	second, err := promptHidden("Confirm " + strings.ToLower(label))
	if err != nil {
		return "", fmt.Errorf("failed to read passphrase: %w", err)
	}
	if !bytes.Equal([]byte(first), []byte(second)) {
		return "", errors.New(color.RedString("Passphrases did not match"))
	}

	return first, nil
}

func promptHidden(label string) (string, error) {
	fmt.Print(promptLabel.Sprint(label) + ": ")
	input, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	if err != nil {
		return "", err
	}
	return string(input), nil
}
