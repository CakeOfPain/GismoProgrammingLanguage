package interpreter

import (
	"fmt"
	"strings"
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

// The crucial part is in the ConsCell branch below. We:
//   1. Interpret the left argument (thus we know its true type).
//   2. Try a wildcard macro: "macroName leftType *".
//      - If found, do NOT interpret the right side and immediately process the macro.
//   3. If no wildcard macro is found, interpret the right argument and try normal type-specific macros.
//   4. If none matches, print an error and return nil.
func (currentScope *Scope) Get(lookupValue Value) Value {
	switch typedValue := lookupValue.(type) {

	case *ConsCell:
		macroName := typedValue.Car.String()
		leftVal := interpretExpression(typedValue.Get(1), currentScope)
		leftTypes := gatherTypeStrings(leftVal)

		// Check for wildcard macro first (right side NOT interpreted).
		for _, leftType := range leftTypes {
			keyWildcard := macroName + " " + leftType + " *"
			if defAny := currentScope.findDefinition(keyWildcard); defAny != nil {
				return processMacro(defAny.definitionValue, currentScope, leftVal, typedValue.Get(2))
			}
		}

		// If no wildcard macro was found, interpret the right side
		rightVal := interpretExpression(typedValue.Get(2), currentScope)
		rightTypes := gatherTypeStrings(rightVal)

		// Check normal macros with both types known
		for _, leftType := range leftTypes {
			for _, rightType := range rightTypes {
				keyFull := macroName + " " + leftType + " " + rightType
				if defFull := currentScope.findDefinition(keyFull); defFull != nil {
					return processMacro(defFull.definitionValue, currentScope, leftVal, rightVal)
				}
			}
		}

		fmt.Printf(
			"ERROR: No match for macro '%s' with left '%s' right '%s'\n",
			macroName,
			leftVal.String(),
			rightVal.String(),
		)
		return nil

	case *Symbol:
		return currentScope.lookupSymbol(typedValue.Value)

	default:
		return nil
	}
}

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
		for defKeyStr, defPointer := range currentScope.definitionsMap {
			builder.WriteString(fmt.Sprintf(
				"%s  %s%s%s: %s%s%s\n",
				indent,
				ansiGreen, defKeyStr, ansiReset,
				ansiYellow, defPointer.definitionValue.String(), ansiReset,
			))
		}
	}

	if currentScope.parentScope != nil {
		builder.WriteString(fmt.Sprintf("%s%sparent scope:%s\n", indent, ansiBlue, ansiReset))
		currentScope.parentScope.stringify(builder, level+1, visited)
	}
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
	return []string{val.GetTypeString()}
}

func processMacro(macroValue Value, currentScope *Scope, left Value, right Value) Value {
	if left == nil && right == nil {
		return macroValue
	}
	macroValue = subSymbol(macroValue, &Symbol{Value: "$1"}, left, true)
	macroValue = subSymbol(macroValue, &Symbol{Value: "$2"}, right, true)
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