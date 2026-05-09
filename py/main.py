from dataclasses import dataclass
from itertools import repeat

@dataclass
class FuzzySearchResult:
    minimumEditDistance: int
    score: float
    index: int
    length: int


def fuzzySearch(test: str, search: str) -> FuzzySearchResult:
    def calculateScore(editDist: int):
        # it is assumed that len(search) is > 0
        matchCount = len(search) - editDist
        result = matchCount / len(search)
        return result

    # find first match: the empty string
    minimumEditDistance = len(search)
    index = 0
    length = 0

    if test == "" or search == "":
        result = FuzzySearchResult(
            minimumEditDistance=minimumEditDistance,
            score= 1 if search == "" else 0,
            index=index,
            length=length,
        )

        return result

    # ED matrix
    columnCount = len(search) + 1
    seedRow = list(range(0, columnCount))
    currentRow = list(repeat(0, columnCount))
    nextCurrentRow = list(repeat(0, columnCount))
    prevRow = seedRow


    potentialEditDistForI = 0
    stopMovingWindow = False
    # move window
    for i in range(0, len(test)):
        if stopMovingWindow:
            break

        # skip if not worth it
        if potentialEditDistForI >= minimumEditDistance:
            potentialEditDistForI -= 2
            continue

        minimumEditDistanceForL = None
        # grow window
        for l in range(1, len(test) - i + 1):
            ti = i + l - 1
            r = l

            currentRow[0] = r

            for sl in range(1, len(search) + 1):
                si = sl - 1
                c = sl

                searchChar = search[si]
                testChar = test[ti]
                if searchChar == testChar:
                    currentRow[c] = prevRow[c - 1]
                else:
                    north = prevRow[c]
                    northWest = prevRow[c - 1]
                    west = currentRow[c - 1]
                    currentRow[c] = 1 + min(
                        north, northWest, west
                    )
            
            editDist = currentRow[len(search)]
            if minimumEditDistanceForL is None or editDist < minimumEditDistanceForL:
                minimumEditDistanceForL = editDist

            prevRow = currentRow

            if editDist <= minimumEditDistance:
                minimumEditDistance = editDist
                index = i
                length = l
            
            if minimumEditDistance == 0:
                stopMovingWindow = True
                break
            
            # rotate rows
            temp = currentRow
            currentRow = nextCurrentRow
            nextCurrentRow = temp

            potentialEditDist = editDist - (
                len(search) - 1
            )

            # clamp window size
            if potentialEditDist > minimumEditDistance:
                break
        
        # reset rows
        prevRow = seedRow

        if minimumEditDistanceForL is None:
            potentialEditDistForI -= 2
        else:
            potentialEditDistForI = minimumEditDistanceForL - 2
    
    score = calculateScore(minimumEditDistance)

    result = FuzzySearchResult(
        minimumEditDistance=minimumEditDistance,
        score=score,
        index=index,
        length=length
    )

    return result




def main():
    with open("../bigS.txt", "r") as bigF:
        bigS = bigF.read()

    def pfsr(s: str, fsr: FuzzySearchResult):
        print(fsr)
        print(s[fsr.index: fsr.index + fsr.length])


    s = bigS[5098:7000]
    pfsr(s, fuzzySearch(s, "dog"))
    import time

    print("searching bigS...")
    start = time.time()
    pfsr(bigS, fuzzySearch(bigS, "what is a good phrase to put in my fuzzy search?"))
    print("time(seconds): " + str(time.time() - start))


if __name__ == "__main__":
    main()
