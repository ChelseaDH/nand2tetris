package parser

import (
	"encoding/xml"
	"fmt"
	"io"
	"strings"

	"github.com/ChelseaDH/JackAnalyser/token"
)

type XmlWriter struct {
	Writer     io.Writer
	identCount int
}

func (x *XmlWriter) writeIdentifier(s string) {
	x.write("<identifier> %s </identifier>\n", s)
}

func (x *XmlWriter) writeKeyword(s string) {
	x.write("<keyword> %s </keyword>\n", s)
}

func (x *XmlWriter) writeSymbol(s string) {
	x.write("<symbol> ")
	xml.EscapeText(x.Writer, []byte(s))
	fmt.Fprint(x.Writer, " </symbol>\n")
}

func (x *XmlWriter) writeIntConst(i int) {
	x.write("<integerConstant> %d </integerConstant>\n", i)
}

func (x *XmlWriter) writeStringConst(s string) {
	x.write("<stringConstant> %s </stringConstant>\n", s)
}

func (x *XmlWriter) writeOpen(s string) {
	x.write("<%s>\n", s)
	x.identCount++
}
func (x *XmlWriter) writeClose(s string) {
	x.identCount--
	x.write("</%s>\n", s)
}

func (x *XmlWriter) write(fs string, args ...interface{}) {
	fmt.Fprintf(x.Writer, strings.Repeat("  ", x.identCount))
	fmt.Fprintf(x.Writer, fs, args...)
}

func (c *JackClass) WriteToXml(x XmlWriter) {
	x.writeOpen("class")
	x.writeKeyword(token.Class.String())
	x.writeIdentifier(c.Name)
	x.writeSymbol(token.LeftBrace.String())

	for _, vd := range c.VarDecs {
		vd.toXml(x)
	}

	for _, s := range c.Subroutines {
		s.toXml(x)
	}

	x.writeSymbol(token.RightBrace.String())
	x.writeClose("class")
}

func (v *ClassVarDec) toXml(x XmlWriter) {
	x.writeOpen("classVarDec")
	var typ string
	if v.Static {
		typ = token.Static.String()
	} else {
		typ = token.Field.String()
	}

	x.writeKeyword(typ)
	v.VarDec.toXml(x)
	x.writeClose("classVarDec")
}

func (v *VarDec) toXml(x XmlWriter) {
	v.Type.toXml(x)
	x.writeIdentifier(v.Name[0])
	for _, n := range v.Name[1:] {
		x.writeSymbol(token.Comma.String())
		x.writeIdentifier(n)
	}

	x.writeSymbol(token.SemiColon.String())
}

func (s *JackSubroutine) toXml(x XmlWriter) {
	x.writeOpen("subroutineDec")
	x.writeKeyword(s.SType.String())
	s.ReturnType.toXml(x)
	x.writeIdentifier(s.SName)

	x.writeSymbol(token.LeftParen.String())
	x.writeOpen("parameterList")
	if len(s.ParamList) > 0 {
		s.ParamList[0].Type.toXml(x)
		x.writeIdentifier(s.ParamList[0].Name)

		for _, p := range s.ParamList[1:] {
			x.writeSymbol(token.Comma.String())
			p.Type.toXml(x)
			x.writeIdentifier(p.Name)
		}
	}
	x.writeClose("parameterList")
	x.writeSymbol(token.RightParen.String())

	x.writeOpen("subroutineBody")
	x.writeSymbol(token.LeftBrace.String())
	for _, v := range s.Vars {
		x.writeOpen("varDec")
		x.writeKeyword(token.Var.String())
		v.Type.toXml(x)
		x.writeIdentifier(v.Name[0])

		for _, n := range v.Name[1:] {
			x.writeSymbol(token.Comma.String())
			x.writeIdentifier(n)
		}
		x.writeSymbol(token.SemiColon.String())
		x.writeClose("varDec")
	}

	x.writeOpen("statements")
	for _, st := range s.Statements {
		st.toXml(x)
	}
	x.writeClose("statements")
	x.writeSymbol(token.RightBrace.String())
	x.writeClose("subroutineBody")
	x.writeClose("subroutineDec")
}

func (s *LetStatement) toXml(x XmlWriter) {
	x.writeOpen("letStatement")
	x.writeKeyword(token.Let.String())
	x.writeIdentifier(s.Name)

	if s.Index != nil {
		x.writeSymbol(token.LeftBracket.String())
		writeExpression(s.Index, x)
		x.writeSymbol(token.RightBracket.String())
	}

	x.writeSymbol(token.Equals.String())
	writeExpression(s.Value, x)
	x.writeSymbol(token.SemiColon.String())
	x.writeClose("letStatement")
}

func (s *IfStatement) toXml(x XmlWriter) {
	x.writeOpen("ifStatement")
	x.writeKeyword(token.If.String())
	x.writeSymbol(token.LeftParen.String())
	writeExpression(s.Condition, x)
	x.writeSymbol(token.RightParen.String())
	x.writeSymbol(token.LeftBrace.String())

	x.writeOpen("statements")
	for _, b := range s.Body {
		b.toXml(x)
	}
	x.writeClose("statements")
	x.writeSymbol(token.RightBrace.String())

	if s.Else != nil {
		x.writeKeyword(token.Else.String())
		x.writeSymbol(token.LeftBrace.String())
		x.writeOpen("statements")
		for _, e := range s.Else {
			e.toXml(x)
		}
		x.writeClose("statements")
		x.writeSymbol(token.RightBrace.String())
	}
	x.writeClose("ifStatement")
}

func (s *WhileStatement) toXml(x XmlWriter) {
	x.writeOpen("whileStatement")
	x.writeKeyword(token.While.String())
	x.writeSymbol(token.LeftParen.String())
	writeExpression(s.Condition, x)
	x.writeSymbol(token.RightParen.String())
	x.writeSymbol(token.LeftBrace.String())

	x.writeOpen("statements")
	for _, b := range s.Body {
		b.toXml(x)
	}
	x.writeClose("statements")

	x.writeSymbol(token.RightBrace.String())
	x.writeClose("whileStatement")
}

func (s *DoStatement) toXml(x XmlWriter) {
	x.writeOpen("doStatement")
	x.writeKeyword(token.Do.String())
	s.Call.toXml(x)
	x.writeSymbol(token.SemiColon.String())
	x.writeClose("doStatement")
}

func (s *ReturnStatement) toXml(x XmlWriter) {
	x.writeOpen("returnStatement")
	x.writeKeyword(token.Return.String())

	if s.Value != nil {
		writeExpression(s.Value, x)
	}

	x.writeSymbol(token.SemiColon.String())
	x.writeClose("returnStatement")
}

func writeExpression(e Expression, x XmlWriter) {
	x.writeOpen("expression")
	e.toXml(x)
	x.writeClose("expression")
}

func (t *BinaryTerm) toXml(x XmlWriter) {
	t.Left.toXml(x)
	x.writeSymbol(t.Operator.String())
	t.Right.toXml(x)
}

func (t *IntegerConst) toXml(x XmlWriter) {
	x.writeOpen("term")
	x.writeIntConst(t.Value)
	x.writeClose("term")
}

func (t *StringConstant) toXml(x XmlWriter) {
	x.writeOpen("term")
	x.writeStringConst(t.Value)
	x.writeClose("term")
}

func (t *BooleanConstant) toXml(x XmlWriter) {
	x.writeOpen("term")
	if t.Value {
		x.writeKeyword(token.True.String())
	} else {
		x.writeKeyword(token.False.String())
	}
	x.writeClose("term")
}

func (t *NullConstant) toXml(x XmlWriter) {
	x.writeOpen("term")
	x.writeKeyword(token.Null.String())
	x.writeClose("term")
}

func (t *ThisConstant) toXml(x XmlWriter) {
	x.writeOpen("term")
	x.writeKeyword(token.This.String())
	x.writeClose("term")
}

func (t *VarName) toXml(x XmlWriter) {
	x.writeOpen("term")
	x.writeIdentifier(t.Name)
	x.writeClose("term")
}

func (t *ArrayAccess) toXml(x XmlWriter) {
	x.writeOpen("term")
	x.writeIdentifier(t.Name)
	x.writeSymbol(token.LeftBracket.String())
	writeExpression(t.Index, x)
	x.writeSymbol(token.RightBracket.String())
	x.writeClose("term")
}

func (t SubroutineCall) toXml(x XmlWriter) {
	if t.IsTerm {
		x.writeOpen("term")
	}

	if t.Name != "" {
		x.writeIdentifier(t.Name)
		x.writeSymbol(token.Dot.String())
	}

	x.writeIdentifier(t.SubName)
	x.writeSymbol(token.LeftParen.String())

	x.writeOpen("expressionList")
	if len(t.Expressions) > 0 {
		writeExpression(t.Expressions[0], x)
		for _, e := range t.Expressions[1:] {
			x.writeSymbol(token.Comma.String())
			writeExpression(e, x)
		}
	}
	x.writeClose("expressionList")
	x.writeSymbol(token.RightParen.String())

	if t.IsTerm {
		x.writeClose("term")
	}
}

func (t *BracketExpression) toXml(x XmlWriter) {
	x.writeOpen("term")
	x.writeSymbol(token.LeftParen.String())
	writeExpression(t.Expression, x)
	x.writeSymbol(token.RightParen.String())
	x.writeClose("term")
}

func (t UnaryTerm) toXml(x XmlWriter) {
	x.writeOpen("term")
	x.writeSymbol(t.Operator.String())
	t.Term.toXml(x)
	x.writeClose("term")
}

func (t *Type) toXml(x XmlWriter) {
	if t.Class == "" {
		x.writeKeyword(t.Token.String())
	} else {
		x.writeIdentifier(t.Class)
	}
}
