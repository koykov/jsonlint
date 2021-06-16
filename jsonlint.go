package jsonlint

import (
	"bytes"
	"errors"

	"github.com/koykov/bytealg"
	"github.com/koykov/fastconv"
)

var (
	// Byte constants.
	bNull  = []byte("null")
	bTrue  = []byte("true")
	bFalse = []byte("false")
	bFmt   = []byte(" \t\n\r")

	// Errors.
	ErrEmptySrc     = errors.New("can't parse empty source")
	ErrUnparsedTail = errors.New("unparsed tail")
	ErrUnexpId      = errors.New("unexpected identifier")
	ErrUnexpEOF     = errors.New("unexpected end of file")
	ErrUnexpEOS     = errors.New("unexpected end of string")

	// Suppress go vet warnings.
	_ = ValidateStr
)

// Validate source bytes.
func Validate(s []byte) (offset int, err error) {
	if len(s) == 0 {
		err = ErrEmptySrc
		return
	}
	s = bytealg.Trim(s, bFmt)
	offset, err = validateGeneric(0, s, offset)
	if err != nil {
		return offset, err
	}
	if offset < len(s) {
		err = ErrUnparsedTail
	}
	return
}

// Validate source string.
func ValidateStr(s string) (int, error) {
	return Validate(fastconv.S2B(s))
}

// Generic validation helper.
func validateGeneric(depth int, s []byte, offset int) (int, error) {
	var err error

	switch {
	case s[offset] == 'n':
		// Check null node.
		if len(s[offset:]) > 3 && bytes.Equal(bNull, s[offset:offset+4]) {
			offset += 4
		} else {
			return offset, ErrUnexpId
		}
	case s[offset] == '{':
		// Check object node.
		offset, err = validateObj(depth+1, s, offset)
	case s[offset] == '[':
		// Check array node.
		offset, err = validateArr(depth+1, s, offset)
	case s[offset] == '"':
		// Check string node.
		e := bytealg.IndexByteAtLR(s, '"', offset+1)
		if e < 0 {
			return len(s), ErrUnexpEOS
		}
		if s[e-1] != '\\' {
			// Good case - string isn't escaped.
			offset = e + 1
		} else {
			// Walk over double quotas and look for unescaped.
			_ = s[len(s)-1]
			for i := e; i < len(s); {
				i = bytealg.IndexByteAtLR(s, '"', i+1)
				if i < 0 {
					e = len(s) - 1
					break
				}
				e = i
				if s[e-1] != '\\' {
					break
				}
			}
			offset = e + 1
		}
	case isDigit(s[offset]):
		// Check number node.
		if len(s[offset:]) > 0 {
			i := offset
			for isDigitDot(s[i]) {
				i++
				if i == len(s) {
					break
				}
			}
			offset = i
		} else {
			return offset, ErrUnexpEOF
		}
	case s[offset] == 't':
		// Check true node.
		if len(s[offset:]) > 3 && bytes.Equal(bTrue, s[offset:offset+4]) {
			offset += 4
		} else {
			return offset, ErrUnexpId
		}
	case s[offset] == 'f':
		// Check false node.
		if len(s[offset:]) > 4 && bytes.Equal(bFalse, s[offset:offset+5]) {
			offset += 5
		} else {
			return offset, ErrUnexpId
		}
	default:
		err = ErrUnexpId
	}

	return offset, err
}

// Object validation helper.
func validateObj(depth int, s []byte, offset int) (int, error) {
	offset++
	var (
		err error
		eof bool
	)
	for offset < len(s) {
		if s[offset] == '}' {
			// End of object.
			offset++
			break
		}
		if offset, eof = skipFmt(s, offset); eof {
			return offset, ErrUnexpEOF
		}
		// Parse key.
		if s[offset] != '"' {
			// Key should be a string wrapped with double quotas.
			return offset, ErrUnexpId
		}
		offset++
		e := bytealg.IndexByteAtLR(s, '"', offset)
		if e < 0 {
			return len(s), ErrUnexpEOS
		}
		if s[e-1] != '\\' {
			// Good case - key isn't escaped.
			offset = e + 1
		} else {
			// Key contains escaped bytes.
			_ = s[len(s)-1]
			for i := e; i < len(s); {
				i = bytealg.IndexByteAtLR(s, '"', i+1)
				if i < 0 {
					e = len(s) - 1
					break
				}
				e = i
				if s[e-1] != '\\' {
					break
				}
			}
			offset = e + 1
		}
		if offset, eof = skipFmt(s, offset); eof {
			return offset, ErrUnexpEOF
		}
		// Check division symbol.
		if s[offset] == ':' {
			offset++
		} else {
			return offset, ErrUnexpId
		}
		if offset, eof = skipFmt(s, offset); eof {
			return offset, ErrUnexpEOF
		}
		// Parse value. It may be an arbitrary type.
		if offset, err = validateGeneric(depth, s, offset); err != nil {
			return offset, err
		}
		if offset, eof = skipFmt(s, offset); eof {
			return offset, ErrUnexpEOF
		}
		// Check end of object again.
		if s[offset] == '}' {
			offset++
			break
		}
		// Check separate symbol.
		if s[offset] == ',' {
			offset++
		} else {
			return offset, ErrUnexpId
		}
		if offset, eof = skipFmt(s, offset); eof {
			return offset, ErrUnexpEOF
		}
	}
	return offset, err
}

// Array validation helper.
func validateArr(depth int, s []byte, offset int) (int, error) {
	offset++
	var (
		err error
		eof bool
	)
	for offset < len(s) {
		if s[offset] == ']' {
			// End of array.
			offset++
			break
		}
		if offset, eof = skipFmt(s, offset); eof {
			return offset, ErrUnexpEOF
		}
		// Parse the value.
		if offset, err = validateGeneric(depth, s, offset); err != nil {
			return offset, err
		}
		if offset, eof = skipFmt(s, offset); eof {
			return offset, ErrUnexpEOF
		}
		if s[offset] == ']' {
			// End of the array caught.
			offset++
			break
		}
		// Check separate symbol.
		if s[offset] == ',' {
			offset++
		} else {
			return offset, ErrUnexpId
		}
		if offset, eof = skipFmt(s, offset); eof {
			return offset, ErrUnexpEOF
		}
	}
	return offset, nil
}

// Skip formatting symbols like tabs, spaces, ...
//
// Returns the next non-format symbol index.
func skipFmt(s []byte, offset int) (int, bool) {
loop:
	if offset >= len(s) {
		return offset, true
	}
	c := s[offset]
	if c != bFmt[0] && c != bFmt[1] && c != bFmt[2] && c != bFmt[3] {
		return offset, false
	}
	offset++
	goto loop
}

// Check if given byte is a part of the number.
func isDigit(c byte) bool {
	return (c >= '0' && c <= '9') || c == '-' || c == '+' || c == 'e' || c == 'E'
}

// Check if given is a part of the number, including dot.
func isDigitDot(c byte) bool {
	return isDigit(c) || c == '.'
}
