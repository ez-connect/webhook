export interface Plugin {
	name: string;
	version: string;
	author: string;
	description: string;
	options: Record<string, unknown>;

	init(): Promise<void>;
	run(r: Request): Promise<Response>;
	destroy(): Promise<void>;
}
