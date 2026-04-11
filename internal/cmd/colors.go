package cmd

import (
	"fmt"
	"os"

	"golang.org/x/image/colornames"
	"golang.org/x/term"
)

// colorEnabled checks whether we should emit ANSI color escape codes.
// Honors the NO_COLOR convention (https://no-color.org) and disables
// coloring when stdout is not a terminal (e.g. piped to a file).
func colorEnabled() bool {
	if _, ok := os.LookupEnv("NO_COLOR"); ok {
		return false
	}
	return term.IsTerminal(int(os.Stdout.Fd()))
}

// colorizeName returns name wrapped in an ANSI 24-bit foreground color
// escape matching the CSS3 color of the same name. If name is not a
// known color, or if color output is disabled, name is returned as-is.
func colorizeName(name string) string {
	if name == "" || !colorEnabled() {
		return name
	}
	c, ok := colornames.Map[name]
	if !ok {
		return name
	}
	// \x1b[38;2;R;G;Bm  <text>  \x1b[0m
	return fmt.Sprintf("\x1b[38;2;%d;%d;%dm%s\x1b[0m", c.R, c.G, c.B, name)
}
