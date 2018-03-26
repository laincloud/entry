package util

import (
	"reflect"
	"testing"
)

func TestLineAddCharacter(t *testing.T) {
	cases := []struct {
		line       *line
		in         rune
		wantBuffer []rune
		wantCursor int
	}{
		{
			line: &line{
				buffer: []rune("hello, "),
				cursor: 7,
			},
			in:         rune('世'),
			wantBuffer: []rune("hello, 世"),
			wantCursor: 8,
		},
		{
			line: &line{
				buffer: []rune("world"),
				cursor: 0,
			},
			in:         rune('你'),
			wantBuffer: []rune("你world"),
			wantCursor: 1,
		},
		{
			line: &line{
				buffer: []rune("你好, orld"),
				cursor: 4,
			},
			in:         rune('w'),
			wantBuffer: []rune("你好, world"),
			wantCursor: 5,
		},
	}

	for _, c := range cases {
		c.line.AddCharacter(c.in)
		if !reflect.DeepEqual(c.line.buffer, c.wantBuffer) {
			t.Errorf("line: %v, after AddCharacter(%v), line.buffer: %v, want: %v.", c.line, c.in, c.line.buffer, c.wantBuffer)
		}
		if c.line.cursor != c.wantCursor {
			t.Errorf("line: %v, after AddCharacter(%v), line.cursor: %v, want: %v.", c.line, c.in, c.line.cursor, c.wantCursor)
		}
	}
}

func TestLineBackspace(t *testing.T) {
	cases := []struct {
		line       *line
		wantBuffer []rune
		wantCursor int
	}{
		{
			line: &line{
				buffer: []rune("hello"),
				cursor: 5,
			},
			wantBuffer: []rune("hell"),
			wantCursor: 4,
		},
		{
			line: &line{
				buffer: []rune("hello"),
				cursor: 0,
			},
			wantBuffer: []rune("hello"),
			wantCursor: 0,
		},
		{
			line: &line{
				buffer: []rune("hello"),
				cursor: 2,
			},
			wantBuffer: []rune("hllo"),
			wantCursor: 1,
		},
	}

	for _, c := range cases {
		c.line.Backspace()
		if !reflect.DeepEqual(c.line.buffer, c.wantBuffer) {
			t.Errorf("line: %v, after Backspace(), line.buffer: %v, want: %v.", c.line, c.line.buffer, c.wantBuffer)
		}
		if c.line.cursor != c.wantCursor {
			t.Errorf("line: %v, after Backspace(), line.cursor: %v, want: %v.", c.line, c.line.cursor, c.wantCursor)
		}
	}
}

func TestLineDeleteCharacterUnderCursor(t *testing.T) {
	cases := []struct {
		line       *line
		wantBuffer []rune
		wantCursor int
	}{
		{
			line: &line{
				buffer: []rune("hello"),
				cursor: 5,
			},
			wantBuffer: []rune("hello"),
			wantCursor: 5,
		},
		{
			line: &line{
				buffer: []rune("hello"),
				cursor: 0,
			},
			wantBuffer: []rune("ello"),
			wantCursor: 0,
		},
		{
			line: &line{
				buffer: []rune("hello"),
				cursor: 2,
			},
			wantBuffer: []rune("helo"),
			wantCursor: 2,
		},
	}

	for _, c := range cases {
		c.line.DeleteCharacterUnderCursor()
		if !reflect.DeepEqual(c.line.buffer, c.wantBuffer) {
			t.Errorf("line: %v, after DeleteCharacterUnderCursor(), line.buffer: %v, want: %v.", c.line, c.line.buffer, c.wantBuffer)
		}
		if c.line.cursor != c.wantCursor {
			t.Errorf("line: %v, after DeleteCharacterUnderCursor(), line.cursor: %v, want: %v.", c.line, c.line.cursor, c.wantCursor)
		}
	}
}

func TestLineDeleteCharactersAfterCursor(t *testing.T) {
	cases := []struct {
		line       *line
		wantBuffer []rune
		wantCursor int
	}{
		{
			line: &line{
				buffer: []rune("hello"),
				cursor: 5,
			},
			wantBuffer: []rune("hello"),
			wantCursor: 5,
		},
		{
			line: &line{
				buffer: []rune("hello, world"),
				cursor: 0,
			},
			wantBuffer: []rune{},
			wantCursor: 0,
		},
		{
			line: &line{
				buffer: []rune("hello, world"),
				cursor: 5,
			},
			wantBuffer: []rune("hello"),
			wantCursor: 5,
		},
	}

	for _, c := range cases {
		c.line.DeleteCharactersAfterCursor()
		if !reflect.DeepEqual(c.line.buffer, c.wantBuffer) {
			t.Errorf("line: %v, after DeleteCharactersAfterCursor(), line.buffer: %v, want: %v.", c.line, c.line.buffer, c.wantBuffer)
		}
		if c.line.cursor != c.wantCursor {
			t.Errorf("line: %v, after DeleteCharactersAfterCursor(), line.cursor: %v, want: %v.", c.line, c.line.cursor, c.wantCursor)
		}
	}
}

func TestLineDeleteCharactersBeforeCursor(t *testing.T) {
	cases := []struct {
		line       *line
		wantBuffer []rune
		wantCursor int
	}{
		{
			line: &line{
				buffer: []rune("hello, world"),
				cursor: 12,
			},
			wantBuffer: []rune{},
			wantCursor: 0,
		},
		{
			line: &line{
				buffer: []rune("hello, world"),
				cursor: 0,
			},
			wantBuffer: []rune("hello, world"),
			wantCursor: 0,
		},
		{
			line: &line{
				buffer: []rune("hello, world"),
				cursor: 5,
			},
			wantBuffer: []rune(", world"),
			wantCursor: 0,
		},
	}

	for _, c := range cases {
		c.line.DeleteCharactersBeforeCursor()
		if !reflect.DeepEqual(c.line.buffer, c.wantBuffer) {
			t.Errorf("line: %v, after DeleteCharactersBeforeCursor(), line.buffer: %v, want: %v.", c.line, c.line.buffer, c.wantBuffer)
		}
		if c.line.cursor != c.wantCursor {
			t.Errorf("line: %v, after DeleteCharactersBeforeCursor(), line.cursor: %v, want: %v.", c.line, c.line.cursor, c.wantCursor)
		}
	}
}

func TestLineDeleteOneWordBeforeCursor(t *testing.T) {
	cases := []struct {
		line       *line
		wantBuffer []rune
		wantCursor int
	}{
		{
			line: &line{
				buffer: []rune("hello, world"),
				cursor: 12,
			},
			wantBuffer: []rune("hello, "),
			wantCursor: 7,
		},
		{
			line: &line{
				buffer: []rune("hello, world"),
				cursor: 0,
			},
			wantBuffer: []rune("hello, world"),
			wantCursor: 0,
		},
		{
			line: &line{
				buffer: []rune("hello, world"),
				cursor: 9,
			},
			wantBuffer: []rune("hello, rld"),
			wantCursor: 7,
		},
		{
			line: &line{
				buffer: []rune("hello, world"),
				cursor: 7,
			},
			wantBuffer: []rune("world"),
			wantCursor: 0,
		},
	}

	for _, c := range cases {
		c.line.DeleteOneWordBeforeCursor()
		if !reflect.DeepEqual(c.line.buffer, c.wantBuffer) {
			t.Errorf("line: %v, after DeleteOneWordBeforeCursor(), line.buffer: %v, want: %v.", c.line, c.line.buffer, c.wantBuffer)
		}
		if c.line.cursor != c.wantCursor {
			t.Errorf("line: %v, after DeleteOneWordBeforeCursor(), line.cursor: %v, want: %v.", c.line, c.line.cursor, c.wantCursor)
		}
	}
}

func TestLineGoHead(t *testing.T) {
	cases := []struct {
		line       *line
		wantBuffer []rune
		wantCursor int
	}{
		{
			line: &line{
				buffer: []rune("hello, world"),
				cursor: 12,
			},
			wantBuffer: []rune("hello, world"),
			wantCursor: 0,
		},
		{
			line: &line{
				buffer: []rune("hello, world"),
				cursor: 0,
			},
			wantBuffer: []rune("hello, world"),
			wantCursor: 0,
		},
		{
			line: &line{
				buffer: []rune("hello, world"),
				cursor: 9,
			},
			wantBuffer: []rune("hello, world"),
			wantCursor: 0,
		},
	}

	for _, c := range cases {
		c.line.GoHead()
		if !reflect.DeepEqual(c.line.buffer, c.wantBuffer) {
			t.Errorf("line: %v, after GoHead(), line.buffer: %v, want: %v.", c.line, c.line.buffer, c.wantBuffer)
		}
		if c.line.cursor != c.wantCursor {
			t.Errorf("line: %v, after GoHead(), line.cursor: %v, want: %v.", c.line, c.line.cursor, c.wantCursor)
		}
	}
}

func TestLineGoEnd(t *testing.T) {
	cases := []struct {
		line       *line
		wantBuffer []rune
		wantCursor int
	}{
		{
			line: &line{
				buffer: []rune("hello, world"),
				cursor: 12,
			},
			wantBuffer: []rune("hello, world"),
			wantCursor: 12,
		},
		{
			line: &line{
				buffer: []rune("hello, world"),
				cursor: 0,
			},
			wantBuffer: []rune("hello, world"),
			wantCursor: 12,
		},
		{
			line: &line{
				buffer: []rune("hello, world"),
				cursor: 9,
			},
			wantBuffer: []rune("hello, world"),
			wantCursor: 12,
		},
	}

	for _, c := range cases {
		c.line.GoEnd()
		if !reflect.DeepEqual(c.line.buffer, c.wantBuffer) {
			t.Errorf("line: %v, after GoEnd(), line.buffer: %v, want: %v.", c.line, c.line.buffer, c.wantBuffer)
		}
		if c.line.cursor != c.wantCursor {
			t.Errorf("line: %v, after GoEnd(), line.cursor: %v, want: %v.", c.line, c.line.cursor, c.wantCursor)
		}
	}
}

func TestLineGoBackOneCharacter(t *testing.T) {
	cases := []struct {
		line       *line
		wantBuffer []rune
		wantCursor int
	}{
		{
			line: &line{
				buffer: []rune("hello, world"),
				cursor: 12,
			},
			wantBuffer: []rune("hello, world"),
			wantCursor: 11,
		},
		{
			line: &line{
				buffer: []rune("hello, world"),
				cursor: 0,
			},
			wantBuffer: []rune("hello, world"),
			wantCursor: 0,
		},
		{
			line: &line{
				buffer: []rune("hello, world"),
				cursor: 9,
			},
			wantBuffer: []rune("hello, world"),
			wantCursor: 8,
		},
	}

	for _, c := range cases {
		c.line.GoBackOneCharacter()
		if !reflect.DeepEqual(c.line.buffer, c.wantBuffer) {
			t.Errorf("line: %v, after GoBackOneCharacter(), line.buffer: %v, want: %v.", c.line, c.line.buffer, c.wantBuffer)
		}
		if c.line.cursor != c.wantCursor {
			t.Errorf("line: %v, after GoBackOneCharacter(), line.cursor: %v, want: %v.", c.line, c.line.cursor, c.wantCursor)
		}
	}
}

func TestLineGoForwardOneCharacter(t *testing.T) {
	cases := []struct {
		line       *line
		wantBuffer []rune
		wantCursor int
	}{
		{
			line: &line{
				buffer: []rune("hello, world"),
				cursor: 12,
			},
			wantBuffer: []rune("hello, world"),
			wantCursor: 12,
		},
		{
			line: &line{
				buffer: []rune("hello, world"),
				cursor: 0,
			},
			wantBuffer: []rune("hello, world"),
			wantCursor: 1,
		},
		{
			line: &line{
				buffer: []rune("hello, world"),
				cursor: 9,
			},
			wantBuffer: []rune("hello, world"),
			wantCursor: 10,
		},
	}

	for _, c := range cases {
		c.line.GoForwardOneCharacter()
		if !reflect.DeepEqual(c.line.buffer, c.wantBuffer) {
			t.Errorf("line: %v, after GoForwardOneCharacter(), line.buffer: %v, want: %v.", c.line, c.line.buffer, c.wantBuffer)
		}
		if c.line.cursor != c.wantCursor {
			t.Errorf("line: %v, after GoForwardOneCharacter(), line.cursor: %v, want: %v.", c.line, c.line.cursor, c.wantCursor)
		}
	}
}

func TestLineGoBackOneWord(t *testing.T) {
	cases := []struct {
		line       *line
		wantBuffer []rune
		wantCursor int
	}{
		{
			line: &line{
				buffer: []rune("hello, !@#$world"),
				cursor: 16,
			},
			wantBuffer: []rune("hello, !@#$world"),
			wantCursor: 11,
		},
		{
			line: &line{
				buffer: []rune("hello, !@#$world"),
				cursor: 0,
			},
			wantBuffer: []rune("hello, !@#$world"),
			wantCursor: 0,
		},
		{
			line: &line{
				buffer: []rune("hello, !@#$world"),
				cursor: 12,
			},
			wantBuffer: []rune("hello, !@#$world"),
			wantCursor: 11,
		},
		{
			line: &line{
				buffer: []rune("hello, !@#$world"),
				cursor: 11,
			},
			wantBuffer: []rune("hello, !@#$world"),
			wantCursor: 0,
		},
	}

	for _, c := range cases {
		c.line.GoBackOneWord()
		if !reflect.DeepEqual(c.line.buffer, c.wantBuffer) {
			t.Errorf("line: %v, after GoBackOneWord(), line.buffer: %v, want: %v.", c.line, c.line.buffer, c.wantBuffer)
		}
		if c.line.cursor != c.wantCursor {
			t.Errorf("line: %v, after GoBackOneWord(), line.cursor: %v, want: %v.", c.line, c.line.cursor, c.wantCursor)
		}
	}
}

func TestLineGoForwardOneWord(t *testing.T) {
	cases := []struct {
		line       *line
		wantBuffer []rune
		wantCursor int
	}{
		{
			line: &line{
				buffer: []rune("hello, !@#$world"),
				cursor: 16,
			},
			wantBuffer: []rune("hello, !@#$world"),
			wantCursor: 16,
		},
		{
			line: &line{
				buffer: []rune("hello, !@#$world"),
				cursor: 0,
			},
			wantBuffer: []rune("hello, !@#$world"),
			wantCursor: 5,
		},
		{
			line: &line{
				buffer: []rune("hello, !@#$world"),
				cursor: 3,
			},
			wantBuffer: []rune("hello, !@#$world"),
			wantCursor: 5,
		},
		{
			line: &line{
				buffer: []rune("hello, !@#$world!@#$"),
				cursor: 5,
			},
			wantBuffer: []rune("hello, !@#$world!@#$"),
			wantCursor: 16,
		},
	}

	for _, c := range cases {
		c.line.GoForwardOneWord()
		if !reflect.DeepEqual(c.line.buffer, c.wantBuffer) {
			t.Errorf("line: %v, after GoForwardOneWord(), line.buffer: %v, want: %v.", c.line, c.line.buffer, c.wantBuffer)
		}
		if c.line.cursor != c.wantCursor {
			t.Errorf("line: %v, after GoForwardOneWord(), line.cursor: %v, want: %v.", c.line, c.line.cursor, c.wantCursor)
		}
	}
}

func TestTermEscape(t *testing.T) {
	cases := []struct {
		in   []byte
		want []rune
	}{
		{
			in:   []byte("你好, world"),
			want: []rune("你好, world"),
		},
		{
			in:   []byte{'h', 'e', 'l', 'l', 'o', asciiDEL},
			want: []rune("hell"),
		},
		{
			in:   []byte{'h', 'e', 'l', 'l', 'o', asciiSOH},
			want: []rune("hello"),
		},
		{
			in:   []byte{'w', 'o', 'r', 'l', 'd', asciiSOH, 'h', 'e', 'l', 'l', 'o', ',', ' '},
			want: []rune("hello, world"),
		},
		{
			in:   append([]byte{',', ' ', 'w', 'o', 'r', 'l', 'd', asciiSOH}, []byte("你好")...),
			want: []rune("你好, world"),
		},
		{
			in:   []byte{'w', 'o', 'r', 'l', 'd', asciiSOH, ',', ' ', asciiSOH, 'h', 'e', 'l', 'l', 'o'},
			want: []rune("hello, world"),
		},
		{
			in:   []byte{'h', 'e', 'l', 'l', 'o', asciiENQ},
			want: []rune("hello"),
		},
		{
			in:   []byte{'h', 'e', 'l', 'l', 'o', asciiENQ, ',', ' ', 'w', 'o', 'r', 'l', 'd'},
			want: []rune("hello, world"),
		},
		{
			in:   []byte{'w', 'o', 'r', asciiSOH, 'h', 'e', 'l', 'l', 'o', ',', ' ', asciiENQ, 'l', 'd'},
			want: []rune("hello, world"),
		},
		{
			in:   []byte{'h', 'e', 'l', 'l', 'o', asciiESC, asciiLeftSquareBracket, asciiD, asciiNAK, 'h', 'e', 'l', 'l', asciiENQ, ',', ' ', 'w', 'o', 'r', 'l', 'd'},
			want: []rune("hello, world"),
		},
		{
			in:   []byte{'w', 'o', 'r', asciiSOH, 'h', 'e', 'l', 'l', 'o', ',', ' ', asciiESC, asciiLeftSquareBracket, asciiC, asciiESC, asciiLeftSquareBracket, asciiC, asciiESC, asciiLeftSquareBracket, asciiC, 'l', 'd'},
			want: []rune("hello, world"),
		},
		{
			in:   []byte{'h', 'e', 'l', 'l', 'o', ',', ' ', ' ', asciiETB, 'h', 'e', 'l', 'l', 'o', ',', ' ', 'w', 'o', 'r', 'l', 'd'},
			want: []rune("hello, world"),
		},
		{
			in:   []byte{'h', 'e', 'l', 'l', 'o', ',', ' ', 'w', 'o', 'r', 'l', 'd', asciiETB, 'w', 'o', 'r', 'l', 'd'},
			want: []rune("hello, world"),
		},
		{
			in:   []byte{'h', 'e', 'l', 'l', 'o', ',', ' ', 'w', 'o', 'r', 'l', 'd', asciiESC, asciiLeftSquareBracket, asciiD, asciiETB, 'w', 'o', 'r', 'l'},
			want: []rune("hello, world"),
		},
		{
			in:   []byte{'h', 'e', 'l', 'l', 'o', ',', ' ', asciiFF, 'w', 'o', 'r', 'l', 'd'},
			want: []rune("hello, world"),
		},
		{
			in:   []byte{'h', 'e', 'l', 'l', 'o', asciiSOH, asciiEOT, asciiEOT, asciiEOT, asciiEOT, asciiEOT, 'h', 'e', 'l', 'l', 'o', ',', ' ', 'w', 'o', 'r', 'l', 'd'},
			want: []rune("hello, world"),
		},
		{
			in:   append([]byte("你好"), append([]byte{asciiSOH, asciiEOT, asciiEOT}, []byte("你好, world")...)...),
			want: []rune("你好, world"),
		},
		{
			in:   []byte{'h', 'e', 'l', 'l', 'o', 'w', asciiSTX, ',', ' ', asciiENQ, 'o', 'r', 'l', 'd'},
			want: []rune("hello, world"),
		},
		{
			in:   []byte{'l', 'l', 'o', asciiSOH, 'h', 'e', asciiACK, asciiACK, asciiACK, ',', ' ', 'w', 'o', 'r', 'l', 'd'},
			want: []rune("hello, world"),
		},
		{
			in:   []byte{'l', 'l', 'o', ',', ' ', asciiESC, asciib, 'h', 'e', asciiENQ, 'w', 'o', 'r', 'l', 'd'},
			want: []rune("hello, world"),
		},
	}

	for _, c := range cases {
		got := TermEscape(c.in)
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("TermEscape(%v) == %v, want: %v.", c.in, got, c.want)
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
