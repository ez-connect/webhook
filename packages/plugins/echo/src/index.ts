import type { Plugin } from '@webhook/core';

export default class Echo implements Plugin {
	name = 'echo';
	version = '1.0.0';
	author = 'plugin-echo';
	description = 'Simple echo plugin that responds with the request body';
	options: Record<string, unknown> = {};

	async init(): Promise<void> {
		// No-op for the echo plugin, but provided for completeness.
	}

	async run(req: Request): Promise<Response> {
		try {
			return Response.json({
				url: req.url,
				method: req.method,
				headers: Object.fromEntries(req.headers),
				body: await req.json(),
			});
		} catch (err) {
			const message = err instanceof Error ? err.message : String(err);
			return new Response(JSON.stringify({ error: 'echo_failed', message }), {
				status: 500,
				headers: { 'Content-Type': 'application/json' },
			});
		}
	}

	async destroy(): Promise<void> {
		// No resources to free for echo plugin
	}
}
