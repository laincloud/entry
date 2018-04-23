package term

import (
	"reflect"
	"testing"
)

func TestEscapeInput(t *testing.T) {
	cases := []struct {
		in   []byte
		want []byte
	}{
		{
			in:   []byte("你好, world"),
			want: []byte("你好, world"),
		},
		{
			in:   []byte{'h', 'e', 'l', 'l', 'o', asciiDEL},
			want: []byte("hell"),
		},
		{
			in:   []byte{'h', 'e', 'l', 'l', 'o', asciiSOH},
			want: []byte("hello"),
		},
		{
			in:   []byte{'w', 'o', 'r', 'l', 'd', asciiSOH, 'h', 'e', 'l', 'l', 'o', ',', ' '},
			want: []byte("hello, world"),
		},
		{
			in:   append([]byte{',', ' ', 'w', 'o', 'r', 'l', 'd', asciiSOH}, []byte("你好")...),
			want: []byte("你好, world"),
		},
		{
			in:   []byte{'w', 'o', 'r', 'l', 'd', asciiSOH, ',', ' ', asciiSOH, 'h', 'e', 'l', 'l', 'o'},
			want: []byte("hello, world"),
		},
		{
			in:   []byte{'h', 'e', 'l', 'l', 'o', asciiENQ},
			want: []byte("hello"),
		},
		{
			in:   []byte{'h', 'e', 'l', 'l', 'o', asciiENQ, ',', ' ', 'w', 'o', 'r', 'l', 'd'},
			want: []byte("hello, world"),
		},
		{
			in:   []byte{'w', 'o', 'r', asciiSOH, 'h', 'e', 'l', 'l', 'o', ',', ' ', asciiENQ, 'l', 'd'},
			want: []byte("hello, world"),
		},
		{
			in:   []byte{'h', 'e', 'l', 'l', 'o', asciiESC, asciiLeftSquareBracket, asciiD, asciiNAK, 'h', 'e', 'l', 'l', asciiENQ, ',', ' ', 'w', 'o', 'r', 'l', 'd'},
			want: []byte("hello, world"),
		},
		{
			in:   []byte{'w', 'o', 'r', asciiSOH, 'h', 'e', 'l', 'l', 'o', ',', ' ', asciiESC, asciiLeftSquareBracket, asciiC, asciiESC, asciiLeftSquareBracket, asciiC, asciiESC, asciiLeftSquareBracket, asciiC, 'l', 'd'},
			want: []byte("hello, world"),
		},
		{
			in:   []byte{'h', 'e', 'l', 'l', 'o', ',', ' ', ' ', asciiETB, 'h', 'e', 'l', 'l', 'o', ',', ' ', 'w', 'o', 'r', 'l', 'd'},
			want: []byte("hello, world"),
		},
		{
			in:   []byte{'h', 'e', 'l', 'l', 'o', ',', ' ', 'w', 'o', 'r', 'l', 'd', asciiETB, 'w', 'o', 'r', 'l', 'd'},
			want: []byte("hello, world"),
		},
		{
			in:   []byte{'h', 'e', 'l', 'l', 'o', ',', ' ', 'w', 'o', 'r', 'l', 'd', asciiESC, asciiLeftSquareBracket, asciiD, asciiETB, 'w', 'o', 'r', 'l'},
			want: []byte("hello, world"),
		},
		{
			in:   []byte{'h', 'e', 'l', 'l', 'o', ',', ' ', asciiFF, 'w', 'o', 'r', 'l', 'd'},
			want: []byte("hello, world"),
		},
		{
			in:   []byte{'h', 'e', 'l', 'l', 'o', asciiSOH, asciiEOT, asciiEOT, asciiEOT, asciiEOT, asciiEOT, 'h', 'e', 'l', 'l', 'o', ',', ' ', 'w', 'o', 'r', 'l', 'd'},
			want: []byte("hello, world"),
		},
		{
			in:   append([]byte("你好"), append([]byte{asciiSOH, asciiEOT, asciiEOT}, []byte("你好, world")...)...),
			want: []byte("你好, world"),
		},
		{
			in:   []byte{'h', 'e', 'l', 'l', 'o', 'w', asciiSTX, ',', ' ', asciiENQ, 'o', 'r', 'l', 'd'},
			want: []byte("hello, world"),
		},
		{
			in:   []byte{'l', 'l', 'o', asciiSOH, 'h', 'e', asciiACK, asciiACK, asciiACK, ',', ' ', 'w', 'o', 'r', 'l', 'd'},
			want: []byte("hello, world"),
		},
		{
			in:   []byte{'l', 'l', 'o', ',', ' ', asciiESC, asciib, 'h', 'e', asciiENQ, 'w', 'o', 'r', 'l', 'd'},
			want: []byte("hello, world"),
		},
	}

	for _, c := range cases {
		got := EscapeInput(c.in)
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("EscapeInput(%v) == %v, want: %v.", c.in, got, c.want)
		}
	}
}

func TestEscapeTabCompletion(t *testing.T) {
	cases := []struct {
		in   []byte
		want []byte
	}{
		{
			in:   []byte{asciiBEL},
			want: []byte{},
		},
		{
			in:   []byte("hello world"),
			want: []byte{},
		},
		{
			in:   []byte("hello "),
			want: []byte("hello "),
		},
		{
			in:   []byte("hello"),
			want: []byte("hello"),
		},
	}

	for _, c := range cases {
		got := EscapeTabCompletion(c.in)
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("EscapeTabCompletion(%v) == %v, want: %v.", c.in, got, c.want)
		}
	}
}

func TestEscapeHistoryCommand(t *testing.T) {
	cases := []struct {
		in   []byte
		want []byte
	}{
		{
			in:   []byte{asciiBEL},
			want: []byte{},
		},
		{
			in:   []byte("hello world"),
			want: []byte("hello world"),
		},
	}

	for _, c := range cases {
		got := EscapeHistoryCommand(c.in)
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("EscapeHistoryCommand(%v) == %v, want: %v.", c.in, got, c.want)
		}
	}
}

func TestHasUpArrowSuffix(t *testing.T) {
	cases := []struct {
		in   []byte
		want bool
	}{
		{
			in:   []byte("he"),
			want: false,
		},
		{
			in:   []byte("hello world"),
			want: false,
		},
		{
			in:   UpArrow,
			want: true,
		},
		{
			in:   append([]byte("hello"), UpArrow...),
			want: true,
		},
		{
			in:   append(append([]byte("hello"), UpArrow...), []byte("world")...),
			want: false,
		},
	}

	for _, c := range cases {
		got := HasUpArrowSuffix(c.in)
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("HasUpArrowSuffix(%v) == %v, want: %v.", c.in, got, c.want)
		}
	}
}

func TestHasDownArrowSuffix(t *testing.T) {
	cases := []struct {
		in   []byte
		want bool
	}{
		{
			in:   []byte("he"),
			want: false,
		},
		{
			in:   []byte("hello world"),
			want: false,
		},
		{
			in:   DownArrow,
			want: true,
		},
		{
			in:   append([]byte("hello"), DownArrow...),
			want: true,
		},
		{
			in:   append(append([]byte("hello"), DownArrow...), []byte("world")...),
			want: false,
		},
	}

	for _, c := range cases {
		got := HasDownArrowSuffix(c.in)
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("HasDownArrowSuffix(%v) == %v, want: %v.", c.in, got, c.want)
		}
	}
}

func TestIsBell(t *testing.T) {
	cases := []struct {
		in   []byte
		want bool
	}{
		{
			in:   []byte("he"),
			want: false,
		},
		{
			in:   []byte{asciiBEL},
			want: true,
		},
		{
			in:   append([]byte("hello"), asciiBEL),
			want: false,
		},
	}

	for _, c := range cases {
		got := IsBell(c.in)
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("IsBell(%v) == %v, want: %v.", c.in, got, c.want)
		}
	}
}

func TestIsCR(t *testing.T) {
	cases := []struct {
		in   []byte
		want bool
	}{
		{
			in:   []byte("he"),
			want: false,
		},
		{
			in:   []byte{asciiCR},
			want: true,
		},
		{
			in:   append([]byte("hello"), asciiCR),
			want: false,
		},
	}

	for _, c := range cases {
		got := IsCR(c.in)
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("IsCR(%v) == %v, want: %v.", c.in, got, c.want)
		}
	}
}

func TestIsTab(t *testing.T) {
	cases := []struct {
		in   []byte
		want bool
	}{
		{
			in:   []byte("he"),
			want: false,
		},
		{
			in:   []byte{asciiHT},
			want: true,
		},
		{
			in:   append([]byte("hello"), asciiHT),
			want: false,
		},
	}

	for _, c := range cases {
		got := IsTab(c.in)
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("IsTab(%v) == %v, want: %v.", c.in, got, c.want)
		}
	}
}
