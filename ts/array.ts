import { requireGreaterThanZero, requireNonNegative, requireSafeInteger } from "./checks";

export function fillArr<T = undefined>(
    arr: T[],
    from: number,
    to: number,
    contents: T | ((index: number) => T)
) {
    for (let i = from; i < to; i++) {
        if (contents instanceof Function) {
            arr[i] = contents(i);
        } else {
            arr[i] = contents;
        }
    }
}

export function genArr<T = undefined>(
    length: number,
    contents?: T | ((index: number) => T)
): T[] {
    const result = [];
    for (let i = 0; i < length; i++) {
        if (contents instanceof Function) {
            result.push(contents(i));
        } else {
            result.push(contents);
        }
    }
    return result;
}

export function extendArr<T>(
    arr: T[],
    targetLength: number,
    contents?: T | ((index:number) => T)
) {
    while (arr.length < targetLength) {
        if (contents instanceof Function) {
            arr.push(contents(arr.length));
        } else {
            arr.push(contents);
        }
    }
}

export function asReadonly<T>(arr: readonly T[]): readonly T[] {
    return arr;
}

export function asWritable<T>(arr: readonly T[]): T[] {
    return arr as T[];
}

/**
 * Copy data from one array to another.
 * @param source What to copy from.
 * @param destination What to copy to (can be the same as source).
 * @param sourceStart Where to copy from.
 * @param destinationStart Where to copy to.
 * @param length How much to copy.
 */
export function arrayCopy(
    source: readonly unknown[],
    destination: unknown[],
    sourceStart: number,
    destinationStart: number,
    length: number
) {
    requireNonNegative(requireSafeInteger(sourceStart));
    requireNonNegative(requireSafeInteger(destinationStart));
    requireGreaterThanZero(requireSafeInteger(length));

    if (sourceStart >= source.length) {
        throw new Error(
            `sourceStart (${sourceStart}) out of range (< source length (${source.length}))`
        );
    }
    if (destinationStart >= destination.length) {
        throw new Error(
            `destinationStart (${destinationStart}) out of range (< destination length (${destination.length}))`
        );
    }

    if (length === 0) return;

    const sourceEnd = sourceStart + length;
    if (sourceEnd > source.length) {
        throw new Error(
            `sourceStart + length out of range (<= source length (${source.length}))`
        );
    }

    const destinationEnd = destinationStart + length;
    if (destinationEnd > destination.length) {
        throw new Error(
            `destinationStart + length out of range (<= destination length (${destination.length}))`
        );
    }

    if (source === destination) {
        destination.copyWithin(destinationStart, sourceStart, sourceEnd);
    } else {
        for (let i = 0; i < length; i++) {
            destination[i + destinationStart] = source[i + sourceStart];
        }
    }
}
