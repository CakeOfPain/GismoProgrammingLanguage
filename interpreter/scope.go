package interpreter

import (
	"fmt"
	"strings"

	"gismolang.org/compiler/tokenizer"
)

type Definition struct {
	definitionName  string
	definitionValue Value
}

type Scope struct {
	parentScope    *Scope
	definitionsMap map[string]*Definition
	localBindings  map[string]Value
	allowExports   bool
}

func NewScope(parentScope *Scope) *Scope {
	return &Scope{
		parentScope:    parentScope,
		definitionsMap: make(map[string]*Definition),
		localBindings:  make(map[string]Value),
		allowExports:   false,
	}
}

func NewEmptyScope() *Scope {
	newScope := NewScope(nil)
	for _, builtinSymbol := range Builtins() {
		newScope.Define(&Symbol{Value: builtinSymbol.identifier}, builtinSymbol)
	}
	return newScope
}

func (currentScope *Scope) ExportDefinition(defKey Value, defValue Value) {
	if currentScope == nil {
		return
	}
	if currentScope.allowExports {
		currentScope.Define(defKey, defValue)
	} else if currentScope.parentScope != nil {
		currentScope.parentScope.ExportDefinition(defKey, defValue)
	}
}

func (currentScope *Scope) DefineLocal(localKey Value, localValue Value) {
	currentScope.localBindings[localKey.String()] = localValue
}

func (currentScope *Scope) SetLocal(localKey Value, localValue Value) {
	if currentScope == nil {
		panic("SetLocal called with no scope")
	}
	localKeyString := localKey.String()
	if _, found := currentScope.localBindings[localKeyString]; found ||
		currentScope.GetLocal(localKey).GetTypeString() == "Nil" {
		currentScope.localBindings[localKeyString] = localValue
		return
	}
	if currentScope.parentScope != nil {
		currentScope.parentScope.SetLocal(localKey, localValue)
		return
	}
	panic(fmt.Sprintf("No existing local for '%s'", localKeyString))
}

func (currentScope *Scope) GetLocal(localKey Value) Value {
	if currentScope == nil {
		return &Nil{}
	}
	localKeyString := localKey.String()
	if foundValue, ok := currentScope.localBindings[localKeyString]; ok {
		return foundValue
	}
	return currentScope.parentScope.GetLocal(localKey)
}

func (currentScope *Scope) Get(lookupValue Value) Value {
	switch typedValue := lookupValue.(type) {
	case *ConsCell:
		macroName := typedValue.Car.String()
		operatorToken := typedValue.Car.GetToken()

		if typedValue.Length() == 2 {
			return currentScope.applyUnaryMacro(macroName, typedValue.Get(1), operatorToken)
		} else if typedValue.Length() >= 3 {
			return currentScope.applyBinaryMacro(macroName, typedValue.Get(1), typedValue.Get(2), operatorToken)
		}
		return nil

	case *Symbol:
		return currentScope.lookupSymbol(typedValue.Value)

	default:
		return nil
	}
}


// Define sets or updates a definition in this scope.
// If the definition key is a single symbol, interpret its value immediately.
func (currentScope *Scope) Define(defKey Value, defValue Value) {
	definitionLookupKey := generateKey(defKey)
	if !strings.Contains(definitionLookupKey, " ") {
		defValue = interpretExpression(defValue, currentScope)
	}
	currentScope.definitionsMap[definitionLookupKey] = &Definition{
		definitionName:  definitionLookupKey,
		definitionValue: defValue,
	}
}

func (currentScope *Scope) String() string {
	var builder strings.Builder
	visited := make(map[*Scope]bool)
	currentScope.stringify(&builder, 0, visited)
	return builder.String()
}

func (currentScope *Scope) stringify(builder *strings.Builder, level int, visited map[*Scope]bool) {
	if currentScope == nil || visited[currentScope] {
		return
	}
	visited[currentScope] = true

	const (
		ansiReset  = "\033[0m"
		ansiBlue   = "\033[1;34m"
		ansiGreen  = "\033[1;32m"
		ansiYellow = "\033[1;33m"
	)

	indent := strings.Repeat("  ", level)
	builder.WriteString(fmt.Sprintf("%s%sscope level %d:%s\n", indent, ansiBlue, level, ansiReset))

	if len(currentScope.definitionsMap) == 0 {
		builder.WriteString(fmt.Sprintf("%s  (no definitions)\n", indent))
	} else {
		for defKeyStr, defPtr := range currentScope.definitionsMap {
			builder.WriteString(fmt.Sprintf(
				"%s  %s%s%s: %s%s%s\n",
				indent,
				ansiGreen, defKeyStr, ansiReset,
				ansiYellow, defPtr.definitionValue.String(), ansiReset,
			))
		}
	}
	if currentScope.parentScope != nil {
		builder.WriteString(fmt.Sprintf("%s%sparent scope:%s\n", indent, ansiBlue, ansiReset))
		currentScope.parentScope.stringify(builder, level+1, visited)
	}
}

func (currentScope *Scope) applyUnaryMacro(macroName string, rawLeft Value, operatorToken *tokenizer.Token) Value {
    leftVal := interpretExpression(rawLeft, currentScope)
    leftTypes := gatherTypeStrings(leftVal)

    for _, leftType := range leftTypes {
        defKey := macroName + " " + leftType
        if foundDef := currentScope.findDefinition(defKey); foundDef != nil {
            resolvedLeft := resolveValueForType(leftVal, leftType)
            
            // Construct $$ -> (OP LEFT)
            wholeExpr := &ConsCell{
                Car: &Symbol{Value: macroName, BaseValue: BaseValue{Token: operatorToken}},
                Cdr: &ConsCell{
                    Car: resolvedLeft,
                    Cdr: &Nil{},
                },
                BaseValue: BaseValue{Token: operatorToken},
            }

            return processMacro(foundDef.definitionValue, currentScope, resolvedLeft, &Nil{}, wholeExpr)
        }
    }
    
    RuntimeError(operatorToken, "No match for unary macro '%s' with type '%s'", macroName, leftVal.GetTypeString())
    return nil
}


func (currentScope *Scope) applyBinaryMacro(macroName string, rawLeft Value, rawRight Value, operatorToken *tokenizer.Token) Value {
    leftVal := interpretExpression(rawLeft, currentScope)
    leftTypes := gatherTypeStrings(leftVal)

    // Check wildcard matches
    for _, leftType := range leftTypes {
        keyWildcard := macroName + " " + leftType + " *"
        if defAny := currentScope.findDefinition(keyWildcard); defAny != nil {
            resolvedLeft := resolveValueForType(leftVal, leftType)
            
            // Construct $$ -> (OP LEFT RIGHT)
            wholeExpr := &ConsCell{
                Car: &Symbol{Value: macroName, BaseValue: BaseValue{Token: operatorToken}},
                Cdr: &ConsCell{
                    Car: resolvedLeft,
                    Cdr: &ConsCell{
                        Car: rawRight, // Raw right for wildcard
                        Cdr: &Nil{},
                    },
                },
                BaseValue: BaseValue{Token: operatorToken},
            }

            return processMacro(defAny.definitionValue, currentScope, resolvedLeft, rawRight, wholeExpr)
        }
    }

    rightVal := interpretExpression(rawRight, currentScope)
    rightTypes := gatherTypeStrings(rightVal)

    // Check exact matches
    for _, leftType := range leftTypes {
        for _, rightType := range rightTypes {
            keyFull := macroName + " " + leftType + " " + rightType
            if defFull := currentScope.findDefinition(keyFull); defFull != nil {
                resolvedLeft := resolveValueForType(leftVal, leftType)
                resolvedRight := resolveValueForType(rightVal, rightType)
                
                // Construct $$ -> (OP LEFT RIGHT)
                wholeExpr := &ConsCell{
                    Car: &Symbol{Value: macroName, BaseValue: BaseValue{Token: operatorToken}},
                    Cdr: &ConsCell{
                        Car: resolvedLeft,
                        Cdr: &ConsCell{
                            Car: resolvedRight,
                            Cdr: &Nil{},
                        },
                    },
                    BaseValue: BaseValue{Token: operatorToken},
                }

                return processMacro(defFull.definitionValue, currentScope, resolvedLeft, resolvedRight, wholeExpr)
            }
        }
    }

    RuntimeError(
        operatorToken, 
        "No match for macro '%s' with left '%s' (%s) and right '%s' (%s)",
        macroName,
        leftVal.String(), leftVal.GetTypeString(),
        rightVal.String(), rightVal.GetTypeString(),
    )
    return nil
}

func (currentScope *Scope) lookupSymbol(symbolName string) Value {
	for searchScope := currentScope; searchScope != nil; searchScope = searchScope.parentScope {
		if foundDefinition, ok := searchScope.definitionsMap[symbolName]; ok {
			return foundDefinition.definitionValue
		}
	}
	return nil
}

func (currentScope *Scope) findDefinition(definitionKey string) *Definition {
	for searchScope := currentScope; searchScope != nil; searchScope = searchScope.parentScope {
		if foundDef, ok := searchScope.definitionsMap[definitionKey]; ok {
			return foundDef
		}
	}
	return nil
}

func gatherTypeStrings(val Value) []string {
    if tv, ok := val.(*TypedValue); ok {
        allTypes := []string{tv.TypeValue.String()}
        for _, fallback := range tv.TypeFallbacks {
            allTypes = append(allTypes, fallback.String())
        }
        return allTypes
    }

    if u, ok := val.(*Union); ok {
        allTypes := []string{"Union"}
        
        for _, v := range u.Values {
            allTypes = append(allTypes, gatherTypeStrings(v)...)
        }
        return allTypes
    }

    return []string{val.GetTypeString()}
}

func processMacro(macroValue Value, currentScope *Scope, left Value, right Value, wholeExpression Value) Value {
	if left == nil && right == nil {
		return macroValue
	}
	macroValue = subSymbol(macroValue, &Symbol{Value: "$1"}, left, true)
	macroValue = subSymbol(macroValue, &Symbol{Value: "$2"}, right, true)
    
    // NEW: Substitute $$
    macroValue = subSymbol(macroValue, &Symbol{Value: "$$"}, wholeExpression, true)

	return interpretExpression(macroValue, currentScope)
}

func generateKey(definitionValue Value) string {
	switch typedValue := definitionValue.(type) {
	case *ConsCell:
		keyString := typedValue.Car.String()
		for i := 1; i < typedValue.Length(); i++ {
			keyString += " " + typedValue.Get(i).String()
		}
		return keyString
	case *Symbol:
		return typedValue.Value
	default:
		return "@UNDEFINED"
	}
}

// Helper function to extract the specific value that triggered the type match
func resolveValueForType(val Value, targetType string) Value {
    // 1. Handle TypedValue: Check its declared Type and Fallbacks
    if tv, ok := val.(*TypedValue); ok {
        if tv.TypeValue.String() == targetType {
            return tv
        }
        for _, fallback := range tv.TypeFallbacks {
            if fallback.String() == targetType {
                return tv
            }
        }
        return nil
    }

    // 2. Handle Union: Recursively search for the matching content
    if u, ok := val.(*Union); ok {
        // If the macro specifically matched "Union", return the whole container
        if targetType == "Union" {
            return u
        }
        // Otherwise, find the inner value that matches the target type
        for _, inner := range u.Values {
            if found := resolveValueForType(inner, targetType); found != nil {
                return found
            }
        }
        return nil
    }

    // 3. Handle Standard Types (Integer, String, etc.)
    if val.GetTypeString() == targetType {
        return val
    }

    return nil
}
