import { JSON5 } from 'bun';

import { exit } from 'node:process';

export interface Config {
	server: ServerConfig;
	hooks: Record<string, HookConfig>;
}

export interface ServerConfig {
	hostname?: string;
	port?: number;
	maxPoolSize?: number;
}

export interface HookConfig {
	plugin: string;
	options: Record<string, unknown>;
}

export const defaultConfig: Config = {
	server: {
		hostname: '0.0.0.0',
		port: 3000,
		maxPoolSize: 10,
	},
	hooks: {},
};

export async function loadConfig(filename: string): Promise<Config> {
	try {
		const text = await Bun.file(filename).text();
		return JSON5.parse(text) as Config;
	} catch (err) {
		console.error(`Load config error: ${err}`);
		exit(1);
	}
}

export function listHooks(config: Config) {
	for (const [path, { plugin }] of Object.entries(config.hooks)) {
		console.log(path, '-->', plugin);
	}
}
