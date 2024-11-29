package interpreter

import (
	"fmt"

	"gismolang.org/compiler/config"
)

var Builtins = []BuiltinFunction{
    {callback: addInt, identifier: "$ADD"},
    {callback: subInt, identifier: "$SUB"},
    {callback: mulInt, identifier: "$MUL"},
    {callback: divInt, identifier: "$DIV"},
    {callback: modInt, identifier: "$MOD"},
    {callback: shiftLeftInt, identifier: "$SHL"},
    {callback: shiftRightInt, identifier: "$SHR"},
    {callback: bitwiseAnd, identifier: "$BWA"},
    {callback: bitwiseOr, identifier: "$BWO"},
    {callback: construct, identifier: "$CONS"},
    {callback: car, identifier: "$CAR"},
    {callback: cdr, identifier: "$CDR"},
    {callback: printValue, identifier: "$PRINT"},
    {callback: printlnValue, identifier: "$PRINTLN"},
    {callback: printScope, identifier: "$SCOPE"},
    {callback: typedef, identifier: "$TYPEDEF"},
    {callback: typeof, identifier: "$TYPEOF"},
    {callback: untype, identifier: "$UNTYPE"},
    {callback: quote, identifier: "$QUOTE"},
    {callback: replace, identifier: "$REPLACE"},
    {callback: eval, identifier: "$EVAL"},
    {callback: lambda, identifier: "$LAMBDA"},
    {callback: catString, identifier: "$CAT"},
    {callback: lenString, identifier: "$STRLEN"},
    {callback: charString, identifier: "$CHAR"},
    {callback: stringify, identifier: "$STR"},
    {callback: ifFunc, identifier: "$IF"},
    {callback: greater, identifier: "$GREATER"},
    {callback: equals, identifier: "$EQUALS"},
    {callback: get, identifier: "$GET"},
    {callback: set, identifier: "$SET"},
    {callback: def, identifier: "$DEF"},
    {callback: write2Output, identifier: "$WRITE"},
    {callback: writeByte2Output, identifier: "$WRITEB"},
    {callback: niler, identifier: "$NIL"},
}

func getArgsList(args Value) []Value {
    var values []Value = []Value{}

    if consCell, ok := args.(*ConsCell); ok {
        if consCell.Get(0).String() == "," {
            values = append(values, getArgsList(consCell.Get(1))...)
            values = append(values, consCell.Get(2))
        } else {
            values = append(values, args)
        }
    } else {
        values = append(values, args)
    }

    return values
}

func binaryIntOp(args Value, scope *Scope, op func(a, b int64) int64) Value {
    argsList := getArgsList(args)
    left := interpretExpression(argsList[0], scope)
    right := interpretExpression(argsList[1], scope)

    if leftInt, ok := left.(*Integer); ok {
        if rightInt, ok := right.(*Integer); ok {
            return &Integer{Value: op(leftInt.Value, rightInt.Value)}
        }
    }
    return &Nil{}
}

func addInt(args Value, scope *Scope) Value {
    return binaryIntOp(args, scope, func(a, b int64) int64 { return a + b })
}

func subInt(args Value, scope *Scope) Value {
    return binaryIntOp(args, scope, func(a, b int64) int64 { return a - b })
}

func mulInt(args Value, scope *Scope) Value {
    return binaryIntOp(args, scope, func(a, b int64) int64 { return a * b })
}

func divInt(args Value, scope *Scope) Value {
    return binaryIntOp(args, scope, func(a, b int64) int64 { return a / b })
}

func modInt(args Value, scope *Scope) Value {
    return binaryIntOp(args, scope, func(a, b int64) int64 { return a % b })
}

func shiftLeftInt(args Value, scope *Scope) Value {
    return binaryIntOp(args, scope, func(a, b int64) int64 { return a << uint(b) })
}

func shiftRightInt(args Value, scope *Scope) Value {
    return binaryIntOp(args, scope, func(a, b int64) int64 { return a >> uint(b) })
}

func bitwiseAnd(args Value, scope *Scope) Value {
    return binaryIntOp(args, scope, func(a, b int64) int64 { return a & b })
}

func bitwiseOr(args Value, scope *Scope) Value {
    return binaryIntOp(args, scope, func(a, b int64) int64 { return a | b })
}

func construct(args Value, scope *Scope) Value {
    argsList := getArgsList(args)
    left := interpretExpression(argsList[0], scope)
    right := interpretExpression(argsList[1], scope)
	return &ConsCell{
		Car: left,
		Cdr: right,
	}
}

func car(args Value, scope *Scope) Value {
    argsList := getArgsList(args)
    value := interpretExpression(argsList[0], scope)
	if consCell, ok := value.(*ConsCell); ok {
		return consCell.Car
	}
	return &Nil{}
}

func cdr(args Value, scope *Scope) Value {
    argsList := getArgsList(args)
    value := interpretExpression(argsList[0], scope)
	if consCell, ok := value.(*ConsCell); ok {
		return consCell.Cdr
	}
	return &Nil{}
}

func printValue(args Value, scope *Scope) Value {
    value := interpretExpression(args, scope)
    fmt.Print(value)
    return &Nil{}
}


func printlnValue(args Value, scope *Scope) Value {
    value := interpretExpression(args, scope)
    fmt.Println(value)
    return &Nil{}
}

func printScope(args Value, scope *Scope) Value {
    fmt.Println(scope)
    return &Nil{}
}


func typedef(args Value, scope *Scope) Value {
    argsList := getArgsList(args)
    left := interpretExpression(argsList[0], scope)
    right := argsList[1]
	return &TypedValue{
		Value: left,
		TypeValue: right,
	}
}

func typeof(args Value, scope *Scope) Value {
    argsList := getArgsList(args)
    left := interpretExpression(argsList[0], scope)
    right := argsList[1]
    if left.GetTypeString() == right.String() {
        return &Integer{
            Value: 1,
        }
    }

	return &Nil{}
}


func untype(args Value, scope *Scope) Value {
    value := interpretExpression(args, scope)
    if typeValue, ok := value.(*TypedValue); ok {
        return typeValue.Value
    }
	return &Nil{}
}


func quote(args Value, scope *Scope) Value {
	return args
}

func replace(args Value, scope *Scope) Value {
    argsList := getArgsList(args)
    expression := interpretExpression(argsList[0], scope)
    symbol := argsList[1]
    replacement := interpretExpression(argsList[2], scope)

    if symbol, ok := symbol.(*Symbol); ok {
        return subSymbol(expression, symbol, replacement, false)
    }

    return &Nil{}
}

func eval(args Value, scope *Scope) Value {
	return interpretExpression(interpretExpression(args, scope), scope)
}

func lambda(args Value, scope *Scope) Value {
    argsList := getArgsList(args)
    left := argsList[0]
    body := argsList[1]
    if sym, ok := left.(*Symbol); ok {
        return BuiltinFunction{
            callback: func(value Value, callerScope *Scope) Value {
                arg := &ConsCell{
                    Car: &Symbol{
                        Value: "@call",
                    },
                    Cdr: &ConsCell{
                        Car: &Symbol{
                            Value: "$QUOTE",
                        },
                        Cdr: &ConsCell{
                            Car: interpretExpression(value, callerScope),
                            Cdr: &Nil{},
                        },
                    },
                }
                return interpretExpression(subSymbol(body, sym, arg, false), scope)
            },
            identifier: "COMPOSED",
        }
    }
    return &Nil{}
}

func catString(args Value, scope *Scope) Value {
    argsList := getArgsList(args)
    left := interpretExpression(argsList[0], scope)
    right := interpretExpression(argsList[1], scope)

	return &String{
        Value: left.String() + right.String(),
	}
}

func lenString(args Value, scope *Scope) Value {
    argsList := getArgsList(args)
    value := interpretExpression(argsList[0], scope)
	return &Integer{
        Value: int64(len(value.String())),
	}
}

func charString(args Value, scope *Scope) Value {
    argsList := getArgsList(args)
    left := interpretExpression(argsList[0], scope)
    right := interpretExpression(argsList[1], scope)
    if index, ok := right.(*Integer); ok {
        return &Integer{
            Value: int64(left.String()[index.Value]),
        }
    }
    return &Nil{}
}

func stringify(args Value, scope *Scope) Value {
    argsList := getArgsList(args)
    value := interpretExpression(argsList[0], scope)
    if code, ok := value.(*Integer); ok {
        return &String{
            Value: string([]byte{byte(code.Value)}),
        }
    }
    return &Nil{}
}

func equals(args Value, scope *Scope) Value {
    argsList := getArgsList(args)
    left := interpretExpression(argsList[0], scope)
    right := interpretExpression(argsList[1], scope)
    if left.String() == right.String() {
        return &Integer{
            Value: 1,
        }
    }
    return &Nil{}
}

func greater(args Value, scope *Scope) Value {
    argsList := getArgsList(args)
    left := interpretExpression(argsList[0], scope)
    right := interpretExpression(argsList[1], scope)
    if leftInteger, ok := left.(*Integer); ok {
        if rightInteger, ok := right.(*Integer); ok {
            if leftInteger.Value > rightInteger.Value {
                return &Integer{
                    Value: 1,
                }
            }
        }
    }
    return &Nil{}
}

func ifFunc(args Value, scope *Scope) Value {
    argsList := getArgsList(args)
    cond := argsList[0]
    interpretCond := interpretExpression(cond, scope)
    if interpretCond.GetTypeString() == "Nil" {
        return interpretExpression(argsList[2], scope)
    }
    
    return interpretExpression(argsList[1], scope)
}

func get(args Value, scope *Scope) Value {
    argsList := getArgsList(args)
    index := argsList[0]
    return scope.GetLocal(index)
}

func set(args Value, scope *Scope) Value {
    argsList := getArgsList(args)
    index := argsList[0]
    value := interpretExpression(argsList[1], scope)
    scope.SetLocal(index, value)
    return &Nil{}
}

func def(args Value, scope *Scope) Value {
    argsList := getArgsList(args)
    index := argsList[0]
    value := interpretExpression(argsList[1], scope)
    scope.DefineLocal(index, value)
    return &Nil{}
}

func write2Output(args Value, scope *Scope) Value {
    argsList := getArgsList(args)
    message := interpretExpression(argsList[0], scope).String()
    if config.OutputEnabled {
        config.OutputFile.WriteString(message)
    }
    return &Nil{}
}


func writeByte2Output(args Value, scope *Scope) Value {
    argsList := getArgsList(args)
    result := interpretExpression(argsList[0], scope)
    if number, ok := result.(*Integer); ok && config.OutputEnabled {
        config.OutputFile.Write([]byte{byte(number.Value)})
    }
    return &Nil{}
}

func niler(args Value, scope *Scope) Value {
    return &Nil{}
}