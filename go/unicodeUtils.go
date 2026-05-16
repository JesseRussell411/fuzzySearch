package main

import (
	"unicode/utf8"
)

func bytesInRune_Utf8(start byte) (count int, validStartingByte bool) {
	validStartingByte = true
	if start&0b1000_0000 == 0 {
		count = 1
	} else if start&0b0100_0000 == 0 {
		validStartingByte = false
		count = 1
	} else {
		for i := 2; i < 8; i++ {
			mask := byte(0b1000_0000) >> i
			if start&mask == 0 {
				count = i
				return
			}
		}
		count = 8
	}
	return
}

func runeAtByteInString(s string, b int) (r rune, l int) {
	subString := s[b:]
	r, l = utf8.DecodeRuneInString(subString)
	return
}
