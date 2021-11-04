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
	reader *bufio.Reader

	cr    rune // current character
	isEOF bool
	err   error
}

func NewLexer(input io.Reader) *Lexer {
	return &Lexer{
		reader: bufio.NewReader(input),
	}
}

func (l *Lexer) nextRune() {
	r, _, err := l.reader.ReadRune()
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

func (l *Lexer) Next() (token.Token, string, error) {
	l.nextRune()
	l.skipWhitespace()

	if l.isEOF {
		return token.End, "", nil
	} else if l.err != nil {
		return token.Error, "", l.err
	}

	// If a scan returns either nothing or false then an error occurred as a result of calling lexer.nextRune()
	// In this case, fall though to shared error/EOF handling
	switch {
	case isLetter(l.cr):
		tok, value := l.scanIdentifier()
		if tok != token.Error {
			return tok, value, nil
		}
		break

	case unicode.IsDigit(l.cr):
		tok, value, err := l.scanNumber()
		if tok != token.Error || err != nil {
			return tok, value, err
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
				return token.Error, "", err
			} else {
				return l.Next()
			}
		} else {
			l.reader.UnreadRune()
			return token.Div, "/", nil
		}
		break

	default:
		char := string(l.cr)
		symbol, ok := token.SymbolMap[char]
		if ok {
			return symbol, char, nil
		}

		return token.Error, char, fmt.Errorf("cannot lex %s character", char)
	}

	if l.isEOF {
		return token.End, "", nil
	} else if l.err != nil {
		return token.Error, "", l.err
	} else {
		return token.Error, "", errors.New("unexpected error occurred")
	}
}

// Looks for both keywords and identifiers.
// Scans until the next rune that is not a letter, _, or number is found.
func (l *Lexer) scanIdentifier() (token.Token, string) {
	runes := []rune{l.cr}

	for {
		l.nextRune()
		if l.err != nil {
			return token.Error, ""
		}

		if isLetter(l.cr) || unicode.IsDigit(l.cr) {
			runes = append(runes, l.cr)
		} else {
			l.reader.UnreadRune()
			break
		}
	}

	ident := string(runes)
	keyword, ok := token.KeywordMap[ident]
	if ok {
		return keyword, ""
	}

	return token.Identifier, ident
}

// Scans until the next non-number is found.
// Only accepts a number within the bounds of integerMin and integerMax.
func (l *Lexer) scanNumber() (token.Token, string, error) {
	runes := []rune{l.cr}

	for {
		l.nextRune()
		if l.err != nil {
			return token.Error, "", nil
		}

		if unicode.IsDigit(l.cr) {
			runes = append(runes, l.cr)
		} else {
			l.reader.UnreadRune()
			break
		}
	}

	s := string(runes)
	n, err := strconv.Atoi(s)
	if err != nil {
		return token.Error, "", err
	}

	if n < integerMin || n > integerMax {
		return token.Error, "", fmt.Errorf("integer constants must be between %d and %d, %d provided", integerMin, integerMax, n)
	}

	return token.IntConst, s, nil
}

// Scans a string literal, discarding the quote characters at the start and end.
func (l *Lexer) scanString() (token.Token, string, error) {
	var runes []rune

	for {
		l.nextRune()
		if l.isEOF {
			return token.Error, "", errors.New("EOF found before closing quote for string literal")
		} else if l.err != nil {
			return token.Error, "", l.err
		}

		if l.cr == '"' {
			break
		}

		runes = append(runes, l.cr)
	}

	return token.StringConst, string(runes), nil
}

// Scans and discards runes until a newline or EOF is found.
// Returns true if successfully found end of comment.
func (l *Lexer) scanInLineComment() bool {
	_, err := l.reader.ReadBytes('\n')

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
