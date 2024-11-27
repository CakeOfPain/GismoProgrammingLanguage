package tokentype

// TokenType represents the type of a token in the compiler's tokenizer.
type TokenType int

// Token types in the tokenizer.
const (
	None TokenType = iota            // No token (default value)
	Identifier                       // Variable, function, or identifier
	Operator                         // Operator symbols (e.g., +, -, *, /)
	Number                           // Number (e.g. 1, 2, 3, 4, 5)
	String                           // String literals
	LParent                          // Left Parenthesis '('
	RParent                          // Right Parenthesis ')'
	LCurlyParent                     // Left Curly Brace '{'
	RCurlyParent                     // Right Curly Brace '}'
	LSquaredParent                   // Left Square Bracket '['
	RSquaredParent                   // Right Square Bracket ']'
	Semicolon                        // Semicolon ';'
	Newline                          // Newline character
	Module                           // Module keyword or token
)

// String returns a string representation of the TokenType.
func (tokenType TokenType) String() string {
	switch tokenType {
	case None:
		return "<None>"
	case Identifier:
		return "<Identifier>"
	case Operator:
		return "<Operator>"
	case String:
		return "<String>"
	case Number:
		return "<Number>"
	case LParent:
		return "<LParent>"
	case RParent:
		return "<RParent>"
	case LCurlyParent:
		return "<LCurlyParent>"
	case RCurlyParent:
		return "<RCurlyParent>"
	case LSquaredParent:
		return "<LSquaredParent>"
	case RSquaredParent:
		return "<RSquaredParent>"
	case Semicolon:
		return "<Semicolon>"
	case Newline:
		return "<Newline>"
	case Module:
		return "<Module>"
	default:
		return "<Unknown TokenType>" // Fallback for unrecognized token types
	}
}