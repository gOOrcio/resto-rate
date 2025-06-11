# Resto Rate - Logging System

This document explains the comprehensive logging system implemented across the Resto Rate application.

## Overview

The logging system provides:
- **Configurable log levels** via environment variables
- **Structured logging** with contextual information
- **Different behavior** for development vs production
- **Client and server-side logging** with unified interface
- **Pretty printing** in development, JSON in production

## Log Levels

The system supports six log levels in order of verbosity:

1. **`trace`** - Most verbose, for detailed debugging
2. **`debug`** - Development debugging information
3. **`info`** - General application information (default for production)
4. **`warn`** - Warnings about potential issues
5. **`error`** - Errors that need attention
6. **`fatal`** - Critical errors requiring immediate action

## Configuration

### Environment Variables

Add these to your `.env` file:

```bash
# Log level: trace, debug, info, warn, error, fatal
LOG_LEVEL=debug          # Default: debug in development, info in production

# Pretty print logs (true/false)
LOG_PRETTY=true          # Default: true in development, false in production

# Enable file logging (true/false)
LOG_FILE=false           # Default: false

# Log directory (when file logging is enabled)
LOG_DIR=./logs           # Default: ./logs
```

### Default Behavior

- **Development**: `LOG_LEVEL=debug`, `LOG_PRETTY=true`
- **Production**: `LOG_LEVEL=info`, `LOG_PRETTY=false`
- **Test**: `LOG_LEVEL=warn`

## Usage Examples

### Basic Logging

```typescript
import { createLoggerFromEnv } from '@resto-rate/logger';

const logger = createLoggerFromEnv('my-service');

logger.trace('Detailed trace information', { userId: '123' });
logger.debug('Debug information', { query: 'SELECT * FROM users' });
logger.info('General information', { event: 'user_login' });
logger.warn('Warning message', { memory: '85%' });
logger.error('Error occurred', { error: 'Connection failed' });
logger.fatal('Critical error', { code: 500 });
```

### Custom Logger Configuration

```typescript
import { createLogger } from '@resto-rate/logger';

const logger = createLogger({
	level: 'info',
	service: 'my-custom-service',
	environment: 'production',
	pretty: false,
});
```

### Child Loggers with Context

```typescript
const parentLogger = createLoggerFromEnv('api');
const requestLogger = parentLogger.child({
	requestId: 'req-123',
	userId: 'user-456',
	ip: '192.168.1.1'
});

requestLogger.info('Processing request'); 
// Output includes: requestId, userId, ip automatically
```

## Implementation Across Components

### API Server (`apps/api`)

The API server uses structured logging throughout:

```typescript
// Main server
import { createLogger, getLoggingConfig } from '@resto-rate/logger';

const loggingConfig = getLoggingConfig();
const logger = createLogger({
	level: loggingConfig.level,
	service: 'api',
	environment: process.env.NODE_ENV,
	pretty: loggingConfig.pretty,
});

logger.info('API Server started', {
	url: `http://localhost:3001`,
	environment: 'development',
	logLevel: 'debug'
});
```

```typescript
// Route handlers
import { createLoggerFromEnv } from '@resto-rate/logger';

const logger = createLoggerFromEnv('user-routes');

export async function createUser(userData) {
	logger.debug('Creating user', { username: userData.username });
	try {
		const user = await userService.create(userData);
		logger.info('User created successfully', { userId: user.id });
		return user;
	} catch (error) {
		logger.error('Failed to create user', { error, userData });
		throw error;
	}
}
```

### Web Frontend (`apps/web`)

Client-side logging with SSR-safe browser compatibility:

```typescript
// In $lib/logger.ts - Safe logger creation for web apps
import { createLogger, createLoggerFromEnv } from '@resto-rate/logger';
import { browser } from '$app/environment';

export function createWebLogger(service: string, level = 'debug') {
	return createLogger({
		level,
		service,
		environment: browser && window?.location?.hostname === 'localhost' ? 'development' : 'production',
		pretty: true,
	});
}

// Pre-configured loggers
export const apiLogger = createWebLogger('web-api');
export const pageLogger = createWebLogger('page');
```

```typescript
// API client using pre-configured logger
import { apiLogger } from '$lib/logger';

async function apiRequest(url: string) {
	apiLogger.debug('API Request', { method: 'GET', url });
	try {
		const response = await fetch(url);
		apiLogger.debug('API Response', { status: response.status });
		return response;
	} catch (error) {
		apiLogger.error('API request failed', { url, error });
		throw error;
	}
}
```

```svelte
<!-- Svelte components - SSR-safe logging -->
<script lang="ts">
	import { createPageLogger } from '$lib/logger';
	
	// Safe for SSR - no browser API access during initialization
	const logger = createPageLogger('restaurants');

	async function loadData() {
		logger.debug('Loading restaurants');
		try {
			const restaurants = await api.getRestaurants();
			logger.info('Restaurants loaded', { count: restaurants.length });
		} catch (error) {
			logger.error('Failed to load restaurants', { error });
		}
	}
</script>
```

## SSR Safety

The logging system is designed to work safely with SvelteKit's Server-Side Rendering:

### ✅ Safe Practices

```typescript
// ✅ Safe - uses SvelteKit's browser store
import { createPageLogger } from '$lib/logger';
const logger = createPageLogger('my-page');

// ✅ Safe - deferred browser API access
import { apiLogger } from '$lib/logger';
apiLogger.info('This works in SSR');

// ✅ Safe - no browser APIs at module level
import { createLoggerFromEnv } from '@resto-rate/logger';
const logger = createLoggerFromEnv('service');
```

### ❌ Unsafe Practices (Fixed)

```typescript
// ❌ Unsafe - direct window access during module initialization
const logger = createLogger({
	environment: window.location.hostname === 'localhost' ? 'development' : 'production'
});

// ❌ Unsafe - localStorage access during SSR
const level = localStorage.getItem('log-level');
```

### Browser API Access

The logger safely handles browser APIs:
- **localStorage**: Checked only when actually logging (not during initialization)
- **window**: Only accessed after verifying it exists
- **console**: Wrapped in safe checks for environments where it might not exist

## Browser Console Output

In development, browser logs are formatted for readability:

```
[2024-12-20T10:30:45.123Z] DEBUG (web-api): API Request {"method":"GET","url":"/api/restaurants"}
[2024-12-20T10:30:45.200Z] DEBUG (web-api): API Response {"status":200,"statusText":"OK"}
[2024-12-20T10:30:45.205Z] INFO (restaurants-page): Restaurants loaded {"count":5}
```

## Server Console Output

### Development (Pretty)
```
10:30:45 DEBUG (api): Processing request
    requestId: "req-123"
    userId: "user-456"
    method: "GET"
    url: "/api/users"

10:30:45 INFO (api): Request completed
    requestId: "req-123"
    duration: "15ms"
    status: 200
```

### Production (JSON)
```json
{"level":30,"time":1703073045123,"service":"api","environment":"production","requestId":"req-123","userId":"user-456","method":"GET","url":"/api/users","msg":"Processing request"}
{"level":30,"time":1703073045138,"service":"api","environment":"production","requestId":"req-123","duration":"15ms","status":200,"msg":"Request completed"}
```

## Testing the Logging System

### Test Script

Run the logging test script to see all log levels in action:

```bash
cd apps/api
bun run test:logging
```

This will demonstrate:
- All log levels (trace through fatal)
- Different logger configurations
- Child loggers with context
- Environment-based configuration

### Changing Log Levels

1. **Via Environment File**:
   ```bash
   # In .env file
   LOG_LEVEL=info
   ```

2. **Via Environment Variable**:
   ```bash
   LOG_LEVEL=warn bun run dev
   ```

3. **Via Browser Console** (for client-side):
   ```javascript
   localStorage.setItem('resto-rate-log-level', 'info');
   // Refresh page
   ```

## Advanced Features

### Contextual Logging

Use child loggers to automatically include context:

```typescript
const requestLogger = logger.child({
	requestId: generateId(),
	userId: session.userId,
	ip: request.ip,
});

// All subsequent logs include this context automatically
requestLogger.info('Processing payment');
requestLogger.error('Payment failed', { amount, currency });
```

### Service-Specific Loggers

Different services can have their own loggers:

```typescript
const authLogger = createLoggerFromEnv('auth-service');
const dbLogger = createLoggerFromEnv('database');
const emailLogger = createLoggerFromEnv('email-service');
```

### Performance Logging

Log performance metrics with structured data:

```typescript
const startTime = Date.now();
// ... operation
const duration = Date.now() - startTime;

logger.info('Operation completed', {
	operation: 'user_search',
	duration: `${duration}ms`,
	resultCount: results.length,
	cacheHit: false
});
```

## Best Practices

1. **Use appropriate log levels**:
   - `trace`: Very detailed debugging (SQL queries, etc.)
   - `debug`: Development debugging
   - `info`: Important application events
   - `warn`: Potential issues
   - `error`: Actual errors
   - `fatal`: Critical system failures

2. **Include relevant context**:
   ```typescript
   logger.error('Database query failed', {
   	query: 'SELECT * FROM users',
   	error: error.message,
   	duration: '1250ms',
   	userId: session.userId
   });
   ```

3. **Use child loggers for requests**:
   ```typescript
   const requestLogger = logger.child({ requestId });
   // Use requestLogger throughout the request lifecycle
   ```

4. **Don't log sensitive information**:
   ```typescript
   // ❌ Bad
   logger.debug('User login', { password: user.password });
   
   // ✅ Good
   logger.debug('User login', { username: user.username });
   ```

5. **Log at the right granularity**:
   - Too little: Hard to debug issues
   - Too much: Log noise, performance impact

## Production Considerations

- **Set `LOG_LEVEL=info`** to reduce log volume
- **Set `LOG_PRETTY=false`** for machine-readable JSON
- **Consider `LOG_FILE=true`** for persistent logging
- **Monitor log volume** and storage
- **Set up log aggregation** (ELK stack, etc.)

## Troubleshooting

### Common Issues

1. **Logs not appearing**:
   - Check `LOG_LEVEL` is set correctly
   - Ensure logger is imported properly
   - Verify environment variables are loaded

2. **Too many/few logs**:
   - Adjust `LOG_LEVEL` in `.env`
   - Check if you're using the right log level for messages

3. **Browser logs not formatted**:
   - Ensure `pretty: true` in logger config
   - Check browser developer tools console

### Debug Commands

```bash
# Test current log level
LOG_LEVEL=trace bun run test:logging

# Check environment variables
bun run dev | grep -i log

# View all log levels
cd apps/api && bun run test:logging
```

This logging system provides comprehensive observability across your entire Resto Rate application while being flexible and configurable for different environments. 