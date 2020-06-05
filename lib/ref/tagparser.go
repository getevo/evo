package ref

import (
	"errors"
	"unicode"
)

const valueQuote = '"'
const keyValueDelimiter = ':'

// Parse parses a custom tag string of struct.
//
// Strict mode is a mode selector:
// It raises an error when an unacceptable custom tag string is given if the mode is true.
// On the other hand, if the mode is false, it immediately returns the processed results until just before the invalid custom tag syntax. It never raises any error.
func ParseTags(tagString string, isStrict bool) (map[string]string, error) {
	key := make([]rune, 0, 100)
	keyCursor := 0
	value := make([]rune, 0, 100)
	valueCursor := 0

	inKeyParsing := true
	isEscaping := false

	tagKeyValue := make(map[string]string)
	tagSetMarker := make(map[string]bool)

	tagRunes := []rune(tagString)
	tagRunesLen := len(tagRunes)
	for i := 0; i < tagRunesLen; i++ {
		r := tagRunes[i]

		if inKeyParsing {
			if unicode.IsSpace(r) {
				if keyCursor > 0 {
					if isStrict {
						return nil, errors.New("invalid custom tag syntax: key must not contain any white space, but it contains")
					}
					// give up when key contain any white space
					break
				}
				continue
			}

			if r == valueQuote {
				if isStrict {
					return nil, errors.New("invalid custom tag syntax: key must not contain any double quote, but it contains")
				}
				// give up when key contains any double quote
				break
			}

			if r == keyValueDelimiter {
				if keyCursor <= 0 {
					if isStrict {
						return nil, errors.New("invalid custom tag syntax: key must not be empty, but it gets empty")
					}

					// give up when key is empty
					break
				}

				inKeyParsing = false
				i++
				if i >= tagRunesLen {
					if isStrict {
						return nil, errors.New("invalid custom tag syntax: value must not be empty, but it gets empty")
					}
					// give up when value is empty
					break
				}
				if tagRunes[i] != valueQuote {
					if isStrict {
						return nil, errors.New("invalid custom tag syntax: quote for value is missing")
					}
					// give up when value isn't wrapped by double quote
					break
				}
				continue
			}
			key = append(key, r)
			keyCursor++
			continue
		}

		// value parsing
		if !isEscaping && r == valueQuote {
			keyStr := string(key[:keyCursor])
			if !tagSetMarker[keyStr] {
				tagSetMarker[keyStr] = true
				tagKeyValue[keyStr] = string(value[:valueCursor])
			}
			key = key[:0]
			keyCursor = 0
			value = value[:0]
			valueCursor = 0
			inKeyParsing = true
			continue
		}

		if r == '\\' {
			if isEscaping {
				value = append(value, r)
				valueCursor++
				isEscaping = false
				continue
			}
			isEscaping = true
			continue
		}
		value = append(value, r)
		isEscaping = false
		valueCursor++
	}

	if inKeyParsing && keyCursor > 0 && isStrict {
		return nil, errors.New("invalid custom tag syntax: a delimiter of key and value is missing")
	}

	if !inKeyParsing && valueCursor > 0 && isStrict {
		return nil, errors.New("invalid custom tag syntax: a value is not terminated with quote")
	}

	return tagKeyValue, nil
}
