package hw02unpackstring

import (
	"errors"
	"strings"
)

var ErrInvalidString = errors.New("invalid string")

func isDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

func isEscape(r rune) bool {
	return r == '\\'
}

func Unpack(source string) (string, error) {
	var resp strings.Builder
	resp.Grow(len(source))
	var startIndex int // index of the first character to copy
	var endIndex int   // index beyond the last character to copy
	var last rune      // last character value

	wasDigit := false
	wasEscape := false
	for index, char := range source {
		isEscape := isEscape(char)
		isDigit := isDigit(char)

		// if the previous character was an escape
		if wasEscape {
			// make sure the current character is escapable
			if !isEscape && !isDigit {
				return "", ErrInvalidString
			}
			// escape
			isEscape, isDigit = false, false
		}

		// if current character is a digit
		if isDigit {
			// if the first character of the string or follows the digit
			if index == 0 || wasDigit {
				return "", ErrInvalidString
			}
		}

		if wasEscape || isDigit {
			// write string up to a symbol before the digit or previous escape
			resp.WriteString(source[startIndex:endIndex])
		}

		if isDigit {
			// write previous character digit number of times
			digit := int(char - '0')
			if digit > 0 {
				resp.WriteString(strings.Repeat(string(last), digit))
			}
		} else {
			// current character is not a digit
			// if we previously escaped or wrote a character digit of times
			if wasEscape || wasDigit {
				// start from current position again
				startIndex = index
			}
			// save last index
			endIndex = index
			if !isEscape {
				// save last character if not escape
				// we may multiply it next if it isn't
				last = char
			}
		}

		// if current character is special, save which one
		wasEscape = isEscape
		wasDigit = isDigit
	}

	// ending with escape
	if wasEscape {
		return "", ErrInvalidString
	}

	// copy rest of the string if the last character wasn't a digit
	if !wasDigit {
		resp.WriteString(source[startIndex:])
	}
	return resp.String(), nil
}
