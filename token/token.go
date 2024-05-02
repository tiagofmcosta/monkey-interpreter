package token

type Type string

type Token struct {
	Type    Type
	Literal string
}

const (
	Illegal = "ILLEGAL"
	Eof     = "EOF"

	Ident = "IDENT"
	Int   = "INT"

	Assign   = "="
	Plus     = "+"
	Minus    = "-"
	Bang     = "!"
	Slash    = "/"
	Asterisk = "*"
	Lt       = "<"
	Gt       = ">"
	Eq       = "=="
	NotEq    = "!="

	Semicolon = ";"
	Comma     = ","

	LParen = "("
	RParen = ")"
	LBrace = "{"
	RBrace = "}"

	Function = "FUNCTION"
	Let      = "LET"
	True     = "TRUE"
	False    = "FALSE"
	If       = "IF"
	Else     = "ELSE"
	Return   = "RETURN"
)

var keywords = map[string]Type{
	"fn":     Function,
	"let":    Let,
	"true":   True,
	"false":  False,
	"if":     If,
	"else":   Else,
	"return": Return,
}

func LookupIdent(ident string) Type {
	if t, ok := keywords[ident]; ok {
		return t
	}
	return Ident
}
