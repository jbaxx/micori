package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"strings"
)

type Token int

const (
	// Special tokens
	ILLEGAL Token = iota
	EOF
	WS

	// Literals
	IDENT

	// Misc characters
	ASTERISK
	COMMA

	// Keywords
	SELECT
	FROM
)

type SelectStatement struct {
	Fields    []string
	TableName string
}

func isWhitespace(ch rune) bool {
	return ch == ' ' || ch == '\t' || ch == '\n'
}

func isLetter(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

func isDigit(ch rune) bool {
	return ch >= '0' && ch <= '9'
}

var eof = rune(0)

type Scanner struct {
	r *bufio.Reader
}

func NewScanner(r io.Reader) *Scanner {
	return &Scanner{r: bufio.NewReader(r)}
}

func (s *Scanner) read() rune {
	ch, _, err := s.r.ReadRune()
	if err != nil {
		return eof
	}
	return ch
}

func (s *Scanner) unread() { _ = s.r.UnreadRune() }

func (s *Scanner) Scan() (tok Token, literal string) {
	// Read the next rune
	ch := s.read()

	// If we see whitespace then consume all contiguous whitespace
	// If we see a letter then consume as an ident or reserved word
	if isWhitespace(ch) {
		s.unread()
		return s.scanWhitespace()
	} else if isLetter(ch) {
		s.unread()
		return s.scanIdent()
	}

	// Otherwise read the individual character
	switch ch {
	case eof:
		return EOF, ""
	case '*':
		return ASTERISK, string(ch)
	case ',':
		return COMMA, string(ch)
	}

	return ILLEGAL, string(ch)

}

func (s *Scanner) scanWhitespace() (tok Token, literal string) {
	// Create a buffer and read the current character into it
	var buf bytes.Buffer
	buf.WriteRune(s.read())

	// Read every subsequent whitespace character into the buffer
	// Non-whitespace characters and EOF will cause the loop to exit
	for {
		ch := s.read()
		if ch == eof {
			break
		}
		if !isWhitespace(ch) {
			s.unread()
			break
		}
		buf.WriteRune(ch)
	}

	return WS, buf.String()
}

func (s *Scanner) scanIdent() (tok Token, literal string) {
	var buf bytes.Buffer
	buf.WriteRune(s.read())

	// Read every subsequent ident character into the buffer
	// Non-dent character and EOF will cause the loop to exit
	for {
		ch := s.read()
		if ch == eof {
			break
		}
		if !isLetter(ch) && !isDigit(ch) && ch != '_' {
			s.unread()
			break
		}
		_, _ = buf.WriteRune(ch)
	}

	switch strings.ToUpper(buf.String()) {
	case "SELECT":
		return SELECT, buf.String()
	case "FROM":
		return FROM, buf.String()
	}

	// Otherwise return as a regular identifier
	return IDENT, buf.String()
}

type Parser struct {
	s   *Scanner
	buf Buffer
}

type Buffer struct {
	tok Token  // last read token
	lit string // last reat literal
	n   int    // vuffer size (max=1)
}

func NewParser(r io.Reader) *Parser {
	return &Parser{s: NewScanner(r)}
}

// scan returns the next token from the underlying scanner
// If a token has been unscanned then read that instead
func (p *Parser) scan() (Token, string) {
	// If we have a token on the buffer, then return it
	if p.buf.n != 0 {
		p.buf.n = 0
		return p.buf.tok, p.buf.lit
	}

	// Otherwise read the next token from the scanner
	tok, lit := p.s.Scan()

	// Save it to the buffer in vase we unscan later
	p.buf.tok, p.buf.lit = tok, lit

	return tok, lit
}

// unscan pushes the previously read token back onto the buffer
func (p *Parser) unscan() { p.buf.n = 1 }

func (p *Parser) scanIgnoreWhitespace() (Token, string) {
	tok, lit := p.scan()
	if tok == WS {
		tok, lit = p.scan()
	}

	return tok, lit
}

func (p *Parser) Parse() (*SelectStatement, error) {
	stmt := &SelectStatement{}

	if tok, lit := p.scanIgnoreWhitespace(); tok != SELECT {
		return nil, fmt.Errorf("found %q, expected SELECT", lit)
	}

	for {
		// Read a field
		tok, lit := p.scanIgnoreWhitespace()
		if tok != IDENT && tok != ASTERISK {
			return nil, fmt.Errorf("found %q, expected field", lit)
		}
		stmt.Fields = append(stmt.Fields, lit)

		// If the next token is not a comma then break the loop
		if tok, _ := p.scanIgnoreWhitespace(); tok != COMMA {
			p.unscan()
			break
		}
	}

	if tok, lit := p.scanIgnoreWhitespace(); tok != FROM {
		return nil, fmt.Errorf("found %q, expected FROM", lit)
	}

	tok, lit := p.scanIgnoreWhitespace()
	if tok != IDENT {
		return nil, fmt.Errorf("found %q, expected table name", lit)
	}

	stmt.TableName = lit

	return stmt, nil
}

func main() {

	r := strings.NewReader("select name, street, address, from data")

	parser := NewParser(r)

	p, err := parser.Parse()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%+v\n", p)
}
