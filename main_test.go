package main

import (
	"fmt"
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

func TestPrecedence(t *testing.T) {
	ops := []rune{'^', '*', '/', '+', '-'}
	for _, r := range ops {
		if _, ok := precedence[r]; !ok {
			t.Errorf("%v should have been in precedence", string(r))
		}
	}
}

func TestParser(t *testing.T) {
	testcases := map[string]string{
		"2 + 4":             "2 4 +",
		"3 + 4 * 2":         "3 4 2 * +",
		"(445+(354*95463))": "445 354 95463 * +",
		"4 ^ 3":             "4 3 ^",
		"3 + 4 * 2 / ( 1 - 5 ) ^ 2 ^ 3": "3 4 2 * 1 5 - 2 3 ^ ^ / +",
	}

	for given, expect := range testcases {
		p := newParser(nil)
		for _, r := range given {
			p.HandleRune(r)
		}
		p.Finish()

		if got := p.RPN(); got != expect {
			t.Errorf("result: '%v' != expected: '%v'", got, expect)
		}
	}
}
