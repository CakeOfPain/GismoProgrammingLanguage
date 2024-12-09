package parser

import (
	"fmt"

	"gismolang.org/compiler/tokenizer"
	"gismolang.org/compiler/tokenizer/tokentype"
)

// SyntaxNode represents a node in the abstract syntax tree (AST).
type SyntaxNode struct {
	Operator  *SyntaxNode
	Arguments []*SyntaxNode
	Value     *tokenizer.Token
}

// NewValueNode creates a new syntax node for a literal value.
func NewValueNode(value *tokenizer.Token) *SyntaxNode {
	return &SyntaxNode{
		Operator:  nil,
		Arguments: nil,
		Value:     value,
	}
}

// NewSExpression creates a new syntax node for an S-expression with an operator and arguments.
func NewSExpression(operator *SyntaxNode, arguments []*SyntaxNode) *SyntaxNode {
	return &SyntaxNode{
		Operator:  operator,
		Arguments: arguments,
		Value:     nil,
	}
}

// Parse generates an AST from a list of tokens.
func Parse(tokens []*tokenizer.Token, source string) *SyntaxNode {
	r := CreateTokenReader(tokens)
	return NewSExpression(NewValueNode(tokenizer.ModuleToken(source)), parseExpressions(&r))
}

// parseExpressions parses a series of expressions.
func parseExpressions(r *TokenReader) []*SyntaxNode {
	var expressions []*SyntaxNode
	for {
		node := parseExpression(r, 0)
		if node != nil {
			expressions = append(expressions, node)
		}
		if r.PeekNext(0).TokenType != tokentype.Newline && r.PeekNext(0).TokenType != tokentype.Semicolon {
			break
		}
		r.Next() // consume newline or semicolon
	}
	return expressions
}


// parseExpression parses a single expression, handling operators and precedence.
func parseExpression(r *TokenReader, parentPrecedence int) *SyntaxNode {
    var left *SyntaxNode

    // Handle unary operators
    if isUnaryOperator(r) {
        operator := r.Next()
        operand := parseLiteral(r)
        if operand == nil {
            left = NewValueNode(operator)
        } else {
            left = NewSExpression(NewValueNode(operator), []*SyntaxNode{operand})
        }
    } else {
        // Parse literal
        left = parseLiteral(r)
    }

    for {
        // Handle function calls with higher precedence
        if r.PeekNext(0).TokenType == tokentype.LParent {
            precedence := tokenizer.FunctionCallPrecedence // Assign higher precedence than dot operator
            if precedence < parentPrecedence {
                break
            }
            left = parseParentCall(r, left)
            continue
        }

		// Handle curly function calls with higher precedence
        if r.PeekNext(0).TokenType == tokentype.LCurlyParent {
            precedence := tokenizer.CurlyCallPrecedence // Assign higher precedence than dot operator
            if precedence < parentPrecedence {
                break
            }
            left = parseCurlyParentCall(r, left)
            continue
        }

        precedence := r.PeekNext(0).BinPrec
        if precedence == 0 || precedence < parentPrecedence {
            break
        }
        operator := r.Next()
        right := parseExpression(r, precedence+1)
        if right == nil {
            left = NewSExpression(NewValueNode(operator), []*SyntaxNode{left})
        } else {
            left = NewSExpression(NewValueNode(operator), []*SyntaxNode{left, right})
        }
    }

    return left
}

func parseParentCall(r *TokenReader, left *SyntaxNode) *SyntaxNode {
    lparent := r.Next()
    if r.PeekNext(0).TokenType == tokentype.RParent {
        r.Next()
		lparent.Alias = "@call"
        return NewSExpression(NewValueNode(lparent), []*SyntaxNode{left})
    } else {
        arguments := parseExpression(r, 0)
        if r.Next().TokenType != tokentype.RParent {
            panic("Expected closing parenthesis")
        }
		lparent.Alias = "@call"
        return NewSExpression(NewValueNode(lparent), []*SyntaxNode{left, arguments})
    }
}

func parseCurlyParentCall(r *TokenReader, left *SyntaxNode) *SyntaxNode {
    lparent := r.Next()
    if r.PeekNext(0).TokenType == tokentype.RCurlyParent {
        r.Next() 
		lparent.Alias = "@callCurly"
        return NewSExpression(NewValueNode(lparent), []*SyntaxNode{left})
    } else {
        arguments := parseExpressions(r)
        if r.Next().TokenType != tokentype.RCurlyParent {
            panic("Expected closing curly parenthesis")
        }
		lparent.Alias = "@callCurly"
        return NewSExpression(NewValueNode(lparent), append([]*SyntaxNode{left}, arguments...))
    }
}

// parseLiteral handles literals such as identifiers, strings, and parentheses expressions.
func parseLiteral(r *TokenReader) *SyntaxNode {
	switch r.PeekNext(0).TokenType {
	case tokentype.Identifier, tokentype.Operator, tokentype.String, tokentype.Number:
		return NewValueNode(r.Next())
	case tokentype.LParent:
		r.Next() // consume '('
		expr := parseExpression(r, 0)
		if r.Next().TokenType != tokentype.RParent {
			panic("Missing closing parenthesis")
		}
		return expr
	case tokentype.LCurlyParent:
		operator := r.Next()
		statements := parseExpressions(r)
		if r.Next().TokenType != tokentype.RCurlyParent {
			panic("Missing closing curly brace")
		}
		operator.Alias = "@begin"
		return NewSExpression(NewValueNode(operator), statements)
	default:
		return nil
	}
}

// Helper to determine if the next token is a unary operator.
func isUnaryOperator(r *TokenReader) bool {
	return r.PeekNext(0).TokenType == tokentype.Operator
}

// String returns a string representation of the syntax node.
func (syntaxNode SyntaxNode) String() string {
	if syntaxNode.Value != nil {
		if syntaxNode.Value.TokenType == tokentype.String {
			return "\"" + syntaxNode.Value.Alias + "\""
		}
		return syntaxNode.Value.Alias
	}

	var arguments string
	for i, arg := range syntaxNode.Arguments {
		if arg == nil {
			continue
		}
		arguments += fmt.Sprint(arg)
		if (i + 1) < len(syntaxNode.Arguments) {
			arguments += " "
		} else {
			arguments = " " + arguments
		}
	}

	if syntaxNode.Operator == nil {
		return "(nil" + arguments + ")"
	}

	return "(" + fmt.Sprint(syntaxNode.Operator) + arguments + ")"
}

// GetOperationType returns the token type of the operator in the syntax node.
func (syntaxNode SyntaxNode) GetOperationType() tokentype.TokenType {
	if syntaxNode.Operator.Value == nil {
		return tokentype.None
	}
	return syntaxNode.Operator.Value.TokenType
}