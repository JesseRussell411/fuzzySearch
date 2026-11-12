import { genArr } from "./array.ts";

let currentRow = genArr(8, 0);
let prevRow: readonly number[] = genArr(8, 0);
let seedRow: readonly number[] = genArr(8, i => i);

function minEditDistance_generate(a: string, b: string, other?: {stopAtZero?: boolean}): void
{
    const stopAtZero = other?.stopAtZero ?? false;
    // based on this youtube video: https://www.youtube.com/watch?v=We3YDTzNXEk
    
    //     a b c d e f <- string a
    //   0 1 2 3 4 5 6
    // a 1 0 1 2 3 4 5
    // z 2 1 1 2 3 4 5
    // c 3 2 2 1 2 3 4
    // f 4 3 3 2 2 3(3) <- minimum edit distance
    // ^
    //  ` - string b
    if (currentRow.length < b.length + 1) {
        currentRow = genArr(b.length + 1, 0);
        prevRow = genArr(b.length + 1, 0);
        seedRow = genArr(b.length + 1, i => i);
    }

    let nextCurrentRow = prevRow as number[];
    prevRow = seedRow;


    outer: for (let ai = 0; ai < a.length; ai++) {
        const ri = ai + 1;
        currentRow[0] = ri;

        for (let bi = 0; bi < b.length; bi++) {
            const ci = bi + 1;

            let score = 0;
            if (a.charAt(ai) === b.charAt(bi)) {
                score = prevRow[ci - 1];
            }
            else {
                const w = currentRow[ci - 1];
                const nw = prevRow[ci - 1];
                const n = prevRow[ci];
                score = 1 + Math.min(w, nw, n);
            }
            currentRow[ci] = score;
            if (stopAtZero && score === 0) break outer;
        }
        prevRow = currentRow;
        const temp = currentRow;
        currentRow = nextCurrentRow;
        nextCurrentRow = temp;
    }
    // based on this youtube video: https://www.youtube.com/watch?v=We3YDTzNXEk
    
    //     a b c d e f <- string a
    //   0 1 2 3 4 5 6
    // a 1 0 1 2 3 4 5
    // z 2 1 1 2 3 4 5
    // c 3 2 2 1 2 3 4
    // f 4 3 3 2 2 3(3) <- minimum edit distance
    // ^
    //     a b c d e f
    //   0 1 2 3 4 5 6
    // a 1 0 1 2 3 4 5
    // z 2 1 1 2 3 4 5
    // c 3 2 2 1 2 3 4
    // f 4 3 3 2 2 3(3) <- value to return
}
export function getEditDistance(a: string, b: string) {
    minEditDistance_generate(a, b);

    const result = prevRow[b.length];
    return result;
}

export function getMinimumSubLengthEditDistance(search: string, test: string): {
    length: number,
    editDistance: number,
} {
    if (search.length === 0)
        return {length: 0, editDistance: 0};

    if (test.length === 0)
        return {length: 0, editDistance: search.length};
    
    minEditDistance_generate(search, test, {stopAtZero: true});

    let length = 1;
    let editDistance = prevRow[length];

    for (let l = length; l <= test.length; l++) {
        if (prevRow[l] < editDistance) {
            editDistance = prevRow[l];
            length = l;
            if (editDistance === 0) break;
        }
    }

    return {length, editDistance}
}






export function levensteinGrid(a: string, b: string): string[][] {
    const result = genArr(
        a.length + 1, 
        (y) => {
            if (y === 0) {
                return genArr(b.length + 1, x => x); // seed row
            } else {
                return genArr(
                    b.length + 1,
                    x => x === 0 ? y : 0 // seed column
                );
            }
        }
    );

    const charCodes_a = genArr(a.length, i => a.charCodeAt(i));
    const charCodes_b = genArr(b.length, i => b.charCodeAt(i));

    for (let x = 1; x <= b.length; x++) {
        for (let y = 1; y <= a.length; y++) {
            if (charCodes_a[y-1] === charCodes_b[x-1]) {
                result[y][x] = result[y-1][x-1]
            } else {
                result[y][x] = Math.min(
                    result[y-1][x-0],
                    result[y-0][x-1],
                    result[y-1][x-1],
                ) + 1;
            }
        }
    }


    return result;
}

