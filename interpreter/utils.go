package interpreter

import (
	"strconv"
	"strings"

	"gismolang.org/compiler/parser"
	"gismolang.org/compiler/tokenizer/tokentype"
)

func syntaxNode2Value(expression *parser.SyntaxNode) Value {
    // If it's a leaf node (Literal)
	if expression.Value != nil {
		tok := expression.Value // This is the *tokenizer.Token
		
		switch expression.Value.TokenType {
		case tokentype.Number:
			tokenVal := expression.Value.Value
			if strings.Contains(tokenVal, ".") {
				if strings.HasPrefix(tokenVal, ".") { tokenVal = "0" + tokenVal }
				value, _ := strconv.ParseFloat(tokenVal, 64)
				return &Float{Value: value, BaseValue: BaseValue{Token: tok}}
			}
			value, _ := strconv.ParseInt(tokenVal, 0, 64)
			return &Integer{Value: value, BaseValue: BaseValue{Token: tok}}
		case tokentype.String:
			return &String{Value: expression.Value.Value, BaseValue: BaseValue{Token: tok}}
		case tokentype.Operator, tokentype.Identifier, tokentype.LParent, tokentype.LCurlyParent:
			return &Symbol{Value: expression.Value.Alias, BaseValue: BaseValue{Token: tok}}
		case tokentype.Module:
			return &Symbol{Value: expression.Value.Alias, BaseValue: BaseValue{Token: tok}}
		}
		return &Nil{BaseValue: BaseValue{Token: tok}}
	}

    // If it's a list (ConsCell)
    // We try to use the operator's token as the token for the whole ConsCell
	var arguments Value = &Nil{}
	for i := len(expression.Arguments) - 1; i >= 0; i-- {
		arg := expression.Arguments[i]
		arguments = &ConsCell{
			Car: syntaxNode2Value(arg),
			Cdr: arguments,
            // Assign token from argument if possible, or nil
            BaseValue: BaseValue{Token: arg.Value}, 
		}
	}
    
    // The main ConsCell gets the token from the Operator
	return &ConsCell{
		Car: syntaxNode2Value(expression.Operator),
		Cdr: arguments,
        BaseValue: BaseValue{Token: expression.Operator.Value},
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