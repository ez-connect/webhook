import { resolve, isAbsolute } from 'node:path';
import { pathToFileURL } from 'node:url';

import { defaultConfig, type Config, type HookConfig } from './config';
import { PluginManager } from './plugin/plugin-manager';

let pluginManager: PluginManager;

export async function serve(config: Config = defaultConfig) {
	const pm = new PluginManager(config.server.maxPoolSize);

	const { hooks = {} } = config;
	// Register all configured hooks with their plugins (resolve plugin paths to file URLs)
	for (const [path, hook] of Object.entries(hooks)) {
		let pluginPath = hook.plugin;
		// If this is already a file: URL, use it directly. Otherwise resolve to an absolute path
		// and convert to a file: URL so dynamic import can resolve it reliably.
		if (!pluginPath.startsWith('file:')) {
			if (!isAbsolute(pluginPath)) {
				// Resolve relative paths against the current working directory
				// which is the typical location the server is launched from.
				pluginPath = resolve(process.cwd(), pluginPath);
			}
			pluginPath = pathToFileURL(pluginPath).href;
		}

		pm.registerPlugin(path, pluginPath);
	}

	pluginManager = pm;

	const { hostname = '0.0.0.0', port = 3000 } = config.server;
	const server = Bun.serve({
		hostname,
		port,
		routes: {
			'/': new Response('Dashboard'),
			'/health': new Response('OK'),
			'/:hook': (req) => handleHook(req, hooks),
		},
		fetch(_: Request) {
			return new Response('Not found', { status: 404 });
		},
	});

	console.log(`Webhook server running on http://${hostname}:${port}`);
	for (const [path, hook] of Object.entries(hooks)) {
		console.log(`  ${path} --> ${hook.plugin}`);
	}

	return server;
}

async function handleHook(
	req: Request,
	hooks: Record<string, HookConfig>,
): Promise<Response> {
	const url = new URL(req.url);
	const path = url.pathname;

	if (req.method !== 'POST') {
		return new Response('Method not allowed', { status: 405 });
	}

	if (!(path in hooks)) {
		return new Response('Not found', { status: 404 });
	}

	console.log(`POST ${path}`);

	try {
		// Run the plugin
		const response = await pluginManager.runPlugin(path, req);
		return response;
	} catch (error) {
		const message = error instanceof Error ? error.message : String(error);
		console.error(`Error handling hook '${path}':`, message);

		return new Response(
			JSON.stringify({
				error: 'Failed to process hook',
				message,
			}),
			{
				status: 500,
				headers: { 'Content-Type': 'application/json' },
			},
		);
	}
}
