package interpreter

import (
	"fmt"
	"os"
	"path/filepath"

	"gismolang.org/compiler/config"
	"gismolang.org/compiler/parser"
	"gismolang.org/compiler/tokenizer"
)

var fileLoadCache = make(map[string]Value)

func Builtins() []BuiltinFunction {
    return []BuiltinFunction{
        // Arithmetic
        {callback: addInt, identifier: "$ADD"},
        {callback: subInt, identifier: "$SUB"},
        {callback: mulInt, identifier: "$MUL"},
        {callback: divInt, identifier: "$DIV"},
        {callback: modInt, identifier: "$MOD"},
        {callback: convInt, identifier: "$INT"},

        // Bitwise
        {callback: bitwiseAnd, identifier: "$BWA"},
        {callback: bitwiseOr, identifier: "$BWO"},
        {callback: shiftLeftInt, identifier: "$SHL"},
        {callback: shiftRightInt, identifier: "$SHR"},

        // List/Cons
        {callback: car, identifier: "$CAR"},
        {callback: cdr, identifier: "$CDR"},
        {callback: construct, identifier: "$CONS"},

        // String
        {callback: catString, identifier: "$CAT"},
        {callback: charString, identifier: "$CHAR"},
        {callback: lenString, identifier: "$STRLEN"},
        {callback: stringify, identifier: "$STR"},

        // Vector
        {callback: vectorCreate, identifier: "$VECTOR"},
        {callback: vectorGet, identifier: "$VECTOR_GET"},
        {callback: vectorSet, identifier: "$VECTOR_SET"},
        {callback: vectorLen, identifier: "$VECTOR_LEN"},
        {callback: vectorResize, identifier: "$VECTOR_RESIZE"},

        // Comparison
        {callback: equals, identifier: "$EQUALS"},
        {callback: greater, identifier: "$GREATER"},

        // Type
        {callback: typedef, identifier: "$TYPEDEF"},
        {callback: typeof, identifier: "$TYPEOF"},
        {callback: untype, identifier: "$UNTYPE"},
        {callback: unionizer, identifier: "$UNION"},

        // Control Flow
        {callback: ifFunc, identifier: "$IF"},
        {callback: whiler, identifier: "$WHILE"},
        {callback: foreacher, identifier: "$FOREACH"},

        // Scope & Variables
        {callback: def, identifier: "$DEF"},
        {callback: get, identifier: "$GET"},
        {callback: set, identifier: "$SET"},

        // I/O
        {callback: printValue, identifier: "$PRINT"},
        {callback: printlnValue, identifier: "$PRINTLN"},
        {callback: write2Output, identifier: "$WRITE"},
        {callback: writeByte2Output, identifier: "$WRITEB"},

        // Meta & Evaluation
        {callback: quote, identifier: "$QUOTE"},
        {callback: eval, identifier: "$EVAL"},
        {callback: lambda, identifier: "$LAMBDA"},
        {callback: replace, identifier: "$REPLACE"},

        // Misc
        {callback: flatter, identifier: "$FLATTEN"},
        {callback: raiser, identifier: "$RAISE"},
        {callback: niler, identifier: "$NIL"},
        {callback: iotainator, identifier: "$IOTA"},
        {callback: exporter, identifier: "$EXPORT"},
        {callback: loadFile, identifier: "$LOAD"},
        {callback: printScope, identifier: "$SCOPE"},
        {callback: catSym, identifier: "$SYMCAT"},
    }
}

// Helper: Handles Integers only (for Bitwise/Shift/Mod)
func binaryIntOp(args Value, scope *Scope, op func(a, b int64) int64) Value {
    argsList := getArgsList(args)
    if len(argsList) < 2 {
        return &Nil{}
    }
    left := interpretExpression(argsList[0], scope)
    right := interpretExpression(argsList[1], scope)

    if leftInt, ok := left.(*Integer); ok {
        if rightInt, ok := right.(*Integer); ok {
            return &Integer{Value: op(leftInt.Value, rightInt.Value)}
        }
    }
    return &Nil{}
}

// Helper: Handles Integers OR Floats (for Add/Sub/Mul/Div)
func binaryNumericOp(args Value, scope *Scope, intOp func(a, b int64) int64, floatOp func(a, b float64) float64) Value {
    argsList := getArgsList(args)
    if len(argsList) < 2 {
        return &Nil{}
    }
    left := interpretExpression(argsList[0], scope)
    right := interpretExpression(argsList[1], scope)

    // Type Promotion: If either side is Float, do Float math
    if _, ok := left.(*Float); ok {
        return performFloatOp(left, right, floatOp)
    }
    if _, ok := right.(*Float); ok {
        return performFloatOp(left, right, floatOp)
    }

    // Default: Integer math
    if leftInt, ok := left.(*Integer); ok {
        if rightInt, ok := right.(*Integer); ok {
            return &Integer{Value: intOp(leftInt.Value, rightInt.Value)}
        }
    }
    return &Nil{}
}

// Helper: Casts mixed numbers to float and executes operation
func performFloatOp(left, right Value, op func(a, b float64) float64) Value {
    var lVal, rVal float64

    // Cast Left
    switch v := left.(type) {
    case *Float:
        lVal = v.Value
    case *Integer:
        lVal = float64(v.Value)
    default:
        return &Nil{}
    }

    // Cast Right
    switch v := right.(type) {
    case *Float:
        rVal = v.Value
    case *Integer:
        rVal = float64(v.Value)
    default:
        return &Nil{}
    }

    return &Float{Value: op(lVal, rVal)}
}

func addInt(args Value, scope *Scope) Value {
    return binaryNumericOp(args, scope,
        func(a, b int64) int64 { return a + b },
        func(a, b float64) float64 { return a + b },
    )
}

func subInt(args Value, scope *Scope) Value {
    return binaryNumericOp(args, scope,
        func(a, b int64) int64 { return a - b },
        func(a, b float64) float64 { return a - b },
    )
}

func mulInt(args Value, scope *Scope) Value {
    return binaryNumericOp(args, scope,
        func(a, b int64) int64 { return a * b },
        func(a, b float64) float64 { return a * b },
    )
}

func divInt(args Value, scope *Scope) Value {
    return binaryNumericOp(args, scope,
        func(a, b int64) int64 { return a / b },
        func(a, b float64) float64 { return a / b },
    )
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
    if len(argsList) < 2 {
        return &Nil{}
    }
    left := interpretExpression(argsList[0], scope)
    right := interpretExpression(argsList[1], scope)
    return &ConsCell{
        Car:       left,
        Cdr:       right,
        BaseValue: BaseValue{Token: left.GetToken()},
    }
}

func car(args Value, scope *Scope) Value {
    argsList := getArgsList(args)
    if len(argsList) < 1 {
        return &Nil{}
    }
    value := interpretExpression(argsList[0], scope)
    if consCell, ok := value.(*ConsCell); ok {
        return consCell.Car
    }
    return &Nil{}
}

func cdr(args Value, scope *Scope) Value {
    argsList := getArgsList(args)
    if len(argsList) < 1 {
        return &Nil{}
    }
    value := interpretExpression(argsList[0], scope)
    if consCell, ok := value.(*ConsCell); ok {
        return consCell.Cdr
    }
    return &Nil{}
}

func printValue(args Value, scope *Scope) Value {
    argsList := getArgsList(args)
    if len(argsList) < 1 {
        return &Nil{}
    }
    value := interpretExpression(argsList[0], scope)
    fmt.Print(value)
    return &Nil{}
}

func printlnValue(args Value, scope *Scope) Value {
    argsList := getArgsList(args)
    if len(argsList) < 1 {
        fmt.Println()
        return &Nil{}
    }
    value := interpretExpression(argsList[0], scope)
    fmt.Println(value)
    return &Nil{}
}

func printScope(args Value, scope *Scope) Value {
    fmt.Println(scope)
    return &Nil{}
}

func typedef(args Value, scope *Scope) Value {
    argsList := getArgsList(args)
    if len(argsList) < 2 {
        return &Nil{}
    }
    left := interpretExpression(argsList[0], scope)
    right := argsList[1]
    return &TypedValue{
        Value:         left,
        TypeValue:     right,
        TypeFallbacks: argsList[2:],
    }
}

func typeof(args Value, scope *Scope) Value {
    argsList := getArgsList(args)
    if len(argsList) < 2 {
        return &Nil{}
    }
    left := interpretExpression(argsList[0], scope)
    right := argsList[1]
    if left.GetTypeString() == right.String() {
        return &Integer{
            Value: 1,
        }
    }

    if typedValue, ok := left.(*TypedValue); ok {
        for _, t := range typedValue.TypeFallbacks {
            if t.String() == right.String() {
                return &Integer{
                    Value: 1,
                }
            }
        }
    }

    return &Nil{}
}

func untype(args Value, scope *Scope) Value {
    argsList := getArgsList(args)
    if len(argsList) < 1 {
        return &Nil{}
    }
    value := interpretExpression(argsList[0], scope)
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
    if len(argsList) < 3 {
        return &Nil{}
    }
    expression := interpretExpression(argsList[0], scope)
    symbol := argsList[1]
    replacement := interpretExpression(argsList[2], scope)

    if sym, ok := symbol.(*Symbol); ok {
        return subSymbol(expression, sym, replacement, false)
    }

    return &Nil{}
}

func eval(args Value, scope *Scope) Value {
    argsList := getArgsList(args)
    if len(argsList) < 1 {
        return &Nil{}
    }
    return interpretExpression(interpretExpression(argsList[0], scope), scope)
}

func lambda(args Value, scope *Scope) Value {
    argsList := getArgsList(args)
    if len(argsList) < 2 {
        return &Nil{}
    }
    left := argsList[0]
    body := argsList[1]
    if sym, ok := left.(*Symbol); ok {
        return BuiltinFunction{
            callback: func(value Value, callerScope *Scope) Value {
                value = interpretExpression(value, callerScope)
                arg := &ConsCell{
                    Car: &Symbol{
                        Value: "@call",
                    },
                    Cdr: &ConsCell{
                        Car: &Symbol{
                            Value: "$QUOTE",
                        },
                        Cdr: &ConsCell{
                            Car: value,
                            Cdr: &Nil{},
                        },
                    },
                    BaseValue: BaseValue{Token: value.GetToken()},
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
    if len(argsList) < 2 {
        return &Nil{}
    }
    left := interpretExpression(argsList[0], scope)
    right := interpretExpression(argsList[1], scope)

    return &String{
        Value: left.String() + right.String(),
    }
}

func lenString(args Value, scope *Scope) Value {
    argsList := getArgsList(args)
    if len(argsList) < 1 {
        return &Integer{Value: 0}
    }
    value := interpretExpression(argsList[0], scope)
    return &Integer{
        Value: int64(len(value.String())),
    }
}

func charString(args Value, scope *Scope) Value {
    argsList := getArgsList(args)
    if len(argsList) < 2 {
        return &Nil{}
    }
    left := interpretExpression(argsList[0], scope)
    right := interpretExpression(argsList[1], scope)
    if index, ok := right.(*Integer); ok {
        s := left.String()
        if index.Value < 0 || int(index.Value) >= len(s) {
            return &Nil{}
        }
        return &Integer{
            Value: int64([]byte(s)[int(index.Value)]),
        }
    }
    return &Nil{}
}

func stringify(args Value, scope *Scope) Value {
    argsList := getArgsList(args)
    if len(argsList) < 1 {
        return &Nil{}
    }
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
    if len(argsList) < 2 {
        return &Nil{}
    }
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
    if len(argsList) < 2 {
        return &Nil{}
    }
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
    if len(argsList) < 3 {
        return &Nil{}
    }
    cond := argsList[0]
    interpretCond := interpretExpression(cond, scope)
    if interpretCond.GetTypeString() == "Nil" {
        return interpretExpression(argsList[2], scope)
    }

    return interpretExpression(argsList[1], scope)
}

func get(args Value, scope *Scope) Value {
    argsList := getArgsList(args)
    if len(argsList) < 1 {
        return &Nil{}
    }
    index := argsList[0]
    return scope.GetLocal(index)
}

func set(args Value, scope *Scope) Value {
    argsList := getArgsList(args)
    if len(argsList) < 2 {
        return &Nil{}
    }
    index := argsList[0]
    value := interpretExpression(argsList[1], scope)
    scope.SetLocal(index, value)
    return &Nil{}
}

func def(args Value, scope *Scope) Value {
    argsList := getArgsList(args)
    if len(argsList) < 2 {
        return &Nil{}
    }
    index := argsList[0]
    value := interpretExpression(argsList[1], scope)
    scope.DefineLocal(index, value)
    return &Nil{}
}

func write2Output(args Value, scope *Scope) Value {
    argsList := getArgsList(args)
    if len(argsList) < 1 {
        return &Nil{}
    }
    message := interpretExpression(argsList[0], scope).String()
    if config.OutputEnabled && config.OutputFile != nil {
        config.OutputFile.WriteString(message)
    }
    return &Nil{}
}

func writeByte2Output(args Value, scope *Scope) Value {
    argsList := getArgsList(args)
    if len(argsList) < 1 {
        return &Nil{}
    }
    result := interpretExpression(argsList[0], scope)
    if number, ok := result.(*Integer); ok && config.OutputEnabled && config.OutputFile != nil {
        config.OutputFile.Write([]byte{byte(number.Value)})
    }
    return &Nil{}
}

func niler(args Value, scope *Scope) Value {
    return &Nil{}
}
func loadFile(args Value, scope *Scope) Value {
    argsList := getArgsList(args)
    if len(argsList) < 1 {
        return &Nil{}
    }

    rawPath := interpretExpression(argsList[0], scope).String()

    absPath, err := filepath.Abs(rawPath)
    if err != nil {
        return &Nil{}
    }

    canonicalPath, err := filepath.EvalSymlinks(absPath)
    if err != nil {
        return &Nil{}
    }

    var programValue Value

    if cached, found := fileLoadCache[canonicalPath]; found {
        programValue = cached
    } else {
        bytes, err := os.ReadFile(canonicalPath)
        if err != nil {
            return &Nil{}
        }

        tokens := tokenizer.Tokenize(string(bytes), canonicalPath)
        ast := parser.Parse(tokens, canonicalPath)
        programValue = syntaxNode2Value(ast)

        fileLoadCache[canonicalPath] = programValue
    }

    if consCell, ok := programValue.(*ConsCell); ok {
        for i := 1; i < consCell.Length(); i++ {
            interpretExpression(consCell.Get(i), scope)
        }
    } else {
        interpretExpression(programValue, scope)
    }

    return &Nil{}
}

func exporter(args Value, scope *Scope) Value {
    argsList := getArgsList(args)
    if len(argsList) < 2 {
        return &Nil{}
    }
    key := interpretExpression(argsList[0], scope)
    // key := argsList[0]
    value := interpretExpression(argsList[1], scope)
    scope.ExportDefinition(key, value)
    return &Nil{}
}

func whiler(args Value, scope *Scope) Value {
    argsList := getArgsList(args)
    if len(argsList) < 2 {
        return &Nil{}
    }
    cond := argsList[0]
    body := argsList[1]
    for interpretExpression(cond, scope).GetTypeString() != "Nil" {
        interpretExpression(body, scope)
    }
    return &Nil{}
}

func flatter(args Value, scope *Scope) Value {
    argsList := getArgsList(args)
    if len(argsList) < 2 {
        return &Nil{}
    }
    value := interpretExpression(argsList[0], scope)
    seperator := interpretExpression(argsList[1], scope)
    return &Vector{
        Elements: flattenBySeparator(value, seperator.String()),
    }
}

func vectorCreate(args Value, scope *Scope) Value {
    argsList := getArgsList(args)
    if len(argsList) < 1 {
        return &Nil{}
    }
    sizeVal := interpretExpression(argsList[0], scope)
    if sizeInt, ok := sizeVal.(*Integer); ok {
        if sizeInt.Value < 0 {
            return &Nil{}
        }
        elements := make([]Value, int(sizeInt.Value))
        for i := range elements {
            elements[i] = &Nil{}
        }
        return &Vector{
            Elements: elements,
        }
    }
    return &Nil{}
}

func vectorGet(args Value, scope *Scope) Value {
    argsList := getArgsList(args)
    if len(argsList) < 2 {
        return &Nil{}
    }
    vecVal := interpretExpression(argsList[0], scope)
    idxVal := interpretExpression(argsList[1], scope)
    if vec, ok := vecVal.(*Vector); ok {
        if idx, ok := idxVal.(*Integer); ok {
            if idx.Value >= 0 && idx.Value < int64(len(vec.Elements)) {
                return vec.Elements[int(idx.Value)]
            }
        }
    }
    return &Nil{}
}

func vectorSet(args Value, scope *Scope) Value {
    argsList := getArgsList(args)
    if len(argsList) < 3 {
        return &Nil{}
    }
    vecVal := interpretExpression(argsList[0], scope)
    idxVal := interpretExpression(argsList[1], scope)
    valueVal := interpretExpression(argsList[2], scope)
    if vec, ok := vecVal.(*Vector); ok {
        if idx, ok := idxVal.(*Integer); ok {
            if idx.Value >= 0 && idx.Value < int64(len(vec.Elements)) {
                vec.Elements[int(idx.Value)] = valueVal
                return valueVal
            }
        }
    }
    return &Nil{}
}

func vectorLen(args Value, scope *Scope) Value {
    argsList := getArgsList(args)
    if len(argsList) < 1 {
        return &Nil{}
    }
    vecVal := interpretExpression(argsList[0], scope)
    if vec, ok := vecVal.(*Vector); ok {
        return &Integer{Value: int64(len(vec.Elements))}
    }
    return &Nil{}
}

func vectorResize(args Value, scope *Scope) Value {
    argsList := getArgsList(args)
    if len(argsList) < 2 {
        return &Nil{}
    }
    vecVal := interpretExpression(argsList[0], scope)
    newSizeVal := interpretExpression(argsList[1], scope)
    if vec, ok := vecVal.(*Vector); ok {
        if newSize, ok := newSizeVal.(*Integer); ok {
            if newSize.Value < 0 {
                return &Nil{}
            }
            oldLenInt := len(vec.Elements)
            newSizeInt := int(newSize.Value)
            if newSizeInt == oldLenInt {
                return vec
            }
            if newSizeInt > oldLenInt {
                extension := make([]Value, newSizeInt-oldLenInt)
                for i := range extension {
                    extension[i] = &Nil{}
                }
                vec.Elements = append(vec.Elements, extension...)
            } else {
                vec.Elements = vec.Elements[:newSizeInt]
            }
            return vec
        }
    }
    return &Nil{}
}

func isolator(value Value, scope *Scope) Value {
    isolatedScope := NewEmptyScope()
    return interpretExpression(value, isolatedScope)
}

func raiser(args Value, scope *Scope) Value {
    argsList := getArgsList(args)
    if len(argsList) < 1 {
        return &Nil{}
    }

    // 1. The marked value
    targetValue := argsList[0]

    // 2. The Error Message
    message := "An error occurred"
    if len(argsList) > 1 {
        message = interpretExpression(argsList[1], scope).String()
    }

    RuntimeError(targetValue.GetToken(), message)

    return &Nil{}
}

func unionizer(args Value, scope *Scope) Value {
    argsList := getArgsList(args)
    values := make([]Value, len(argsList))

    for i, arg := range argsList {
        values[i] = interpretExpression(arg, scope)
    }

    return &Union{
        Values: values,
    }
}

func foreacher(args Value, scope *Scope) Value {
    argsList := getArgsList(args)
    if len(argsList) < 3 {
        return &Nil{}
    }
    collectionVal := interpretExpression(argsList[0], scope)
    variableName := interpretExpression(argsList[1], scope)
    body := argsList[2]

    if varSym, ok := variableName.(*Symbol); ok {
        switch collection := collectionVal.(type) {
        case *Vector:
            for _, element := range collection.Elements {
                newBody := subSymbol(body, varSym, element, true)
                interpretExpression(newBody, scope)
            }
        case *ConsCell:
            for i := 0; i < collection.Length(); i++ {
                element := collection.Get(i)
                newBody := subSymbol(body, varSym, element, true)
                interpretExpression(newBody, scope)
            }
        case *String:
            for _, b := range []byte(collection.Value) {
                element := &Integer{Value: int64(b), BaseValue: BaseValue{Token: collection.GetToken()}}
                newBody := subSymbol(body, varSym, element, true)
                interpretExpression(newBody, scope)
            }
        }
    }
    return &Nil{}
}

func iotainator(args Value, scope *Scope) Value {
    currentValue := config.IotaValue
    config.IotaValue++
    return &Integer{
        Value: int64(currentValue),
    }
}

func convInt(args Value, scope *Scope) Value {
    argsList := getArgsList(args)
    if len(argsList) < 1 {
        return &Nil{}
    }
    value := interpretExpression(argsList[0], scope)

    switch v := value.(type) {
    case *Integer:
        return v
    case *Float:
        return &Integer{
            Value: int64(v.Value),
        }
    case *String:
        var intValue int64
        _, err := fmt.Sscanf(v.Value, "%d", &intValue)
        if err == nil {
            return &Integer{
                Value: intValue,
            }
        }
    }

    return &Nil{}
}

func catSym(args Value, scope *Scope) Value {
    argsList := getArgsList(args)
    if len(argsList) < 2 {
        return &Nil{}
    }
    left := interpretExpression(argsList[0], scope)
    right := interpretExpression(argsList[1], scope)

    new_symbol := left.String() + right.String()

    return &Symbol{
        Value: new_symbol,
        BaseValue: BaseValue{
            Token: left.GetToken(),
        },
    }
}