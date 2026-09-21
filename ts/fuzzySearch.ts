import { genArr, extendArr } from "./array";
import { toCharArray, toCharCodeArray } from "./string";

export type FuzzySearchMatch = {
    index: number,
    length: number,
    score: number,
    minimumEditDistance: number,
}

let seedRow = genArr(8, i => i);
let currentRow = genArr(8, 0);
let nextCurrentRow = genArr(8, 0);
let searchCharCodes = genArr(8, 0);

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

    // if search is longer than test, the min edit distance between them will be equal to the difference in length.
    if (search.length - test.length > appliedMaximumEditDistance) {
        // in this case, that difference is so great that the min edit distance can't possibly be smaller than the max edit dist, so a match won't be found
        return undefined as Return;
    }

    // if search equals test, obviously test, itself, would be the best match, with an edit dist of 0
    if (search === test) {
        return {
            index: 0,
            length: search.length,
            minimumEditDistance: 0,
            score: 1
        }
    }

    //#endregion

    // hopefully this diagram will help explain what's happening
    // the goal is to calculate the levenshtein distance between search and multiple lengths of the substring at the current index in test
    //
    //     a b c d e f <--search string
    //  [0 1 2 3 4 5 6] <- seedRow (and initial prevRow)
    // a 1 0 1 2 3 4(5) <- - - - - - - - - - - -levenshtein distance between "abcdef" and "a"
    // z 2 1 1 2 3 4(5) <- - - - - - - - - - - -levenshtein distance between "abcdef" and "az"
    // c 3 2 2 1 2 3(4) <- - - - - - - - - - - -levenshtein distance between "abcdef" and "azc"
    // f[4 3 3 2 2 3(3)]<- prevRow - "(3)"<- - -levenshtein distance between "abcdef" and "azcf"
    // g[5 4 4 3 3 3(4)]<- currentRow - "(4)"<--levenshtein distance between "abcdef" and "azcfg"
    // ^
    //  ` - substring of test
    //
    // note that the table in this diagram is the same table produced to calculate the levenshtein distance (edit distance) between "abcdef" and "azcfd"
    // but a serendipitous trait of this table is that in order to calculate the edit distance between these two strings, the edit distance between each sub-length string is also calculated
    // the edit distance between "abcdef" and "a" is found first, then that row in the table is used to calculate the edit distance between "abcdef and "az", then that row is used to calculate the next
    // etc.
    //


    /** How many columns the table has. */
    const columnCount = search.length + 1;

    if (seedRow.length < columnCount) {
        extendArr(seedRow,         columnCount, i => i);
        extendArr(currentRow,      columnCount, 0);
        extendArr(nextCurrentRow,  columnCount, 0);
    }

    /** The previous row */
    let prevRow = seedRow

    //#region cache char codes
    for (let i = 0; i < search.length; i++) {
        searchCharCodes[i] = search.charCodeAt(i);
    }
    //#endregion

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
            const testCharCode = test.charCodeAt(ti);


            //     a b c d e f
            //   0 1 2 3 4 5 6
            // a 1 0 1 2 3 4 5
            // z 2 1 1 2 3 4 5
            // c 3 2 2 1 2 3 4
            // f 4 3 3 2 2 3 3
            //   ^
            //    `- populate the seed column
            currentRow[0] = r;

            // if (Object.hasOwn(firstCharCache, testChar
            /** length of search string so far (the column of the matrix) */
            let sl = 1;
            for (;sl <= search.length; sl++) {
                /** index within search */
                let si = sl - 1;
                /** the current column in the matrix */
                const c = sl;

                const searchChar = searchCharCodes[si];
                if (searchChar === testCharCode) {
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
            const potentialEditDist = l - search.length + 1;
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
