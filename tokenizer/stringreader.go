package tokenizer

type StringReader struct {
    runes []rune // Store the input as a slice of runes
    ptr   int    // Pointer to the current position in the rune slice
    
    // Position of the *next* character (the one PeekNext sees)
    Line   int
    Column int

    // Position of the *most recently consumed* character (the one Next returned)
    // We use this in nextToken to know exactly where 'current' came from.
    LastLine   int
    LastColumn int
}

// CreateStringReader converts the input string to a rune slice once.
func CreateStringReader(str string) StringReader {
    return StringReader{
        runes:      []rune(str),
        ptr:        0,
        Line:       1, // Start at Line 1
        Column:     1, // Start at Column 1
        LastLine:   1,
        LastColumn: 1,
    }
}

// PeekNext returns the rune at the current pointer + index without advancing the pointer.
func (stringReader *StringReader) PeekNext(index int) rune {
    if (stringReader.ptr + index) >= len(stringReader.runes) {
        return '\000' // Return the null rune if out of bounds
    }
    return stringReader.runes[stringReader.ptr+index]
}

// Next advances the pointer, updates positions, and returns the current rune.
func (stringReader *StringReader) Next() rune {
    if stringReader.PeekNext(0) == '\000' {
        return '\000' // Return null rune if at the end
    }

    // 1. Capture the position of the character we are about to consume
    stringReader.LastLine = stringReader.Line
    stringReader.LastColumn = stringReader.Column

    // 2. Get the character and advance pointer
    char := stringReader.runes[stringReader.ptr]
    stringReader.ptr += 1

    // 3. Update the Line/Column for the *next* character
    if char == '\n' {
        stringReader.Line++
        stringReader.Column = 1
    } else {
        stringReader.Column++
    }

    return char
}