package interpreter

import "strconv"

type Value interface {
    GetTypeString() string
    String() string
}

type Integer struct {
    Value int64
}

func (integer Integer) GetTypeString() string {
    return "int"
}

func (integer Integer) String() string {
    return strconv.FormatInt(integer.Value, 10)
}

type String struct {
    Value string
}

func (str String) String() string {
    return str.Value
}

func (str *String) GetTypeString() string {
    return "string"
}

type Symbol struct {
    Value string
}

func (sym Symbol) String() string {
    return string(sym.Value)
}

func (sym *Symbol) GetTypeString() string {
    return "symbol"
}

type TypedValue struct {
    Value Value
    TypeValue Value
}

func (typeValue TypedValue) String() string {
    return "<" + typeValue.TypeValue.String() + " " + typeValue.Value.String() + ">"
}

func (typedValue *TypedValue) GetTypeString() string {
    return typedValue.TypeValue.String()
}

type BuiltinFunction struct {
    callback func(value Value, scope *Scope) Value
    identifier string
}

func (fn BuiltinFunction) GetTypeString() string {
    return "builtin"
}

func (builtinFunction BuiltinFunction) String() string {
    return "@buildin:" + builtinFunction.identifier
}

type ConsCell struct {
    Car Value
    Cdr Value
}

func (consCell *ConsCell) String() string {
    if consCell == nil {
        return "nil"
    }
    return "(" + consCell.stringElements() + ")"
}

func (consCell *ConsCell) stringElements() string {
    if consCell == nil {
        return ""
    }

    var result string
    if consCell.Car != nil {
        result += consCell.Car.String()
    } else {
        result += "nil"
    }

    switch cdr := consCell.Cdr.(type) {
    case *ConsCell:
        result += " " + cdr.stringElements()
    case *Nil:
        // Proper list ends here; do nothing
    default:
        // Improper list: print the cdr after a dot
        if cdr != nil {
            result += " . " + cdr.String()
        } else {
            result += " . nil"
        }
    }
    return result
}

func (fn ConsCell) GetTypeString() string {
    return "ConsCell"
}

type Nil struct {
}

func (_ *Nil) String() string {
    return "nil"
}

func (fn Nil) GetTypeString() string {
    return "Nil"
}

func (consCell *ConsCell) Length() int {
    if consCell, ok := consCell.Cdr.(*ConsCell); ok {
        return consCell.Length() + 1
    }
    return 1
}

func (consCell *ConsCell) Get(index int) Value {
    if index < 1 {
        return consCell.Car
    }
    if consCell, ok := consCell.Cdr.(*ConsCell); ok {
        return consCell.Get(index-1)
    }
    return &Nil{}
}