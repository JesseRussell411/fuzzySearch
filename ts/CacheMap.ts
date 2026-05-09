export class CacheMap<K, V> extends Map<K, V> {
    private _maxSize: number;
    public get maxSize() {
        return this._maxSize;
    }
    public set maxSize(value: number) {
        while (this.size >= value) {
            for(const key of this.keys()) {
                this.delete(key);
                break;
            }
        }
        this._maxSize = value;
    }

    public constructor(maxSize: number) {
        super();
        this._maxSize = maxSize;
    }

    public set(key: K, value: V) {
        if (this.size >= this.maxSize) {
            for (const key of this.keys()) {
                this.delete(key);
                break;
            }
        }
        if (this.size >= this.maxSize) {
            throw Error("fuck!");
        }

        return super.set(key, value);
    }
}