import { Descriptor } from "as-wasi/assembly";

export function init(): void {
	// Initialization logic if needed
}

// Read all standard input iteratively
export function run(): void {
	const stdin = new Descriptor(0);
	const input = stdin.readAll();
	if (input == null) return;

	const stdout = new Descriptor(1);
	stdout.write(input);
}

export function destroy(): void {
	// Cleanup logic if needed
}
