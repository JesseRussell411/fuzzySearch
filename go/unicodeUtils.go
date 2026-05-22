package main

import (
	"unicode/utf8"
)

func bytesInRune_Utf8(start byte) (count int, validStartingByte bool) {
	// _, result := utf8.DecodeRune([]byte{start, 0b1000_0000, 0b1000_0000, 0b1000_0000})
	// count = result
	// validStartingByte = true
	// return

	validStartingByte = true
	count = 1
	if start&0b1000_0000 == 0 {
		count = 1
	} else if start&0b0100_0000 == 0 {
		validStartingByte = false
		count = 1
	} else {
		for i := 2; i < 5; i++ {
			mask := byte(0b1000_0000) >> i
			if start&mask == 0 {
				count = i
				return
			}
		}
		validStartingByte = false
	}
	return
}

func runeAtByteInString(s string, b int) (r rune, l int) {
	subString := s[b:]
	r, l = utf8.DecodeRuneInString(subString)
	return
}
