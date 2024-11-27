package interpreter


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