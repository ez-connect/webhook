# 🎣 Webhook Server

A high-performance webhook server built with [Bun](https://bun.com) with plugin support, lazy loading, and intelligent resource pooling.

## Features

- **🚀 Built with Bun** - Lightning-fast JavaScript runtime
- **🔌 Plugin Architecture** - Easily extend with TypeScript/JavaScript plugins
- **💾 Lazy Loading** - Load plugins only when needed
- **🏊 LRU Pool Management** - Configurable plugin pool with automatic eviction
- **📊 Dashboard** - Beautiful web-based monitoring interface
- **⚡ Hot Reload Ready** - Support for dynamic plugin loading
- **🛠️ CLI Tools** - Easy-to-use command-line interface for plugin management

## Project Structure

```
packages/
├── core/           # Core plugin interface and base classes
│   └── src/
│       └── plugin.ts
├── server/         # Webhook server implementation
│   ├── src/
│   │   ├── index.ts           # CLI entry point
│   │   ├── app.ts             # Bun server implementation
│   │   ├── cli.ts             # CLI commands handler
│   │   ├── config.ts          # Configuration manager
│   │   └── plugin-manager.ts  # Plugin loading and pooling
│   └── webhook.config.json    # Configuration file
├── plugin-echo/    # Example echo plugin
└── plugin-imou/    # Example IMOU plugin
```

## Installation

```bash
# Install dependencies
bun install

# From the server package directory
cd packages/server
bun install
```

## Configuration

Plugins are configured in `webhook.config.json`:

```json
{
  "server": {
    "port": 3000,
    "maxPoolSize": 10
  },
  "hooks": {
    "my-hook": {
      "plugin": "my-plugin-name",
      "path": "./plugins/my-plugin.ts",
      "options": {}
    }
  }
}
```

## Usage

### Starting the Server

```bash
# Start the webhook server
bun run src/index.ts serve

# Or using npm script
bun run dev
```

The server will start on `http://0.0.0.0:3000` (configurable).

### Managing Plugins

#### List All Plugins

```bash
bun run src/index.ts plugins ls
```

Output:
```
📋 Registered Plugins:

Hook Name          │ Plugin Path            │ Options
──────────────────────────────────────────────────────────────
my-hook            │ my-plugin-name         │ {}
another-hook       │ another-plugin         │ {}
```

#### Add a Plugin

```bash
bun run src/index.ts plugins add <hook-name> <plugin-path> [options]
```

Example:
```bash
bun run src/index.ts plugins add imou-webhook packages/plugin-imou/src/index.ts
bun run src/index.ts plugins add proxy ./plugins/proxy.ts '{"target":"http://localhost:8080"}'
```

#### Remove a Plugin

```bash
bun run src/index.ts plugins rm <hook-name>
```

Example:
```bash
bun run src/index.ts plugins rm imou-webhook
```

### Help

```bash
# General help
bun run src/index.ts help

# Plugins management help
bun run src/index.ts plugins help
```

## API Endpoints

### Dashboard

```
GET /
```

Returns HTML dashboard with server statistics and registered hooks.

### Health Check

```
GET /health
```

Returns JSON with server health status:

```json
{
  "status": "ok",
  "timestamp": "2024-01-01T00:00:00.000Z",
  "poolSize": 2
}
```

### Webhook Hook

```
POST /:hook-name
```

Forwards the request to the registered plugin and returns the plugin's response.

Example:
```bash
curl -X POST http://localhost:3000/my-hook \
  -H "Content-Type: application/json" \
  -d '{"data": "value"}'
```

## Creating Plugins

### Plugin Interface

All plugins must extend the `Plugin` class from `@webhook/core`:

```typescript
import { Plugin } from '@webhook/core';

export default class MyPlugin extends Plugin {
  constructor() {
    super({
      name: 'my-plugin',
      version: '1.0.0',
      author: 'Your Name',
      description: 'My awesome webhook plugin',
      options: {},
    });
  }

  override async init() {
    // Initialize plugin resources
    console.log('Plugin initialized');
  }

  override async run(req: Request): Promise<Response> {
    // Handle the webhook request
    if (req.method !== 'POST') {
      return new Response('Method not allowed', { status: 405 });
    }

    const body = await req.json();

    // Process the webhook
    const result = await processWebhook(body);

    return Response.json({
      status: 'success',
      data: result,
    });
  }

  async destroy() {
    // Clean up plugin resources
    console.log('Plugin destroyed');
  }
}
```

### Example: Echo Plugin

```typescript
import { Plugin } from '@webhook/core';

export default class EchoPlugin extends Plugin {
  constructor() {
    super({
      name: 'echo',
      version: '1.0.0',
      description: 'Echo webhook responses',
    });
  }

  override async init() {
    console.log('Echo plugin initialized');
  }

  override async run(req: Request): Promise<Response> {
    const body = await req.text();

    return new Response(body, {
      status: 200,
      headers: { 'Content-Type': 'text/plain' },
    });
  }
}
```

## Plugin Pool Management

The plugin manager uses an LRU (Least Recently Used) cache to manage plugins in memory:

- **Lazy Loading**: Plugins are only loaded when a hook is first called
- **Automatic Eviction**: When the pool reaches max size, the least recently used plugin is evicted
- **Configurable Size**: Set `maxPoolSize` in `webhook.config.json`
- **Cleanup**: Evicted plugins have their `destroy()` method called

Example flow:
```
1. Request arrives for hook 'my-hook'
2. Plugin manager checks if plugin is loaded
3. If not loaded, plugin is dynamically imported and initialized
4. Plugin is added to LRU cache
5. If cache is full, least used plugin is evicted
6. Plugin processes the request
```

## Configuration Reference

### Server Config

```json
{
  "server": {
    "port": 3000,              // Server port (default: 3000)
    "maxPoolSize": 10          // Max plugins in memory (default: 10)
  }
}
```

### Hook Config

```json
{
  "hooks": {
    "hook-name": {
      "plugin": "plugin-id",   // Plugin identifier
      "path": "./path/to/plugin.ts",  // Path to plugin file
      "options": {             // Plugin-specific options
        "key": "value"
      }
    }
  }
}
```

## Performance Tips

1. **Monitor Pool Size**: Use the dashboard to check how many plugins are loaded
2. **Set Appropriate Pool Size**: Balance between memory usage and frequent reloading
3. **Optimize Plugin Init**: Keep initialization lightweight
4. **Use Caching**: Cache expensive operations within plugins

## Troubleshooting

### Plugin Not Loading

Check the configuration:
```bash
bun run src/index.ts plugins ls
```

Ensure the path is correct and the plugin exports a default class.

### Pool Size Growing

Monitor with:
```bash
curl http://localhost:3000/health
```

If pool size keeps growing, increase `maxPoolSize` or check for plugin leaks.

### Server Not Starting

Check the port is available:
```bash
# Try a different port by modifying webhook.config.json
```

## Development

### Building from Source

```bash
bun install
cd packages/server
bun run src/index.ts serve
```

### Running Tests

```bash
bun test
```

### Type Checking

```bash
bun run tsc --noEmit
```

## Examples

### Complete Setup

1. **Initialize config**:
```bash
bun run src/index.ts plugins ls  # Creates default config
```

2. **Add plugins**:
```bash
bun run src/index.ts plugins add echo packages/plugin-echo/src/index.ts
bun run src/index.ts plugins add imou packages/plugin-imou/src/index.ts
```

3. **Start server**:
```bash
bun run src/index.ts serve
```

4. **Test webhook**:
```bash
curl -X POST http://localhost:3000/echo \
  -H "Content-Type: application/json" \
  -d '{"message": "Hello, Webhook!"}'
```

### Using with Docker

```dockerfile
FROM oven/bun:latest

WORKDIR /app

COPY package.json .
COPY bun.lockb .
COPY packages packages

RUN bun install --production

EXPOSE 3000

CMD ["bun", "run", "packages/server/src/index.ts", "serve"]
```

## License

MIT

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Support

For issues and questions, please open an issue on the GitHub repository.

---

Built with ❤️ using [Bun](https://bun.com)
