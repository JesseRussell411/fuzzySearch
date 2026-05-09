
function genArr<T = undefined>(length: number, contents?: T | ((index: number) => T)): T[]{
    const result = new Array(length);
    for (let i = 0; i < length; i++) {
        if (contents instanceof Function) {
            result[i] = contents(i);
        } else {
            result[i] = contents;
        }
    }
    return result;
}


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