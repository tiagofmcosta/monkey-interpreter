package evaluator

import (
	"monkey-interpreter/ast"
	"monkey-interpreter/object"
	"monkey-interpreter/token"
)

var (
	Null  = &object.Null{}
	True  = &object.Boolean{Value: true}
	False = &object.Boolean{Value: false}
)

func Eval(node ast.Node) object.Object {
	switch node := node.(type) {
	// Statements
	case *ast.Program:
		return evalStatements(node.Statements)
	case *ast.ExpressionStatement:
		return Eval(node.Expression)

	// Expressions
	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}
	case *ast.BooleanLiteral:
		return nativeBoolToBooleanObject(node.Value)
	case *ast.PrefixExpression:
		right := Eval(node.Right)
		return evalPrefixExpression(node.Operator, right)
	case *ast.InfixExpression:
		left := Eval(node.Left)
		right := Eval(node.Right)
		return evalInfixExpression(node.Operator, left, right)
	}

	return nil
}

func evalStatements(statements []ast.Statement) object.Object {
	var result object.Object

	for _, statement := range statements {
		result = Eval(statement)
	}

	return result
}

func nativeBoolToBooleanObject(input bool) *object.Boolean {
	if input {
		return True
	}

	return False
}

func evalPrefixExpression(operator string, right object.Object) object.Object {
	switch operator {
	case token.Bang:
		return evalBangOperatorExpression(right)
	case token.Minus:
		return evalMinusPrefixOperatorExpression(right)
	default:
		return Null
	}
}

func evalMinusPrefixOperatorExpression(right object.Object) object.Object {
	if right.Type() != object.IntegerObj {
		return Null
	}

	value := right.(*object.Integer).Value
	return &object.Integer{Value: -value}
}

func evalBangOperatorExpression(right object.Object) object.Object {
	switch right {
	case True:
		return False
	case False:
		return True
	case Null:
		return True
	default:
		return False
	}
}

func evalInfixExpression(operator string, left, right object.Object) object.Object {
	switch {
	case left.Type() == object.IntegerObj && right.Type() == object.IntegerObj:
		return evalIntegerInfixExpression(operator, left, right)
	case operator == token.Equal:
		return nativeBoolToBooleanObject(left == right)
	case operator == token.NotEqual:
		return nativeBoolToBooleanObject(left != right)
	default:
		return Null
	}
}

func evalIntegerInfixExpression(operator string, left, right object.Object) object.Object {
	leftVal := left.(*object.Integer).Value
	rightVal := right.(*object.Integer).Value

	switch operator {
	case token.Plus:
		return &object.Integer{Value: leftVal + rightVal}
	case token.Minus:
		return &object.Integer{Value: leftVal - rightVal}
	case token.Asterisk:
		return &object.Integer{Value: leftVal * rightVal}
	case token.Slash:
		return &object.Integer{Value: leftVal / rightVal}
	case token.LesserThan:
		return nativeBoolToBooleanObject(leftVal < rightVal)
	case token.GreaterThan:
		return nativeBoolToBooleanObject(leftVal > rightVal)
	case token.Equal:
		return nativeBoolToBooleanObject(leftVal == rightVal)
	case token.NotEqual:
		return nativeBoolToBooleanObject(leftVal != rightVal)
	default:
		return Null
	}
}
