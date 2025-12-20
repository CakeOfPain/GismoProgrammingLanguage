package interpreter

import (
	"gismolang.org/compiler/parser"
)
func Interpret(expressions *parser.SyntaxNode) {
    sexpressions := syntaxNode2Value(expressions)
    interpretModule(sexpressions)
}

func interpretModule(value Value) {
    scope := NewEmptyScope()

    if consCell, ok := value.(*ConsCell); ok {
        length := consCell.Length()
        for i:=1; i < length; i++ {
            arg := consCell.Get(i);
            interpretExpression(arg, scope);
        }
    }
}

func interpretExpression(value Value, scope *Scope) Value {
    switch v := value.(type) {
    case *ConsCell:
        operator := v.Get(0).String()
        switch operator {
        case "@callCurly":
             // ... [Logic remains unchanged] ...
             var result Value = &Nil{}
             arglen := v.Length()
             for i:=2; i < arglen; i++ {
                 result = &ConsCell{Car: v.Get(2+arglen-i-1), Cdr: result}
             }
             v = &ConsCell{
                 Car: v.Get(0),
                 Cdr: &ConsCell{
                     Car: v.Get(1),
                     Cdr: &ConsCell{
                         Car: &ConsCell{Car: &Symbol{Value: "@begin"}, Cdr: result},
                         Cdr: &Nil{},
                     },
                 },
             }
        case "@call":
            function := interpretExpression(v.Get(1), scope)
            arguments := v.Get(2)
            if builtinFunction, ok := function.(BuiltinFunction); ok {
                return builtinFunction.callback(arguments, scope)
            }
        case "::=":
            scope.Define(v.Get(1), v.Get(2))
            return &Nil{}
        case "@begin", "Module":
            // ... [Logic remains unchanged] ...
            var result Value = &Nil{}
            scope.allowExports = true
            newScope := NewScope(scope)
            arglen := v.Length()
            for i:=1; i < arglen; i++ {
                result = interpretExpression(v.Get(i), newScope)
            }
            scope.allowExports = false
            return result
        }
        
        // Try to resolve definition or macro
        result := scope.Get(v)
        if result == nil {
            // IMPROVED ERROR HANDLING
            RuntimeError(v.GetToken(), "Could not interpret expression: %s", v.String())
        }
		
        return result
    case *Symbol:
        result := scope.Get(v)
        if result == nil {
            return v
        }
        return result
    }

    return value
}