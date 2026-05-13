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
	score               float64
	runeOffset          int
	byteOffset          int
	runeCount           int
	byteCount           int
}

type FuzzySearchOptions struct {
	minimumScore        float64
	maximumEditDistance int
}

func fuzzySearchWithMaximumEditDistance(test, search string, maximumEditDistance int) FuzzySearchMatch {
	return fuzzySearchWithOptions(test, search, FuzzySearchOptions{minimumScore: 0.0, maximumEditDistance: maximumEditDistance})
}
func fuzzySearchWithMinimumScore(test, search string, minimumScore float64) FuzzySearchMatch {
	return fuzzySearchWithOptions(test, search, FuzzySearchOptions{minimumScore: minimumScore, maximumEditDistance: -1})
}
func fuzzySearch(test, search string) FuzzySearchMatch {
	return fuzzySearchWithOptions(test, search, FuzzySearchOptions{minimumScore: 0.0, maximumEditDistance: -1})
}
func fuzzySearchWithOptions(test, search string, options FuzzySearchOptions) FuzzySearchMatch {
	searchLength := utf8.RuneCountInString(search)

	minimumScore := 0.0
	maximumEditDistance := math.MaxInt

	minimumScore = max(0.0, min(1.0, options.minimumScore))
	if options.maximumEditDistance >= 0 {
		maximumEditDistance = options.maximumEditDistance
	}

	minimumMatchesForMinimumScore := math.Ceil(minimumScore * float64(searchLength))
	maximumEditDistanceForMinimumScore := searchLength - int(minimumMatchesForMinimumScore)

	appliedMaximumEditDistance := min(maximumEditDistance, maximumEditDistanceForMinimumScore)

	// default values for empty string
	minimumEditDistance := searchLength
	runeOffset := 0
	byteOffset := 0
	runeCount := 0
	byteCount := 0

	if search == "" || test == "" {
		var score float64
		if search == "" {
			score = 1.0
		} else {
			score = 0.0
		}

		return FuzzySearchMatch{
			minimumEditDistance: minimumEditDistance,
			score:               score,
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

	// window byte offset
	wb := 0
	// window rune offset
	wi := 0

	potentialEditDistanceForWI := 0
	for {
		// byte count of rune at start of window
		windowRuneLength, _ := bytesInRune_Utf8(test[wb])

		// byte count of current rune in test
		testRuneLength := windowRuneLength

		// edit distance of best substring in window (the one with the lowest ED)
		minimumEditDistanceFromWI := math.MaxInt
		// rune count of best substring in window
		lOfMinimumEditDistanceFromWI := 0
		// byte count of best substring in window
		bcOfMinimumEditDistanceFromWI := 0

		// window byte count
		wbc := 0

		// test byte offset
		tb := wb
		// test rune offset
		ti := wi
		// window length
		wl := 1

		// check if it's worth checking the windows position
		if potentialEditDistanceForWI >= minimumEditDistance || potentialEditDistanceForWI > appliedMaximumEditDistance {
			// if not: skip to the next one
			goto nextWindowPosition // this is all the way down here because goto's can't jump over variable declarations
		}

		for {
			wbc += testRuneLength
			if wbc >= len(test)-wb {
				break
			}
			// ED matrix row
			r := wl

			// populate seed column
			row[0] = r

			// search rune offset
			si := 0
			// search byte offset
			sb := 0
			for {
				searchRuneLength, _ := bytesInRune_Utf8(search[sb])
				c := si + 1

				// instead of converting the bytes in search and test
				// to runes, just check if the bytes match
				// #region check bytes for match
				match := searchRuneLength == testRuneLength

				if match {
					for rb := 0; rb < testRuneLength; rb++ {
						match = search[sb+rb] == test[tb+rb]
						if !match {
							break
						}
					}
				}

				// searchRune, _ := runeAtByteInString(search, sb)
				// testRune, _ := runeAtByteInString(test, tb)
				// match := testRune == searchRune
				//#endregion

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
				if sb >= len(search) || si >= searchLength {
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
			if potentialEditDist > minimumEditDistance || potentialEditDist > appliedMaximumEditDistance {
				break
			}

			tb += testRuneLength
			ti++
			wl++

			testRuneLength, _ = bytesInRune_Utf8(test[tb])
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

	nextWindowPosition:
		// update potential edit distance for window position
		if minimumEditDistanceFromWI == math.MaxInt {
			potentialEditDistanceForWI -= 2
		} else {
			potentialEditDistanceForWI = minimumEditDistanceFromWI - 2
		}

		wb += windowRuneLength
		wi++

		if wb >= len(test) {
			break
		}
	}

	matchCount := searchLength - minimumEditDistance
	score := float64(matchCount) / float64(searchLength)

	if score < minimumScore || minimumEditDistance > appliedMaximumEditDistance {
		// remove undefined behavior
		result := FuzzySearchMatch{
			minimumEditDistance: searchLength,
		}

		return result
	}

	result := FuzzySearchMatch{
		minimumEditDistance: minimumEditDistance,
		score:               score,
		runeOffset:          runeOffset,
		byteOffset:          byteOffset,
		runeCount:           runeCount,
		byteCount:           byteCount,
	}

	return result
}

func main() {

	data, err := os.ReadFile("../bigS.txt")
	if err != nil {
		panic("couldn't read bigS: " + err.Error())
	}

	// bigS := "oh 𐀄 shit𐀄s this guy 𐀄hello 𐀄" <-- for testing multi byte runes
	bigS := string(data)
	println(bytesInRune_Utf8(bigS[3]))

	search := "duff's device is a thing"
	searchBytes := []byte(search)

	searchBytes[1] = 0b1010_1010
	// searchBytes[2] = 0b1010_1010
	// search = string(searchBytes)
	println("search: ", search)

	for range 20 {
		start := time.Now()

		fsm := fuzzySearchWithMinimumScore(
			bigS,
			search,
			0.7,
		)
		stop := time.Now()
		elapsed := float64(stop.Sub(start).Microseconds()) / 1000.0

		ss := bigS[fsm.byteOffset : fsm.byteOffset+fsm.byteCount]
		fmt.Println(fsm, ss)
		println(elapsed)
	}

}

func use(...any) {
}

func bytesInRune_Utf8(start byte) (count int, validStartingByte bool) {
	validStartingByte = true
	if start&0b1000_0000 == 0 {
		count = 1
	} else if start&0b0100_0000 == 0 {
		validStartingByte = false
		count = 1
	} else if start&0b0010_0000 == 0 {
		count = 2
	} else if start&0b0001_0000 == 0 {
		count = 3
	} else if start&0b0000_1000 == 0 {
		count = 4
	} else if start&0b0000_0100 == 0 {
		count = 5
	} else if start&0b0000_0010 == 0 {
		count = 6
	} else if start&0b0000_0001 == 0 {
		count = 7
	} else {
		count = 8
	}

	return
}

func runeAtByteInString(s string, b int) (r rune, l int) {
	subString := s[b:]
	r, l = utf8.DecodeRuneInString(subString)
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
