package main

import (
	"unicode/utf8"
)

func bytesInRune_utf8(start byte) (count int, validStartingByte bool) {
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

func BytesInRune_utf8(start byte) int {
	byteCount, _ := bytesInRune_utf8(start)
	return byteCount
}

func IsValidStartingByte_utf8(b byte) bool {
	_, valid := bytesInRune_utf8(b)
	return valid

}

func subString_utf8(str string, start, length int) (substring string, actualStart int) {
	for !IsValidStartingByte_utf8(str[start]) && start > 0 {
		start--
	}

	if start+length < len(str) {
		for !IsValidStartingByte_utf8(str[start+length]) && length > 0 {
			length--
		}
	}

	substring = str[start : start+length]
	actualStart = start
	return
}

func breakIntoSubstrings_utf8(str string, count int) []string {
	length := len(str) / count
	remainder := len(str) % count
	totalLength := 0
	result := make([]string, count)

	for i := range count {
		subLength := length
		if i < remainder {
			subLength += 1
		}

		subString, actualStart := subString_utf8(str, totalLength, subLength)

		actualLength := len(subString)
		arctualStartOffset := actualStart - totalLength

		totalLength += actualLength - arctualStartOffset

		result[i] = subString
	}

	return result
}

func runeAtByteInString(s string, b int) (r rune, l int) {
	subString := s[b:]
	r, l = utf8.DecodeRuneInString(subString)
	return
}

func IsTrailingByte(b byte) bool {
	return b&0b1100_0000 == 0b1000_0000
}
