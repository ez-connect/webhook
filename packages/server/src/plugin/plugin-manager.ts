import type { Plugin } from '@webhook/core';

import { LRUCache } from './cache';

export class PluginManager {
	private pool: LRUCache<string, Plugin>;
	private pluginPaths: Map<string, string> = new Map();

	constructor(maxPoolSize: number = 10) {
		this.pool = new LRUCache(maxPoolSize, (name: string, plugin: Plugin) => {
			plugin.destroy().catch((error: Error) => {
				console.error(`Error destroying evicted plugin '${name}':`, error);
			});
		});
	}

	registerPlugin(name: string, path: string): void {
		this.pluginPaths.set(name, path);
	}

	async loadPlugin(name: string): Promise<Plugin> {
		// Check if already loaded
		const cached = this.pool.get(name);
		if (cached) {
			return cached;
		}

		const pluginPath = this.pluginPaths.get(name);
		if (!pluginPath) {
			throw new Error(`Plugin '${name}' not found in registry`);
		}

		try {
			// Dynamic import
			const module = await import(pluginPath);
			const PluginClass = module.default;

			if (!PluginClass) {
				throw new Error(`Plugin '${name}' does not export a default class`);
			}

			const plugin = new PluginClass();

			// Initialize plugin
			await plugin.init();

			// Store in pool (LRU will handle eviction)
			this.pool.set(name, plugin);

			return plugin;
		} catch (error) {
			const message = error instanceof Error ? error.message : String(error);
			throw new Error(`Failed to load plugin '${name}': ${message}`);
		}
	}

	async unloadPlugin(name: string): Promise<void> {
		const plugin = this.pool.get(name);
		if (!plugin) {
			return;
		}

		try {
			await plugin.destroy();
		} catch (error) {
			console.error(`Error destroying plugin '${name}':`, error);
		}

		this.pool.delete(name);
	}

	async runPlugin(name: string, request: Request): Promise<Response> {
		const plugin = await this.loadPlugin(name);
		return plugin.run(request);
	}

	getLoadedPlugins(): string[] {
		return this.pool.keys();
	}

	isPluginLoaded(name: string): boolean {
		return this.pool.has(name);
	}

	async unloadAll(): Promise<void> {
		const plugins = this.pool.values();
		for (const plugin of plugins) {
			try {
				await plugin.destroy();
			} catch (error) {
				console.error('Error destroying plugin:', error);
			}
		}
		this.pool.clear();
	}

	getPoolSize(): number {
		return this.pool.size();
	}

	setMaxPoolSize(size: number): void {
		if (size <= 0) {
			throw new Error('Max pool size must be greater than 0');
		}

		// Preserve existing plugin instances where possible.
		// Capture current entries, rebuild a new LRUCache with the new size,
		// and re-insert previously loaded plugins (the LRUCache will evict
		// if the new size is smaller).
		const oldKeys = this.pool.keys();
		const oldValues = this.pool.values();
		const oldEntries: Array<{ name: string; plugin?: Plugin }> = oldKeys.map(
			(k, i) => ({
				name: k,
				plugin: oldValues[i],
			}),
		);

		this.pool = new LRUCache(size, (name: string, plugin: Plugin) => {
			plugin.destroy().catch((error: Error) => {
				console.error(`Error destroying evicted plugin '${name}':`, error);
			});
		});

		for (const entry of oldEntries) {
			if (!entry.plugin) continue;
			try {
				this.pool.set(entry.name, entry.plugin);
			} catch {
				// Ignore migration errors - the new pool will manage evictions.
			}
		}
	}
}

export const pm = new PluginManager();
