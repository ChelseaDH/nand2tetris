package parser

import (
	"os"
	"reflect"
	"testing"

	"github.com/ChelseaDH/JackAnalyser/lexer"
	"github.com/ChelseaDH/JackAnalyser/token"
)

type parserTest struct {
	filePath    string
	expectedAst *JackClass
	expectErr   bool
}

var parserTests = []parserTest{
	{
		filePath: "../TestFiles/class.jack",
		expectedAst: &JackClass{
			Name: "test",
			VarDecs: []ClassVarDec{
				{
					Static: true,
					VarDec: VarDec{
						Type: Type{Token: token.Int},
						Name: "x",
					},
				},
				{
					Static: true,
					VarDec: VarDec{
						Type: Type{Token: token.Int},
						Name: "y",
					},
				},
				{
					Static: false,
					VarDec: VarDec{
						Type: Type{Token: token.Boolean},
						Name: "valid",
					},
				},
			},
			Subroutines: []JackSubroutine{
				{
					SType: token.Constructor,
					ReturnType: Type{
						Token: token.Identifier,
						Class: "test",
					},
					SName: "New",
					ParamList: []Param{
						{
							Type: Type{
								Token: token.Int,
							},
							Name: "x",
						},
						{
							Type: Type{
								Token: token.Int,
							},
							Name: "y",
						},
					},
				},
				{
					SType: token.Function,
					ReturnType: Type{
						Token: token.Void,
					},
					SName: "testing",
					Vars: []VarDec{
						{
							Type: Type{
								Token: token.Int,
							},
							Name: "a",
						},
					},
				},
			},
		},
		expectErr: false,
	},
}

func TestParser_Parse(t *testing.T) {
	for _, test := range parserTests {
		file, _ := os.Open(test.filePath)
		defer file.Close()

		parser := NewParser(lexer.NewLexer(file))
		ast, err := parser.Parse()

		if !reflect.DeepEqual(test.expectedAst, ast) {
			t.Errorf("output class %v not equal to expected class %v for file %s", ast, test.expectedAst, test.filePath)
		}

		if err == nil && test.expectErr {
			t.Errorf("expected an error but none returned for file %s", test.filePath)
		}

		if err != nil && !test.expectErr {
			t.Errorf("did not expect an error, but %q returned for file %s", err, test.filePath)
		}
	}
}
