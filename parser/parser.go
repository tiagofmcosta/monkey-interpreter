package parser

import (
	"fmt"
	"monkey-interpreter/ast"
	"monkey-interpreter/lexer"
	"monkey-interpreter/token"
	"strconv"
)

const (
	_ int = iota
	Lowest
	Equals        // ==
	LesserGreater // < or >
	Sum           // +
	Product       // *
	Prefix        // -X or !X
	Call          // myFunction(X)
)

var precedences = map[token.Type]int{
	token.Equal:       Equals,
	token.NotEqual:    Equals,
	token.LesserThan:  LesserGreater,
	token.GreaterThan: LesserGreater,
	token.Plus:        Sum,
	token.Minus:       Sum,
	token.Slash:       Product,
	token.Asterisk:    Product,
	token.LParen:      Call,
}

type Parser struct {
	lexer  *lexer.Lexer
	errors []string

	curToken  token.Token
	peekToken token.Token

	prefixParseFns map[token.Type]prefixParseFn
	infixParseFns  map[token.Type]infixParseFn
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		lexer:  l,
		errors: []string{},
	}

	p.prefixParseFns = make(map[token.Type]prefixParseFn)
	p.registerPrefix(token.Ident, p.parseIdentifier)
	p.registerPrefix(token.Int, p.parseIntegerLiteral)
	p.registerPrefix(token.Bang, p.parsePrefixExpression)
	p.registerPrefix(token.Minus, p.parsePrefixExpression)
	p.registerPrefix(token.True, p.parseBoolean)
	p.registerPrefix(token.False, p.parseBoolean)
	p.registerPrefix(token.LParen, p.parseGroupedExpression)
	p.registerPrefix(token.If, p.parseIfExpression)
	p.registerPrefix(token.Function, p.parseFunctionLiteral)

	p.infixParseFns = make(map[token.Type]infixParseFn)
	p.registerInfix(token.Plus, p.parseInfixExpression)
	p.registerInfix(token.Minus, p.parseInfixExpression)
	p.registerInfix(token.Slash, p.parseInfixExpression)
	p.registerInfix(token.Asterisk, p.parseInfixExpression)
	p.registerInfix(token.Equal, p.parseInfixExpression)
	p.registerInfix(token.NotEqual, p.parseInfixExpression)
	p.registerInfix(token.LesserThan, p.parseInfixExpression)
	p.registerInfix(token.GreaterThan, p.parseInfixExpression)
	p.registerInfix(token.LParen, p.parseCallExpression)

	// Read two tokens, so curToken and peekToken are both set
	p.nextToken()
	p.nextToken()

	return p
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) peekError(t token.Type) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead", t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.lexer.NextToken()
}

func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	for !p.curTokenIs(token.Eof) {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}

	return program
}

func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.Let:
		return p.parseLetStatement()
	case token.Return:
		return p.parseReturnStatement()
	default:
		return p.parseExpressionStatement()
	}
}

func (p *Parser) parseLetStatement() *ast.LetStatement {
	stmt := &ast.LetStatement{Token: p.curToken}

	if !p.expectPeek(token.Ident) {
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(token.Assign) {
		return nil
	}

	p.nextToken()

	stmt.Value = p.parseExpression(Lowest)

	if p.peekTokenIs(token.Semicolon) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.curToken}

	p.nextToken()

	stmt.ReturnValue = p.parseExpression(Lowest)

	if p.peekTokenIs(token.Semicolon) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) expectPeek(t token.Type) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	}

	p.peekError(t)
	return false
}

func (p *Parser) curTokenIs(t token.Type) bool {
	return p.curToken.Type == t
}

func (p *Parser) peekTokenIs(t token.Type) bool {
	return p.peekToken.Type == t
}

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

func (p *Parser) registerPrefix(tokenType token.Type, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType token.Type, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{Token: p.curToken}

	stmt.Expression = p.parseExpression(Lowest)

	if p.peekTokenIs(token.Semicolon) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) noPrefixParseFnError(t token.Type) {
	msg := fmt.Sprintf("no prefix parse function for %s found", t)
	p.errors = append(p.errors, msg)
}

func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}

	return Lowest
}

func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}

	return Lowest
}

func (p *Parser) parseExpression(precedence int) ast.Expression {
	prefix := p.prefixParseFns[p.curToken.Type]

	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}

	leftExp := prefix()

	for !p.peekTokenIs(token.Semicolon) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}

		p.nextToken()

		leftExp = infix(leftExp)
	}

	return leftExp
}

func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	lit := &ast.IntegerLiteral{Token: p.curToken}

	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	lit.Value = value

	return lit
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	expression := &ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}

	p.nextToken()

	expression.Right = p.parseExpression(Prefix)

	return expression
}

func (p *Parser) parseBoolean() ast.Expression {
	return &ast.BooleanLiteral{Token: p.curToken, Value: p.curTokenIs(token.True)}
}

func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken()

	expr := p.parseExpression(Lowest)

	if !p.expectPeek(token.RParen) {
		return nil
	}

	return expr
}

func (p *Parser) parseIfExpression() ast.Expression {
	expression := &ast.IfExpression{Token: p.curToken}

	if !p.expectPeek(token.LParen) {
		return nil
	}

	p.nextToken()

	expression.Condition = p.parseExpression(Lowest)

	if !p.expectPeek(token.RParen) {
		return nil
	}

	if !p.expectPeek(token.LBrace) {
		return nil
	}

	expression.Consequence = p.parseBlockStatement()

	if p.peekTokenIs(token.Else) {
		p.nextToken()

		if !p.expectPeek(token.LBrace) {
			return nil
		}

		expression.Alternative = p.parseBlockStatement()
	}

	return expression
}

func (p *Parser) parseFunctionLiteral() ast.Expression {
	lit := &ast.FunctionLiteral{Token: p.curToken}

	if !p.expectPeek(token.LParen) {
		return nil
	}

	lit.Parameters = p.parseFunctionParameters()

	if !p.expectPeek(token.LBrace) {
		return nil
	}

	lit.Body = p.parseBlockStatement()

	return lit
}

func (p *Parser) parseFunctionParameters() []*ast.Identifier {
	var identifiers []*ast.Identifier

	if p.peekTokenIs(token.RParen) {
		p.nextToken()
		return identifiers
	}

	p.nextToken()

	ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	identifiers = append(identifiers, ident)

	for p.peekTokenIs(token.Comma) {
		p.nextToken()
		p.nextToken()
		ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		identifiers = append(identifiers, ident)
	}

	if !p.expectPeek(token.RParen) {
		return nil
	}

	return identifiers
}

func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: p.curToken}
	block.Statements = []ast.Statement{}

	p.nextToken()

	for !p.curTokenIs(token.RBrace) && !p.curTokenIs(token.Eof) {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.nextToken()
	}

	return block
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expression := &ast.InfixExpression{
		Token:    p.curToken,
		Left:     left,
		Operator: p.curToken.Literal,
	}

	precedence := p.curPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)

	return expression
}

func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	exp := &ast.CallExpression{Token: p.curToken, Function: function}
	exp.Arguments = p.parseCallArguments()
	return exp
}

func (p *Parser) parseCallArguments() []ast.Expression {
	var args []ast.Expression

	if p.peekTokenIs(token.RParen) {
		p.nextToken()
		return args
	}

	p.nextToken()
	args = append(args, p.parseExpression(Lowest))

	for p.peekTokenIs(token.Comma) {
		p.nextToken()
		p.nextToken()
		args = append(args, p.parseExpression(Lowest))
	}

	if !p.expectPeek(token.RParen) {
		return nil
	}

	return args
}
