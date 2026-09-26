package main

import (
	"fmt"
	"math"
	"sync"
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

func EmptyFuzzySearchMatch(search string) FuzzySearchMatch {
	searchLen := utf8.RuneCountInString(search)
	score := 0.0
	minimumEditDistance := searchLen
	if searchLen > 0 {
		score = 1.0
	}
	return FuzzySearchMatch{
		minimumEditDistance: minimumEditDistance,
		score:               score,
		runeOffset:          0,
		byteOffset:          0,
		runeCount:           0,
		byteCount:           0,
	}
}

type FuzzySearchParams struct {
	testString          string
	searchString        string
	minimumScore        float64
	maximumEditDistance int
	cacheDepth          int
	takeProgress        func(FuzzySearchProgress) bool
	takeTestRuneCount   func(int)
	rootCache           *rowCache
	byteCount           int
	threadCount         int
	searchLength        int
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

func (self FuzzySearchParams) TakeTestRuneCount(value func(int)) FuzzySearchParams {
	self.takeTestRuneCount = value
	return self
}

func (self FuzzySearchParams) ThreadCount(value int) FuzzySearchParams {
	self.threadCount = value
	return self
}

func (self FuzzySearchParams) priv_SearchLength(value int) FuzzySearchParams {
	self.searchLength = value
	return self
}

func (self FuzzySearchParams) priv_RootCache(value *rowCache) FuzzySearchParams {
	self.rootCache = value
	return self
}

type rowCache struct {
	row      []int
	children sync.Map
}

var nilCache *rowCache = &rowCache{}

func newRowCache() *rowCache {
	return &rowCache{}
}

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
	rootCache := params.rootCache
	takeTestRuneCount := params.takeTestRuneCount
	if rootCache == nil {
		rootCache = newRowCache()
	}
	threadCount := params.threadCount
	searchLength := params.searchLength
	//#endregion

	if searchLength <= 0 {
		searchLength = utf8.RuneCountInString(search)
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

		if takeTestRuneCount != nil {
			testRuneCount := utf8.RuneCountInString(test)
			takeTestRuneCount(testRuneCount)
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

	//#region multi-threading
	if threadCount > 1 {
		substrings := BreakIntoSubstrings_utf8(test, threadCount)
		resultChannels := make([]chan FuzzySearchMatch, threadCount)
		runeCounts := make([]int, threadCount)

		for i, substring := range substrings {
			c := Gorun(
				FuzzySearchWith(substring, search).
					ThreadCount(0).
					priv_SearchLength(searchLength).
					TakeTestRuneCount(func(c int) { runeCounts[i] = c }).
					Run,
			)
			resultChannels[i] = c
		}

		result := FuzzySearchMatch{
			minimumEditDistance: searchLength,
		}

		// TODO! check the sections that are cut in half by splitting the string
		// the ,quick,brown, fox,jumps, over, the ,lazy ,dog
		//  |      |
		// this section in between the split
		// this is gonna be really hard
		// and even harder with utf-8

		byteOffset := 0
		runeOffset := 0
		for i, c := range resultChannels {
			subResult := <-c
			if subResult.minimumEditDistance < result.minimumEditDistance {
				subResult.byteOffset += byteOffset
				subResult.runeOffset += runeOffset
				result = subResult
			}

			byteOffset += len(substrings[i])
			runeOffset += runeCounts[i]
		}
	}
	//#endregion

	columnCount := searchLength + 1
	seedRow := make([]int, columnCount)
	row := make([]int, columnCount)
	nextRow := make([]int, columnCount)
	for i := range seedRow {
		seedRow[i] = i
	}

	search_utf32 := make([]byte, searchLength*4)
	searchRuneLengths := make([]int, searchLength)

	sb := 0
	i := 0
	for {
		runeLength := BytesInRune_utf8(search[sb])
		searchRuneLengths[i] = runeLength
		for runeByte := range runeLength {
			search_utf32[i*4+runeByte] = search[sb+runeByte]
		}

		sb += runeLength
		i++

		if i >= searchLength {
			break
		}
	}
	fmt.Println(search_utf32)
	fmt.Println(searchRuneLengths)

	prevRow := seedRow

	// window byte offset
	wb := 0
	// window rune offset
	wi := 0

	potentialEditDistanceForWI := 0
	for {
		// byte count of rune at start of window
		windowRuneLength := BytesInRune_utf8(test[wb])

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
			testRune, _ := RuneAtByteInString(test, tb)
			// ED matrix row
			r := wl
			cacheAny, cacheHit := prevCache.children.Load(testRune)

			if cacheHit {
				cache := cacheAny.(*rowCache)
				prevRow = cache.row
				prevCache = cache
			} else {
				// populate seed column
				row[0] = r

				for si := range searchLength {
					// searchRuneLength := BytesInRune_utf8(search[sb])
					searchRuneLength := searchRuneLengths[si]
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
							match = search_utf32[si*4+rb] == test[tb+rb]
							if !match {
								break
							}
						}
					}

					// searchRune, _ := runeAtByteInString(search, sb)
					// match := testRune == searchRune
					//#endregion
					editDistMatched := UnsafeBoundlessSliceGet_int(prevRow, uintptr(c-1))

					north := UnsafeBoundlessSliceGet_int(prevRow, uintptr(c))
					northWest := UnsafeBoundlessSliceGet_int(prevRow, uintptr(c-1))
					west := UnsafeBoundlessSliceGet_int(row, uintptr(c-1))
					editDistUnMatched := 1 + min(north, northWest, west)

					// condition at the end slightly faster
					if match {
						UnsafeBoundlessSliceSet_int(row, uintptr(c), editDistMatched)
					} else {
						UnsafeBoundlessSliceSet_int(row, uintptr(c), editDistUnMatched)
					}
				}

				// populate cache
				if r <= cacheDepth {
					cacheRow := make([]int, len(row))
					copy(cacheRow, row)
					cache := &rowCache{row: cacheRow}
					// prevCache.children[testRune] = cache
					prevCache.children.Store(testRune, cache)
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
			editDist := UnsafeBoundlessSliceGet_int(prevRow, uintptr(searchLength))

			if editDist <= minimumEditDistanceFromWI {
				minimumEditDistanceFromWI = editDist
				lOfMinimumEditDistanceFromWI = wl
				bcOfMinimumEditDistanceFromWI = wbc
			}

			// clamp window size
			potentialEditDist := wl - searchLength + 1
			if potentialEditDist > minimumEditDistance || potentialEditDist > appliedMaximumEditDistance {
				break
			}

			tb += testRuneLength
			ti++
			wl++

			testRuneLength = BytesInRune_utf8(test[tb])
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

	if takeTestRuneCount != nil {
		testRuneCount := wi - 1 + utf8.RuneCountInString(test[wb:])
		takeTestRuneCount(testRuneCount)
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
