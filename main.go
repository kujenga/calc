package main

// a simple command line calculator
//
// based on https://en.wikipedia.org/wiki/Shunting-yard_algorithm

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"unicode"
	"unicode/utf8"
)

// see: https://en.wikipedia.org/wiki/Exponentiation_by_squaring
func pow(x, n int) int {
	if n == 0 {
		return 1
	}
	if n == 1 {
		return x
	}
	if n%2 == 0 { // even
		return pow(x*x, n/2)
	}
	// odd
	return x * pow(x*x, (n-1)/2)
}

var precedence = map[rune]int{
	'^': 4,
	'*': 3,
	'/': 3,
	'+': 2,
	'-': 2,
}

type opfunc func(a, b int) int

var opfuncs = map[rune]opfunc{
	'^': func(a, b int) int {
		return pow(a, b)
	},
	'*': func(a, b int) int {
		return a * b
	},
	'/': func(a, b int) int {
		return a / b
	},
	'+': func(a, b int) int {
		return a + b
	},
	'-': func(a, b int) int {
		return a - b
	},
}

// if an operator is not in here, it is left associative
var ra = map[rune]bool{
	'^': true,
}

func main() {
	p := newParser(os.Stdin, os.Stdout)
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

func (t *token) IsOperator() bool {
	if len(t.val) != 1 {
		return false
	}
	r, _ := utf8.DecodeRuneInString(t.val)
	_, ok := precedence[r]
	return ok
}

func (t *token) ToRune() rune {
	r, _ := utf8.DecodeRuneInString(t.val)
	return r
}

func (t *token) ToInt() int {
	i, err := strconv.Atoi(t.val)
	if err != nil {
		panic(err.Error())
	}
	return i
}

// parser

type parser struct {
	r      *bufio.Reader
	w      *bufio.Writer
	acc    string
	stack  []*rune
	sptr   int
	output []*token
	optr   int
}

func newParser(reader io.Reader, writer io.Writer) *parser {
	if reader == nil {
		reader = os.Stdin
	}
	if writer == nil {
		writer = os.Stdout
	}
	return &parser{
		r:      bufio.NewReader(reader),
		w:      bufio.NewWriter(writer),
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
		for {
			ptr := p.Spop()
			if ptr == nil {
				p.SignalInvalid(')')
				break
			}
			if *ptr == '(' {
				break
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

func (p *parser) Eval() int {
	// eval stack
	if p.optr == 0 {
		return 0
	}

	s := make([]int, p.optr)
	sptr := 0
	pop := func() int {
		sptr--
		return s[sptr]
	}
	push := func(i int) {
		s[sptr] = i
		sptr++
	}

	for _, t := range p.output[:p.optr] {
		if t.numeric {
			push(t.ToInt())
		} else if t.IsOperator() {
			b := pop()
			a := pop()
			o := opfuncs[t.ToRune()](a, b)
			push(o)
		}
	}
	return pop()
}

func (p *parser) Parse(str string) {
	for _, r := range str {
		p.HandleRune(r)
	}
	p.Finish()
}

func (p *parser) Reset() {
	p.sptr = 0
	p.optr = 0
	fmt.Fprint(p.w, "> ")
}

func (p *parser) Run() {
	fmt.Fprint(p.w, "> ")
	for {
		r, _, err := p.r.ReadRune()
		if err != nil {
			fmt.Printf("error: %v\n", err)
			break
		}

		done := p.HandleRune(r)
		if done {
			fmt.Fprintln(p.w, "rpn:", p.RPN(), "=", p.Eval())
			p.Reset()
		}
	}
}
