package term

// Input denotes user input from terminal
type Input struct {
	buffer []rune
	cursor int
}

// NewInput return an initialized *Input
func NewInput() *Input {
	return &Input{
		buffer: make([]rune, 0),
		cursor: 0,
	}
}

// AddCharacter add one character to input
func (in *Input) AddCharacter(r rune) {
	in.buffer = append(in.buffer[:in.cursor], append([]rune{r}, in.buffer[in.cursor:]...)...)
	in.cursor++
}

// Backspace delete the last character to input
func (in *Input) Backspace() {
	if in.cursor == 0 {
		return
	}

	in.buffer = append(in.buffer[:in.cursor-1], in.buffer[in.cursor:]...)
	in.cursor--
}

// DeleteCharacterUnderCursor delete the character under the cursor
func (in *Input) DeleteCharacterUnderCursor() {
	if in.cursor == len(in.buffer) {
		return
	}

	in.buffer = append(in.buffer[:in.cursor], in.buffer[in.cursor+1:]...)
}

// DeleteCharactersAfterCursor delete the character after the cursor
func (in *Input) DeleteCharactersAfterCursor() {
	// Ctrl-k, delete characters after the cursor
	in.buffer = in.buffer[:in.cursor]
}

// DeleteCharactersBeforeCursor delete characters before the cursor
func (in *Input) DeleteCharactersBeforeCursor() {
	in.buffer = in.buffer[in.cursor:]
	in.cursor = 0
}

// DeleteOneWordBeforeCursor delete word before the cursor
func (in *Input) DeleteOneWordBeforeCursor() {
	if in.cursor == 0 {
		return
	}

	hasSpace := false
	hasNonSpace := false
	for i := in.cursor - 1; i >= 0; i-- {
		if in.buffer[i] == ' ' {
			if !hasNonSpace {
				continue
			}

			in.buffer = append(in.buffer[:i+1], in.buffer[in.cursor:]...)
			in.cursor = i + 1
			hasSpace = true
			break
		} else {
			hasNonSpace = true
		}
	}

	if !hasSpace {
		in.buffer = in.buffer[in.cursor:]
		in.cursor = 0
	}
}

// GoHead go to the beginning of the input
func (in *Input) GoHead() {
	in.cursor = 0
}

// GoEnd go to the end of the input
func (in *Input) GoEnd() {
	in.cursor = len(in.buffer)
}

// GoBackOneCharacter go back one character
func (in *Input) GoBackOneCharacter() {
	if in.cursor > 0 {
		in.cursor--
	}
}

// GoForwardOneCharacter go forward one character
func (in *Input) GoForwardOneCharacter() {
	if in.cursor < len(in.buffer) {
		in.cursor++
	}
}

// GoBackOneWord go back one word
func (in *Input) GoBackOneWord() {
	if in.cursor == 0 {
		return
	}

	hasNonDelimiter := false
	for i := in.cursor - 1; i >= 0; i-- {
		if !isAlphabetOrNumeric(in.buffer[i]) {
			if !hasNonDelimiter {
				continue
			}

			in.cursor = i + 1
			return
		}

		hasNonDelimiter = true
	}

	in.cursor = 0
}

// GoForwardOneWord go forward one word
func (in *Input) GoForwardOneWord() {
	if in.cursor == len(in.buffer) {
		return
	}

	hasNonDelimiter := false
	for i := in.cursor; i < len(in.buffer); i++ {
		if !isAlphabetOrNumeric(in.buffer[i]) {
			if !hasNonDelimiter {
				continue
			}

			in.cursor = i
			return
		}

		hasNonDelimiter = true
	}

	in.cursor = len(in.buffer)
}

func isAlphabetOrNumeric(r rune) bool {
	if (r >= 48 && r <= 57) || (r >= 65 && r <= 90) || (r >= 97 && r <= 122) {
		return true
	}

	return false
}
