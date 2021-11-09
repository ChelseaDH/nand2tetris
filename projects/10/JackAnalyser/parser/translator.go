package parser

import (
	"fmt"
	"io"
	"strings"
	"unicode"

	"github.com/ChelseaDH/JackAnalyser/token"
)

type instructionWriter interface {
	Write(string)
	getCondCount() int
	incrementCondCount()
}

type TestWriter struct {
	output    []string
	condCount int
}

func (w *TestWriter) Write(s string) {
	w.output = append(w.output, s)
}
func (w *TestWriter) getCondCount() int {
	return w.condCount
}
func (w *TestWriter) incrementCondCount() {
	w.condCount++
}

type FileWriter struct {
	File      io.Writer
	condCount int
}

func (w *FileWriter) Write(s string) {
	fmt.Fprintf(w.File, fmt.Sprintf("%s\n", s))
}
func (w *FileWriter) getCondCount() int {
	return w.condCount
}
func (w *FileWriter) incrementCondCount() {
	w.condCount++
}

type variableKind int

const (
	Field variableKind = iota
	Static
	Argument
	Local
)

var variableKindMap = map[variableKind]string{
	Field:    "field",
	Static:   "static",
	Argument: "argument",
	Local:    "local",
}

type variable struct {
	typ      Type
	kind     variableKind
	position int
}

type ClassScope struct {
	Name            string
	SymbolTable     map[string]variable
	SubroutineTable map[string]token.Token
	FieldCount      int
}

func (c *JackClass) toVm(writer instructionWriter) {
	symbolTable := make(map[string]variable)
	fieldCount := 0
	staticCount := 0

	for _, vd := range c.VarDecs {
		var kind variableKind
		var position int
		if vd.Static {
			kind = Static
			position = staticCount
			staticCount++
		} else {
			kind = Field
			position = fieldCount
			fieldCount++
		}

		symbolTable[vd.VarDec.Name] = variable{
			typ:      vd.VarDec.Type,
			kind:     kind,
			position: position,
		}
	}

	subroutineTable := make(map[string]token.Token)
	for _, s := range c.Subroutines {
		subroutineTable[fmt.Sprintf("%s.%s", c.Name, s.SName)] = s.SType
	}

	cs := ClassScope{
		Name:            c.Name,
		SymbolTable:     symbolTable,
		SubroutineTable: subroutineTable,
		FieldCount:      fieldCount,
	}

	for _, s := range c.Subroutines {
		s.toVm(cs, writer)
	}
}

func (s *JackSubroutine) toVm(scope ClassScope, writer instructionWriter) {
	argCount := 0
	localCount := 0
	symbolTable := make(map[string]variable)

	// Store reference to the object the method is being called on
	if s.SType == token.Method {
		symbolTable[token.This.String()] = variable{
			typ: Type{
				Token: token.Identifier,
				Class: s.SName,
			},
			kind:     Argument,
			position: argCount,
		}
		argCount++
	}

	// Add function arguments to symbol table
	if len(s.ParamList) > 0 {
		for _, p := range s.ParamList {
			symbolTable[p.Name] = variable{
				typ:      p.Type,
				kind:     Argument,
				position: argCount,
			}
			argCount++
		}
	}

	// Add local vars to symbol table
	for _, v := range s.Vars {
		symbolTable[v.Name] = variable{
			typ:      v.Type,
			kind:     Local,
			position: localCount,
		}
		localCount++
	}

	writer.Write(fmt.Sprintf("function %s.%s %d", scope.Name, s.SName, localCount))

	// For constructor:
	// Allocate memory required to store the object
	// Anchor return address to 'THIS'
	if s.SType == token.Constructor {
		writer.Write(fmt.Sprintf("push constant %d", scope.FieldCount))
		writer.Write("call Memory.alloc 1")
		writer.Write("pop pointer 0")
	}

	// For method call:
	// Address of object the function is being called on is store in argument 0
	if s.SType == token.Method {
		writer.Write(fmt.Sprintf("push %s 0", variableKindMap[Argument]))
		writer.Write("pop pointer 0")
	}

	for _, st := range s.Statements {
		st.toVm(scope, symbolTable, writer)
	}
}

func (s LetStatement) toVm(classScope ClassScope, routineScope map[string]variable, writer instructionWriter) {
	variableInScope := findVariableInScope(s.Name, classScope, routineScope)

	// Handle assigning to an array
	// Avoid conflicting use of pointer if Value is an array access expression
	if s.Index != nil {
		writer.Write(fmt.Sprintf("push %s %d", variableKindMap[variableInScope.kind], variableInScope.position))
		s.Index.toVm(classScope, routineScope, writer)
		writer.Write("add") // top stack value = base addr of array + Index

		s.Value.toVm(classScope, routineScope, writer)
		writer.Write("pop temp 0") // top stack value = Value

		writer.Write("pop pointer 1") // THAT points to base addr of array + Index
		writer.Write("push temp 0")
		writer.Write("pop that 0")
		return
	}

	s.Value.toVm(classScope, routineScope, writer)
	writeVariable("pop", variableInScope, writer)
}

func (s IfStatement) toVm(classScope ClassScope, routineScope map[string]variable, writer instructionWriter) {
	label1 := generateCondLabel(classScope.Name, writer)
	label2 := generateCondLabel(classScope.Name, writer)

	s.Condition.toVm(classScope, routineScope, writer)
	writer.Write(OperatorMap[token.Not])
	writer.Write(fmt.Sprintf("if-goto %s", label1))

	for _, b := range s.Body {
		b.toVm(classScope, routineScope, writer)
	}

	writer.Write(fmt.Sprintf("goto %s", label2))
	writer.Write(fmt.Sprintf("label %s", label1))

	for _, e := range s.Else {
		e.toVm(classScope, routineScope, writer)
	}

	writer.Write(fmt.Sprintf("label %s", label2))
}

func (s WhileStatement) toVm(classScope ClassScope, routineScope map[string]variable, writer instructionWriter) {
	label1 := generateCondLabel(classScope.Name, writer)
	label2 := generateCondLabel(classScope.Name, writer)

	writer.Write(fmt.Sprintf("label %s", label1))
	s.Condition.toVm(classScope, routineScope, writer)
	writer.Write(OperatorMap[token.Not])
	writer.Write(fmt.Sprintf("if-goto %s", label2))

	for _, b := range s.Body {
		b.toVm(classScope, routineScope, writer)
	}

	writer.Write(fmt.Sprintf("goto %s", label1))
	writer.Write(fmt.Sprintf("label %s", label2))
}

func (s DoStatement) toVm(classScope ClassScope, routineScope map[string]variable, writer instructionWriter) {
	s.Call.toVm(classScope, routineScope, writer)
	// dump return value of call
	writer.Write("pop temp 0")
}

func (s ReturnStatement) toVm(classScope ClassScope, routineScope map[string]variable, writer instructionWriter) {
	if s.Value == nil {
		// Functions must return a value to the stack, use dummy value when there is no return
		IntegerConst{Value: 0}.toVm(classScope, routineScope, writer)
	} else {
		s.Value.toVm(classScope, routineScope, writer)
	}

	writer.Write("return")
}

func (t BinaryTerm) toVm(classScope ClassScope, routineScope map[string]variable, writer instructionWriter) {
	t.Left.toVm(classScope, routineScope, writer)
	t.Right.toVm(classScope, routineScope, writer)

	if t.Operator == token.Mult || t.Operator == token.Div {
		writer.Write(fmt.Sprintf("call %s 2", OperatorMap[t.Operator]))
	} else {
		writer.Write(OperatorMap[t.Operator])
	}
}

func (i IntegerConst) toVm(classScope ClassScope, routineScope map[string]variable, writer instructionWriter) {
	writer.Write(fmt.Sprintf("push constant %d", i.Value))
}

func (s StringConstant) toVm(classScope ClassScope, routineScope map[string]variable, writer instructionWriter) {
	IntegerConst{Value: len(s.Value)}.toVm(classScope, routineScope, writer)
	writer.Write("call String.new 1")

	for _, c := range s.Value {
		IntegerConst{Value: int(c)}.toVm(classScope, routineScope, writer)
		writer.Write("call String.appendChar 2")
	}
}

func (b BooleanConstant) toVm(classScope ClassScope, routineScope map[string]variable, writer instructionWriter) {
	if b.Value {
		UnaryTerm{
			Operator: token.Minus,
			Term:     IntegerConst{Value: 1},
		}.toVm(classScope, routineScope, writer)
	} else {
		IntegerConst{Value: 0}.toVm(classScope, routineScope, writer)
	}
}

func (n NullConstant) toVm(classScope ClassScope, routineScope map[string]variable, writer instructionWriter) {
	IntegerConst{Value: 0}.toVm(classScope, routineScope, writer)
}

func (t ThisConstant) toVm(classScope ClassScope, routineScope map[string]variable, writer instructionWriter) {
	writer.Write(fmt.Sprintf("push pointer 0"))
}

func (v VarName) toVm(classScope ClassScope, routineScope map[string]variable, writer instructionWriter) {
	findAndWriteVariableInScope(v.Name, "push", classScope, routineScope, writer)
}

func (a ArrayAccess) toVm(classScope ClassScope, routineScope map[string]variable, writer instructionWriter) {
	array := findVariableInScope(a.Name, classScope, routineScope)

	a.Index.toVm(classScope, routineScope, writer)
	writer.Write(fmt.Sprintf("push %s %d", variableKindMap[array.kind], array.position))

	writer.Write("add")
	writer.Write("pop pointer 1")
	writer.Write("push that 0")
}

func (s SubroutineCall) toVm(classScope ClassScope, routineScope map[string]variable, writer instructionWriter) {
	var name string
	args := len(s.Arguments)

	if s.ClassName != "" {
		if !unicode.IsUpper(rune(s.ClassName[0])) {
			object := findVariableInScope(s.ClassName, classScope, routineScope)
			writeVariable("push", object, writer)
			args++
			name = fmt.Sprintf("%s.%s", object.typ.Class, s.SubName)
		} else {
			name = fmt.Sprintf("%s.%s", s.ClassName, s.SubName)
		}
	} else {
		// Function passed without a className must exist on the current class
		name = fmt.Sprintf("%s.%s", classScope.Name, s.SubName)

		sr, found := classScope.SubroutineTable[name]
		if !found {
			panic(fmt.Sprintf("subroutine with name %s not found in class %s", s.SubName, classScope.Name))
		}

		if sr == token.Method {
			writer.Write("push pointer 0")
			args++
		}
	}

	for _, a := range s.Arguments {
		a.toVm(classScope, routineScope, writer)
	}
	writer.Write(fmt.Sprintf("call %s %d", name, args))
}

func (b BracketExpression) toVm(classScope ClassScope, routineScope map[string]variable, writer instructionWriter) {
	b.Expression.toVm(classScope, routineScope, writer)
}

func (t UnaryTerm) toVm(classScope ClassScope, routineScope map[string]variable, writer instructionWriter) {
	t.Term.toVm(classScope, routineScope, writer)

	if t.Operator == token.Minus {
		writer.Write("neg")
	} else {
		writer.Write(OperatorMap[t.Operator])
	}
}

func generateCondLabel(className string, writer instructionWriter) string {
	label := fmt.Sprintf("COND_%s_%d", strings.ToUpper(className), writer.getCondCount())
	writer.incrementCondCount()
	return label
}

func findVariableInScope(name string, classScope ClassScope, routineScope map[string]variable) variable {
	var variableInScope variable
	variableInScope, found := routineScope[name]
	if !found {
		variableInScope, found = classScope.SymbolTable[name]
		if !found {
			panic(fmt.Errorf("undeclared variable %s", name))
		}
	}

	return variableInScope
}

func writeVariable(op string, variable variable, writer instructionWriter) {
	if variable.kind == Field {
		writer.Write(fmt.Sprintf("%s %s %d", op, token.This.String(), variable.position))
	} else {
		writer.Write(fmt.Sprintf("%s %s %d", op, variableKindMap[variable.kind], variable.position))
	}
}

func findAndWriteVariableInScope(name string, op string, classScope ClassScope, routineScope map[string]variable, writer instructionWriter) {
	variableInScope := findVariableInScope(name, classScope, routineScope)
	writeVariable(op, variableInScope, writer)
}

func WriteClassToFile(class *JackClass, file io.Writer) {
	class.toVm(&FileWriter{File: file})
}

// Operators
var OperatorMap = map[token.Token]string{
	token.Plus:        "add",
	token.Minus:       "sub",
	token.Equals:      "eq",
	token.GreaterThan: "gt",
	token.LessThan:    "lt",
	token.And:         "and",
	token.Or:          "or",
	token.Not:         "not",

	token.Mult: "Math.multiply",
	token.Div:  "Math.divide",
}
