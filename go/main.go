package main

import (
	"fmt"
	"math"
	"os"
	"time"
	"unicode/utf8"
	"unsafe"
)

type FuzzySearchMatch struct {
	minimumEditDistance int
	score               float64
	runeOffset          int
	byteOffset          int
	runeCount           int
	byteCount           int
}

type FuzzySearchParams struct {
	testString          string
	searchString        string
	minimumScore        float64
	maximumEditDistance int
	cacheDepth          int
	takeProgress        func(FuzzySearchProgress) bool
}

type FuzzySearchProgress struct {
	byteOffset int
	runeOffset int
	bestMatch  FuzzySearchMatch
}

func FuzzySearch(test, search string) FuzzySearchMatch {
	return FuzzySearchWith(test, search).Run()
}
func FuzzySearchWith(test, search string) FuzzySearchParams {
	return FuzzySearchParams{
		testString:          test,
		searchString:        search,
		minimumScore:        0.0,
		maximumEditDistance: math.MaxInt,
		cacheDepth:          2,
	}
}

func (self FuzzySearchParams) Run() FuzzySearchMatch {
	result := fuzzySearchFromBuilder(self)
	return result
}

func (self FuzzySearchParams) MinScore(value float64) FuzzySearchParams {
	self.minimumScore = value
	return self
}

func (self FuzzySearchParams) MaxEditDistance(value int) FuzzySearchParams {
	self.maximumEditDistance = value
	return self
}

func (self FuzzySearchParams) CacheDepth(value int) FuzzySearchParams {
	self.cacheDepth = value
	return self
}

func (self FuzzySearchParams) TakeProgress(value func(FuzzySearchProgress) bool) FuzzySearchParams {
	self.takeProgress = value
	return self
}

type rowCache struct {
	row      []int
	children map[rune]rowCache
}

var nilCache rowCache = rowCache{children: make(map[rune]rowCache)}

func calcScore(editDistance, searchLength int) float64 {
	matchCount := searchLength - editDistance
	score := float64(matchCount) / float64(searchLength)
	return score
}

func fuzzySearchFromBuilder(params FuzzySearchParams) FuzzySearchMatch {
	//#region take options
	test := params.testString
	search := params.searchString
	minimumScore := 0.0
	maximumEditDistance := math.MaxInt
	cacheDepth := params.cacheDepth
	minimumScore = max(0.0, min(1.0, params.minimumScore))
	maximumEditDistance = max(0, params.maximumEditDistance)
	takeProgress := params.takeProgress
	//#endregion

	searchLength := utf8.RuneCountInString(search)

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
	rootCache := rowCache{children: make(map[rune]rowCache)}

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

		prevCache := rootCache

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
			testRune, _ := runeAtByteInString(test, tb)
			// ED matrix row
			r := wl
			cache, cacheHit := prevCache.children[testRune]

			if cacheHit {
				prevRow = cache.row
				prevCache = cache
			} else {
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
							// searchByte := unsafeBoundlessStringGet(search, uintptr(sb+rb))
							// testByte := unsafeBoundlessStringGet(test, uintptr(tb+rb))
							// match = searchByte == testByte
							match = search[sb+rb] == test[tb+rb]
							if !match {
								break
							}
						}
					}

					// searchRune, _ := runeAtByteInString(search, sb)
					// match := testRune == searchRune
					//#endregion

					if match {
						// row[c] = prevRow[c-1]
						editDist := unsafeBoundlessSliceGet_int(prevRow, uintptr(c-1))
						unsafeBoundlessSliceSet_int(row, uintptr(c), editDist)
					} else {
						// north := prevRow[c]
						// northWest := prevRow[c-1]
						// west := row[c-1]

						// row[c] = 1 + min(north, northWest, west)

						north := unsafeBoundlessSliceGet_int(prevRow, uintptr(c))
						northWest := unsafeBoundlessSliceGet_int(prevRow, uintptr(c-1))
						west := unsafeBoundlessSliceGet_int(row, uintptr(c-1))

						unsafeBoundlessSliceSet_int(row, uintptr(c), 1+min(north, northWest, west))
					}

					si++
					sb += searchRuneLength
					if sb >= len(search) || si >= searchLength {
						break
					}
				}

				// populate cache
				if r <= cacheDepth {
					cacheRow := make([]int, len(row))
					copy(cacheRow, row)
					cache = rowCache{row: cacheRow, children: make(map[rune]rowCache)}
					prevCache.children[testRune] = cache
					prevCache = cache
				} else {
					prevCache = nilCache
				}
				// rotate rows
				prevRow = row
				temp := row
				row = nextRow
				nextRow = temp
			}

			// editDist := prevRow[searchLength]
			editDist := unsafeBoundlessSliceGet_int(prevRow, uintptr(searchLength))

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

		//#region progress report
		if takeProgress != nil {
			bestMatch := FuzzySearchMatch{minimumEditDistance: searchLength}

			if minimumEditDistance <= maximumEditDistance {
				score := calcScore(minimumEditDistance, searchLength)
				if score >= minimumScore {
					bestMatch.minimumEditDistance = minimumEditDistance
					bestMatch.score = score

					bestMatch.byteCount = byteCount
					bestMatch.byteOffset = byteOffset
					bestMatch.runeCount = runeCount
					bestMatch.runeOffset = runeOffset
				}
			}

			progress := FuzzySearchProgress{
				byteOffset: wb,
				runeOffset: wi,
				bestMatch:  bestMatch,
			}

			stop := takeProgress(progress)
			if stop {
				break
			}
		}
		//#endregion

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

	score := calcScore(minimumEditDistance, searchLength)

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

	slice := []int{1, 2, 3}

	println("slice item:", unsafeBoundlessSliceGet_int(slice, 1))
	unsafeBoundlessSliceSet_int(slice, 1, 20)
	println("slice item now:", unsafeBoundlessSliceGet_int(slice, 1))

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

		fsm := FuzzySearchWith(
			bigS,
			search,
		).CacheDepth(2).MinScore(0.7).Run()
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

func unsafeBoundlessSliceGet_int(slc []int, i uintptr) int {
	start := uintptr(unsafe.Pointer(&slc[0]))
	target := start + unsafe.Sizeof(int(0))*i
	targetPointer := unsafe.Pointer(target)
	targetValue := *(*int)(targetPointer)
	return targetValue
}

func unsafeBoundlessSliceSet_int(slc []int, i uintptr, value int) {
	start := uintptr(unsafe.Pointer(&slc[0]))
	target := start + unsafe.Sizeof(int(0))*i
	targetPointer := unsafe.Pointer(target)
	*(*int)(targetPointer) = value
}

func unsafeBoundlessStringGet(str string, i uintptr) byte {
	data := unsafe.StringData(str)
	target := uintptr(unsafe.Pointer(data)) + i
	targetPointer := unsafe.Pointer(target)
	targetValue := *(*byte)(targetPointer)
	return targetValue
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
