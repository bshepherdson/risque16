package main

import (
	"fmt"
	"io"
	"strconv"
	"strings"
)

type Parser struct {
	s   *Scanner
	buf struct {
		tok Token  // Last read token.
		lit string // Last read literal
		n   int    // buffer size (max=1)
	}
}

// NewParser returns a new Parser instance.
func NewParser(filename string, r io.Reader) *Parser {
	return &Parser{s: NewScanner(filename, r)}
}

// scan returns the next token from the underlying scanner.
// If a token has been unscanned then read that instead.
func (p *Parser) scan() (Token, string) {
	// If we have a token on the buffer, then return it.
	if p.buf.n != 0 {
		p.buf.n = 0
		return p.buf.tok, p.buf.lit
	}

	// Otherwise read the next token from the scanner.
	tok, lit := p.s.Scan()

	// Save it to the buffer in case we unscan later.
	p.buf.tok, p.buf.lit = tok, lit
	return tok, lit
}

// Unscan pushes previously read token back onto the buffer.
func (p *Parser) unscan() {
	p.buf.n = 1
}

// scanIgnoreWhitespace is a wrapper that skips whitespace tokens.
// NEWLINE is not a whitespace token according to this; those are important.
func (p *Parser) scanIgnoreWhitespace() (Token, string) {
	tok, lit := p.scan()
	if tok == WS {
		tok, lit = p.scan()
	}
	return tok, lit
}

func (p *Parser) wrapError(e error) error {
	return fmt.Errorf("Parse error at %s   %v", p.s.Location(), e)
}

// Actual top-level parser. Returns our AST object.
func (p *Parser) Parse() (*AST, error) {
	lines := make([]Assembled, 0, 256)
	for {
		tok, lit := p.scanIgnoreWhitespace()
		if tok == DOT {
			l, err := p.parseDirective()
			if err != nil {
				return nil, p.wrapError(err)
			}
			lines = append(lines, l)
		} else if tok == IDENT { // Should be an instruction.
			upper := strings.ToUpper(lit)
			if _, ok := instructions[upper]; ok {
				l, err := p.parseInstruction(upper)
				if err != nil {
					return nil, p.wrapError(err)
				}
				lines = append(lines, l)
			}
		} else if tok == COLON { // Label definition
			tok, lit = p.scan() // WS not allowed.
			if tok != IDENT {
				lines = append(lines, &LabelDef{lit})
			}
		} else if tok == NEWLINE {
			continue
		} else if tok == EOF {
			break
		} else {
			return nil, p.wrapError(fmt.Errorf("Unexpected %s", tokenNames[tok]))
		}
	}
	return &AST{lines}, nil
}

func (p *Parser) parseDirective() (Assembled, error) {
	dir, lit := p.scan() // No whitespace after the .
	if dir != IDENT {
		return nil, fmt.Errorf("Expected directive command after dot, but found %s", tokenNames[dir])
	}

	switch strings.ToUpper(lit) {
	case "DAT":
		// Comma-separated expressions.
		args, err := p.parseExprList(true /* strings allowed */)
		if err != nil {
			return nil, fmt.Errorf("Failed to parse .DAT values: %v", err)
		}
		if !p.consume(NEWLINE) {
			t, lit := p.scanIgnoreWhitespace()
			return nil, fmt.Errorf("Unexpected %s '%s' at end of line", tokenNames[t], lit)
		}
		return &DatBlock{args}, nil

	case "ORG":
		expr, err := p.parseSimpleExpr()
		if err != nil {
			return nil, fmt.Errorf("Bad expression for .ORG: %v", err)
		}
		if !p.consume(NEWLINE) {
			t, lit := p.scanIgnoreWhitespace()
			return nil, fmt.Errorf("Unexpected %s '%s' at end of line", tokenNames[t], lit)
		}
		return &Org{expr}, nil

	case "FILL":
		values, err := p.parseExprList(false /* no strings */)
		if err != nil {
			return nil, fmt.Errorf("Failed to parse .FILL arguments: %v", err)
		}
		if len(values) != 2 {
			return nil, fmt.Errorf(".FILL requires two arguments, found %d", len(values))
		}
		if !p.consume(NEWLINE) {
			t, lit := p.scanIgnoreWhitespace()
			return nil, fmt.Errorf("Unexpected %s '%s' at end of line", tokenNames[t], lit)
		}
		return &FillBlock{values[1], values[0]}, nil

	case "RESERVE":
		loc := p.s.Location()
		expr, err := p.parseSimpleExpr()
		if err != nil {
			return nil, fmt.Errorf("Bad expression for .RESERVE: %v", err)
		}
		if !p.consume(NEWLINE) {
			t, lit := p.scanIgnoreWhitespace()
			return nil, fmt.Errorf("Unexpected %s '%s' at end of line", tokenNames[t], lit)
		}
		return &FillBlock{&Constant{0, loc}, expr}, nil

	case "DEFINE":
		t, lit := p.scanIgnoreWhitespace()
		if t != IDENT {
			return nil, fmt.Errorf(".DEFINE's first argument must be an identifier; found %s", tokenNames[t])
		}

		if !p.consumeComma() {
			return nil, fmt.Errorf("No comma after .DEFINE identifier")
		}

		expr, err := p.parseSimpleExpr()
		if err != nil {
			return nil, fmt.Errorf("Bad expression for .DEFINE: %v", err)
		}
		if !p.consume(NEWLINE) {
			t, lit := p.scanIgnoreWhitespace()
			return nil, fmt.Errorf("Unexpected %s '%s' at end of line", tokenNames[t], lit)
		}
		return &SymbolDef{lit, expr}, nil

		// TODO: Macros
	}

	return nil, fmt.Errorf("Unknown directive: %s", lit)
}

// "Simple expression" is kind of a misnomer; it's actually any expression other
// than a string literal, since those are only allowed in
func (p *Parser) parseSimpleExpr() (Expression, error) {
	// TODO: More complete expression parsing.
	loc := p.s.Location()
	tok, lit := p.scanIgnoreWhitespace()
	switch tok {
	case IDENT:
		return &LabelUse{lit, loc}, nil
	case NUMBER:
		n, err := strconv.Atoi(lit)
		if err != nil {
			return nil, err
		}
		return &Constant{uint16(n), loc}, nil
	}
	return nil, fmt.Errorf("Found %s while parsing expression", tokenNames[tok])
}

func (p *Parser) parseExpr() ([]Expression, error) {
	// Either a string literal or a simple expression.
	loc := p.s.Location()
	tok, lit := p.scanIgnoreWhitespace()
	if tok == STRING {
		b := make([]Expression, len(lit))
		for i, c := range lit {
			b[i] = &Constant{uint16(c), loc}
		}
		return b, nil
	}
	// Unscan, otherwise, and try again.
	p.unscan()

	expr, err := p.parseSimpleExpr()
	if err != nil {
		return nil, err
	}
	buf := make([]Expression, 1)
	buf[0] = expr
	return buf, nil
}

func (p *Parser) parseExprList(allowStringLiterals bool) ([]Expression, error) {
	buf := make([]Expression, 0, 16)
	for {
		if allowStringLiterals {
			exprs, err := p.parseExpr()
			if err != nil {
				return nil, err
			}
			for _, e := range exprs {
				buf = append(buf, e)
			}
		} else {
			expr, err := p.parseSimpleExpr()
			if err != nil {
				return nil, err
			}
			buf = append(buf, expr)
		}

		// Now look for either a comma or a newline.
		if !p.consumeComma() {
			break
		}
	}
	if len(buf) == 0 {
		return nil, fmt.Errorf("Empty expression list")
	}
	return buf, nil
}

// Returns false if we can't find that token next.
func (p *Parser) consume(t Token) bool {
	tok, _ := p.scanIgnoreWhitespace()
	if tok != t {
		p.unscan()
	}
	return tok == t
}

func (p *Parser) consumeComma() bool {
	return p.consume(COMMA)
}

// Instruction parsing.
func (p *Parser) parseInstruction(opcode string) (Assembled, error) {
	// Special case for PUSH, POP, LDMIA, STMIA, LDR and STR.
	// They have their own rules for bracketing.
	if opcode == "PUSH" || opcode == "POP" {
		return p.parsePushPop(opcode)
	}
	if opcode == "LDMIA" || opcode == "STMIA" {
		return p.parseMultiStoreLoad(opcode)
	}
	if opcode == "LDR" || opcode == "STR" {
		return p.parseLoadStore(opcode)
	}

	// Parsing regular instructions: comma-separated list of arguments.
	args, err := p.parseArgList(opcode)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse argument list: %v", err)
	}
	return &Instruction{opcode, args, p.s.Location()}, nil
}

func (p *Parser) parseArgList(opcode string) ([]*Arg, error) {
	args := make([]*Arg, 0, 3)

	for {
		// Parse an arg. Register, PC, SP, literal, label expression.
		done := false
		r, err := p.parseReg()
		if err == nil {
			args = append(args, &Arg{kind: AT_REG, reg: r})
			done = true
		}

		if !done {
			lit, err := p.parseLiteral()
			if err == nil {
				args = append(args, &Arg{kind: AT_LITERAL, lit: lit})
				done = true
			}
		}

		if !done {
			expression, err := p.parseSimpleExpr()
			if err == nil {
				args = append(args, &Arg{kind: AT_LABEL, label: expression})
				done = true
			}
		}

		if !done {
			t, _ := p.scanIgnoreWhitespace()
			if t == PC {
				args = append(args, &Arg{kind: AT_PC})
			} else if t == SP {
				args = append(args, &Arg{kind: AT_SP})
			} else {
				// Found something unexpected.lit
				return nil, fmt.Errorf("Expected argument, but found %s", tokenNames[t])
			}
			done = true
		}

		// Now we expect a comma or newline.
		t, _ := p.scanIgnoreWhitespace()
		if t == NEWLINE {
			break
		} else if t != COMMA {
			return nil, fmt.Errorf("Expected comma or end of arg list, but found %s", tokenNames[t])
		}
	}
	return args, nil
}

func (p *Parser) parsePushPop(opcode string) (Assembled, error) {
	regs, lrpc, err := p.parseRlist(opcode, true)
	if err != nil {
		return nil, fmt.Errorf("Error parsing register list for %s: %v", opcode, err)
	}
	t, _ := p.scanIgnoreWhitespace()
	if t != NEWLINE {
		return nil, fmt.Errorf("Unexpected %s at end of line", tokenNames[t])
	}
	return &StackOp{regs, opcode == "PUSH", lrpc, 0xffff}, nil
}

func (p *Parser) parseMultiStoreLoad(opcode string) (Assembled, error) {
	base, err := p.parseReg()
	if err != nil {
		return nil, fmt.Errorf("Failed to parse base register of %s: %v", opcode, err)
	}

	if !p.consume(BANG) {
		return nil, fmt.Errorf("Bang (!) required after %s base register", opcode)
	}

	if !p.consumeComma() {
		t, _ := p.scanIgnoreWhitespace()
		return nil, fmt.Errorf("Expected comma and register list for %s, but found %s", opcode, tokenNames[t])
	}

	regs, lrpc, err := p.parseRlist(opcode, false)
	if err != nil {
		return nil, fmt.Errorf("Error parsing register list for %s: %v", opcode, err)
	}
	if lrpc {
		return nil, fmt.Errorf("LR and PC not allowed in register list for %s", opcode)
	}

	t, _ := p.scanIgnoreWhitespace()
	if t != NEWLINE {
		return nil, fmt.Errorf("Unexpected %s at end of line", tokenNames[t])
	}
	return &StackOp{regs, opcode == "STMIA", false, base}, nil
}

func (p *Parser) parseReg() (uint16, error) {
	t, lit := p.scanIgnoreWhitespace()
	if t == REGISTER {
		r, err := strconv.Atoi(lit[1:])
		if err != nil {
			return 0, fmt.Errorf("Failed to parse number in register: %s", lit)
		}
		return uint16(r), nil
	}
	p.unscan()
	return 0, fmt.Errorf("Expected register, but found %s", tokenNames[t])
}

func (p *Parser) parseRlist(opcode string, pclrAllowed bool) (uint16, bool, error) {
	var regs uint16
	var pclr bool

	if !p.consume(LBRACE) {
		return 0, false, fmt.Errorf("Could not parse Rlist")
	}

	// Now a comma-separated list of regs and PC or LR.
	for {
		r, err := p.parseReg()
		if err == nil && 0 <= r && r < 8 {
			// Include the reg and continue.
			regs = regs | (1 << uint(r))
			continue
		} else if r < 0 || r >= 8 {
			return 0, false, fmt.Errorf("No such register: r%d (only r0-r7)", r)
		}

		// If we're still here, try parsing PC or LR.
		t := ILLEGAL
		if pclrAllowed {
			t, _ := p.scanIgnoreWhitespace()
			if t == LR {
				if opcode == "POP" || opcode == "LDMIA" {
					return 0, false, fmt.Errorf("LR not allowed in %s; maybe you meant PC?", opcode)
				}
				pclr = true
			} else if t == PC {
				if opcode == "PUSH" || opcode == "STMIA" {
					return 0, false, fmt.Errorf("PC not allowed in %s; maybe you meant LR?", opcode)
				}
				pclr = true
			} else {
				// Bad next token.
				return 0, false, fmt.Errorf("Expected register, PC or LR in register list, but found %s", tokenNames[t])
			}
		} else {
			return 0, false, fmt.Errorf("Expected register in register list, but found %s", tokenNames[t])
		}

		// Now look for a comma or RBRACE.
		if p.consume(RBRACE) {
			return regs, pclr, nil
		}

		// If not RBRACE, expect a comma.
		if !p.consumeComma() {
			t, _ := p.scanIgnoreWhitespace()
			return 0, false, fmt.Errorf("Expected } or comma in register list, but found %s", tokenNames[t])
		}
	}
}

func (p *Parser) parseLiteral() (Expression, error) {
	if !p.consume(HASH) {
		return nil, fmt.Errorf("Failed to parse # for literal")
	}
	return p.parseSimpleExpr()
}

func (p *Parser) parseLoadStore(opcode string) (Assembled, error) {
	// Always a base register, comma, and square brackets.
	// But it's one of a few possibilities:
	// [Rb]
	// [Rb], #lit
	// [Rb], Ra
	// [Rb, #lit]
	// [Rb, Ra]
	// [PC, #U8] - only LDR
	// [SP, #U8]

	dest, err := p.parseReg()
	if err != nil {
		return nil, fmt.Errorf("Expected source/destination register for %s: %v", opcode, err)
	}

	if !p.consumeComma() {
		t, _ := p.scanIgnoreWhitespace()
		return nil, fmt.Errorf("Couldn't find comma in %s, found %s", opcode, tokenNames[t])
	}

	// There are three parts, of which only one is required:
	// - base register is required
	// - following literal or register (pre-increment)
	// - post-square bracket literal or register (post-increment)
	// base register might be PC or SP.
	if !p.consume(LBRAC) {
		t, _ := p.scanIgnoreWhitespace()
		return nil, fmt.Errorf("Expected [ in %s, but found %s", opcode, tokenNames[t])
	}

	t, _ := p.scanIgnoreWhitespace()
	if t == PC || t == SP {
		// Special case. Always a pre-incrementing literal.
		if !p.consumeComma() {
			t, _ = p.scanIgnoreWhitespace()
			return nil, fmt.Errorf("Expected comma in %s, but found %s", opcode, tokenNames[t])
		}

		lit, err := p.parseLiteral()
		if err != nil {
			return nil, fmt.Errorf("Error parsing literal offset in %s: %v", opcode, err)
		}

		if !p.consume(RBRAC) {
			t, _ = p.scanIgnoreWhitespace()
			return nil, fmt.Errorf("Expected ] in %s, but found %s", opcode, tokenNames[t])
		}
		if !p.consume(NEWLINE) {
			t, _ = p.scanIgnoreWhitespace()
			return nil, fmt.Errorf("Unexpected %s at end of line", tokenNames[t])
		}

		return &LoadStore{opcode == "STR", dest, 0xffff, lit, 0xffff, nil, 0xffff, t == SP}, nil
	} else {
		// Regular register.
		p.unscan()
		base, err := p.parseReg()

		out := &LoadStore{opcode == "STR", dest, base, nil, 0xffff, nil, 0xffff, false}

		// Next is a comma or ].
		t, _ := p.scanIgnoreWhitespace()
		if t == COMMA {
			// Try to parse a literal.
			out.preLit, err = p.parseLiteral()
			if err != nil {
				out.preReg, err = p.parseReg()
				if err != nil {
					return nil, fmt.Errorf("Expected pre-indexed value, but failed to parse.")
				}
			}

			if !p.consume(RBRAC) {
				return nil, fmt.Errorf("Expected closing ] after base register")
			}
		} else if t != RBRAC {
			return nil, fmt.Errorf("Expected comma or ] after base register")
		}

		// Next is a comma or EOL.
		t, _ = p.scanIgnoreWhitespace()
		if t == COMMA {
			// post-incrementing is real.
			out.postLit, err = p.parseLiteral()
			if err != nil {
				out.postReg, err = p.parseReg()
				if err != nil {
					return nil, fmt.Errorf("Expected literal or register for post-increment: %v", err)
				}
			}
		}

		t, _ = p.scanIgnoreWhitespace()
		if t == NEWLINE {
			return nil, fmt.Errorf("Unexpected %s at end of line", tokenNames[t])
		}

		return out, nil
	}
}
