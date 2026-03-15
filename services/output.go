package services

import (
	"os"

	"github.com/cli/go-gh/v2/pkg/term"
)

// NewFormatter creates a Formatter based on the format string.
func NewFormatter(format string) Formatter {
	if format == "json" {
		return NewJSONFormatter(os.Stdout)
	}
	t := term.FromEnv()
	isTTY := t.IsTerminalOutput()
	width, _, _ := t.Size()
	if width <= 0 {
		width = 80
	}
	return NewPlainFormatter(os.Stdout, isTTY, width)
}
