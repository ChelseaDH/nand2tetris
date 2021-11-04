package lexer

import (
	"bufio"
	"os"
	"strings"
	"testing"

	"github.com/ChelseaDH/JackAnalyser/token"
)

type lexerNextComp struct {
	token token.Token
	value string
}

type lexerStringTest struct {
	input       string
	expectation lexerNextComp
}

var nextStringTests = []lexerStringTest{
	{
		input: "\" A broken string",
		expectation: lexerNextComp{
			token: token.Error,
		},
	},
	{
		input: "/** A broken API block comment",
		expectation: lexerNextComp{
			token: token.Error,
		},
	},
	{
		input: "/** A broken block comment",
		expectation: lexerNextComp{
			token: token.Error,
		},
	},
	{
		input: "23",
		expectation: lexerNextComp{
			token: token.IntConst,
			value: "23",
		},
	},
	{
		input: `"hello"`,
		expectation: lexerNextComp{
			token: token.StringConst,
			value: "hello",
		},
	},
}

func TestNext_String(t *testing.T) {
	for _, test := range nextStringTests {
		reader := bufio.NewReader(strings.NewReader(test.input))
		lexer := NewLexer(reader)

		tok, value, err := lexer.Next()

		if tok != test.expectation.token {
			t.Errorf("output token %q not equal to expected token %q for %s", tok, test.expectation.token, test.input)
		}

		if value != test.expectation.value {
			t.Errorf("output value '%s' not equal to expected value '%s' for %s", value, test.expectation.value, test.input)
		}

		if err == nil && test.expectation.token == token.Error {
			t.Errorf("expected an error but none returned for %s", test.input)
		}

		if err != nil && test.expectation.token != token.Error {
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
				token: token.Error,
				value: "!",
			},
		},
	},
	{
		filePath: "../TestFiles/comments.jack",
		expectations: []lexerNextComp{
			{
				token: token.End,
			},
		},
	},
	{
		filePath: "../TestFiles/endOfFile.jack",
		expectations: []lexerNextComp{
			{
				token: token.Function,
			},
			{
				token: token.End,
			},
		},
	},
	{
		filePath: "../TestFiles/functionTest.jack",
		expectations: []lexerNextComp{
			{
				token: token.Function,
			},
			{
				token: token.Int,
			},
			{
				token: token.Identifier,
				value: "Test",
			},
			{
				token: token.LeftParen,
				value: "(",
			},
			{
				token: token.Int,
			},
			{
				token: token.Identifier,
				value: "x",
			},
			{
				token: token.Comma,
				value: ",",
			},
			{
				token: token.Int,
			},
			{
				token: token.Identifier,
				value: "y",
			},
			{
				token: token.RightParen,
				value: ")",
			},
			{
				token: token.LeftBrace,
				value: "{",
			},
			{
				token: token.Return,
			},
			{
				token: token.Identifier,
				value: "x",
			},
			{
				token: token.Plus,
				value: "+",
			},
			{
				token: token.Identifier,
				value: "y",
			},
			{
				token: token.SemiColon,
				value: ";",
			},
			{
				token: token.RightBrace,
				value: "}",
			},
			{
				token: token.End,
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
			tok, value, err := lexer.Next()

			if tok != expectation.token {
				t.Errorf("output token %q not equal to expected token %q in file %s", tok, expectation.token, test.filePath)
			}

			if value != expectation.value {
				t.Errorf("output value '%s' not equal to expected value '%s' in file %s", value, expectation.value, test.filePath)
			}

			if err == nil && expectation.token == token.Error {
				t.Errorf("expected an error but none returned in file %s", test.filePath)
			}

			if err != nil && expectation.token != token.Error {
				t.Errorf("did not expect an error, but %q returned in file %s", err, test.filePath)
			}
		}
	}
}
