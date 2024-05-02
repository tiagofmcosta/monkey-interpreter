package lexer

import (
	"monkey-interpreter/token"
)

type Lexer struct {
	input        string
	position     int
	readPosition int
	char         byte
}

func New(input string) *Lexer {
	lexer := &Lexer{input: input}
	lexer.readChar()
	return lexer
}

func (lexer *Lexer) NextToken() token.Token {
	var tok token.Token

	lexer.skipWhitespace()

	switch lexer.char {
	case '=':
		if lexer.peekChar() == '=' {
			char := lexer.char
			lexer.readChar()
			tok = token.Token{Type: token.Equal, Literal: string(char) + string(lexer.char)}
		} else {
			tok = newToken(token.Assign, lexer.char)
		}
	case '+':
		tok = newToken(token.Plus, lexer.char)
	case '-':
		tok = newToken(token.Minus, lexer.char)
	case '!':
		if lexer.peekChar() == '=' {
			char := lexer.char
			lexer.readChar()
			tok = token.Token{Type: token.NotEqual, Literal: string(char) + string(lexer.char)}
		} else {
			tok = newToken(token.Bang, lexer.char)
		}
	case '/':
		tok = newToken(token.Slash, lexer.char)
	case '*':
		tok = newToken(token.Asterisk, lexer.char)
	case '<':
		tok = newToken(token.LesserThan, lexer.char)
	case '>':
		tok = newToken(token.GreaterThan, lexer.char)
	case ';':
		tok = newToken(token.Semicolon, lexer.char)
	case ',':
		tok = newToken(token.Comma, lexer.char)
	case '(':
		tok = newToken(token.LParen, lexer.char)
	case ')':
		tok = newToken(token.RParen, lexer.char)
	case '{':
		tok = newToken(token.LBrace, lexer.char)
	case '}':
		tok = newToken(token.RBrace, lexer.char)
	case 0:
		tok.Literal = ""
		tok.Type = token.Eof
	default:
		if isLetter(lexer.char) {
			tok.Literal = lexer.readIdentifier()
			tok.Type = token.LookupIdent(tok.Literal)
			return tok
		} else if isDigit(lexer.char) {
			tok.Type = token.Int
			tok.Literal = lexer.readNumber()
			return tok
		} else {
			tok = newToken(token.Illegal, lexer.char)
		}
	}

	lexer.readChar()
	return tok
}

func (lexer *Lexer) readChar() {
	if lexer.readPosition >= len(lexer.input) {
		lexer.char = 0
	} else {
		lexer.char = lexer.input[lexer.readPosition]
	}
	lexer.position = lexer.readPosition
	lexer.readPosition++
}

func (lexer *Lexer) readIdentifier() string {
	position := lexer.position
	for isLetter(lexer.char) {
		lexer.readChar()
	}

	return lexer.input[position:lexer.position]
}

func (lexer *Lexer) readNumber() string {
	position := lexer.position
	for isDigit(lexer.char) {
		lexer.readChar()
	}

	return lexer.input[position:lexer.position]
}

func (lexer *Lexer) skipWhitespace() {
	for lexer.char == ' ' || lexer.char == '\t' || lexer.char == '\n' || lexer.char == '\r' {
		lexer.readChar()
	}
}

func (lexer *Lexer) peekChar() byte {
	if lexer.readPosition >= len(lexer.input) {
		return 0
	} else {
		return lexer.input[lexer.readPosition]
	}
}

func newToken(tokenType token.Type, char byte) token.Token {
	return token.Token{Type: tokenType, Literal: string(char)}
}

func isLetter(char byte) bool {
	return 'a' <= char && char <= 'z' || 'A' <= char && char <= 'Z' || char == '_'
}

func isDigit(char byte) bool {
	return '0' <= char && char <= '9'
}
