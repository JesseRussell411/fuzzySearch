package main

import (
	"fmt"
	"math"
	"unicode/utf8"
)

type FuzzySearchMatch struct {
	minimumEditDistance int
	index               int
	byteOffset          int
	length              int
	byteCount           int
}

func fuzzySearch(test, search string) FuzzySearchMatch {
	searchLength := utf8.RuneCountInString(search)

	// default values for empty string
	minimumEditDistance := searchLength
	charIndex := 0
	byteIndex := 0
	charLength := 0
	byteLength := 0

	if search == "" || test == "" {
		return FuzzySearchMatch{
			minimumEditDistance: minimumEditDistance,
			index:               charIndex,
			byteOffset:          byteIndex,
			length:              charLength,
			byteCount:           byteLength,
		}
	}

	columnCount := searchLength + 1
	seedRow := make([]int, columnCount)
	row := make([]int, columnCount)
	nextRow := make([]int, columnCount)
	for i := range seedRow {
		seedRow[i] = i
	}

	prevRow := seedRow

	wb := 0
	wi := 0

	potentialEditDistanceForWI := 0
	for {
		windowRuneLength := bytesInRune(test[wb])
		testRuneLength := windowRuneLength

		minimumEditDistanceForWI := math.MaxInt
		lOfMinimumEditDistanceForWI := 0
		bcOfMinimumEditDistanceForWI := 0

		wbc := 0

		tb := wb
		ti := wi
		wl := 1

		// check if it's worth checking the windows position
		if potentialEditDistanceForWI >= minimumEditDistance {
			// if not: skip to the next one
			potentialEditDistanceForWI -= 2
			goto nextWindowPosition
		}

		for {
			wbc += testRuneLength
			if wbc >= len(test)-wb {
				break
			}
			r := wl

			// populate seed column
			row[0] = r
			println("r ", r)

			si := 0
			sb := 0
			for {
				// TODO? cache this in a lookup table?
				searchRuneLength := bytesInRune(search[sb])
				c := si + 1
				println("c", c)

				testRune, _ := runeAtByteInString(test, tb)
				searchRune, _ := runeAtByteInString(search, sb)
				if searchRune == testRune {
					row[c] = prevRow[c-1]
				} else {
					north := prevRow[c]
					northWest := prevRow[c-1]
					west := row[c-1]
					println("n ", north)
					println("nw", northWest)
					println(" w", west)

					row[c] = 1 + min(north, northWest, west)
				}
				println("ed--", row[c])

				si++
				sb += searchRuneLength
				if sb >= len(search) {
					break
				}
			}

			editDist := row[searchLength]
			// rotate rows
			prevRow = row
			temp := row
			row = nextRow
			nextRow = temp

			if editDist <= minimumEditDistanceForWI {
				minimumEditDistanceForWI = editDist
				lOfMinimumEditDistanceForWI = wl
				bcOfMinimumEditDistanceForWI = wbc
			}

			// clamp window size
			potentialEditDist := editDist - (searchLength - wl)
			if potentialEditDist > minimumEditDistance {
				break
			}

			tb += testRuneLength
			ti++
			wl++

			testRuneLength = bytesInRune(test[tb])
		}
		// reset rows
		prevRow = seedRow

		if minimumEditDistanceForWI < minimumEditDistance {
			minimumEditDistance = minimumEditDistanceForWI
			byteIndex = wb
			charIndex = wi
			byteLength = bcOfMinimumEditDistanceForWI
			charLength = lOfMinimumEditDistanceForWI
		}

		// check for perfect match
		if minimumEditDistance == 0 {
			// perfect match found
			break
		}

		// update potential edit distance for window position
		if minimumEditDistanceForWI == math.MaxInt {
			potentialEditDistanceForWI -= 2
		} else {
			potentialEditDistanceForWI = minimumEditDistanceForWI - 2
		}

	nextWindowPosition:
		wb += windowRuneLength
		wi++

		if wb >= len(test) {
			break
		}
	}

	result := FuzzySearchMatch{
		minimumEditDistance: minimumEditDistance,
		index:               charIndex,
		byteOffset:          byteIndex,
		length:              charLength,
		byteCount:           byteLength,
	}

	return result
}

func main() {
	s := "the fox jumps"
	fsm := fuzzySearch(s, "fax")
	ss := s[fsm.index : fsm.index+fsm.length]
	fmt.Println(fsm, ss)
}

func use(...any) {
}

func bytesInRune(start byte) int {
	if start&0b1000_0000 == 0 {
		return 1
	} else if start&0b0100_0000 == 0 {
		return 0
	} else if start&0b0010_0000 == 0 {
		return 2
	} else if start&0b0001_0000 == 0 {
		return 3
	} else if start&0b0000_1000 == 0 {
		return 4
	} else if start&0b0000_0100 == 0 {
		return 5
	} else if start&0b0000_0010 == 0 {
		return 6
	} else if start&0b0000_0001 == 0 {
		return 7
	} else {
		return 8
	}
}

func runeAtByteInString(s string, b int) (r rune, l int) {
	l = bytesInRune(s[b])
	bytes := make([]byte, l)

	for i := 0; i < l; i++ {
		bytes[i] = s[i+b]
	}

	r, l = utf8.DecodeRune(bytes)

	return
}

// import "strings"
// import "unicode/utf8"

// var s = "abc"

// type BufferedString struct {
// 	s     string
// 	// reader strings.IndexByte()
// 	b1   []rune
// 	b2   []rune
// 	b1i  int
// 	b2i  int
// 	move1 bool
// }

// func NewBufferedString(s string, size int) BufferedString {
// 	new := BufferedString{s: s, b1: make([]rune, 0, size), b2: make([]rune, 0, size)}
// 	return new
// }

// func (bs *BufferedString) charAt(i int) {
// 	[]byte(s)

// 	size := len(bs.b1)
// 	use(size)
// 	// if (strings.NewReader(

// }

// func use(...any) {}

// func (s string) runeAtByte(int byteIndex)

// func runeSize(b byte) int {

// }
