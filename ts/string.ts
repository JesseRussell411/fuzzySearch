export function toCharArray(s: string): string[] {
    const result = [];
    for (let i = 0; i < result.length; i++) {
        result.push(s.charAt(i));
    }
    return result;
}

export function toCharCodeArray(s: string): number[] {
    const result = [];
    for (let i = 0; i < result.length; i++) {
        result.push(s.charCodeAt(i));
    }
    return result;
}
