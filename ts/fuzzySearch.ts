import { genArr } from "./array";

export type FuzzySearchMatch = {
    index: number,
    length: number,
    score: number,
    minimumEditDistance: number,
}

let seedRow = genArr(8, i => i);
let currentRow = genArr(8, 0);
let nextCurrentRow = genArr(8, 0);

export function fuzzySearch<
    MS extends (number | undefined) = undefined,
    MED extends (number | undefined) = undefined
>(
    test: string,
    search: string,
    other?: {
        minimumScore?: MS,
        maximumEditDistance?: MED,
        startingIndex?: number,
        mode?: "bestMatch" | "firstMatch"
        /**
         * @param index The index that is about to be searched.
         * @param bestMatch The best match found so far.
         * @returns If true: stops searching.
         */
        onProgress?: (index: number, bestMatch: FuzzySearchMatch) => boolean | undefined
    }
): (MS extends number ? (MS extends 0 ? never : undefined)
        : MED extends number ? undefined
            : never)
| FuzzySearchMatch {
    type Return = ReturnType<typeof fuzzySearch<MS, MED>>;
    //#region parameters
    /** minimum score allowed */
    const minimumScore = other?.minimumScore ?? 0;
    /** maximum edit distance allowed */
    const maximumEditDistance = other?.maximumEditDistance ?? Infinity;
    const startingIndex = other?.startingIndex ?? 0;

    /** minimum number of matching characters that achieves the minimum score */
    const minimumMatchesForScore = Math.ceil(search.length * minimumScore);
    /** maximum edit distance that achieves the minimum score */
    const maximumEditDistanceForScore = search.length - minimumMatchesForScore;

    /** maximum edit distance allowed that achieves the minimum score */
    const appliedMaximumEditDistance = Math.min(
        maximumEditDistance,
        maximumEditDistanceForScore,
    )
    //#endregion

    // find the first match
    //    ... the empty string
    /** index of best match so far */
    let index = 0;
    /** length of best match so far */
    let length = 0;
    /** min edit dist of best match so far */
    let minimumEditDistance = search.length;

    //#region easy guards
    if (search === "" || test === "") {
        return {
            index,
            length,
            score: search === "" ? 1 : 0,
            minimumEditDistance
        };
    }
    // /!\ at this point it's assumed search and test both contain at least 1 character /!\

    if (search.length - test.length > appliedMaximumEditDistance){
        return undefined as Return;
    }

    // indexOf is 2 orders of magnitude faster than this function
    // this means it can check for a perfect match ahead of time
    // and if it's a perfect match, then fuzzy searching won't find anything better
    // so short circuit and return here
    const indexOf_index = test.indexOf(search);
    if (indexOf_index >= 0) {
        return {
            index: indexOf_index,
            length: search.length,
            minimumEditDistance: 0,
            score: 1
        }
    } else if (appliedMaximumEditDistance === 0) {
        return undefined as Return;
    }

    // /!\ at this point it's known that the lowest edit distance to be found is 1 /!\
    //#endregion

    //     a b c d e f <- search string
    //  [0 1 2 3 4 5 6] <- seedRow (and initial prevRow)
    // a 1 0 1 2 3 4 5
    // z[2 1 1 2 3 4 5] <- prevRow
    // c[3 2 2 1 2 3 4] <- currentRow
    // f 4 3 3 2 2 3(3) <- minimum edit distance for length of substring so far
    // g 5 4 4 3 3 3(4) <- minimum edit distance for length of substring so far
    // g 6 5 5 4 4 4(4) <- minimum edit distance for length of substring so far
    // ^
    //  ` - substring of test
    const columnCount = search.length + 1;
    if (seedRow.length < columnCount) {
        seedRow = genArr(columnCount, i => i);
        currentRow = genArr(columnCount, 0);
        nextCurrentRow = genArr(columnCount, 0);
    }
    let prevRow = seedRow

    /** the best possible edit distance to find at the window's current position */
    let potentialEditDistanceForI = 0;
    substringIndex: for (let i = startingIndex; i < test.length; i++) {
        // check if this window position is worth searching
        if (
            potentialEditDistanceForI >= minimumEditDistance
            || potentialEditDistanceForI > appliedMaximumEditDistance
        ) {
            potentialEditDistanceForI -= 2;
            continue;
        }

        //#region update progress
        const stop = other?.onProgress?.(
            i,
            {index, length, minimumEditDistance, score: calcScore(minimumEditDistance)}
        );
        if (stop) break;
        //#endregion

        /** the best edit distance found at this window position
         * used to update `potentialEditDistanceForI`
         */
        let minimumEditDistanceFromI = Infinity;
        let lOfMinimumEditDistanceFromI = 0;

        //#region expand window
        /** length of substring of test so far */
        let l = 1;
        for (; l <= test.length - i; l++) {
            /** index within test */
            const ti = i + l - 1;
            /** the current row in the matrix */
            const r = l;


            //     a b c d e f
            //   0 1 2 3 4 5 6
            // a 1 0 1 2 3 4 5
            // z 2 1 1 2 3 4 5
            // c 3 2 2 1 2 3 4
            // f 4 3 3 2 2 3 3
            //   ^
            //    `- populate the seed column
            currentRow[0] = r;

            /** length of search string so far (the column of the matrix) */
            let sl = 1;
            for (;sl <= search.length; sl++) {
                /** index within search */
                let si = sl - 1;
                /** the current column in the matrix */
                const c = sl;

                const searchChar = search.charAt(si);
                const testChar = test.charAt(ti);
                if (searchChar === testChar) {
                    currentRow[c] = prevRow[c - 1];
                } else {
                    const north = prevRow[c];
                    const northWest = prevRow[c - 1];
                    const west = currentRow[c - 1];
                    currentRow[c] = 1 + Math.min(north, northWest, west);
                }
            }

            const editDist = currentRow[search.length];

            prevRow = currentRow;


            // rotate rows
            let temp = currentRow;
            currentRow = nextCurrentRow;
            nextCurrentRow = temp;
            // const editDist = currentRow[search.length];
            if (editDist <= minimumEditDistanceFromI) {
                minimumEditDistanceFromI = editDist
                lOfMinimumEditDistanceFromI = l
            }

            // clamp window size
            /** The best potential edit distance to be found by continuing this loop */
            const potentialEditDist = editDist - (search.length - l);
            if (
                potentialEditDist > minimumEditDistance
                || potentialEditDist > appliedMaximumEditDistance
            ){
                break;
            }
        }
        //#endregion
        if (minimumEditDistanceFromI < minimumEditDistance) {
            // better match found

            minimumEditDistance = minimumEditDistanceFromI;
            index = i;
            length = lOfMinimumEditDistanceFromI;
        }

        // edit distance can't get better than 1 thanks to the indexOf check at the start
        if (minimumEditDistance <= 1) break substringIndex;

        //#region firstMatch mode
        if (other?.mode === "firstMatch" && minimumEditDistance <= appliedMaximumEditDistance) {
            const score = calcScore(minimumEditDistance)
            if (score >= minimumScore) {
                if (length > search.length){
                    const bestWithinFirst = fuzzySearch(
                        test.substring(
                            index,
                            // TODO is this right?
                            index + length + minimumEditDistance
                        ),
                        search,
                        {
                            maximumEditDistance: maximumEditDistance,
                            minimumScore: minimumScore,
                        }
                    );
                    if (bestWithinFirst === undefined) {
                        return {index, length, minimumEditDistance, score}
                    } else {
                        return {
                            ...bestWithinFirst,
                            index: index + bestWithinFirst?.index
                        }
                    }
                } else {
                    return {index, length, minimumEditDistance, score}
                }
            }
        }
        //#endregion

        // reset rows
        prevRow = seedRow;

        // potential edit distance of next window position
        if (minimumEditDistanceFromI === Infinity) {
            potentialEditDistanceForI -= 2;
        } else {
            potentialEditDistanceForI = minimumEditDistanceFromI - 2;
        }
    }

    const score = calcScore(minimumEditDistance);

    if (score >= (minimumScore) && minimumEditDistance <= (maximumEditDistance)) {
        return {index, length, score, minimumEditDistance};
    } {
        return undefined as Return;
    }


    function calcScore(editDistance: number) {
        // it is assumed that search.length is > 0
        const matchCount = search.length - editDistance
        const result = matchCount / search.length;
        return result;
    }
}
