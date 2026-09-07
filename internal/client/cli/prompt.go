package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
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
