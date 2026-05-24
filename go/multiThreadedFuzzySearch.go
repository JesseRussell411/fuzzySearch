package main

type Output struct {
	result FuzzySearchMatch
	index  int
}

func multiThreadedFuzzySearch(test, search string) FuzzySearchMatch {
	strings := splitStringMultipleTimes(test, 4)
	resultChannels := make([]chan FuzzySearchMatch, 0)
	rowCache := newRowCache()

	for _, s := range strings {
		c := make(chan FuzzySearchMatch)
		resultChannels = append(resultChannels, c)
		go (func() {
			result := FuzzySearchWith(s, search).setRootCache(rowCache).Run()
			c <- result
		})()
	}

	bestMatch := EmptyFuzzySearchMatch(search)

	for _, c := range resultChannels {
		match := <-c

		if match.minimumEditDistance < bestMatch.minimumEditDistance {
			bestMatch = match
		}
		if match.minimumEditDistance == 0 {
			break
		}
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

	for !IsValidStartingByte_utf8(s[byteOffset]) {
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
