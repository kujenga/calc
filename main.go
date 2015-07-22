package main

// a simple command line calculator
//
// based on https://en.wikipedia.org/wiki/Shunting-yard_algorithm

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"unicode"
)

var precedence = map[rune]int{
	'^': 4,
	'*': 3,
	'/': 3,
	'+': 2,
	'-': 2,
}

// if an operator is not in here, it is left associative
var ra = map[rune]bool{
	'^': true,
}

func main() {
	p := newParser(os.Stdin)
	p.Run()
}

// token

type token struct {
	val     string
	numeric bool
}

func newToken(val string, numeric bool) *token {
	return &token{
		val:     val,
		numeric: numeric,
	}
}

// parser

type parser struct {
	r      *bufio.Reader
	acc    string
	stack  []*rune
	sptr   int
	output []*token
	optr   int
}

func newParser(reader io.Reader) *parser {
	return &parser{
		r:      bufio.NewReader(reader),
		stack:  make([]*rune, 1024),
		sptr:   0,
		output: make([]*token, 1024),
		optr:   0,
	}
}

func (p *parser) Spush(t *rune) {
	p.stack[p.sptr] = t
	p.sptr++
}
func (p *parser) Speek() *rune {
	if p.sptr == 0 {
		return nil
	}
	return p.stack[p.sptr-1]
}
func (p *parser) Spop() *rune {
	if p.sptr == 0 {
		return nil
	}
	p.sptr--
	return p.stack[p.sptr] // no inline decrement?
}

func (p *parser) Opush(o *token) {
	p.output[p.optr] = o
	p.optr++
}
func (p *parser) OpushR(r *rune) {
	p.Opush(newToken(string(*r), false))
}
func (p *parser) Opop() *token {
	p.optr--
	return p.output[p.optr] // no inline decrement?
}

func (p *parser) String() string {
	acc := "["
	for _, t := range p.output[:p.optr] {
		acc = fmt.Sprintf("%v %v", acc, t.val)
	}
	acc += " ]\t["

	for _, r := range p.stack[:p.sptr] {
		acc = fmt.Sprintf("%v %v", acc, string(*r))
	}
	acc += " ]"
	return acc
}

func (p *parser) RPN() string {
	acc := ""
	for _, t := range p.output[:p.optr] {
		acc = fmt.Sprintf("%v %v", acc, t.val)
	}
	if acc != "" {
		acc = acc[1:]
	}
	return acc
}

func (p *parser) SignalInvalid(r rune) {
	p.r.Reset(os.Stdin)
	fmt.Printf("invalid input: %#U\n", r)
}

func (p *parser) HandleOperator(o1 rune) {
	o1p := precedence[o1]
	for {
		ptr := p.Speek()
		if ptr == nil {
			break
		}
		o2 := *ptr

		o2p, ok := precedence[o2]
		if ok && (!ra[o2] && o1p <= o2p || ra[o2] && o1p < o2p) {
			// remove from stack, add to output
			p.OpushR(p.Spop())
		} else {
			break
		}
	}
	p.Spush(&o1)
}

func (p *parser) HandleRune(r rune) bool {
	// accumulating digits of a number
	if unicode.IsDigit(r) {
		p.acc += string(r)
		return false
	}

	// the previous number is finished, push it to the stack
	if p.acc != "" {
		p.Opush(newToken(p.acc, true))
		p.acc = ""
	}

	// handle operators, which are in the precedence value map
	if _, ok := precedence[r]; ok {
		p.HandleOperator(r)
		return false
	}

	// test for value of the next symbol
	switch r {
	case '(':
		p.Spush(&r)

	case ')':
		// Until we see left paren, pop the stack onto the output queue
	rparen:
		for {
			ptr := p.Spop()
			if ptr == nil {
				p.SignalInvalid(')')
				break rparen
			}
			if *ptr == '(' {
				break rparen
			}
			p.OpushR(ptr)
			continue
		}

	case '\n':
		p.Finish()
		return true

	default:
		if !unicode.IsSpace(r) {
			fmt.Println("hit default")
			p.SignalInvalid(r)
		}
	}
	return false
}

func (p *parser) Finish() {
	// if a number is sitting in the accumulator, handle it
	if p.acc != "" {
		p.Opush(newToken(p.acc, true))
		p.acc = ""
	}

	// finish
	for {
		ptr := p.Spop()
		if ptr == nil {
			break
		}
		if _, ok := precedence[*ptr]; !ok {
			p.SignalInvalid(*ptr)
			return
		}
		p.OpushR(ptr)
	}
}

func (p *parser) Evaluate() int {

	return 0
}

func (p *parser) Reset() {
	p.sptr = 0
	p.optr = 0
	fmt.Print("> ")
}

func (p *parser) Run() {
	fmt.Print("> ")
	for {
		r, _, err := p.r.ReadRune()
		if err != nil {
			fmt.Printf("error: %v\n", err)
			break
		}

		done := p.HandleRune(r)
		if done {
			fmt.Println("rpn:", p.RPN(), "=", p.Evaluate())
			p.Reset()
		}
	}
}
