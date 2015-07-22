package main

import (
	"bytes"
	"io"
	"regexp"
	"testing"
)

func TestPow(t *testing.T) {
	if r := pow(2, 4); r != 16 {
		t.Errorf("2^4 is 16, not %v", r)
	}
}

func TestParserStack(t *testing.T) {
	p := newParser(nil, nil)

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

var testcases = map[string]string{
	"":                  "",
	"2 + 4":             "2 4 +",
	"3 + 4 * 2":         "3 4 2 * +",
	"(445+(354*95463))": "445 354 95463 * +",
	"4 ^ 3":             "4 3 ^",
	"3 + 4 * 2 / ( 1 - 5 ) ^ 2 ^ 3": "3 4 2 * 1 5 - 2 3 ^ ^ / +",
}

func TestParser(t *testing.T) {
	for given, expect := range testcases {
		p := newParser(nil, nil)
		p.Parse(given)

		// test that string debugging method matches expected form
		if m, err := regexp.Match(`\[ (.*? )*?\]\t\[ \]`, []byte(p.String())); !m || err != nil {
			t.Errorf("failed to match or error occured %v", err)
		}

		if got := p.RPN(); got != expect {
			t.Errorf("result: '%v' != expected: '%v'", got, expect)
		}
	}
}

func TestParserInput(t *testing.T) {
	r, w := io.Pipe()
	output := bytes.NewBuffer(nil)
	p := newParser(r, output)
	go p.Run()

	for given := range testcases {
		given += "\n" // add a newline for termination

		n, err := w.Write([]byte(given))
		if n != len([]byte(given)) || err != nil {
			t.Errorf("error in writing %v bytes: %v", n, err)
		}

		// TODO: check output
	}
}

var testoutcomes = map[string]int{
	"2 + 4":                       6,
	"3 + 4 * 2":                   11,
	"(445+(354*95463))":           33794347,
	"( 1 - 5 ) ^ 2":               16,
	"(8 + 4 * 2) / ( 1 - 5 ) ^ 2": 1,
}

func TestParserEvaluation(t *testing.T) {
	p := newParser(nil, nil)
	for given, expect := range testoutcomes {
		p.Parse(given)

		if res := p.Eval(); res != expect {
			t.Errorf("result %v didn't equal %v", res, expect)
		}
		p.Reset()
	}
}
