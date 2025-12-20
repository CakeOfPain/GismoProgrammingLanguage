package interpreter

import (
	"strconv"
	"strings"

	"gismolang.org/compiler/parser"
	"gismolang.org/compiler/tokenizer/tokentype"
)

func syntaxNode2Value(expression *parser.SyntaxNode) Value {
    if(expression.Value != nil) {
        switch(expression.Value.TokenType) {
        case tokentype.Number:
            tokenVal := expression.Value.Value
            
            // Check for Dot -> Float (Handles "1.23" and ".123")
            if strings.Contains(tokenVal, ".") {
                // If starts with dot, prepend 0 for safety if needed
                if strings.HasPrefix(tokenVal, ".") {
                    tokenVal = "0" + tokenVal
                }
                value, _ := strconv.ParseFloat(tokenVal, 64)
                return &Float{
                    Value: value,
                }
            }
            
            // Check for Integer (Base 10, Hex, Bin, Octal)
            // base=0 automatically detects 0x, 0b, 0o
            value, _ := strconv.ParseInt(tokenVal, 0, 64)
            return &Integer{
                Value: value,
            }

        case tokentype.String:
            return &String{
                Value: expression.Value.Value,
            }
        case tokentype.Operator, tokentype.Identifier, tokentype.LParent, tokentype.LCurlyParent:
            return &Symbol{
                Value: expression.Value.Alias,
            }
        case tokentype.Module:
            return &Symbol{
                Value: expression.Value.Alias,
            }
        }
        // OTHERWISE (SHOULD NOT HAPPEN)
        return &Nil{}
    }

    var arguments Value = &Nil{}
    for i := len(expression.Arguments)-1; i>=0; i-- {
        arg := expression.Arguments[i]
        arguments = &ConsCell{
            Car: syntaxNode2Value(arg),
            Cdr: arguments,
        }
    }

    return &ConsCell{
        Car: syntaxNode2Value(expression.Operator),
        Cdr: arguments,
    }
}

func subSymbol(value Value, sym *Symbol, sub Value, limited bool) Value {
    switch v := value.(type) {
    case *ConsCell:
        if limited && isMacroVariableAssignment(v) {
            return v
        }
        return &ConsCell{
            Car: subSymbol(v.Car, sym, sub, limited),
            Cdr: subSymbol(v.Cdr, sym, sub, limited),
        }
    case *Symbol:
        if v.Value == sym.Value {
            return sub
        }
        break
    }

    return value
}

func isMacroVariableAssignment(expr *ConsCell) bool {
    if expr.Car.String() != "::=" {
        return false
    }
    if arg, ok := expr.Cdr.(*ConsCell); ok {
        if arg.Car.GetTypeString() == "ConsCell" {
            return true
        }
    }
    return false
}

func flattenBySeparator(value Value, separator string) []Value {
    if consCell, ok := value.(*ConsCell); ok {
        first := consCell.Get(0)
        if first.GetTypeString() == "symbol" && first.String() == separator {
            return append(flattenBySeparator(consCell.Get(1), separator), consCell.Get(2))
        }
    }

    return []Value{value}
}

func getArgsList(args Value) []Value {
    return flattenBySeparator(args, ",")
}