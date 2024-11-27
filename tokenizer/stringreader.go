package tokenizer

type StringReader struct {
	runes []rune // Store the input as a slice of runes
	ptr   int    // Pointer to the current position in the rune slice
}

// CreateStringReader converts the input string to a rune slice once.
func CreateStringReader(str string) StringReader {
	return StringReader{
		runes: []rune(str), // Convert the string to a rune slice once
		ptr:   0,
	}
}

// PeekNext returns the rune at the current pointer + index without advancing the pointer.
func (stringReader *StringReader) PeekNext(index int) rune {
	if (stringReader.ptr + index) >= len(stringReader.runes) {
		return '\000' // Return the null rune if out of bounds
	}
	return stringReader.runes[stringReader.ptr+index]
}

// Next advances the pointer and returns the current rune.
func (stringReader *StringReader) Next() rune {
	if stringReader.PeekNext(0) == '\000' {
		return '\000' // Return null rune if at the end
	}
	stringReader.ptr += 1
	return stringReader.runes[stringReader.ptr-1] // Return the current rune after advancing the pointer
}