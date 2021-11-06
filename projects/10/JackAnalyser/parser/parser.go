package parser

import (
	"fmt"
	"strconv"

	"github.com/ChelseaDH/JackAnalyser/lexer"
	"github.com/ChelseaDH/JackAnalyser/token"
)

type Parser struct {
	lexer *lexer.Lexer

	current          token.Token
	next             token.Token
	value, nextValue string
}

func NewParser(lexer *lexer.Lexer) Parser {
	return Parser{
		lexer: lexer,
	}
}

func (p *Parser) Parse() (j *JackClass, err error) {
	defer func() {
		recovered := recover()
		if recovered != nil {
			x, ok := recovered.(error)
			if x != nil && ok {
				err = x
			}
			panic(recovered)
		}
	}()

	p.advance()
	p.expect(token.Class)
	p.expect(token.Identifier)
	name := p.value
	p.expect(token.LeftBrace)

	var variables []ClassVarDec
	vars := p.parseClassVarDec()
	for vars != nil {
		variables = append(variables, vars...)
		vars = p.parseClassVarDec()
	}

	var subroutines []JackSubroutine
	subroutine := p.parseSubroutine()
	for subroutine != nil {
		subroutines = append(subroutines, *subroutine)
		subroutine = p.parseSubroutine()
	}

	p.expect(token.RightBrace)

	return &JackClass{
		Name:        name,
		VarDecs:     variables,
		Subroutines: subroutines,
	}, nil
}

func (p *Parser) advance() {
	var err error
	p.current, p.value = p.next, p.nextValue
	p.next, p.nextValue, err = p.lexer.Next()
	if err != nil {
		panic(err)
	}
}

func (p *Parser) expect(token token.Token) {
	if p.next != token {
		panic(fmt.Errorf("expected token '%s', got '%s' (value: %s)", token.String(), p.next.String(), p.nextValue))
	}

	p.advance()
}

func (p *Parser) accept(token token.Token) bool {
	return p.next == token
}

func (p *Parser) parseType() Type {
	switch p.next {
	case token.Int, token.Char, token.Boolean:
		p.advance()
		return Type{Token: p.current}
	case token.Identifier:
		p.advance()
		return Type{Token: p.current, Class: p.value}
	default:
		panic(fmt.Errorf("unknown type %s", p.current.String()))
	}
}

func (p *Parser) parseClassVarDec() []ClassVarDec {
	if !p.accept(token.Static) && !p.accept(token.Field) {
		return nil
	}
	static := p.accept(token.Static)
	p.advance()
	typ := p.parseType()
	p.expect(token.Identifier)
	name := p.value

	vars := []ClassVarDec{
		{
			Static: static,
			VarDec: VarDec{
				Type: typ,
				Name: name,
			},
		},
	}

	for p.accept(token.Comma) {
		p.advance()
		p.expect(token.Identifier)
		vars = append(vars, ClassVarDec{
			Static: static,
			VarDec: VarDec{
				Type: typ,
				Name: p.value,
			},
		})
	}

	p.expect(token.SemiColon)
	return vars
}

func (p *Parser) parseSubroutine() *JackSubroutine {
	var sTyp token.Token
	switch p.next {
	case token.Constructor, token.Function, token.Method:
		sTyp = p.next
		p.advance()
	default:
		return nil
	}

	var typ Type
	if p.accept(token.Void) {
		typ = Type{Token: token.Void}
		p.advance()
	} else {
		typ = p.parseType()
	}

	p.expect(token.Identifier)
	name := p.value
	p.expect(token.LeftParen)

	var params []Param
	var paramType Type
	switch p.next {
	case token.Int, token.Char, token.Boolean, token.Identifier:
		paramType = p.parseType()
		p.expect(token.Identifier)
		params = append(params, Param{
			Type: paramType,
			Name: p.value,
		})

		for p.accept(token.Comma) {
			p.advance()
			paramType = p.parseType()
			p.expect(token.Identifier)

			params = append(params, Param{
				Type: paramType,
				Name: p.value,
			})
		}
	default:
	}

	p.expect(token.RightParen)
	vars, statements := p.parseSubroutineBody()

	return &JackSubroutine{
		SType:      sTyp,
		ReturnType: typ,
		SName:      name,
		ParamList:  params,
		Vars:       vars,
		Statements: statements,
	}
}

func (p *Parser) parseParam() *Param {
	var typ Type
	switch p.next {
	case token.Int, token.Char, token.Boolean, token.Identifier:
		typ = p.parseType()
	default:
		return nil
	}
	p.expect(token.Identifier)

	return &Param{
		Type: typ,
		Name: p.value,
	}
}

func (p *Parser) parseSubroutineBody() ([]VarDec, []Statement) {
	p.expect(token.LeftBrace)

	var variables []VarDec
	vars := p.parseVarDec()
	for vars != nil {
		variables = append(variables, vars...)
		vars = p.parseVarDec()
	}

	statements := p.parseStatements()
	p.expect(token.RightBrace)

	return variables, statements
}

func (p *Parser) parseVarDec() []VarDec {
	if p.accept(token.Var) {
		p.advance()
	} else {
		return nil
	}

	typ := p.parseType()
	p.expect(token.Identifier)

	vars := []VarDec{
		{
			Type: typ,
			Name: p.value,
		},
	}

	for p.accept(token.Comma) {
		p.advance()
		p.expect(token.Identifier)
		vars = append(vars, VarDec{
			Type: typ,
			Name: p.value,
		},
		)
	}

	p.expect(token.SemiColon)
	return vars
}

func (p *Parser) parseStatement() Statement {
	statements := []func() Statement{
		p.parseLetStatement,
		p.parseIfStatement,
		p.parseWhileStatement,
		p.parseDoStatement,
		p.parseReturnStatement,
	}

	for _, fn := range statements {
		statement := fn()
		if statement != nil {
			return statement
		}
	}

	return nil
}

func (p *Parser) parseStatements() []Statement {
	statements := []Statement{}
	statement := p.parseStatement()
	for statement != nil {
		statements = append(statements, statement)
		statement = p.parseStatement()
	}

	return statements
}

func (p *Parser) parseLetStatement() Statement {
	if p.accept(token.Let) {
		p.advance()
	} else {
		return nil
	}

	p.expect(token.Identifier)
	name := p.value

	var index Expression
	if p.accept(token.LeftBracket) {
		p.advance()
		index = p.parseExpression()
		p.expect(token.RightBracket)
	}

	p.expect(token.Equals)
	expression := p.parseExpression()
	p.expect(token.SemiColon)

	return &LetStatement{
		Name:  name,
		Index: index,
		Value: expression,
	}
}

func (p *Parser) parseIfStatement() Statement {
	if p.accept(token.If) {
		p.advance()
	} else {
		return nil
	}

	p.expect(token.LeftParen)
	condition := p.parseExpression()
	p.expect(token.RightParen)

	p.expect(token.LeftBrace)
	body := p.parseStatements()
	p.expect(token.RightBrace)

	var elseBody []Statement
	if p.accept(token.Else) {
		p.advance()
		p.expect(token.LeftBrace)
		elseBody = p.parseStatements()
		p.expect(token.RightBrace)
	}

	return &IfStatement{
		Condition: condition,
		Body:      body,
		Else:      elseBody,
	}
}

func (p *Parser) parseWhileStatement() Statement {
	if p.accept(token.While) {
		p.advance()
	} else {
		return nil
	}

	p.expect(token.LeftParen)
	condition := p.parseExpression()
	p.expect(token.RightParen)

	p.expect(token.LeftBrace)
	statements := p.parseStatements()
	p.expect(token.RightBrace)

	return &WhileStatement{
		Condition: condition,
		Body:      statements,
	}
}

func (p *Parser) parseDoStatement() Statement {
	if p.accept(token.Do) {
		p.advance()
	} else {
		return nil
	}

	p.expect(token.Identifier)
	call := p.parseSubroutineCall()
	p.expect(token.SemiColon)

	return &DoStatement{Call: call}
}

func (p *Parser) parseReturnStatement() Statement {
	if p.accept(token.Return) {
		p.advance()
	} else {
		return nil
	}

	var value Expression
	if !p.accept(token.SemiColon) {
		value = p.parseExpression()
	}

	p.expect(token.SemiColon)

	return &ReturnStatement{Value: value}
}

func (p *Parser) parseSubroutineCall() SubroutineCall {
	name := p.value

	var subName string
	if p.accept(token.Dot) {
		p.advance()
		p.expect(token.Identifier)
		subName = p.value
	} else {
		subName = name
		name = ""
	}

	p.expect(token.LeftParen)
	var expressions []Expression
	if !p.accept(token.RightParen) {
		expressions = []Expression{p.parseExpression()}
		for p.accept(token.Comma) {
			p.advance()
			expressions = append(expressions, p.parseExpression())
		}
	}
	p.expect(token.RightParen)

	return SubroutineCall{
		Name:        name,
		SubName:     subName,
		Expressions: expressions,
	}
}

func (p *Parser) parseExpression() Expression {
	term := p.parseTerm()

	switch p.next {
	case token.Plus, token.Minus, token.Mult, token.Div, token.And, token.Or,
		token.LessThan, token.GreaterThan, token.Equals:
		op := p.next
		p.advance()
		return &BinaryTerm{
			Left:     term,
			Operator: op,
			Right:    p.parseExpression(),
		}

	default:
		return term
	}
}

func (p *Parser) parseTerm() Expression {
	switch p.next {
	case token.IntConst:
		i, err := strconv.Atoi(p.nextValue)
		if err != nil {
			panic(err)
		}
		p.advance()
		return &IntegerConst{Value: i}

	case token.StringConst:
		p.advance()
		return &StringConstant{Value: p.value}

	case token.True:
		p.advance()
		return &BooleanConstant{Value: true}

	case token.False:
		p.advance()
		return &BooleanConstant{Value: false}

	case token.Null:
		p.advance()
		return &NullConstant{}

	case token.This:
		p.advance()
		return &ThisConstant{}

	case token.LeftParen:
		p.advance()
		expression := p.parseExpression()
		p.expect(token.RightParen)

		return &BracketExpression{Expression: expression}

	case token.Minus, token.Not:
		p.advance()
		op := p.current
		return UnaryTerm{
			Operator: op,
			Term:     p.parseTerm(),
		}

	case token.Identifier:
		p.advance()
		name := p.value
		if p.accept(token.LeftBracket) {
			p.advance()
			expression := p.parseExpression()
			p.expect(token.RightBracket)

			return &ArrayAccess{
				Name:  name,
				Index: expression,
			}
		}

		if p.accept(token.Dot) || p.accept(token.LeftParen) {
			return p.parseSubroutineCall()
		}

		return &VarName{Name: name}

	default:
		panic(fmt.Errorf("unexpected token whilst parsing term: %s", p.next))
	}
}
