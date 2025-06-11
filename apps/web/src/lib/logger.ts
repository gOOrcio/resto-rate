import { createBrowserLogger, createBrowserLoggerFromEnv, type LogLevel } from '@resto-rate/logger';
import { browser } from '$app/environment';

// Create a safe logger for the web app
export function createWebLogger(service: string, level: LogLevel = 'debug') {
	return createBrowserLogger({
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
export const webLogger = createBrowserLoggerFromEnv('web');

// Utility to create component-specific loggers
export function createComponentLogger(componentName: string) {
	return createWebLogger(`component-${componentName}`);
}

// Utility to create page-specific loggers
export function createPageLogger(pageName: string) {
	return createWebLogger(`page-${pageName}`);
} 