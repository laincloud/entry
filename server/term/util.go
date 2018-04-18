package term

import (
	"bytes"
)

const (
	asciiSOH               = 1
	asciiSTX               = 2
	asciiEOT               = 4
	asciiENQ               = 5
	asciiACK               = 6
	asciiBEL               = 7
	asciiBS                = 8
	asciiHT                = 9
	asciiVT                = 11
	asciiFF                = 12
	asciiCR                = 13
	asciiNAK               = 21
	asciiETB               = 23
	asciiESC               = 27 // Control sequence prefix
	asciiA                 = 65
	asciiB                 = 66
	asciiC                 = 67
	asciiD                 = 68
	asciiLeftSquareBracket = 91 // Control sequence prefix
	asciib                 = 98
	asciif                 = 102
	asciiDEL               = 127
)

var (
	// UpArrow denotes up arrow
	UpArrow = []byte{asciiESC, asciiLeftSquareBracket, asciiA}
	// DownArrow denotes down arrow
	DownArrow = []byte{asciiESC, asciiLeftSquareBracket, asciiB}
)

// EscapeInput escape special characters such as SOH, ENQ and DEL in user input
func EscapeInput(input []byte) []byte {
	rs := bytes.Runes(input)
	in := NewInput()
	for i := 0; i < len(rs); i++ {
		switch {
		case rs[i] == asciiSOH:
			// Ctrl-a, go to the beginning of the line
			in.GoHead()
		case rs[i] == asciiSTX:
			// Ctrl-b, go back one character
			in.GoBackOneCharacter()
		case rs[i] == asciiEOT:
			// Ctrl-d, delete the character under the cursor
			in.DeleteCharacterUnderCursor()
		case rs[i] == asciiENQ:
			// Ctrl-e, go to the end of the line
			in.GoEnd()
		case rs[i] == asciiACK:
			// Ctrl-f, go forward one character
			in.GoForwardOneCharacter()
		case rs[i] == asciiVT:
			// Ctrl-k, delete characters after the cursor
			in.DeleteCharactersAfterCursor()
		case rs[i] == asciiFF:
			// Ctrl-l, clear the screen
		case rs[i] == asciiNAK:
			// Ctrl-u, delete characters before the cursor
			in.DeleteCharactersBeforeCursor()
		case rs[i] == asciiETB:
			// Ctrl-w, delete word before the cursor
			in.DeleteOneWordBeforeCursor()
		case rs[i] == asciiESC && i < len(rs)-2 && rs[i+1] == asciiLeftSquareBracket && rs[i+2] == asciiC:
			// Right arrow, go forward one character
			in.GoForwardOneCharacter()
			i += 2
		case rs[i] == asciiESC && i < len(rs)-2 && rs[i+1] == asciiLeftSquareBracket && rs[i+2] == asciiD:
			// Left arrow: go back one character
			in.GoBackOneCharacter()
			i += 2
		case rs[i] == asciiESC && i < len(rs)-1 && rs[i+1] == asciib:
			// Alt-b, go back one word
			in.GoBackOneWord()
			i++
		case rs[i] == asciiESC && i < len(rs)-1 && rs[i+1] == asciif:
			// Alt-f, go forward one word
			in.GoForwardOneWord()
			i++
		case rs[i] == asciiBS || rs[i] == asciiDEL:
			// Backspace, delete one character before the cursor
			in.Backspace()
		default:
			in.AddCharacter(rs[i])
		}
	}

	return []byte(string(in.buffer))
}

// EscapeTabCompletion escape tab completion
func EscapeTabCompletion(src []byte) []byte {
	if IsBell(src) {
		return []byte{}
	}

	for i, b := range src {
		if b == ' ' && (i != len(src)-1) {
			return []byte{}
		}
	}

	return src
}

// EscapeHistoryCommand escape history command
func EscapeHistoryCommand(input []byte) []byte {
	if IsBell(input) {
		return []byte{}
	}

	return input
}

// HasUpArrowSuffix test whether input has up arrow suffix
func HasUpArrowSuffix(input []byte) bool {
	if len(input) < 3 {
		return false
	}

	for i, b := range UpArrow {
		if input[len(input)-len(UpArrow)+i] != b {
			return false
		}
	}

	return true
}

// HasDownArrowSuffix test whether input has down arrow suffix
func HasDownArrowSuffix(input []byte) bool {
	if len(input) < 3 {
		return false
	}

	for i, v := range DownArrow {
		if input[len(input)-len(DownArrow)+i] != v {
			return false
		}
	}

	return true
}

// IsBell test whether input is Bell
func IsBell(bs []byte) bool {
	if len(bs) == 1 && bs[0] == asciiBEL {
		return true
	}

	return false
}

// IsCR test whether input is CR
func IsCR(input []byte) bool {
	if len(input) == 1 && input[0] == asciiCR {
		return true
	}

	return false
}

// IsTab test whether input is Tab
func IsTab(input []byte) bool {
	if len(input) == 1 && input[0] == asciiHT {
		return true
	}

	return false
}
