package term

import (
	"reflect"
	"testing"
)

func TestInputAddCharacter(t *testing.T) {
	cases := []struct {
		input      *Input
		in         rune
		wantBuffer []rune
		wantCursor int
	}{
		{
			input: &Input{
				buffer: []rune("hello, "),
				cursor: 7,
			},
			in:         rune('世'),
			wantBuffer: []rune("hello, 世"),
			wantCursor: 8,
		},
		{
			input: &Input{
				buffer: []rune("world"),
				cursor: 0,
			},
			in:         rune('你'),
			wantBuffer: []rune("你world"),
			wantCursor: 1,
		},
		{
			input: &Input{
				buffer: []rune("你好, orld"),
				cursor: 4,
			},
			in:         rune('w'),
			wantBuffer: []rune("你好, world"),
			wantCursor: 5,
		},
	}

	for _, c := range cases {
		c.input.AddCharacter(c.in)
		if !reflect.DeepEqual(c.input.buffer, c.wantBuffer) {
			t.Errorf("input: %v, after AddCharacter(%v), input.buffer: %v, want: %v.", c.input, c.in, c.input.buffer, c.wantBuffer)
		}
		if c.input.cursor != c.wantCursor {
			t.Errorf("input: %v, after AddCharacter(%v), input.cursor: %v, want: %v.", c.input, c.in, c.input.cursor, c.wantCursor)
		}
	}
}

func TestInputBackspace(t *testing.T) {
	cases := []struct {
		input      *Input
		wantBuffer []rune
		wantCursor int
	}{
		{
			input: &Input{
				buffer: []rune("hello"),
				cursor: 5,
			},
			wantBuffer: []rune("hell"),
			wantCursor: 4,
		},
		{
			input: &Input{
				buffer: []rune("hello"),
				cursor: 0,
			},
			wantBuffer: []rune("hello"),
			wantCursor: 0,
		},
		{
			input: &Input{
				buffer: []rune("hello"),
				cursor: 2,
			},
			wantBuffer: []rune("hllo"),
			wantCursor: 1,
		},
	}

	for _, c := range cases {
		c.input.Backspace()
		if !reflect.DeepEqual(c.input.buffer, c.wantBuffer) {
			t.Errorf("input: %v, after Backspace(), input.buffer: %v, want: %v.", c.input, c.input.buffer, c.wantBuffer)
		}
		if c.input.cursor != c.wantCursor {
			t.Errorf("input: %v, after Backspace(), input.cursor: %v, want: %v.", c.input, c.input.cursor, c.wantCursor)
		}
	}
}

func TestInputDeleteCharacterUnderCursor(t *testing.T) {
	cases := []struct {
		input      *Input
		wantBuffer []rune
		wantCursor int
	}{
		{
			input: &Input{
				buffer: []rune("hello"),
				cursor: 5,
			},
			wantBuffer: []rune("hello"),
			wantCursor: 5,
		},
		{
			input: &Input{
				buffer: []rune("hello"),
				cursor: 0,
			},
			wantBuffer: []rune("ello"),
			wantCursor: 0,
		},
		{
			input: &Input{
				buffer: []rune("hello"),
				cursor: 2,
			},
			wantBuffer: []rune("helo"),
			wantCursor: 2,
		},
	}

	for _, c := range cases {
		c.input.DeleteCharacterUnderCursor()
		if !reflect.DeepEqual(c.input.buffer, c.wantBuffer) {
			t.Errorf("input: %v, after DeleteCharacterUnderCursor(), input.buffer: %v, want: %v.", c.input, c.input.buffer, c.wantBuffer)
		}
		if c.input.cursor != c.wantCursor {
			t.Errorf("input: %v, after DeleteCharacterUnderCursor(), input.cursor: %v, want: %v.", c.input, c.input.cursor, c.wantCursor)
		}
	}
}

func TestInputDeleteCharactersAfterCursor(t *testing.T) {
	cases := []struct {
		input      *Input
		wantBuffer []rune
		wantCursor int
	}{
		{
			input: &Input{
				buffer: []rune("hello"),
				cursor: 5,
			},
			wantBuffer: []rune("hello"),
			wantCursor: 5,
		},
		{
			input: &Input{
				buffer: []rune("hello, world"),
				cursor: 0,
			},
			wantBuffer: []rune{},
			wantCursor: 0,
		},
		{
			input: &Input{
				buffer: []rune("hello, world"),
				cursor: 5,
			},
			wantBuffer: []rune("hello"),
			wantCursor: 5,
		},
	}

	for _, c := range cases {
		c.input.DeleteCharactersAfterCursor()
		if !reflect.DeepEqual(c.input.buffer, c.wantBuffer) {
			t.Errorf("input: %v, after DeleteCharactersAfterCursor(), input.buffer: %v, want: %v.", c.input, c.input.buffer, c.wantBuffer)
		}
		if c.input.cursor != c.wantCursor {
			t.Errorf("input: %v, after DeleteCharactersAfterCursor(), input.cursor: %v, want: %v.", c.input, c.input.cursor, c.wantCursor)
		}
	}
}

func TestInputDeleteCharactersBeforeCursor(t *testing.T) {
	cases := []struct {
		input      *Input
		wantBuffer []rune
		wantCursor int
	}{
		{
			input: &Input{
				buffer: []rune("hello, world"),
				cursor: 12,
			},
			wantBuffer: []rune{},
			wantCursor: 0,
		},
		{
			input: &Input{
				buffer: []rune("hello, world"),
				cursor: 0,
			},
			wantBuffer: []rune("hello, world"),
			wantCursor: 0,
		},
		{
			input: &Input{
				buffer: []rune("hello, world"),
				cursor: 5,
			},
			wantBuffer: []rune(", world"),
			wantCursor: 0,
		},
	}

	for _, c := range cases {
		c.input.DeleteCharactersBeforeCursor()
		if !reflect.DeepEqual(c.input.buffer, c.wantBuffer) {
			t.Errorf("input: %v, after DeleteCharactersBeforeCursor(), input.buffer: %v, want: %v.", c.input, c.input.buffer, c.wantBuffer)
		}
		if c.input.cursor != c.wantCursor {
			t.Errorf("input: %v, after DeleteCharactersBeforeCursor(), input.cursor: %v, want: %v.", c.input, c.input.cursor, c.wantCursor)
		}
	}
}

func TestInputDeleteOneWordBeforeCursor(t *testing.T) {
	cases := []struct {
		input      *Input
		wantBuffer []rune
		wantCursor int
	}{
		{
			input: &Input{
				buffer: []rune("hello, world"),
				cursor: 12,
			},
			wantBuffer: []rune("hello, "),
			wantCursor: 7,
		},
		{
			input: &Input{
				buffer: []rune("hello, world"),
				cursor: 0,
			},
			wantBuffer: []rune("hello, world"),
			wantCursor: 0,
		},
		{
			input: &Input{
				buffer: []rune("hello, world"),
				cursor: 9,
			},
			wantBuffer: []rune("hello, rld"),
			wantCursor: 7,
		},
		{
			input: &Input{
				buffer: []rune("hello, world"),
				cursor: 7,
			},
			wantBuffer: []rune("world"),
			wantCursor: 0,
		},
	}

	for _, c := range cases {
		c.input.DeleteOneWordBeforeCursor()
		if !reflect.DeepEqual(c.input.buffer, c.wantBuffer) {
			t.Errorf("input: %v, after DeleteOneWordBeforeCursor(), input.buffer: %v, want: %v.", c.input, c.input.buffer, c.wantBuffer)
		}
		if c.input.cursor != c.wantCursor {
			t.Errorf("input: %v, after DeleteOneWordBeforeCursor(), input.cursor: %v, want: %v.", c.input, c.input.cursor, c.wantCursor)
		}
	}
}

func TestInputGoHead(t *testing.T) {
	cases := []struct {
		input      *Input
		wantBuffer []rune
		wantCursor int
	}{
		{
			input: &Input{
				buffer: []rune("hello, world"),
				cursor: 12,
			},
			wantBuffer: []rune("hello, world"),
			wantCursor: 0,
		},
		{
			input: &Input{
				buffer: []rune("hello, world"),
				cursor: 0,
			},
			wantBuffer: []rune("hello, world"),
			wantCursor: 0,
		},
		{
			input: &Input{
				buffer: []rune("hello, world"),
				cursor: 9,
			},
			wantBuffer: []rune("hello, world"),
			wantCursor: 0,
		},
	}

	for _, c := range cases {
		c.input.GoHead()
		if !reflect.DeepEqual(c.input.buffer, c.wantBuffer) {
			t.Errorf("input: %v, after GoHead(), input.buffer: %v, want: %v.", c.input, c.input.buffer, c.wantBuffer)
		}
		if c.input.cursor != c.wantCursor {
			t.Errorf("input: %v, after GoHead(), input.cursor: %v, want: %v.", c.input, c.input.cursor, c.wantCursor)
		}
	}
}

func TestInputGoEnd(t *testing.T) {
	cases := []struct {
		input      *Input
		wantBuffer []rune
		wantCursor int
	}{
		{
			input: &Input{
				buffer: []rune("hello, world"),
				cursor: 12,
			},
			wantBuffer: []rune("hello, world"),
			wantCursor: 12,
		},
		{
			input: &Input{
				buffer: []rune("hello, world"),
				cursor: 0,
			},
			wantBuffer: []rune("hello, world"),
			wantCursor: 12,
		},
		{
			input: &Input{
				buffer: []rune("hello, world"),
				cursor: 9,
			},
			wantBuffer: []rune("hello, world"),
			wantCursor: 12,
		},
	}

	for _, c := range cases {
		c.input.GoEnd()
		if !reflect.DeepEqual(c.input.buffer, c.wantBuffer) {
			t.Errorf("input: %v, after GoEnd(), input.buffer: %v, want: %v.", c.input, c.input.buffer, c.wantBuffer)
		}
		if c.input.cursor != c.wantCursor {
			t.Errorf("input: %v, after GoEnd(), input.cursor: %v, want: %v.", c.input, c.input.cursor, c.wantCursor)
		}
	}
}

func TestInputGoBackOneCharacter(t *testing.T) {
	cases := []struct {
		input      *Input
		wantBuffer []rune
		wantCursor int
	}{
		{
			input: &Input{
				buffer: []rune("hello, world"),
				cursor: 12,
			},
			wantBuffer: []rune("hello, world"),
			wantCursor: 11,
		},
		{
			input: &Input{
				buffer: []rune("hello, world"),
				cursor: 0,
			},
			wantBuffer: []rune("hello, world"),
			wantCursor: 0,
		},
		{
			input: &Input{
				buffer: []rune("hello, world"),
				cursor: 9,
			},
			wantBuffer: []rune("hello, world"),
			wantCursor: 8,
		},
	}

	for _, c := range cases {
		c.input.GoBackOneCharacter()
		if !reflect.DeepEqual(c.input.buffer, c.wantBuffer) {
			t.Errorf("input: %v, after GoBackOneCharacter(), input.buffer: %v, want: %v.", c.input, c.input.buffer, c.wantBuffer)
		}
		if c.input.cursor != c.wantCursor {
			t.Errorf("input: %v, after GoBackOneCharacter(), input.cursor: %v, want: %v.", c.input, c.input.cursor, c.wantCursor)
		}
	}
}

func TestInputGoForwardOneCharacter(t *testing.T) {
	cases := []struct {
		input      *Input
		wantBuffer []rune
		wantCursor int
	}{
		{
			input: &Input{
				buffer: []rune("hello, world"),
				cursor: 12,
			},
			wantBuffer: []rune("hello, world"),
			wantCursor: 12,
		},
		{
			input: &Input{
				buffer: []rune("hello, world"),
				cursor: 0,
			},
			wantBuffer: []rune("hello, world"),
			wantCursor: 1,
		},
		{
			input: &Input{
				buffer: []rune("hello, world"),
				cursor: 9,
			},
			wantBuffer: []rune("hello, world"),
			wantCursor: 10,
		},
	}

	for _, c := range cases {
		c.input.GoForwardOneCharacter()
		if !reflect.DeepEqual(c.input.buffer, c.wantBuffer) {
			t.Errorf("input: %v, after GoForwardOneCharacter(), input.buffer: %v, want: %v.", c.input, c.input.buffer, c.wantBuffer)
		}
		if c.input.cursor != c.wantCursor {
			t.Errorf("input: %v, after GoForwardOneCharacter(), input.cursor: %v, want: %v.", c.input, c.input.cursor, c.wantCursor)
		}
	}
}

func TestInputGoBackOneWord(t *testing.T) {
	cases := []struct {
		input      *Input
		wantBuffer []rune
		wantCursor int
	}{
		{
			input: &Input{
				buffer: []rune("hello, !@#$world"),
				cursor: 16,
			},
			wantBuffer: []rune("hello, !@#$world"),
			wantCursor: 11,
		},
		{
			input: &Input{
				buffer: []rune("hello, !@#$world"),
				cursor: 0,
			},
			wantBuffer: []rune("hello, !@#$world"),
			wantCursor: 0,
		},
		{
			input: &Input{
				buffer: []rune("hello, !@#$world"),
				cursor: 12,
			},
			wantBuffer: []rune("hello, !@#$world"),
			wantCursor: 11,
		},
		{
			input: &Input{
				buffer: []rune("hello, !@#$world"),
				cursor: 11,
			},
			wantBuffer: []rune("hello, !@#$world"),
			wantCursor: 0,
		},
	}

	for _, c := range cases {
		c.input.GoBackOneWord()
		if !reflect.DeepEqual(c.input.buffer, c.wantBuffer) {
			t.Errorf("input: %v, after GoBackOneWord(), input.buffer: %v, want: %v.", c.input, c.input.buffer, c.wantBuffer)
		}
		if c.input.cursor != c.wantCursor {
			t.Errorf("input: %v, after GoBackOneWord(), input.cursor: %v, want: %v.", c.input, c.input.cursor, c.wantCursor)
		}
	}
}

func TestInputGoForwardOneWord(t *testing.T) {
	cases := []struct {
		input      *Input
		wantBuffer []rune
		wantCursor int
	}{
		{
			input: &Input{
				buffer: []rune("hello, !@#$world"),
				cursor: 16,
			},
			wantBuffer: []rune("hello, !@#$world"),
			wantCursor: 16,
		},
		{
			input: &Input{
				buffer: []rune("hello, !@#$world"),
				cursor: 0,
			},
			wantBuffer: []rune("hello, !@#$world"),
			wantCursor: 5,
		},
		{
			input: &Input{
				buffer: []rune("hello, !@#$world"),
				cursor: 3,
			},
			wantBuffer: []rune("hello, !@#$world"),
			wantCursor: 5,
		},
		{
			input: &Input{
				buffer: []rune("hello, !@#$world!@#$"),
				cursor: 5,
			},
			wantBuffer: []rune("hello, !@#$world!@#$"),
			wantCursor: 16,
		},
	}

	for _, c := range cases {
		c.input.GoForwardOneWord()
		if !reflect.DeepEqual(c.input.buffer, c.wantBuffer) {
			t.Errorf("input: %v, after GoForwardOneWord(), input.buffer: %v, want: %v.", c.input, c.input.buffer, c.wantBuffer)
		}
		if c.input.cursor != c.wantCursor {
			t.Errorf("input: %v, after GoForwardOneWord(), input.cursor: %v, want: %v.", c.input, c.input.cursor, c.wantCursor)
		}
	}
}

func TestIsAlphabetOrNumberic(t *testing.T) {
	cases := []struct {
		in   rune
		want bool
	}{
		{
			in:   '0',
			want: true,
		},
		{
			in:   '9',
			want: true,
		},
		{
			in:   'A',
			want: true,
		},
		{
			in:   'Z',
			want: true,
		},
		{
			in:   'a',
			want: true,
		},
		{
			in:   'z',
			want: true,
		},
		{
			in:   ' ',
			want: false,
		},
	}

	for _, c := range cases {
		got := isAlphabetOrNumeric(c.in)
		if got != c.want {
			t.Errorf("isAlphabetOrNumeric(%v) == %v, want: %v.", c.in, got, c.want)
		}
	}
}
