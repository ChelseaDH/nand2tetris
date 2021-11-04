package token

import (
	"strconv"
)

type Token int

const (
	End Token = iota
	Error

	// Keywords
	Class
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

	// Symbols
	LeftBrace
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

	Identifier
	IntConst
	StringConst
)

var KeywordMap = map[string]Token{
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

var SymbolMap = map[string]Token{
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

var tokens = [...]string{
	// Keywords
	Class:       "class",
	Constructor: "constructor",
	Function:    "function",
	Method:      "method",
	Field:       "field",
	Static:      "static",
	Var:         "var",
	Int:         "int",
	Char:        "char",
	Boolean:     "boolean",
	Void:        "void",
	True:        "true",
	False:       "false",
	Null:        "null",
	This:        "this",
	Let:         "let",
	Do:          "do",
	If:          "if",
	Else:        "else",
	While:       "while",
	Return:      "return",

	// Symbols
	LeftBrace:    "{",
	RightBrace:   "}",
	LeftParen:    "(",
	RightParen:   ")",
	LeftBracket:  "[",
	RightBracket: "]",
	Dot:          ".",
	Comma:        ",",
	SemiColon:    ";",
	Plus:         "+",
	Minus:        "-",
	Mult:         "*",
	Div:          "/",
	And:          "&",
	Or:           "|",
	LessThan:     "<",
	GreaterThan:  ">",
	Equals:       "=",
	Not:          "~",

	Identifier:  "identifier",
	IntConst:    "int constant",
	StringConst: "string constant",

	End:   "end",
	Error: "error",
}

func (tok Token) String() string {
	s := ""
	if 0 <= tok && tok < Token(len(tokens)) {
		s = tokens[tok]
	}
	if s == "" {
		s = "token(" + strconv.Itoa(int(tok)) + ")"
	}
	return s
}
