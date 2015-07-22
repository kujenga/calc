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
	'−': 2,
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
	return acc
}

func (p *parser) SignalInvalid(r rune) {
	p.r.Reset(os.Stdin)
	fmt.Println("invalid input", string(r))
}

func (p *parser) HandleOperator(op rune) {
	prec := precedence[op]
Outer:
	for {
		ptr := p.Speek()
		if ptr == nil {
			break Outer
		}
		next := *ptr

		nPrec, ok := precedence[next]
		if ok && (ra[next] && nPrec < prec || !ra[next] && nPrec <= prec) {
			// remove from stack, add to output
			p.OpushR(p.Spop())
		} else {
			break Outer
		}
	}
	p.Spush(&op)
}

func (p *parser) HandleRParen() {
	for {
		ptr := p.Spop()
		if ptr == nil {
			p.SignalInvalid(')')
			return
		}
		if *ptr == '(' {
			return
		}
		// Until we see left paren, pop the stack onto the output queue
		p.OpushR(ptr)
		continue
	}
}

func (p *parser) HandleRune(r rune) {

}

func (p *parser) Eval() {
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

func (p *parser) Reset() {
	p.sptr = 0
	p.optr = 0
	fmt.Print("> ")
}

func (p *parser) Run() {
	fmt.Print("> ")
	acc := ""
	for {
		r, _, err := p.r.ReadRune()
		if err != nil {
			fmt.Printf("error: %v\n", err)
			break
		}

		// accumulating digits of a number
		if unicode.IsDigit(r) {
			acc += string(r)
			continue
		}

		// the previous number is finished, push it to the stack
		if acc != "" {
			p.Opush(newToken(acc, true))
			acc = ""
		}

		// handle operators, which are in the precedence value map
		if _, ok := precedence[r]; ok {
			p.HandleOperator(r)
			continue
		}

		// test for value of the next symbol
		switch r {
		case '(':
			p.Spush(&r)

		case ')':
			p.HandleRParen()

		case '\n':
			p.Eval()
			fmt.Println("rpn:", p.RPN())
			p.Reset()

		default:
			if !unicode.IsSpace(r) {
				p.SignalInvalid(r)
			}
		}
	}
}
