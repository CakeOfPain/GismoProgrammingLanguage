package interpreter

import (
	"fmt"
	"strings"
)

type Definition struct {
	key string
	value Value
}

type Scope struct {
	previous *Scope
	definitions map[string] *Definition
	locals map[string] Value
	AcceptsExports bool
}

func NewScope(previous *Scope) *Scope {
	return &Scope{
		previous: previous,
		definitions: make(map[string]*Definition),
		locals: make(map[string]Value),
		AcceptsExports: false,
	}
}

func NewEmptyScope() *Scope {
	scope := NewScope(nil)

    for _, builtin := range Builtins() {
        scope.Define(&Symbol{
            Value: builtin.identifier,
        }, builtin)
    }

	return scope
}

func (scope *Scope) ExportDefinition(key Value, value Value) {
	if scope.AcceptsExports {
		scope.Define(key, value)
	} else if scope.previous != nil {
		scope.previous.ExportDefinition(key, value)
	}
}

func (scope *Scope) DefineLocal(key Value, value Value) {
	scope.locals[key.String()] = value
}

func (scope *Scope) SetLocal(key Value, value Value) {
	if _, ok := scope.locals[key.String()]; ok || scope.GetLocal(key).GetTypeString() == "Nil" {
		scope.locals[key.String()] = value
		return
	}
	scope.previous.SetLocal(key, value)
}

func (scope *Scope) GetLocal(key Value) Value {
	if found, ok := scope.locals[key.String()]; ok {
		return found
	}
	if scope.previous == nil {
		return &Nil{}
	}
	return scope.previous.GetLocal(key)
}

func (scope *Scope) Get(value Value, topScope *Scope) Value {
	switch v := value.(type) {
	case *ConsCell:
		key := v.Car.String()
		leftValue := interpretExpression(v.Get(1), scope)
		queryAny := key + " " + leftValue.GetTypeString() + " *"

		var currentScope *Scope = scope
		for currentScope != nil {
			if found, ok := currentScope.definitions[queryAny]; ok {
				return processMacro(found.value, scope, leftValue, v.Get(2))
			}
			currentScope = currentScope.previous
		}

		rightValue := v.Get(2)
		query := key + " " + leftValue.GetTypeString()
		if rightValue.GetTypeString() != "Nil" {
			query += " " + interpretExpression(rightValue, scope).GetTypeString()
		}

		currentScope = scope
		for currentScope != nil {
			if found, ok := currentScope.definitions[query]; ok {
				return processMacro(found.value, scope, leftValue, rightValue)
			}
			currentScope = currentScope.previous
		}

		fmt.Printf("ERROR: Implementation for %s or %s missing!\n", query, queryAny)

		break
	case *Symbol:
		var currentScope *Scope = scope
		for currentScope != nil {
			if found, ok := currentScope.definitions[v.Value]; ok {
				return found.value
			}
			currentScope = currentScope.previous
		}
	}
	return nil
}

func (scope *Scope) Define(key Value, value Value) {
	query := generateKey(key)
	if !strings.Contains(query, " ") {
		value = interpretExpression(value, scope)
	}
	scope.definitions[query] = &Definition{
		key: query,
		value: value,
	}
}

func processMacro(value Value, scope *Scope, left Value, right Value) Value {
	if left == nil && right == nil {
		return value
	}
	value = subSymbol(value, &Symbol{Value: "$1"}, left, true)
	value = subSymbol(value, &Symbol{Value: "$2"}, right, true)
	return interpretExpression(value, scope)
}

func generateKey(value Value) string {
	switch v := value.(type) {
	case *ConsCell:
		key := v.Car.String()
		arglen := v.Length()-1
		for i:=0; i < arglen; i++ {
			arg := v.Get(i+1)
			key += " " + arg.String()
		}
		return key
	case *Symbol:
		return v.Value
	}
	return "@UNDEFINED"
}
func (scope *Scope) String() string {
    var sb strings.Builder
    visited := make(map[*Scope]bool)
    scope.stringify(&sb, 0, visited)
    return sb.String()
}

func (scope *Scope) stringify(sb *strings.Builder, level int, visited map[*Scope]bool) {
    if scope == nil || visited[scope] {
        return
    }
    visited[scope] = true

    // ANSI-Farbcodes
    colorReset := "\033[0m"
    colorScopeLevel := "\033[1;34m" // Fett Blau
    colorKey := "\033[1;32m"        // Fett Grün
    colorValue := "\033[1;33m"      // Fett Gelb
    indent := strings.Repeat("  ", level)

    sb.WriteString(fmt.Sprintf("%s%sScope Level %d:%s\n", indent, colorScopeLevel, level, colorReset))

    if len(scope.definitions) == 0 {
        sb.WriteString(fmt.Sprintf("%s  (keine Definitionen)\n", indent))
    } else {
        for key, def := range scope.definitions {
            sb.WriteString(fmt.Sprintf("%s  %s%s%s: %s%s%s\n",
                indent,
                colorKey, key, colorReset,
                colorValue, def.value.String(), colorReset))
        }
    }

    if scope.previous != nil {
        sb.WriteString(fmt.Sprintf("%s%sVorheriger Scope:%s\n", indent, colorScopeLevel, colorReset))
        scope.previous.stringify(sb, level+1, visited)
    }
}