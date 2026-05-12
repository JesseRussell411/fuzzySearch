export function toCharArray(s: string): string[] {
    const result = new Array<string>(s.length);
    for (let i = 0; i < result.length; i++) {
        result[i] = s.charAt(i);
    }
    return result;
}

export function toCharCodeArray(s: string): number[] {
    const result = new Array<number>(s.length);
    for (let i = 0; i < result.length; i++) {
        result[i] = s.charCodeAt(i);
    }
    return result;
}
