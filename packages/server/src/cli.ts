#!/usr/bin/env bun

import { parseArgs } from 'node:util';

import { serve } from './app';
import { exit } from 'node:process';
import { defaultConfig, listHooks, loadConfig } from './config';

function usage(): void {
	console.log(`Webhook version 0.2.0

Usage: webhook <command> [options]

Commands:
  serve		launch the webhook server
  ls			list all registered hooks
  version	show the version number

Options:
  -c, --config <path>	path to the configuration file
`);

	exit(1);
}

export async function main(): Promise<void> {
	let configFile: string | undefined;
	let cmd: string | undefined;

	try {
		const { values, positionals } = parseArgs({
			args: Bun.argv,
			options: {
				config: {
					short: 'c',
					type: 'string',
				},
			},
			strict: true,
			allowPositionals: true,
		});

		configFile = values.config;
		cmd = positionals.at(-1);
	} catch {
		usage();
	}

	const config = configFile ? await loadConfig(configFile) : defaultConfig;
	if (!configFile) {
		console.warn('No config file provided, using default config');
	}

	switch (cmd) {
		case 'serve':
			await serve(config);
			break;
		case 'ls':
			listHooks(config);
			break;
		default:
			usage();
			break;
	}
}
