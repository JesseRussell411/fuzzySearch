package main

import "unicode/utf8"

type Output struct {
	result FuzzySearchMatch
	index  int
}

func Gorun[R any](f func() R) chan R {
	c := make(chan R)
	go (func() {
		result := f()
		c <- result
	})()
	return c
}

func multiThreadedFuzzySearch(test, search string) FuzzySearchMatch {
	searchLength := utf8.RuneCountInString(search)
	strings := BreakIntoSubstrings_utf8(test, 16)
	resultChannels := make([]chan FuzzySearchMatch, len(strings))
	chunkSizes := make([]int, len(strings))

	rowCache := newRowCache()

	for i, s := range strings {
		// c := make(chan FuzzySearchMatch)
		c := Gorun(
			FuzzySearchWith(s, search).
				priv_RootCache(rowCache).
				TakeTestRuneCount(func(runeCount int) {
					chunkSizes[i] = runeCount
				}).
				Run,
		)

		resultChannels[i] = c
		// go (func() {
		// 	result := FuzzySearchWith(s, search).setRootCache(rowCache).Run()
		// 	c <- result
		// })()
	}

	bestMatch := EmptyFuzzySearchMatch(search)

	i := 0
	for ci, c := range resultChannels {
		match := <-c

		if match.minimumEditDistance < bestMatch.minimumEditDistance {
			bestMatch = match
			bestMatch.byteOffset += i
			if ci > 0 {
				bestMatch.runeOffset += chunkSizes[ci-1]
			}
		}
		use(searchLength)

		// if ci < (len(resultChannels) - 1) {
		// 	// fuzzy search split border
		// 	half := (searchLength + 1) * 8

		// 	border := splitString(
		// 		splitStringRoundForward(
		// 			test,
		// 			max(0, i+len(strings[ci])-half),
		// 		).second,
		// 		min(len(test), half),
		// 	).first

		// 	match := FuzzySearchWith(border, search).setRootCache(rowCache).Run()
		// 	if match.minimumEditDistance < bestMatch.minimumEditDistance || (match.minimumEditDistance == bestMatch.minimumEditDistance && match.runeOffset < bestMatch.runeOffset){
		// 		bestMatch = match

		// 		bestMatch.byteOffset += i
		// 		if ci > 0 {
		// 			bestMatch.runeOffset += chunkSizes[ci-1]
		// 		}
		// 	}
		// }

		if bestMatch.minimumEditDistance == 0 {
			break
		}

		i += len(strings[ci])
	}

	return bestMatch
}

func runFuzzySearch(test, search string, output chan FuzzySearchMatch) {
	result := FuzzySearch(test, search)
	output <- result
}

type splitStringResult struct {
	first  string
	second string
}

func splitString(s string, byteOffset int) splitStringResult {
	originalByteOffset := byteOffset

	for IsTrailingByte(s[byteOffset]) {
		if byteOffset <= 0 {
			return splitStringResult{
				first:  s[0:originalByteOffset],
				second: s[originalByteOffset:],
			}
		}
		byteOffset--
	}

	return splitStringResult{
		first:  s[0:byteOffset],
		second: s[byteOffset:],
	}
}

func splitStringRoundForward(s string, byteOffset int) splitStringResult {
	originalByteOffset := byteOffset

	for IsTrailingByte(s[byteOffset]) {
		if byteOffset >= (len(s) - 1) {
			return splitStringResult{
				first:  s[0:originalByteOffset],
				second: s[originalByteOffset:],
			}
		}
		byteOffset++
	}

	return splitStringResult{
		first:  s[0:byteOffset],
		second: s[byteOffset:],
	}
}

func splitStringInHalf(s string) splitStringResult {
	return splitString(s, len(s)/2)
}

func splitStringMultipleTimes(s string, times int) []string {
	if times > 1 {
		result := make([]string, 0)
		splitResult := splitStringInHalf(s)
		splitsResultsFirst := splitStringMultipleTimes(splitResult.first, times-1)
		splitsResultsSecond := splitStringMultipleTimes(splitResult.second, times-1)

		for _, str := range splitsResultsFirst {
			result = append(result, str)
		}
		for _, str := range splitsResultsSecond {
			result = append(result, str)
		}

		return result
	} else if times == 1 {
		splitResult := splitStringInHalf(s)

		return []string{splitResult.first, splitResult.second}
	} else {
		return []string{s}
	}
}
