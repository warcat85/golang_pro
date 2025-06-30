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

		if wasEscape {
			if !isEscape && !isDigit {
				return "", ErrInvalidString
			}
			isEscape, isDigit = false, false
		}

		if isDigit {
			if index == 0 || wasDigit {
				return "", ErrInvalidString
			}
		}

		if wasEscape || isDigit {
			resp.WriteString(source[startIndex:endIndex])
		}

		if isDigit {
			digit := int(char - '0')
			if digit > 0 {
				resp.WriteString(strings.Repeat(string(last), digit))
			}
		} else {
			if wasEscape || wasDigit {
				startIndex = index
			}
			endIndex = index
			if !isEscape {
				last = char
			}
		}

		wasEscape = isEscape
		wasDigit = isDigit
	}

	if wasEscape {
		return "", ErrInvalidString
	}

	if !wasDigit {
		resp.WriteString(source[startIndex:])
	}
	return resp.String(), nil
}
