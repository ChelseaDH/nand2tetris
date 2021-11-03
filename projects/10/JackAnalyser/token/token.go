package token

import (
	"fmt"
)

type Keyword int

const (
	Class Keyword = iota
	Constructor
	Function
	Method
	Field
	Static
	Var
	Int
	Char
	Boolean
	Void
	True
	False
	Null
	This
	Let
	Do
	If
	Else
	While
	Return
)

var KeywordMap = map[string]Keyword{
	"class":       Class,
	"constructor": Constructor,
	"function":    Function,
	"method":      Method,
	"field":       Field,
	"static":      Static,
	"var":         Var,
	"int":         Int,
	"char":        Char,
	"boolean":     Boolean,
	"void":        Void,
	"true":        True,
	"false":       False,
	"null":        Null,
	"this":        This,
	"let":         Let,
	"do":          Do,
	"if":          If,
	"else":        Else,
	"while":       While,
	"return":      Return,
}

type Symbol int

const (
	LeftBrace Symbol = iota
	RightBrace
	LeftParen
	RightParen
	LeftBracket
	RightBracket
	Dot
	Comma
	SemiColon
	Plus
	Minus
	Mult
	Div
	And
	Or
	LessThan
	GreaterThan
	Equals
	Not
)

var SymbolMap = map[string]Symbol{
	"{": LeftBrace,
	"}": RightBrace,
	"(": LeftParen,
	")": RightParen,
	"[": LeftBracket,
	"]": RightBracket,
	".": Dot,
	",": Comma,
	";": SemiColon,
	"+": Plus,
	"-": Minus,
	"*": Mult,
	"/": Div,
	"&": And,
	"|": Or,
	"<": LessThan,
	">": GreaterThan,
	"=": Equals,
	"~": Not,
}

type Token interface {
	Xml() string
}

type KeywordToken struct {
	Keyword Keyword
	Text    string
}

func (kt *KeywordToken) Xml() string {
	return fmt.Sprintf("<keyword> %s </keyword>", kt.Text)
}

type SymbolToken struct {
	Symbol Symbol
	Text   string
}

func (st *SymbolToken) Xml() string {
	return fmt.Sprintf("<symbol> %s </symbol>", st.Text)
}

type IdentifierToken struct {
	Identifier string
}

func (it *IdentifierToken) Xml() string {
	return fmt.Sprintf("<identifier> %s </identifier>", it.Identifier)
}

type IntConstToken struct {
	IntVal int
}

func (it *IntConstToken) Xml() string {
	return fmt.Sprintf("<integerConstant> %d </integerConstant>", it.IntVal)
}

type StringConstToken struct {
	StringVal string
}

func (st *StringConstToken) Xml() string {
	return fmt.Sprintf("<stringConstant> %s </stringConstant>", st.StringVal)
}

type EndToken struct{}

func (et *EndToken) Xml() string { return "" }
