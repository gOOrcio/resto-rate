// Export shared types and interfaces
export type { LogLevel, LoggerConfig, ContextualLogger } from './shared';
export { getDefaultLogLevel, getConfigFromEnv } from './shared';

// Export server logger (for Node.js/Fastify)
export { createServerLogger, createServerLoggerFromEnv } from './server';

// Export browser logger (for Svelte/Vite)
export { createBrowserLogger, createBrowserLoggerFromEnv } from './browser';

// Backward compatibility exports
export { createServerLogger as createLogger } from './server';
export { createServerLoggerFromEnv as createLoggerFromEnv } from './server';

// Default logger instance for server environments
import { createServerLoggerFromEnv } from './server';
export const logger = createServerLoggerFromEnv('resto-rate'); 