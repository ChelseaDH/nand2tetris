package parser

import (
	"github.com/ChelseaDH/JackAnalyser/token"
)

type Type struct {
	Token token.Token
	Class string
}

type VarDec struct {
	Type Type
	Name string
}

type ClassVarDec struct {
	Static bool
	VarDec VarDec
}

type Param struct {
	Type Type
	Name string
}

type JackSubroutine struct {
	SType      token.Token
	ReturnType Type
	SName      string
	ParamList  []Param
	Vars       []VarDec
	Statements []Statement
}

type JackClass struct {
	Name        string
	VarDecs     []ClassVarDec
	Subroutines []JackSubroutine
}

type Statement interface{
	toVm(classScope ClassScope, routineScope map[string]variable, writer instructionWriter)
}

type LetStatement struct {
	Name  string
	Index Expression
	Value Expression
}

type IfStatement struct {
	Condition Expression
	Body      []Statement
	Else      []Statement
}

type WhileStatement struct {
	Condition Expression
	Body      []Statement
}

type DoStatement struct {
	Call SubroutineCall
}

type ReturnStatement struct {
	Value Expression
}

type Expression interface{
	toVm(classScope ClassScope, routineScope map[string]variable, writer instructionWriter)
}

type BinaryTerm struct {
	Left     Expression
	Operator token.Token
	Right    Expression
}

type IntegerConst struct {
	Value int
}

type StringConstant struct {
	Value string
}

type BooleanConstant struct {
	Value bool
}

type NullConstant struct{}

type ThisConstant struct{}

type VarName struct {
	Name string
}

type ArrayAccess struct {
	Name  string
	Index Expression
}

type SubroutineCall struct {
	ClassName string
	SubName   string
	Arguments []Expression
}

type BracketExpression struct {
	Expression Expression
}

type UnaryTerm struct {
	Operator token.Token
	Term     Expression
}
