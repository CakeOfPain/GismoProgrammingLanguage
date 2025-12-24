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
	Line      int    // NEW: Line number (1-based)
	Column    int    // NEW: Column number (1-based)
	Value     string
	Alias     string
	BinPrec   int
}

// NoneToken is a sentinel token used to represent the absence of a token.
var NoneToken = &Token{
	TokenType: tokentype.None,
	Source:    "Unknown",
	Pos:       0,
	Line:      0,
	Column:    0,
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
		Line:      0,
		Column:    0,
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
	{"/=", 2}, {"%=", 2}, {"#=", 2}, {":=", 2}, {"<-", 2}, {"@", 18}, {".", 17}, {"+", 13}, {"-", 13},
	{"*", 14}, {"/", 14}, {"%", 14}, {",", 3}, {":", 15}, {"=", 2}, {"<", 10}, {">", 10},
	{"&", 8}, {"|", 7},
}

const FunctionCallPrecedence = 16
const CurlyCallPrecedence = 14
const identifierPrecedence = 4

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
	// Capture starting position of the token
	startPos := r.ptr - 1
	startLine := r.Line
	startCol := r.Column - 1 // Previous char was the start

	switch {
	case current == '/' && r.PeekNext(0) == '/':
		skipLineComment(r)
		return nil
	case current == '\\' && (r.PeekNext(0) == '\n' || (r.PeekNext(0) == '\r' && r.PeekNext(1) == '\n')):
		skipLineContinuation(r)
		return nil
	case current == '\n' || current == ';':
		return createToken(tokentype.Newline, source, startPos, startLine, startCol, string(current))
	case unicode.IsSpace(current):
		return nil
	case unicode.IsLetter(current) || strings.ContainsRune("_$\\", current):
		return createIdentifierToken(current, r, source, startPos, startLine, startCol)

	// Check if '.' is the start of a float (.123)
	case current == '.' && unicode.IsDigit(r.PeekNext(0)):
		return createNumberToken(current, r, source, startPos, startLine, startCol)

	case strings.ContainsRune("+-*/=~#:?!%&|,.^<>@", current):
		return createOperatorToken(current, r, source, startPos, startLine, startCol)
	case unicode.IsDigit(current):
		return createNumberToken(current, r, source, startPos, startLine, startCol)
	case current == '"':
		return createStringToken(r, source, startPos, startLine, startCol)
	default:
		if tokenType, exists := oneCharacterTokens[current]; exists {
			return createToken(tokenType, source, startPos, startLine, startCol, string(current))
		}
	}
	return nil
}

// Helper function to create a new token.
func createToken(tokenType tokentype.TokenType, source string, pos, line, col int, value string) *Token {
	return &Token{
		TokenType: tokenType,
		Source:    source,
		Pos:       pos,
		Line:      line,
		Column:    col,
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
func createIdentifierToken(current rune, r *StringReader, source string, pos, line, col int) *Token {
	value := readCharacters(r, current, func(r rune) bool {
		return unicode.IsLetter(r) || unicode.IsDigit(r) || strings.ContainsRune("_$", r)
	})
	return &Token{
		TokenType: tokentype.Identifier,
		Source:    source,
		Pos:       pos,
		Line:      line,
		Column:    col,
		Value:     value,
		Alias:     value,
		BinPrec:   identifierPrecedence,
	}
}

// Creates a number token (Integer, Float, Hex, Bin, Octal).
func createNumberToken(current rune, r *StringReader, source string, pos, line, col int) *Token {
	var builder strings.Builder
	builder.WriteRune(current)

	// CASE 1: Starts with Dot (.123)
	if current == '.' {
		for unicode.IsDigit(r.PeekNext(0)) {
			builder.WriteRune(r.Next())
		}
		return &Token{
			TokenType: tokentype.Number, Source: source, Pos: pos, Line: line, Column: col,
			Value: builder.String(), Alias: builder.String(), BinPrec: identifierPrecedence,
		}
	}

	// CASE 2: Base Prefixes (0x, 0b, 0o)
	if current == '0' {
		peek := r.PeekNext(0)
		if peek == 'x' || peek == 'X' {
			builder.WriteRune(r.Next()) // consume x
			for isHexDigit(r.PeekNext(0)) {
				builder.WriteRune(r.Next())
			}
			return &Token{
				TokenType: tokentype.Number, Source: source, Pos: pos, Line: line, Column: col,
				Value: builder.String(), Alias: builder.String(), BinPrec: identifierPrecedence,
			}
		}
		if peek == 'b' || peek == 'B' {
			builder.WriteRune(r.Next()) // consume b
			for r.PeekNext(0) == '0' || r.PeekNext(0) == '1' {
				builder.WriteRune(r.Next())
			}
			return &Token{
				TokenType: tokentype.Number, Source: source, Pos: pos, Line: line, Column: col,
				Value: builder.String(), Alias: builder.String(), BinPrec: identifierPrecedence,
			}
		}
		if peek == 'o' || peek == 'O' {
			builder.WriteRune(r.Next()) // consume o
			for r.PeekNext(0) >= '0' && r.PeekNext(0) <= '7' {
				builder.WriteRune(r.Next())
			}
			return &Token{
				TokenType: tokentype.Number, Source: source, Pos: pos, Line: line, Column: col,
				Value: builder.String(), Alias: builder.String(), BinPrec: identifierPrecedence,
			}
		}
	}

	// CASE 3: Standard Integer or Float
	for unicode.IsDigit(r.PeekNext(0)) {
		builder.WriteRune(r.Next())
	}

	// Check for Float
	if r.PeekNext(0) == '.' && unicode.IsDigit(r.PeekNext(1)) {
		builder.WriteRune(r.Next()) // Consume '.'
		for unicode.IsDigit(r.PeekNext(0)) {
			builder.WriteRune(r.Next()) // Consume fractional digits
		}
	}

	value := builder.String()
	return &Token{
		TokenType: tokentype.Number,
		Source:    source,
		Pos:       pos,
		Line:      line,
		Column:    col,
		Value:     value,
		Alias:     value,
		BinPrec:   identifierPrecedence,
	}
}

func isHexDigit(r rune) bool {
	return unicode.IsDigit(r) || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')
}

// Creates an operator token.
func createOperatorToken(current rune, r *StringReader, source string, pos, line, col int) *Token {
	value := readCharacters(r, current, func(r rune) bool {
		return strings.ContainsRune("+-*/=~#:?!%&|,.^<>@", r)
	})
	return &Token{
		TokenType: tokentype.Operator,
		Source:    source,
		Pos:       pos,
		Line:      line,
		Column:    col,
		Value:     value,
		Alias:     value,
		BinPrec:   identifierPrecedence,
	}
}

// Creates a string token.
func createStringToken(r *StringReader, source string, pos, line, col int) *Token {
	var builder strings.Builder
	// pos is passed in
	for r.PeekNext(0) != '"' && r.PeekNext(0) != '\000' {
		if r.PeekNext(0) == '\\' {
			r.Next()
			switch r.Next() {
			case '"':
				builder.WriteRune('"')
			case '\\':
				builder.WriteRune('\\')
			case 'n':
				builder.WriteRune('\n')
			case 'r':
				builder.WriteRune('\r')
			case 't':
				builder.WriteRune('\t')
			// ... add other escapes ...
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
		Line:      line,
		Column:    col,
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

// Skips a line continuation.
func skipLineContinuation(r *StringReader) {
	if r.PeekNext(0) == '\n' {
		r.Next()
	} else if r.PeekNext(0) == '\r' && r.PeekNext(1) == '\n' {
		r.Next()
		r.Next()
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

func (token Token) String() string {
	if token.TokenType == tokentype.Operator {
		return fmt.Sprintf("%s(%d) ", token.TokenType, token.BinPrec)
	}
	return fmt.Sprintf("%s ", token.TokenType)
}