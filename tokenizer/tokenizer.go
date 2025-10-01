package tokenizer

import (
	"fmt"
	"strings"
	"unicode"

	"gismolang.org/compiler/tokenizer/tokentype"
)

// Token represents a token parsed by the tokenizer.
type Token struct {
	TokenType tokentype.TokenType
	Source    string
	Pos       int
	Value     string
	Alias     string
	BinPrec   int
}

// NoneToken is a sentinel token used to represent the absence of a token.
var NoneToken = &Token{
	TokenType: tokentype.None,
	Source:    "Unknown",
	Pos:       0,
	Value:     "",
	Alias:     "",
	BinPrec:   0,
}

// ModuleToken creates a module token with the given source.
func ModuleToken(source string) *Token {
	return &Token{
		TokenType: tokentype.Module,
		Source:    source,
		Pos:       0,
		Value:     "Module",
		Alias:     "Module",
		BinPrec:   0,
	}
}

var oneCharacterTokens = map[rune]tokentype.TokenType{
	'{': tokentype.LCurlyParent,
	'}': tokentype.RCurlyParent,
	'(': tokentype.LParent,
	')': tokentype.RParent,
	'[': tokentype.LSquaredParent,
	']': tokentype.RSquaredParent,
}

// Operator precedence maps for binary and unary operators.
var binaryPrecedenceMap = map[string]int{}

// Initialize precedence maps
func init() {
	for _, binaryPrecedence := range binaryPrecedences {
		binaryPrecedenceMap[binaryPrecedence.operator] = binaryPrecedence.precedence
	}
}

var binaryPrecedences = []struct {
	operator   string
	precedence int
}{
	{"::=", 1}, {"=>", 4}, {"->", 15}, {"<<", 11}, {">>", 11}, {"<=", 10}, {">=", 10},
	{"==", 9}, {"!=", 9}, {"&&", 6}, {"||", 5}, {"+=", 2}, {"-=", 2}, {"*=", 2},
	{"/=", 2}, {"%=", 2}, {"#=", 2}, {":=", 2},  {"@", 18}, {".", 17}, {"+", 13}, {"-", 13},
	{"*", 14}, {"/", 14}, {"%", 14}, {",", 4}, {":", 15}, {"=", 2}, {"<", 10}, {">", 10},
	{"&", 8}, {"|", 7},
}

const FunctionCallPrecedence = 16
const CurlyCallPrecedence = 14

// Tokenize converts the input code into a list of tokens.
func Tokenize(code string, source string) []*Token {
	var tokens []*Token
	r := CreateStringReader(code)

	for r.PeekNext(0) != '\000' { // While there are more characters
		current := r.Next()
		if token := nextToken(current, &r, source); token != nil {
			tokens = append(tokens, token)
		}
	}

	mapBinaryPrecedence(tokens)

	return tokens
}

// nextToken identifies and returns the next token from the input stream.
func nextToken(current rune, r *StringReader, source string) *Token {
	switch {
	case current == '/' && r.PeekNext(0) == '/':
		skipLineComment(r)
		return nil
	case current == '\\' && (r.PeekNext(0) == '\n' || (r.PeekNext(0) == '\r' && r.PeekNext(1) == '\n')):
		skipLineContinuation(r)
		return nil
	case current == '\n' || current == ';':
		return createToken(tokentype.Newline, source, r.ptr, string(current))
	case unicode.IsSpace(current):
		return nil
	case unicode.IsLetter(current) || strings.ContainsRune("_$\\", current):
		return createIdentifierToken(current, r, source)
	case strings.ContainsRune("+-*/=~#:?!%&|,.^<>@", current):
		return createOperatorToken(current, r, source)
	case unicode.IsDigit(current):
		return createNumberToken(current, r, source)
	case current == '"':
		return createStringToken(r, source)
	default:
		if tokenType, exists := oneCharacterTokens[current]; exists {
			return createToken(tokenType, source, r.ptr, string(current))
		}
	}
	return nil
}

// Helper function to create a new token.
func createToken(tokenType tokentype.TokenType, source string, pos int, value string) *Token {
	return &Token{
		TokenType: tokenType,
		Source:    source,
		Pos:       pos,
		Value:     value,
	}
}

// Reads a sequence of characters for identifiers or operators.
func readCharacters(r *StringReader, current rune, check func(r rune) bool) string {
	var builder strings.Builder
	builder.WriteRune(current)
	for check(r.PeekNext(0)) {
		builder.WriteRune(r.Next())
	}
	return builder.String()
}

// Creates an identifier token.
func createIdentifierToken(current rune, r *StringReader, source string) *Token {
	value := readCharacters(r, current, func(r rune) bool {
		return unicode.IsLetter(r) || unicode.IsDigit(r) || strings.ContainsRune("_$", r)
	})
	return &Token{
		TokenType: tokentype.Identifier,
		Source:    source,
		Pos:       r.ptr,
		Value:     value,
		Alias:     value,
		BinPrec:   3,
	}
}

// Creates a number token.
func createNumberToken(current rune, r *StringReader, source string) *Token {
	value := readCharacters(r, current, func(r rune) bool {
		return unicode.IsDigit(r)
	})
	return &Token{
		TokenType: tokentype.Number,
		Source:    source,
		Pos:       r.ptr,
		Value:     value,
		Alias:     value,
		BinPrec:   3,
	}
}

// Creates an operator token.
func createOperatorToken(current rune, r *StringReader, source string) *Token {
	value := readCharacters(r, current, func(r rune) bool {
		return strings.ContainsRune("+-*/=~#:?!%&|,.^<>@", r)
	})
	return &Token{
		TokenType: tokentype.Operator,
		Source:    source,
		Pos:       r.ptr,
		Value:     value,
		Alias:     value,
		BinPrec:   3,
	}
}

// Creates a string token.
func createStringToken(r *StringReader, source string) *Token {
	var builder strings.Builder
	pos := r.ptr
	for r.PeekNext(0) != '"' && r.PeekNext(0) != '\000' {
		if r.PeekNext(0) == '\\' {
			r.Next()
			switch(r.Next()) {
			case '"':
				builder.WriteRune('"')
				break;
			case '\\':
				builder.WriteRune('\\')
				break;
			case 'n':
				builder.WriteRune('\n')
				break;
			case 'r':
				builder.WriteRune('\r')
				break;
			case 't':
				builder.WriteRune('\t')
				break;
			case 'a':
				builder.WriteRune('\a')
				break;
			case 'f':
				builder.WriteRune('\f')
				break;
			case 'b':
				builder.WriteRune('\b')
				break;
			case 'v':
				builder.WriteRune('\v')
				break;
			}
			continue
		}
		builder.WriteRune(r.Next())
	}
	r.Next() // Consume closing quote

	value := builder.String()
	return &Token{
		TokenType: tokentype.String,
		Source:    source,
		Pos:       pos,
		Value:     value,
		Alias:     value,
	}
}

// Skips a line comment.
func skipLineComment(r *StringReader) {
	for r.PeekNext(0) != '\n' && r.PeekNext(0) != '\000' {
		r.Next()
	}
}

// Skips a line continuation (for both Unix and Windows).
func skipLineContinuation(r *StringReader) {
	if r.PeekNext(0) == '\n' {
		r.Next() // Unix-style line continuation
	} else if r.PeekNext(0) == '\r' && r.PeekNext(1) == '\n' {
		r.Next(); r.Next() // Windows-style line continuation
	}
}

// Maps binary operator precedences to tokens.
func mapBinaryPrecedence(tokens []*Token) {
	for _, token := range tokens {
		if token.TokenType == tokentype.Operator {
			if precedence, exists := binaryPrecedenceMap[token.Value]; exists {
				token.BinPrec = precedence
			}
		}
	}
}

// String returns a string representation of a token.
func (token Token) String() string {
	if token.TokenType == tokentype.Operator {
		return fmt.Sprintf("%s(%d) ", token.TokenType, token.BinPrec)
	}
	return fmt.Sprintf("%s ", token.TokenType)
}