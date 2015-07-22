package main

import (
	"os"
	"testing"
)

func TestParserStack(t *testing.T) {
	p := newParser(os.Stdin)

	x := 'v'
	p.Spush(&x)
	if v := *p.Spop(); v != x {
		t.Errorf("value incorrect %v", v)
	}
	if p.sptr != 0 {
		t.Errorf("sptr had wrong value: %v", p.sptr)
	}
}

func TestParser(t *testing.T) {
	testcases := map[string]string{
		"2 + 4":             "2 4 +",
		"(445+(354*95463))": "445 354 95463 * +",
		// "3 + 4 * 2 / ( 1 - 5 ) ^ 2 ^ 3": "3 4 2 * 1 5 - 2 3 ^ ^ / +",
	}

	for given, expect := range testcases {
		p := newParser(nil)
		for _, r := range given {
			p.HandleRune(r)
		}
		p.Finish()

		if got := p.RPN(); got != expect {
			t.Errorf("from %v got '%v' != expected '%v'", p.String(), got, expect)
		}
	}
}
