package util

import (
	"testing"
)

func TestGetValidUTF8Length(t *testing.T) {
	testCase := []byte{228, 184, 150, 228, 184}
	if actual := getValidUT8Length(testCase); actual != 3 {
		t.Errorf("Case 1 failed: actual is %d", actual)
	}
	if actual := getValidUT8Length(testCase[3:]); actual != 0 {
		t.Errorf("Case 2 failed: actual is %d", actual)
	}
	if actual := getValidUT8Length(testCase[:0]); actual != 0 {
		t.Errorf("Case 3 failed: actual is %d", actual)
	}
	if actual := getValidUT8Length(testCase[:1]); actual != 0 {
		t.Errorf("Case 4 failed: actual is %d", actual)
	}
}
