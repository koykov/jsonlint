package jsonlint

import (
	"bytes"
	"unsafe"

	"github.com/koykov/bytealg"
	"github.com/koykov/byteconv"
	"github.com/koykov/byteseq"
)

var (
	// Byte constants.
	bNull  = []byte("null")
	bTrue  = []byte("true")
	bFalse = []byte("false")
	bFmt   = []byte(" \t\n\r")
)

// Validate source bytes.
func Validate[T byteseq.Byteseq](x T) (offset int, err error) {
	var p []byte
	if b, ok := byteseq.ToBytes(x); ok {
		p = b
	} else if s, ok := byteseq.ToString(x); ok {
		p = byteconv.S2B(s)
	}
	if len(p) == 0 {
		err = ErrEmptySrc
		return
	}
	p = bytealg.TrimBytes(p, bFmt)
	offset, err = validateGeneric(0, p, offset)
	if err != nil {
		return offset, err
	}
	if offset < len(p) {
		err = ErrUnparsedTail
	}
	return
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
		e := bytealg.IndexByteAtBytes(s, '"', offset+1)
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
				i = bytealg.IndexByteAtBytes(s, '"', i+1)
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
	n := len(s)
	for offset < len(s) {
		if s[offset] == '}' {
			// End of object.
			offset++
			break
		}
		if offset, eof = skipFmtTable(s, n, offset); eof {
			return offset, ErrUnexpEOF
		}
		// Parse key.
		if s[offset] != '"' {
			// Key should be a string wrapped with double quotas.
			return offset, ErrUnexpId
		}
		offset++
		e := bytealg.IndexByteAtBytes(s, '"', offset)
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
				i = bytealg.IndexByteAtBytes(s, '"', i+1)
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
		if offset, eof = skipFmtTable(s, n, offset); eof {
			return offset, ErrUnexpEOF
		}
		// Check division symbol.
		if s[offset] == ':' {
			offset++
		} else {
			return offset, ErrUnexpId
		}
		if offset, eof = skipFmtTable(s, n, offset); eof {
			return offset, ErrUnexpEOF
		}
		// Parse value. It may be an arbitrary type.
		if offset, err = validateGeneric(depth, s, offset); err != nil {
			return offset, err
		}
		if offset, eof = skipFmtTable(s, n, offset); eof {
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
		if offset, eof = skipFmtTable(s, n, offset); eof {
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
	n := len(s)
	for offset < len(s) {
		if s[offset] == ']' {
			// End of array.
			offset++
			break
		}
		if offset, eof = skipFmtTable(s, n, offset); eof {
			return offset, ErrUnexpEOF
		}
		// Parse the value.
		if offset, err = validateGeneric(depth, s, offset); err != nil {
			return offset, err
		}
		if offset, eof = skipFmtTable(s, n, offset); eof {
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
		if offset, eof = skipFmtTable(s, n, offset); eof {
			return offset, ErrUnexpEOF
		}
	}
	return offset, nil
}

// Check if given byte is a part of the number.
func isDigit(c byte) bool {
	return (c >= '0' && c <= '9') || c == '-' || c == '+' || c == 'e' || c == 'E'
}

// Check if given is a part of the number, including dot.
func isDigitDot(c byte) bool {
	return isDigit(c) || c == '.'
}

// Table based approach of skipFmt.
func skipFmtTable(src []byte, n, offset int) (int, bool) {
	_ = src[n-1]
	_ = skipTable[255]
	if n-offset > 512 {
		offset, _ = skipFmtBin8(src, n, offset)
	}
	for ; offset < n && skipTable[src[offset]]; offset++ {
	}
	return offset, offset == n
}

// Binary based approach of skipFmt.
func skipFmtBin8(src []byte, n, offset int) (int, bool) {
	_ = src[n-1]
	_ = skipTable[255]
	if *(*uint64)(unsafe.Pointer(&src[offset])) == binNlSpace7 {
		offset += 8
		for offset < n && *(*uint64)(unsafe.Pointer(&src[offset])) == binSpace8 {
			offset += 8
		}
	}
	return offset, false
}

var (
	skipTable   = [256]bool{}
	binNlSpace7 uint64
	binSpace8   uint64
)

func init() {
	skipTable[' '] = true
	skipTable['\t'] = true
	skipTable['\n'] = true
	skipTable['\t'] = true

	binNlSpace7Bytes, binSpace8Bytes := []byte("\n       "), []byte("        ")
	binNlSpace7, binSpace8 = *(*uint64)(unsafe.Pointer(&binNlSpace7Bytes[0])), *(*uint64)(unsafe.Pointer(&binSpace8Bytes[0]))
}
