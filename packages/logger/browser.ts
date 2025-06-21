import {
	type LoggerConfig,
	type ContextualLogger,
	type LogLevel,
	getDefaultLogLevel,
} from './shared';

class BrowserLogger implements ContextualLogger {
	private context: object = {};
	private actualLevel: LogLevel;

	constructor(
		private config: LoggerConfig,
		context?: object
	) {
		this.context = context || {};
		this.actualLevel = this.getEffectiveLevel();
	}

	private getEffectiveLevel(): LogLevel {
		// Check localStorage for override (useful for debugging)
		try {
			if (typeof localStorage !== 'undefined') {
				const stored = localStorage.getItem('resto-rate-log-level') as LogLevel;
				if (stored && ['trace', 'debug', 'info', 'warn', 'error', 'fatal'].includes(stored)) {
					return stored;
				}
			}
		} catch {
			// localStorage might not be available
		}
		return this.config.level;
	}

	private shouldLog(level: LogLevel): boolean {
		const levels = ['trace', 'debug', 'info', 'warn', 'error', 'fatal'];
		return levels.indexOf(level) >= levels.indexOf(this.actualLevel);
	}

	private formatMessage(level: LogLevel, msg: string, extra?: object): string {
		const timestamp = new Date().toISOString();
		const context = { ...this.context, ...extra };
		const contextStr = Object.keys(context).length > 0 ? ` ${JSON.stringify(context)}` : '';
		return `[${timestamp}] ${level.toUpperCase()} (${this.config.service}): ${msg}${contextStr}`;
	}

	private log(method: 'debug' | 'info' | 'warn' | 'error', message: string) {
		if (console && console[method]) {
			console[method](message);
		}
	}

	trace(msg: string, extra?: object) {
		if (this.shouldLog('trace')) {
			this.log('debug', this.formatMessage('trace', msg, extra));
		}
	}

	debug(msg: string, extra?: object) {
		if (this.shouldLog('debug')) {
			this.log('debug', this.formatMessage('debug', msg, extra));
		}
	}

	info(msg: string, extra?: object) {
		if (this.shouldLog('info')) {
			this.log('info', this.formatMessage('info', msg, extra));
		}
	}

	warn(msg: string, extra?: object) {
		if (this.shouldLog('warn')) {
			this.log('warn', this.formatMessage('warn', msg, extra));
		}
	}

	error(msg: string, extra?: object) {
		if (this.shouldLog('error')) {
			this.log('error', this.formatMessage('error', msg, extra));
		}
	}

	fatal(msg: string, extra?: object) {
		if (this.shouldLog('fatal')) {
			this.log('error', this.formatMessage('fatal', msg, extra));
		}
	}

	child(context: object): ContextualLogger {
		return new BrowserLogger(this.config, { ...this.context, ...context });
	}
}

export function createBrowserLogger(config: LoggerConfig): ContextualLogger {
	return new BrowserLogger(config);
}

export function createBrowserLoggerFromEnv(service: string): ContextualLogger {
	// Simple browser-safe environment detection
	const environment = 'development'; // Default for browser, can be overridden
	const level = getDefaultLogLevel(environment);

	return createBrowserLogger({
		level,
		service,
		environment,
		pretty: true,
	});
}
