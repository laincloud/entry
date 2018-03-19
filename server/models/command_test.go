package models

import (
	"testing"
)

func TestIsRisky(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{
			in:   "vipw",
			want: true,
		},
		{
			in:   "a vipw b",
			want: true,
		},
		{
			in:   "chmod 777",
			want: true,
		},
		{
			in:   "chmod  777",
			want: true,
		},
		{
			in:   "chmod777",
			want: false,
		},
	}

	for _, c := range cases {
		got := isRisky(c.in)
		if got != c.want {
			t.Errorf("isRisky(%v) == %v, want: %v.", c.in, got, c.want)
		}
	}
}
