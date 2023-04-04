package stract

type runeStack []rune

// IsEmpty check if stack is empty
func (s *runeStack) isEmpty() bool {
	return len(*s) == 0
}

// Push a new value onto the stack
func (s *runeStack) push(str rune) {
	*s = append(*s, str) // Simply append the new value to the end of the stack
}

// Pop remove and return top element of stack. Return false if stack is empty.
func (s *runeStack) pop() rune {
	if s.isEmpty() {
		return none
	} else {
		index := len(*s) - 1   // Get the index of the top most element.
		element := (*s)[index] // Index into the slice and obtain the element.
		*s = (*s)[:index]      // Remove it from the stack by slicing it off.
		return element
	}
}

func (s *runeStack) Last() rune {
	if s.isEmpty() {
		return none
	} else {
		return (*s)[len(*s)-1]
	}
}
