import { createLogger, createLoggerFromEnv } from '@resto-rate/logger';
import { browser } from '$app/environment';

// Create a safe logger for the web app that works during SSR
export function createWebLogger(service: string, level: 'trace' | 'debug' | 'info' | 'warn' | 'error' | 'fatal' = 'debug') {
	return createLogger({
		level,
		service,
		environment: browser && window?.location?.hostname === 'localhost' ? 'development' : 'production',
		pretty: true,
	});
}

// Pre-configured loggers for common use cases
export const apiLogger = createWebLogger('web-api');
export const pageLogger = createWebLogger('page');
export const componentLogger = createWebLogger('component');

// Environment-based logger that's safe for SSR
export const webLogger = createLoggerFromEnv('web');

// Utility to create component-specific loggers
export function createComponentLogger(componentName: string) {
	return createWebLogger(`component-${componentName}`);
}

// Utility to create page-specific loggers
export function createPageLogger(pageName: string) {
	return createWebLogger(`page-${pageName}`);
} 