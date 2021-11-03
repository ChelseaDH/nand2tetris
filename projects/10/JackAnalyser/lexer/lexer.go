package lexer

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"unicode"

	"github.com/ChelseaDH/JackAnalyser/token"
)

const (
	integerMin = 0
	integerMax = 32767
)

type Lexer struct {
	Reader *bufio.Reader

	cr    rune // current character
	isEOF bool
	err   error
}

func NewLexer(input io.Reader) Lexer {
	lexer := Lexer{
		Reader: bufio.NewReader(input),
	}

	return lexer
}

func (l *Lexer) nextRune() {
	r, _, err := l.Reader.ReadRune()
	if err != nil {
		l.parseError(err)
	}

	l.cr = r
}

func (l *Lexer) parseError(err error) {
	if err == io.EOF {
		l.isEOF = true
	} else {
		l.err = err
	}
}

func (l *Lexer) Next() (token.Token, error) {
	l.nextRune()
	l.skipWhitespace()

	if l.isEOF {
		return &token.EndToken{}, nil
	} else if l.err != nil {
		return nil, l.err
	}

	// If a scan returns either nothing or false then an error occurred as a result of calling lexer.nextRune()
	// In this case, fall though to shared error/EOF handling
	switch {
	case isLetter(l.cr):
		tok := l.scanIdentifier()
		if tok != nil {
			return tok, nil
		}
		break

	case unicode.IsDigit(l.cr):
		tok, err := l.scanNumber()
		if tok != nil || err != nil {
			return tok, err
		}
		break

	case l.cr == '"':
		return l.scanString()

	case l.cr == '/':
		l.nextRune()
		if l.err != nil {
			break
		}

		if l.cr == '/' {
			scanned := l.scanInLineComment()
			if scanned {
				return l.Next()
			}
		} else if l.cr == '*' {
			err := l.scanBlockComment()
			if err != nil {
				return nil, err
			} else {
				return l.Next()
			}
		} else {
			l.Reader.UnreadRune()
			return &token.SymbolToken{Symbol: token.Div, Text: "/"}, nil
		}
		break

	default:
		char := string(l.cr)
		symbol, ok := token.SymbolMap[char]
		if ok {
			return &token.SymbolToken{Symbol: symbol,
				Text: char,
			}, nil
		}

		return nil, fmt.Errorf("cannot lex %s character", char)
	}

	if l.isEOF {
		return &token.EndToken{}, nil
	} else if l.err != nil {
		return nil, l.err
	} else {
		return nil, errors.New("unexpected error occurred")
	}
}

// Looks for both keywords and identifiers.
// Scans until the next rune that is not a letter, _, or number is found.
func (l *Lexer) scanIdentifier() token.Token {
	runes := []rune{l.cr}

	for {
		l.nextRune()
		if l.err != nil {
			return nil
		}

		if isLetter(l.cr) || unicode.IsDigit(l.cr) {
			runes = append(runes, l.cr)
		} else {
			l.Reader.UnreadRune()
			break
		}
	}

	ident := string(runes)
	keyword, ok := token.KeywordMap[ident]
	if ok {
		return &token.KeywordToken{
			Keyword: keyword,
			Text:    ident,
		}
	}

	return &token.IdentifierToken{Identifier: ident}
}

// Scans until the next non-number is found.
// Only accepts a number within the bounds of integerMin and integerMax.
func (l *Lexer) scanNumber() (token.Token, error) {
	runes := []rune{l.cr}

	for {
		l.nextRune()
		if l.err != nil {
			return nil, nil
		}

		if unicode.IsDigit(l.cr) {
			runes = append(runes, l.cr)
		} else {
			l.Reader.UnreadRune()
			break
		}
	}

	n, err := strconv.Atoi(string(runes))
	if err != nil {
		return nil, err
	}

	if n < integerMin || n > integerMax {
		return nil, fmt.Errorf("integer constants must be between %d and %d, %d provided", integerMin, integerMax, n)
	}

	return &token.IntConstToken{IntVal: n}, nil
}

// Scans a string literal, discarding the quote characters at the start and end.
func (l *Lexer) scanString() (token.Token, error) {
	var runes []rune

	for {
		l.nextRune()
		if l.isEOF {
			return nil, errors.New("EOF found before closing quote for string literal")
		} else if l.err != nil {
			return nil, l.err
		}

		if l.cr == '"' {
			break
		}

		runes = append(runes, l.cr)
	}

	str := string(runes)
	return &token.StringConstToken{StringVal: str}, nil
}

// Scans and discards runes until a newline or EOF is found.
// Returns true if successfully found end of comment.
func (l *Lexer) scanInLineComment() bool {
	_, err := l.Reader.ReadBytes('\n')

	if err != nil {
		l.parseError(err)
		return false
	}

	return true
}

// Scans and discards runes until a * followed immediately by a / is found.
// Returns true if successfully found end of comment.
func (l *Lexer) scanBlockComment() error {
	for {
		l.nextRune()
		if l.err != nil {
			return l.err
		} else if l.isEOF {
			return errors.New("EOF found before end of block comment")
		} else if l.cr != '*' {
			continue
		}

		l.nextRune()
		if l.err != nil {
			return l.err
		} else if l.isEOF {
			return errors.New("EOF found before end of block comment")
		} else if l.cr == '/' {
			return nil
		}
	}
}

func (l *Lexer) skipWhitespace() {
	for unicode.IsSpace(l.cr) {
		l.nextRune()
	}
}

// Keywords can only begin with a letter.
// Identifiers can begin with either a letter or an _.
func isLetter(r rune) bool {
	return unicode.IsLetter(r) || r == '_'
}
