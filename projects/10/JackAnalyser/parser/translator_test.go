package parser

import (
	"reflect"
	"testing"

	"github.com/ChelseaDH/JackAnalyser/token"
)

type expressionTest struct {
	input        string
	classScope   map[string]variable
	routineScope map[string]variable
	expOutput    []string
}

var expressionTests = []expressionTest{
	{
		input: "x + g(2, y, -z) * 5",
		classScope: map[string]variable{
			"x": {
				typ: Type{
					Token: token.Int,
					Class: "",
				},
				kind:     Static,
				position: 0,
			},
		},
		routineScope: map[string]variable{
			"y": {
				typ: Type{
					Token: token.Int,
					Class: "",
				},
				kind:     Argument,
				position: 0,
			},
			"z": {
				typ: Type{
					Token: token.Int,
					Class: "",
				},
				kind:     Local,
				position: 0,
			},
		},
		expOutput: []string{"push static 0", "push 2", "push argument 0", "push local 0", "-", "call g", "push 5", "*", "+"},
	},
}

func TestExpression(t *testing.T) {
	for _, test := range expressionTests {
		expression := ParseExpression(test.input)

		w := &TestWriter{}
		expression.toVm(test.classScope, test.routineScope, w)

		if reflect.DeepEqual(test.expOutput, w.output) {
			t.Errorf("expected output %v got %v", test.expOutput, w.output)
		}
	}
}
