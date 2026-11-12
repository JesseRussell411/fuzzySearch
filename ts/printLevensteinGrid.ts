const a = process.argv[2];
const b = process.argv[3];
if (typeof a !== "string" || typeof b !== "string"){
    console.error("please provide 2 strings");
    process.exit()
}

function *enumerate<T>(iterable: Iterable<T>): Iterable<[number, T]> {
    let i = 0;
    for (const item of iterable) {
        yield [i, item];
        i++;
    }
}

import { levensteinGrid } from "./minEditDist";


const grid = levensteinGrid(a, b);


const columnPadding = 1;
const widths = grid[0].map(
    (_, x) => Math.max(
        ...Array.from(
            {length: grid.length},
            (_, y) => String(grid[y][x]).length + columnPadding
        )
    )
);

const line = [] as String[];
const print = (...args: any[]) => line.push(...args);
function flush() {
    console.log(line.join(""));
    line.length = 0;
}


// print b
print(" ");
print(" ".repeat(widths[0]));
for (const [x, c] of enumerate(b)) {
    const width = widths[x];

    print(c.padStart(width));
}
flush()

// print seedRow
print(" ");
for (const [x, n] of enumerate(grid[0])) {
    const width = widths[x];
    print(String(grid[0][x]).padStart(width));
}
flush()

// print the rest of the rows

for (let y = 1; y < grid.length; y++) {
    print(a[y - 1]);
for (let x = 0; x < grid[0].length; x++) {
    const width = widths[x];
    print(String(grid[y][x]).padStart(width));
}
    flush();
}

