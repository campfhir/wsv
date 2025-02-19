package utils

import (
	"errors"
	"fmt"
	"unsafe"
)

const (
	CharCharacterTabulation     = 0x0009
	CharLineFeed                = 0x000A
	CharLineTabulation          = 0x000B
	CharFormFeed                = 0x000C
	CharCarriageReturn          = 0x000D
	CharSpace                   = 0x0020
	CharNextLine                = 0x0085
	CharNoBreakSpace            = 0x00A0
	CharOghamSpaceMark          = 0x1680
	CharEnQuad                  = 0x2000
	CharEmQuad                  = 0x2001
	CharEnSpace                 = 0x2002
	CharEmSpace                 = 0x2003
	CharThreePerEmSpace         = 0x2004
	CharFourPerEmSpace          = 0x2005
	CharSixPerEmSpace           = 0x2006
	CharFigureSpace             = 0x2007
	CharPunctuationSpace        = 0x2008
	CharThinSpace               = 0x2009
	CharHairSpace               = 0x200A
	CharLineSeparator           = 0x2028
	CharParagraphSeparator      = 0x2029
	CharNarrowNoBreakSpace      = 0x202F
	CharMediumMathematicalSpace = 0x205F
	CharIdeographicSpace        = 0x3000
)

var ptrSize int = 0

func PtrSize() int {
	if ptrSize != 0 {
		return ptrSize
	}

	u_ptr_size := unsafe.Sizeof(uintptr(0))
	if u_ptr_size == 4 {
		ptrSize = 4
	} else if u_ptr_size == 8 {
		ptrSize = 8
	} else {
		panic("cannot determine architecture")
	}

	return ptrSize
}

func RuneToBytes(rs []rune) []byte {
	b := []byte{}
	for _, r := range rs {
		b = append(b, byte(r))
	}
	return b
}

func IsFieldDelimiter(rn rune) bool {
	return (rn == CharCharacterTabulation ||
		rn == CharLineTabulation ||
		rn == CharFormFeed ||
		rn == CharCarriageReturn ||
		rn == CharSpace ||
		rn == CharNextLine ||
		rn == CharNoBreakSpace ||
		rn == CharOghamSpaceMark ||
		rn == CharEnQuad ||
		rn == CharEmQuad ||
		rn == CharEnSpace ||
		rn == CharEmSpace ||
		rn == CharThreePerEmSpace ||
		rn == CharFourPerEmSpace ||
		rn == CharSixPerEmSpace ||
		rn == CharFigureSpace ||
		rn == CharPunctuationSpace ||
		rn == CharThinSpace ||
		rn == CharHairSpace ||
		rn == CharLineSeparator ||
		rn == CharParagraphSeparator ||
		rn == CharNarrowNoBreakSpace ||
		rn == CharMediumMathematicalSpace ||
		rn == CharIdeographicSpace)
}

func IsLiteralEmptyString(b []*byte) bool {
	if len(b) != 4 {
		return false
	}
	b0 := b[0]
	b1 := b[1]
	b2 := b[2]
	b3 := b[3]

	if rune(*b1) != '"' || rune(*b2) != '"' {
		return false
	}

	if b0 != nil && !IsFieldDelimiter(rune(*b0)) {
		return false
	}

	if b3 != nil && !IsFieldDelimiter(rune(*b3)) {
		return false
	}

	return true
}

func GetIndexOfSlice[T any](s []T, i int) (*T, error) {
	if i < 0 {
		return nil, errors.New("index must be be 0 or greater")
	}
	if i > len(s)-1 {
		return nil, fmt.Errorf("index %d is greater than %d", i, len(s)-1)
	}
	return &s[i], nil
}
