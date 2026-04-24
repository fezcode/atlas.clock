package ui

import "strings"

// 5-row phosphor-style big glyphs for the detail view.
var bigDigits = map[rune][]string{
	'0': {" ███ ", " █ █ ", " █ █ ", " █ █ ", " ███ "},
	'1': {"  █  ", " ██  ", "  █  ", "  █  ", " ███ "},
	'2': {" ███ ", "   █ ", " ███ ", " █   ", " ███ "},
	'3': {" ███ ", "   █ ", " ███ ", "   █ ", " ███ "},
	'4': {" █ █ ", " █ █ ", " ███ ", "   █ ", "   █ "},
	'5': {" ███ ", " █   ", " ███ ", "   █ ", " ███ "},
	'6': {" ███ ", " █   ", " ███ ", " █ █ ", " ███ "},
	'7': {" ███ ", "   █ ", "   █ ", "  █  ", "  █  "},
	'8': {" ███ ", " █ █ ", " ███ ", " █ █ ", " ███ "},
	'9': {" ███ ", " █ █ ", " ███ ", "   █ ", " ███ "},
	':': {"     ", "  █  ", "     ", "  █  ", "     "},
	'.': {"     ", "     ", "     ", "     ", "  █  "},
	' ': {"     ", "     ", "     ", "     ", "     "},
}

// renderBigText returns a 5-line styled big-digit rendering.
func renderBigText(input string) string {
	lines := make([]string, 5)
	for _, r := range input {
		digit, ok := bigDigits[r]
		if !ok {
			digit = []string{"     ", "     ", "     ", "     ", "     "}
		}
		for i := 0; i < 5; i++ {
			lines[i] += digit[i]
		}
	}
	// Style each line as a whole so truncation works cleanly downstream.
	for i := range lines {
		lines[i] = sBigDigit.Render(lines[i])
	}
	return strings.Join(lines, "\n")
}
