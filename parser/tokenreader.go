package parser

import (
	"gismolang.org/compiler/tokenizer"
	"gismolang.org/compiler/tokenizer/tokentype"
)

// TokenReader is used to traverse the list of tokens while parsing.
type TokenReader struct {
	tokens []*tokenizer.Token
	ptr    int
}

// CreateTokenReader initializes and returns a new TokenReader.
func CreateTokenReader(tokens []*tokenizer.Token) TokenReader {
	return TokenReader{
		tokens: tokens,
		ptr:    0,
	}
}

// PeekNext returns the token at the current pointer plus the given index offset without advancing the pointer.
// If the index exceeds the token list length, it returns a NoneToken.
func (tr *TokenReader) PeekNext(index int) *tokenizer.Token {
	if (tr.ptr + index) >= len(tr.tokens) {
		return tokenizer.NoneToken
	}
	return tr.tokens[tr.ptr+index]
}

// Next advances the token pointer and returns the next token.
// If the token list has been exhausted, it returns a NoneToken.
func (tr *TokenReader) Next() *tokenizer.Token {
	if tr.PeekNext(0).TokenType == tokentype.None {
		return tokenizer.NoneToken
	}
	tr.ptr++
	return tr.PeekNext(-1) // Return the previously pointed token
}