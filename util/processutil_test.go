// processutil_test.go
// go test "dataserver/util" -v
package util

import (
	"bytes"
	"fmt"
	"testing"
)

/*
func IntMin(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// A test is created by writing a function with a name
// beginning with `Test`.
func TestIntMinBasic(t *testing.T) {
	fmt.Println("Running test")
	ans := IntMin(2, -2)
	if ans != -2 {
		// `t.Error*` will report test failures but continue
		// executing the test. `t.Fatal*` will report test
		// failures and stop the test immediately.
		t.Errorf("IntMin(2, -2) = %d; want -2", ans)
	}
}

func TestIntMinTableDriven(t *testing.T) {
	var tests = []struct {
		a, b int
		want int
	}{
		{0, 1, 0},
		{1, 0, 0},
		{2, -2, -2},
		{0, -1, -1},
		{-1, 0, -1},
	}

	for _, tt := range tests {
		// t.Run enables running "subtests", one for each
		// table entry. These are shown separately
		// when executing `go test -v`.
		testname := fmt.Sprintf("%d,%d", tt.a, tt.b)
		t.Run(testname, func(t *testing.T) {
			ans := IntMin(tt.a, tt.b)
			if ans != tt.want {
				t.Errorf("got %d, want %d", ans, tt.want)
			}
		})
	}
}
*/
func TestCgi(t *testing.T) {
	/*
		var buf bytes.Buffer

		RunCgi(&buf, `ls`, []string{`ls`})
	*/
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "Size: %d MB.", 85)
	s := buf.String() // s == "Size: 85 MB."
	fmt.Println(s)
	RunCgi(&buf, `ls`, []string{`-la`})

	fmt.Printf("CGI OUTPUT: %s\n", buf.String())
}
