package util

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
	asciiVT                = 11
	asciiFF                = 12
	asciiNAK               = 21
	asciiETB               = 23
	asciiESC               = 27 // Control sequence prefix
	asciiC                 = 67
	asciiD                 = 68
	asciiLeftSquareBracket = 91 // Control sequence prefix
	asciib                 = 98
	asciif                 = 102
	asciiDEL               = 127
)

// TermEscape handle special characters such as SOH, ENQ and DEL in ASCII
func TermEscape(src []byte) []rune {
	rs := bytes.Runes(src)
	line := &line{
		buffer: make([]rune, 0),
		cursor: 0,
	}
	for i := 0; i < len(rs); i++ {
		switch {
		case rs[i] == asciiSOH:
			// Ctrl-a, go to the beginning of the line
			line.GoHead()
		case rs[i] == asciiSTX:
			// Ctrl-b, go back one character
			line.GoBackOneCharacter()
		case rs[i] == asciiEOT:
			// Ctrl-d, delete the character under the cursor
			line.DeleteCharacterUnderCursor()
		case rs[i] == asciiENQ:
			// Ctrl-e, go to the end of the line
			line.GoEnd()
		case rs[i] == asciiACK:
			// Ctrl-f, go forward one character
			line.GoForwardOneCharacter()
		case rs[i] == asciiVT:
			// Ctrl-k, delete characters after the cursor
			line.DeleteCharactersAfterCursor()
		case rs[i] == asciiFF:
			// Ctrl-l, clear the screen
		case rs[i] == asciiNAK:
			// Ctrl-u, delete characters before the cursor
			line.DeleteCharactersBeforeCursor()
		case rs[i] == asciiETB:
			// Ctrl-w, delete word before the cursor
			line.DeleteOneWordBeforeCursor()
		case rs[i] == asciiESC && i < len(rs)-2 && rs[i+1] == asciiLeftSquareBracket && rs[i+2] == asciiC:
			// Right arrow, go forward one character
			line.GoForwardOneCharacter()
			i += 2
		case rs[i] == asciiESC && i < len(rs)-2 && rs[i+1] == asciiLeftSquareBracket && rs[i+2] == asciiD:
			// Left arrow: go back one character
			line.GoBackOneCharacter()
			i += 2
		case rs[i] == asciiESC && i < len(rs)-1 && rs[i+1] == asciib:
			// Alt-b, go back one word
			line.GoBackOneWord()
			i++
		case rs[i] == asciiESC && i < len(rs)-1 && rs[i+1] == asciif:
			// Alt-f, go forward one word
			line.GoForwardOneWord()
			i++
		case rs[i] == asciiDEL:
			// Backspace, delete one character before the cursor
			line.Backspace()
		default:
			line.AddCharacter(rs[i])
		}
	}

	return line.buffer
}

type line struct {
	buffer []rune
	cursor int
}

func (l *line) AddCharacter(r rune) {
	l.buffer = append(l.buffer[:l.cursor], append([]rune{r}, l.buffer[l.cursor:]...)...)
	l.cursor++
}

func (l *line) Backspace() {
	if l.cursor == 0 {
		return
	}

	l.buffer = append(l.buffer[:l.cursor-1], l.buffer[l.cursor:]...)
	l.cursor--
}

// DeleteCharacter delete the character under the cursor
func (l *line) DeleteCharacterUnderCursor() {
	if l.cursor == len(l.buffer) {
		return
	}

	l.buffer = append(l.buffer[:l.cursor], l.buffer[l.cursor+1:]...)
}

func (l *line) DeleteCharactersAfterCursor() {
	// Ctrl-k, delete characters after the cursor
	l.buffer = l.buffer[:l.cursor]
}

// DeleteCharactersBeforeCursor delete characters before the cursor
func (l *line) DeleteCharactersBeforeCursor() {
	l.buffer = l.buffer[l.cursor:]
	l.cursor = 0
}

// DeleteOneWordBeforeCursor delete word before the cursor
func (l *line) DeleteOneWordBeforeCursor() {
	if l.cursor == 0 {
		return
	}

	hasSpace := false
	hasNonSpace := false
	for i := l.cursor - 1; i >= 0; i-- {
		if l.buffer[i] == ' ' {
			if !hasNonSpace {
				continue
			}

			l.buffer = append(l.buffer[:i+1], l.buffer[l.cursor:]...)
			l.cursor = i + 1
			hasSpace = true
			break
		} else {
			hasNonSpace = true
		}
	}

	if !hasSpace {
		l.buffer = l.buffer[l.cursor:]
		l.cursor = 0
	}
}

// GoHead go to the beginning of the line
func (l *line) GoHead() {
	l.cursor = 0
}

// GoEnd go to the end of the line
func (l *line) GoEnd() {
	l.cursor = len(l.buffer)
}

// GoBackOneCharacter go back one character
func (l *line) GoBackOneCharacter() {
	if l.cursor > 0 {
		l.cursor--
	}
}

// GoForwardOneCharacter go forward one character
func (l *line) GoForwardOneCharacter() {
	if l.cursor < len(l.buffer) {
		l.cursor++
	}
}

func (l *line) GoBackOneWord() {
	if l.cursor == 0 {
		return
	}

	hasNonDelimiter := false
	for i := l.cursor - 1; i >= 0; i-- {
		if !isAlphabetOrNumeric(l.buffer[i]) {
			if !hasNonDelimiter {
				continue
			}

			l.cursor = i + 1
			return
		}

		hasNonDelimiter = true
	}

	l.cursor = 0
}

func (l *line) GoForwardOneWord() {
	if l.cursor == len(l.buffer) {
		return
	}

	hasNonDelimiter := false
	for i := l.cursor; i < len(l.buffer); i++ {
		if !isAlphabetOrNumeric(l.buffer[i]) {
			if !hasNonDelimiter {
				continue
			}

			l.cursor = i
			return
		}

		hasNonDelimiter = true
	}

	l.cursor = len(l.buffer)
}

func isAlphabetOrNumeric(r rune) bool {
	if (r >= 48 && r <= 57) || (r >= 65 && r <= 90) || (r >= 97 && r <= 122) {
		return true
	}

	return false
}

// TermComplement handle the response of tab completion
func TermComplement(src []byte) []byte {
	if len(src) == 1 && src[0] == asciiBEL {
		return []byte{}
	}

	for _, b := range src {
		if b == ' ' {
			return []byte{}
		}
	}

	return src
}
