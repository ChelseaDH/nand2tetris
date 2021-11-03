package lexer

import (
	"bufio"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/ChelseaDH/JackAnalyser/token"
)

type lexerNextComp struct {
	expectedToken token.Token
	expectErr     bool
}

type lexerStringTest struct {
	input       string
	expectation lexerNextComp
}

var nextStringTests = []lexerStringTest{
	{
		input: "\" A broken string",
		expectation: lexerNextComp{
			expectedToken: nil,
			expectErr:     true,
		},
	},
	{
		input: "/** A broken API block comment",
		expectation: lexerNextComp{
			expectedToken: nil,
			expectErr:     true,
		},
	},
	{
		input: "/** A broken block comment",
		expectation: lexerNextComp{
			expectedToken: nil,
			expectErr:     true,
		},
	},
}

func TestNext_String(t *testing.T) {
	for _, test := range nextStringTests {
		reader := bufio.NewReader(strings.NewReader(test.input))
		lexer := NewLexer(reader)

		tok, err := lexer.Next()

		if !reflect.DeepEqual(tok, test.expectation.expectedToken) {
			t.Errorf("output token %q not equal to expected token %q for %s", tok, test.expectation.expectedToken, test.input)
		}

		if err == nil && test.expectation.expectErr {
			t.Errorf("expected an error but none returned for %s", test.input)
		}

		if err != nil && !test.expectation.expectErr {
			t.Errorf("did not expect an error, but %q returned for %s", err, test.input)
		}
	}
}

type lexerFileTest struct {
	filePath     string
	expectations []lexerNextComp
}

var nextFileTests = []lexerFileTest{
	{
		filePath: "../TestFiles/invalidSymbol.jack",
		expectations: []lexerNextComp{
			{
				expectedToken: nil,
				expectErr:     true,
			},
		},
	},
	{
		filePath: "../TestFiles/comments.jack",
		expectations: []lexerNextComp{
			{
				expectedToken: &token.EndToken{},
			},
		},
	},
	{
		filePath: "../TestFiles/endOfFile.jack",
		expectations: []lexerNextComp{
			{
				expectedToken: &token.KeywordToken{
					Keyword: token.Function,
					Text:    "function",
				},
			},
			{
				expectedToken: &token.EndToken{},
			},
		},
	},
	{
		filePath: "../TestFiles/functionTest.jack",
		expectations: []lexerNextComp{
			{
				expectedToken: &token.KeywordToken{
					Keyword: token.Function,
					Text:    "function",
				},
			},
			{
				expectedToken: &token.KeywordToken{
					Keyword: token.Int,
					Text:    "int",
				},
			},
			{
				expectedToken: &token.IdentifierToken{
					Identifier: "Test",
				},
			},
			{
				expectedToken: &token.SymbolToken{
					Symbol: token.LeftParen,
					Text:   "(",
				},
			},
			{
				expectedToken: &token.KeywordToken{
					Keyword: token.Int,
					Text:    "int",
				},
			},
			{
				expectedToken: &token.IdentifierToken{
					Identifier: "x",
				},
			},
			{
				expectedToken: &token.SymbolToken{
					Symbol: token.Comma,
					Text:   ",",
				},
			},
			{
				expectedToken: &token.KeywordToken{
					Keyword: token.Int,
					Text:    "int",
				},
			},
			{
				expectErr: false,
				expectedToken: &token.IdentifierToken{
					Identifier: "y",
				},
			},
			{
				expectedToken: &token.SymbolToken{
					Symbol: token.RightParen,
					Text:   ")",
				},
			},
			{
				expectedToken: &token.SymbolToken{
					Symbol: token.LeftBrace,
					Text:   "{",
				},
			},
			{
				expectedToken: &token.KeywordToken{
					Keyword: token.Return,
					Text:    "return",
				},
			},
			{
				expectedToken: &token.IdentifierToken{
					Identifier: "x",
				},
			},
			{
				expectedToken: &token.SymbolToken{
					Symbol: token.Plus,
					Text:   "+",
				},
			},
			{
				expectedToken: &token.IdentifierToken{
					Identifier: "y",
				},
			},
			{
				expectedToken: &token.SymbolToken{
					Symbol: token.SemiColon,
					Text:   ";",
				},
			},
			{
				expectedToken: &token.SymbolToken{
					Symbol: token.RightBrace,
					Text:   "}",
				},
			},
			{
				expectedToken: &token.EndToken{},
			},
		},
	},
}

func TestNext_File(t *testing.T) {
	for _, test := range nextFileTests {
		file, _ := os.Open(test.filePath)
		defer file.Close()

		reader := bufio.NewReader(file)
		lexer := NewLexer(reader)

		for _, expectation := range test.expectations {
			tok, err := lexer.Next()

			if !reflect.DeepEqual(tok, expectation.expectedToken) {
				t.Errorf("output token %q not equal to expected token %q in file %s", tok, expectation.expectedToken, test.filePath)
			}

			if err == nil && expectation.expectErr {
				t.Errorf("expected an error but none returned in file %s", test.filePath)
			}

			if err != nil && !expectation.expectErr {
				t.Errorf("did not expect an error, but %q returned in file %s", err, test.filePath)
			}
		}
	}
}
