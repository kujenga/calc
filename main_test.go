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
	_ = "7 + ( 8 * 9 )"
}
