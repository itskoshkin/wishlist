package colors

import (
	"os"
)

const (
	regularWhite  = "\033[37m"
	regularBlack  = "\033[30m"
	regularRed    = "\033[31m"
	regularYellow = "\033[33m"
	regularGreen  = "\033[32m"
	regularCyan   = "\033[36m"
	regularBlue   = "\033[34m"
	regularPurple = "\033[35m"

	highlightedWhite  = "\033[97m"
	highlightedBlack  = "\033[90m"
	highlightedRed    = "\033[91m"
	highlightedYellow = "\033[93m"
	highlightedGreen  = "\033[92m"
	highlightedCyan   = "\033[96m"
	highlightedBlue   = "\033[94m"
	highlightedPurple = "\033[95m"

	white      = highlightedWhite
	lightGray  = regularWhite
	darkGray   = highlightedBlack
	black      = regularBlack
	red        = highlightedRed
	orange     = regularRed
	yellow     = highlightedYellow
	darkYellow = regularYellow
	green      = highlightedGreen
	darkGreen  = regularGreen
	cyan       = regularCyan
	sky        = highlightedCyan
	blue       = highlightedBlue
	darkBlue   = regularBlue
	magenta    = highlightedPurple
	purple     = regularPurple

	bold          = "\033[1m"
	faint         = "\033[2m"
	italic        = "\033[3m"
	underline     = "\033[4m"
	blink         = "\033[5m"
	background    = "\033[7m"
	hidden        = "\033[8m"
	strikethrough = "\033[9m"

	reset = "\033[0m"

	nl  = "\n"            // New line
	pl  = "\033[F"        // Previous line
	cpl = "\033[A\033[2K" // Clear previous line
)

const (
	rgbRed      = "\033[38;2;255;0;0m"
	rgbOrange   = "\033[38;2;255;127;0m"
	rgbGreen    = "\033[38;2;0;255;0m"
	rgbSkyBlue  = "\033[38;2;0;191;255m"
	rgbDarkBlue = "\033[38;2;0;0;255m"
	rgbViolet   = "\033[38;2;148;0;211m"
	rgbPink     = "\033[38;2;255;105;180m"
	rgbBrown    = "\033[38;2;139;69;19m"
	rgbGold     = "\033[38;2;255;215;0m"
	rgbMint     = "\033[38;2;152;251;152m"
)

func IsRGBSupported() bool {
	val := os.Getenv("COLORTERM")
	if val == "truecolor" || val == "24bit" {
		return true
	}

	return false
}

// Colors

func Red(text string) string {
	if IsRGBSupported() {
		return rgbRed + text + reset
	}

	return red + text + reset
}

func Orange(text string) string {
	if IsRGBSupported() {
		return rgbOrange + text + reset
	}

	return yellow + text + reset
}

func Yellow(text string) string { return yellow + text + reset }

func Green(text string) string {
	if IsRGBSupported() {
		return rgbGreen + text + reset
	}

	return green + text + reset
}

func Sky(text string) string { return sky + text + reset }

func Blue(text string) string { return blue + text + reset }

func Purple(text string) string { return purple + text + reset }

func White(text string) string { return white + text + reset }

func Gray(text string) string { return lightGray + text + reset }

func Black(text string) string { return black + text + reset }

// Styles

func Bold(text string) string { return bold + text + reset }

func Italic(text string) string { return italic + text + reset }

func Underline(text string) string { return underline + text + reset }

func Strikethrough(text string) string { return strikethrough + text + reset }

func Background(text string) string { return background + text + reset }
