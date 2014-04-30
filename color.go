package gordon

import (
	"github.com/aybabtme/color/brush"
	"github.com/dotcloud/docker/pkg/term"
)

var Colorize bool

func init() {
	if term.IsTerminal(1) {
		Colorize = true
	}
}

func Green(s string) string{
	if Colorize {
		return brush.Green(s).String()
	}
	return s
}

func Red(s string) string {
	if Colorize {
		return brush.Red(s).String()
	}
	return s
}

func DarkRed(s string) string {
	if Colorize {
		return brush.DarkRed(s).String()
	}
	return s
}

func DarkYellow(s string) string {
	if Colorize {
		return brush.DarkYellow(s).String()
	}
	return s
}

func Yellow(s string) string {
	if Colorize {
		return brush.Yellow(s).String()
	}
	return s
}
