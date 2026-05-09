package main

import (
	"fmt"
	"math"
	"os"
	"time"
	"unicode/utf8"
)

type FuzzySearchMatch struct {
	minimumEditDistance int
	runeOffset          int
	byteOffset          int
	runeCount           int
	byteCount           int
}

func fuzzySearch(test, search string) FuzzySearchMatch {
	searchLength := utf8.RuneCountInString(search)

	// default values for empty string
	minimumEditDistance := searchLength
	runeOffset := 0
	byteOffset := 0
	runeCount := 0
	byteCount := 0

	if search == "" || test == "" {
		return FuzzySearchMatch{
			minimumEditDistance: minimumEditDistance,
			runeOffset:          runeOffset,
			byteOffset:          byteOffset,
			runeCount:           runeCount,
			byteCount:           byteCount,
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
		// TODO handle invalid runes
		testRuneLength := windowRuneLength

		minimumEditDistanceFromWI := math.MaxInt
		lOfMinimumEditDistanceFromWI := 0
		bcOfMinimumEditDistanceFromWI := 0

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

			si := 0
			sb := 0
			for {
				// TODO? cache this in a lookup table?
				searchRuneLength := bytesInRune(search[sb])
				c := si + 1

				match := searchRuneLength == testRuneLength

				if match {
					for rb := 0; rb < testRuneLength; rb++ {
						match = search[sb+rb] == test[tb+rb]
						if !match {
							break
						}
					}
				}

				// testRune := runeAtBytesInString(test, tb, testRuneLength)
				// searchRune := runeAtBytesInString(search, sb, searchRuneLength)
				// TODO handle invalid runes
				if match {
					row[c] = prevRow[c-1]
				} else {
					north := prevRow[c]
					northWest := prevRow[c-1]
					west := row[c-1]

					row[c] = 1 + min(north, northWest, west)
				}

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

			if editDist <= minimumEditDistanceFromWI {
				minimumEditDistanceFromWI = editDist
				lOfMinimumEditDistanceFromWI = wl
				bcOfMinimumEditDistanceFromWI = wbc
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
			// TODO handle invalid runes
		}
		// reset rows
		prevRow = seedRow

		if minimumEditDistanceFromWI < minimumEditDistance {
			minimumEditDistance = minimumEditDistanceFromWI
			byteOffset = wb
			runeOffset = wi
			byteCount = bcOfMinimumEditDistanceFromWI
			runeCount = lOfMinimumEditDistanceFromWI
		}

		// check for perfect match
		if minimumEditDistance == 0 {
			// perfect match found
			break
		}

		// update potential edit distance for window position
		if minimumEditDistanceFromWI == math.MaxInt {
			potentialEditDistanceForWI -= 2
		} else {
			potentialEditDistanceForWI = minimumEditDistanceFromWI - 2
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
		runeOffset:          runeOffset,
		byteOffset:          byteOffset,
		runeCount:           runeCount,
		byteCount:           byteCount,
	}

	return result
}

func main() {

	data, err := os.ReadFile("./bigS.txt")
	if err != nil {
		panic("couldn't read bigS: " + err.Error())
	}

	// bigS := "oh 𐀄 shit𐀄s this guy 𐀄 hello 𐀄"
	bigS := string(data)
	println(bytesInRune(bigS[3]))

	start := time.Now()
	fsm := fuzzySearch(bigS, "what is a good phrase to put in my fuzzy search?")
	stop := time.Now()
	elapsed := stop.Sub(start).Milliseconds()

	ss := bigS[fsm.byteOffset : fsm.byteOffset+fsm.byteCount]
	fmt.Println(fsm, ss)
	println(elapsed)

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

	for i := range l {
		bytes[i] = s[i+b]
	}

	r, l = utf8.DecodeRune(bytes)

	return
}

func runeAtBytesInString(s string, b int, l int) rune {
	bytes := make([]byte, l)

	for i := range l {
		bytes[i] = s[i+b]
	}

	r, _ := utf8.DecodeRune(bytes)

	return r
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
