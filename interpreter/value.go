package interpreter

import (
	"strconv"

	"gismolang.org/compiler/tokenizer"
)

// NEW: Add GetToken to interface
type Value interface {
	GetTypeString() string
	String() string
	GetToken() *tokenizer.Token
}

// Base implementation helper
type BaseValue struct {
    Token *tokenizer.Token
}

func (b BaseValue) GetToken() *tokenizer.Token {
    return b.Token
}

// Update all structs to embed BaseValue or implement GetToken
// -----------------------------------------------------------

type Integer struct {
	BaseValue
	Value int64
}
func (integer Integer) GetTypeString() string { return "int" }
func (integer Integer) String() string        { return strconv.FormatInt(integer.Value, 10) }

type Float struct {
	BaseValue
	Value float64
}
func (f Float) GetTypeString() string { return "float" }
func (f Float) String() string        { return strconv.FormatFloat(f.Value, 'f', -1, 64) }

type String struct {
	BaseValue
	Value string
}
func (str String) String() string        { return str.Value }
func (str *String) GetTypeString() string { return "string" }

type Symbol struct {
	BaseValue
	Value string
}
func (sym Symbol) String() string        { return string(sym.Value) }
func (sym *Symbol) GetTypeString() string { return "symbol" }

type TypedValue struct {
	BaseValue
	Value         Value
	TypeValue     Value
	TypeFallbacks []Value
}
func (typeValue TypedValue) String() string {
	return "<" + typeValue.TypeValue.String() + " " + typeValue.Value.String() + ">"
}
func (typedValue *TypedValue) GetTypeString() string { return typedValue.TypeValue.String() }

type BuiltinFunction struct {
	BaseValue
	callback   func(value Value, scope *Scope) Value
	identifier string
}
func (fn BuiltinFunction) GetTypeString() string { return "builtin" }
func (builtinFunction BuiltinFunction) String() string {
	return "@builtin:" + builtinFunction.identifier
}

type ConsCell struct {
	BaseValue
	Car Value
	Cdr Value
}
func (consCell *ConsCell) String() string {
	if consCell == nil { return "nil" }
	return "(" + consCell.stringElements() + ")"
}
func (consCell *ConsCell) stringElements() string {
	if consCell == nil { return "" }
	var result string
	if consCell.Car != nil { result += consCell.Car.String() } else { result += "nil" }
	switch cdr := consCell.Cdr.(type) {
	case *ConsCell: result += " " + cdr.stringElements()
	case *Nil: // End of list
	default:
		if cdr != nil { result += " . " + cdr.String() } else { result += " . nil" }
	}
	return result
}
func (fn ConsCell) GetTypeString() string { return "ConsCell" }
func (consCell *ConsCell) Length() int {
	if consCell, ok := consCell.Cdr.(*ConsCell); ok { return consCell.Length() + 1 }
	return 1
}
func (consCell *ConsCell) Get(index int) Value {
	if index < 1 { return consCell.Car }
	if consCell, ok := consCell.Cdr.(*ConsCell); ok { return consCell.Get(index - 1) }
	return &Nil{}
}

type Nil struct {
	BaseValue
}
func (_ *Nil) String() string        { return "nil" }
func (fn Nil) GetTypeString() string { return "Nil" }

type Vector struct {
	BaseValue
	Elements []Value
}
func (v *Vector) GetTypeString() string { return "Vector" }
func (v *Vector) String() string {
	str := "["
	for i, el := range v.Elements {
		str += el.String()
		if i < len(v.Elements)-1 { str += ", " }
	}
	str += "]"
	return str
}
func (v *Vector) Length() int { return len(v.Elements) }

type Union struct {
	BaseValue
	Values []Value
}
func (u *Union) GetTypeString() string { return "Union" }
func (u *Union) String() string {
	str := "<Union"
	for _, val := range u.Values { str += " " + val.String() }
	str += ">"
	return str
}