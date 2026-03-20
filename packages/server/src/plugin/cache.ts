export class LRUCache<K, V> {
	private cache: Map<K, V> = new Map();
	private order: K[] = [];
	private maxSize: number;
	private onEvict: (key: K, value: V) => Promise<void> | void;

	constructor(
		maxSize: number,
		onEvict?: (key: K, value: V) => Promise<void> | void,
	) {
		this.maxSize = maxSize;
		this.onEvict = onEvict || (() => {});
	}

	get(key: K): V | undefined {
		if (!this.cache.has(key)) return undefined;

		// Move to end (most recently used)
		this.order = this.order.filter((k) => k !== key);
		this.order.push(key);

		return this.cache.get(key);
	}

	set(key: K, value: V): void {
		// If exists, remove first
		if (this.cache.has(key)) {
			this.order = this.order.filter((k) => k !== key);
		}

		// If at capacity, remove LRU
		if (this.cache.size >= this.maxSize && !this.cache.has(key)) {
			const lruKey = this.order.shift();
			if (lruKey !== undefined) {
				const lruValue = this.cache.get(lruKey);
				if (lruValue !== undefined) {
					this.onEvict(lruKey, lruValue);
				}
				this.cache.delete(lruKey);
			}
		}

		// Add new entry
		this.cache.set(key, value);
		this.order.push(key);
	}

	has(key: K): boolean {
		return this.cache.has(key);
	}

	delete(key: K): boolean {
		this.order = this.order.filter((k) => k !== key);
		return this.cache.delete(key);
	}

	clear(): void {
		this.cache.clear();
		this.order = [];
	}

	size(): number {
		return this.cache.size;
	}

	keys(): K[] {
		return Array.from(this.cache.keys());
	}

	values(): V[] {
		return Array.from(this.cache.values());
	}
}
