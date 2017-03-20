package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"
)

type Token int

const (
	// Special tokens
	ILLEGAL Token = iota
	EOF
	WS
	NEWLINE

	REGISTER // rN
	PC
	SP
	LR
	NUMBER // Immediates, .dat etc.
	IDENT  // Labels
	STRING // String literals

	// Punctuation
	DOT
	HASH
	COLON
	COMMA
	LBRAC
	RBRAC
	LPAREN
	RPAREN
	LBRACE
	RBRACE

	// Operators
	PLUS
	MINUS
	TIMES
	DIVIDE
	LANGLES
	RANGLES
	AND
	OR
	XOR
	NOT
)

var tokenNames = map[Token]string{
	ILLEGAL:  "<ILLEGAL>",
	EOF:      "EOF",
	WS:       "whitespace",
	NEWLINE:  "newline",
	REGISTER: "register",
	PC:       "PC",
	SP:       "SP",
	LR:       "LR",
	NUMBER:   "number",
	IDENT:    "identifier",
	STRING:   "string literal",
	DOT:      "dot",
	HASH:     "#",
	COLON:    ":",
	COMMA:    ",",
	LBRAC:    "[",
	RBRAC:    "]",
	LBRACE:   "{",
	RBRACE:   "}",
	LPAREN:   "(",
	RPAREN:   ")",
	PLUS:     "+",
	MINUS:    "-",
	TIMES:    "*",
	DIVIDE:   "/",
	LANGLES:  "<<",
	RANGLES:  ">>",
	AND:      "&",
	OR:       "|",
	XOR:      "^",
	NOT:      "~",
}

// We'll put this EOF rune on the end of everything.
var eof = rune(0)

// Some character classes.
func isWhitespace(ch rune) bool {
	// TODO: Comments
	return ch == ' ' || ch == '\t'
}

func isLetter(ch rune) bool {
	return ('A' <= ch && ch <= 'Z') || ('a' <= ch && ch <= 'z')
}
func isDigit(ch rune) bool {
	return '0' <= ch && ch <= '9'
}

// Scanner is our lexer.
type Scanner struct {
	r       *bufio.Reader
	file    string
	line    uint
	col     uint
	noCount uint
}

func NewScanner(filename string, r io.Reader) *Scanner {
	return &Scanner{r: bufio.NewReader(r), file: filename, line: 1, col: 0}
}

// read reads the next rune from the buffered reader.
// Returns the rune(0) if an error occurs or io.EOF is returned.
func (s *Scanner) read() rune {
	ch, _, err := s.r.ReadRune()
	if err != nil {
		return eof
	}

	if s.noCount == 0 {
		s.col++
		if ch == '\n' {
			s.col = 0
			s.line++
		}
	} else {
		s.noCount--
	}

	return ch
}

func (s *Scanner) unread() {
	_ = s.r.UnreadRune()
	s.noCount++ // Avoids double-counting when we re-scan.
}

func (s *Scanner) Location() string {
	return fmt.Sprintf("%s:%d:%d", s.file, s.line, s.col)
}

func (s *Scanner) Scan() (Token, string) {
	t, l := s.innerScan()
	fmt.Printf("%s - '%s'\n", tokenNames[t], l)
	return t, l
}

func (s *Scanner) innerScan() (tok Token, lit string) {
	ch := s.read()

	// If we see whitespace, then consume it all.
	if isWhitespace(ch) {
		s.unread()
		return s.scanWhile(isWhitespace, WS)
	} else if isLetter(ch) || ch == '_' {
		s.unread()
		return s.scanIdent()
	} else if isDigit(ch) {
		s.unread()
		// TODO: Other bases.
		// TODO: Negatives.
		return s.scanNumber()
	}

	// Otherwise, scan individual characters.
	switch ch {
	case eof:
		return EOF, ""
	case '.':
		return DOT, string(ch)
	case '#':
		return HASH, string(ch)
	case ':':
		return COLON, string(ch)
	case ',':
		return COMMA, string(ch)
	case '[':
		return LBRAC, string(ch)
	case ']':
		return RBRAC, string(ch)
	case '{':
		return LBRACE, string(ch)
	case '}':
		return RBRACE, string(ch)
	case '(':
		return LPAREN, string(ch)
	case ')':
		return RPAREN, string(ch)
	case '\n':
		return NEWLINE, string(ch)
	case '+':
		return PLUS, string(ch)
	case '-':
		return MINUS, string(ch)
	case '*':
		return TIMES, string(ch)
	case '/':
		return DIVIDE, string(ch)
	case '&':
		return AND, string(ch)
	case '|':
		return OR, string(ch)
	case '^':
		return XOR, string(ch)
	case '~':
		return NOT, string(ch)
	case '<':
		next := s.read()
		if next == '<' {
			return LANGLES, "<<"
		} else {
			return ILLEGAL, string(ch) + string(next)
		}
	case '>':
		next := s.read()
		if next == '>' {
			return RANGLES, ">>"
		} else {
			return ILLEGAL, string(ch) + string(next)
		}
	case ';':
		return s.scanWhile(func(c rune) bool { return c != '\n' }, WS)
	case '"':
		return s.scanStringLiteral()
	}

	fmt.Printf("%v\n", ch)
	return ILLEGAL, string(ch)
}

func (s *Scanner) scanWhile(p func(rune) bool, t Token) (Token, string) {
	// Create a buffer and read the current character into it.
	var buf bytes.Buffer
	// Read the first character, which we know to be whitespace.
	buf.WriteRune(s.read())
	for {
		if ch := s.read(); ch == eof {
			break
		} else if !p(ch) {
			s.unread()
			break
		} else {
			buf.WriteRune(ch)
		}
	}
	return t, buf.String()
}

var keywords = map[string]Token{
	"PC": PC,
	"SP": SP,
	"LR": LR,
	"R0": REGISTER,
	"R1": REGISTER,
	"R2": REGISTER,
	"R3": REGISTER,
	"R4": REGISTER,
	"R5": REGISTER,
	"R6": REGISTER,
	"R7": REGISTER,
}

func (s *Scanner) scanIdent() (tok Token, lit string) {
	var buf bytes.Buffer
	buf.WriteRune(s.read())

	for {
		if ch := s.read(); ch == eof {
			break
		} else if !isLetter(ch) && !isDigit(ch) && ch != '_' {
			s.unread()
			break
		} else {
			_, _ = buf.WriteRune(ch)
		}
	}

	st := buf.String()
	if t, ok := keywords[strings.ToUpper(st)]; ok {
		return t, st
	} else {
		return IDENT, st
	}
}

func (s *Scanner) scanNumber() (tok Token, lit string) {
	var buf bytes.Buffer
	ch := s.read()
	firstZero := ch == '0'
	count := 1
	buf.WriteRune(ch)

	for {
		ch = s.read()
		if isHexDigit(ch) || (firstZero && count == 1 && ch == 'x') {
			buf.WriteRune(ch)
			count++
		} else {
			s.unread()
			return NUMBER, buf.String()
		}
	}
}

func isHexDigit(ch rune) bool {
	return ('0' <= ch && ch <= '9') || ('a' <= ch && ch <= 'f') ||
		('A' <= ch && ch <= 'F')
}

func (s *Scanner) scanStringLiteral() (tok Token, lit string) {
	var buf bytes.Buffer
	// TODO: Escaping.
	for {
		if ch := s.read(); ch == '"' {
			return STRING, buf.String()
		} else if ch == eof {
			return ILLEGAL, buf.String()
		} else {
			_, _ = buf.WriteRune(ch)
		}
	}
}
