package main

import (
	"fmt"
	"math"
	"unicode/utf8"
)

type FuzzySearchMatch struct {
	minimumEditDistance int
	charIndex           int
	byteIndex           int
	charLength          int
	byteLength          int
}

func fuzzySearch(test, search string) FuzzySearchMatch {
	searchLength := utf8.RuneCountInString(search)

	minimumEditDistance := searchLength
	charIndex := 0
	byteIndex := 0
	charLength := 0
	byteLength := 0

	if search == "" || test == "" {
		return FuzzySearchMatch{
			minimumEditDistance: minimumEditDistance,
			charIndex:           charIndex,
			byteIndex:           byteIndex,
			charLength:          charLength,
			byteLength:          byteLength,
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
windowPosition:
	for {
		windowRuneLength := bytesInRune(test[wb])
		testRuneLength := windowRuneLength

		wbc := 0

		tb := wb
		ti := wi
		wl := 1
		r := wl
		minimumEditDistanceForWI := math.MaxInt

		// check if it's worth checking the windows position
		if potentialEditDistanceForWI >= minimumEditDistance {
			// if not: skip to the next one
			potentialEditDistanceForWI -= 2
			goto nextWindowPosition
		}

		for {
			wbc += testRuneLength
			// populate seed column
			row[0] = r

			si := 0
			sb := 0
			for {
				// TODO? cache this in a lookup table?
				searchRuneLength := bytesInRune(search[sb])
				c := si + 1

				testRune, _ := runeAtByteInString(test, tb)
				searchRune, _ := runeAtByteInString(search, sb)
				if searchRune == testRune {
					row[c] = prevRow[c-1]
				} else {
					north := prevRow[c]
					northWest := prevRow[c-1]
					west := row[c-1]

					row[c] = 1 + min(north, northWest, west)
				}
				// rotate rows
				prevRow = row
				temp := row
				row = nextRow
				nextRow = temp

				si++
				sb += searchRuneLength
				if sb > len(search) {
					break
				}
			}
			// reset rows
			prevRow = seedRow

			editDist := row[searchLength]
			minimumEditDistanceForWI = min(minimumEditDistanceForWI, editDist)

			if editDist < minimumEditDistance {
				minimumEditDistance = editDist
				byteIndex = wb
				charIndex = wi
				byteLength = wbc
				charLength = wl
			}

			// clamp window size
			potentialEditDist := editDist - (searchLength - wl)
			if potentialEditDist >= minimumEditDistance {
				break
			}

			// check for perfect match
			if minimumEditDistance == 0 {
				// perfect match found
				break windowPosition
			}

			tb += testRuneLength
			testRuneLength = bytesInRune(test[tb])
			ti++

			wl++
			if wl > searchLength*2 {
				break
			}
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

		if wb > len(test) {
			break
		}
	}

	result := FuzzySearchMatch{
		minimumEditDistance: minimumEditDistance,
		charIndex:           charIndex,
		byteIndex:           byteIndex,
		charLength:          charLength,
		byteLength:          byteLength,
	}

	return result
}

func main() {
	s := "a𐀄bc"

	ss := s[0:1]

	fmt.Println(len("𐀄"))
	fmt.Println()

	for i := 0; i < len(s); {
		r, l := runeAtByteInString(s, i)
		if l == 0 {
			i++
			continue
		}

		fmt.Println(r)
		fmt.Println(l)
		fmt.Println()
		i += l
	}

	use(ss)

	fmt.Println(fuzzySearch("the fox jumps", "fox"))
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
