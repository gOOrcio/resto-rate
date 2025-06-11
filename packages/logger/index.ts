import pino, { type Logger } from 'pino';

export type LogLevel = 'trace' | 'debug' | 'info' | 'warn' | 'error' | 'fatal';

export interface LoggerConfig {
	level: LogLevel;
	service: string;
	environment: 'development' | 'production' | 'test';
	pretty?: boolean;
}

export interface ContextualLogger {
	trace: (msg: string, extra?: object) => void;
	debug: (msg: string, extra?: object) => void;
	info: (msg: string, extra?: object) => void;
	warn: (msg: string, extra?: object) => void;
	error: (msg: string, extra?: object) => void;
	fatal: (msg: string, extra?: object) => void;
	child: (context: object) => ContextualLogger;
}

class PinoLogger implements ContextualLogger {
	constructor(private logger: Logger) {}

	trace(msg: string, extra?: object) {
		this.logger.trace(extra, msg);
	}

	debug(msg: string, extra?: object) {
		this.logger.debug(extra, msg);
	}

	info(msg: string, extra?: object) {
		this.logger.info(extra, msg);
	}

	warn(msg: string, extra?: object) {
		this.logger.warn(extra, msg);
	}

	error(msg: string, extra?: object) {
		this.logger.error(extra, msg);
	}

	fatal(msg: string, extra?: object) {
		this.logger.fatal(extra, msg);
	}

	child(context: object): ContextualLogger {
		return new PinoLogger(this.logger.child(context));
	}
}

class BrowserLogger implements ContextualLogger {
	private context: object = {};
	private actualLevel: LogLevel;
	
	constructor(
		private level: LogLevel,
		private service: string,
		context?: object
	) {
		this.context = context || {};
		this.actualLevel = this.getEffectiveLevel();
	}

	private getEffectiveLevel(): LogLevel {
		// Only check localStorage when we're actually in a browser environment
		if (typeof window !== 'undefined' && typeof localStorage !== 'undefined') {
			try {
				const stored = localStorage.getItem('resto-rate-log-level') as LogLevel;
				if (stored && ['trace', 'debug', 'info', 'warn', 'error', 'fatal'].includes(stored)) {
					return stored;
				}
			} catch {
				// localStorage might not be available in some browser contexts
			}
		}
		return this.level;
	}

	private shouldLog(level: LogLevel): boolean {
		const levels = ['trace', 'debug', 'info', 'warn', 'error', 'fatal'];
		const currentLevelIndex = levels.indexOf(this.actualLevel);
		const messageLevelIndex = levels.indexOf(level);
		return messageLevelIndex >= currentLevelIndex;
	}

	private formatMessage(level: LogLevel, msg: string, extra?: object): string {
		const timestamp = new Date().toISOString();
		const context = { ...this.context, ...extra };
		const contextStr = Object.keys(context).length > 0 ? ` ${JSON.stringify(context)}` : '';
		return `[${timestamp}] ${level.toUpperCase()} (${this.service}): ${msg}${contextStr}`;
	}

	private safeConsoleLog(method: 'debug' | 'info' | 'warn' | 'error', message: string) {
		// Only log when we're in a browser environment
		if (typeof console !== 'undefined' && console[method]) {
			console[method](message);
		}
	}

	trace(msg: string, extra?: object) {
		if (this.shouldLog('trace')) {
			this.safeConsoleLog('debug', this.formatMessage('trace', msg, extra));
		}
	}

	debug(msg: string, extra?: object) {
		if (this.shouldLog('debug')) {
			this.safeConsoleLog('debug', this.formatMessage('debug', msg, extra));
		}
	}

	info(msg: string, extra?: object) {
		if (this.shouldLog('info')) {
			this.safeConsoleLog('info', this.formatMessage('info', msg, extra));
		}
	}

	warn(msg: string, extra?: object) {
		if (this.shouldLog('warn')) {
			this.safeConsoleLog('warn', this.formatMessage('warn', msg, extra));
		}
	}

	error(msg: string, extra?: object) {
		if (this.shouldLog('error')) {
			this.safeConsoleLog('error', this.formatMessage('error', msg, extra));
		}
	}

	fatal(msg: string, extra?: object) {
		if (this.shouldLog('fatal')) {
			this.safeConsoleLog('error', this.formatMessage('fatal', msg, extra));
		}
	}

	child(context: object): ContextualLogger {
		return new BrowserLogger(this.level, this.service, { ...this.context, ...context });
	}
}

// No-op logger for environments where logging isn't available
class NoOpLogger implements ContextualLogger {
	trace() {}
	debug() {}
	info() {}
	warn() {}
	error() {}
	fatal() {}
	child(): ContextualLogger {
		return new NoOpLogger();
	}
}

export function createLogger(config: LoggerConfig): ContextualLogger {
	try {
		// Detect environment - be safe about browser detection
		const isBrowser = typeof window !== 'undefined' && typeof document !== 'undefined';
		
		if (isBrowser) {
			return new BrowserLogger(config.level, config.service);
		}

		// Server-side logger using Pino
		const pinoConfig: pino.LoggerOptions = {
			level: config.level,
			base: {
				service: config.service,
				environment: config.environment,
			},
		};

		// Pretty printing for development
		if (config.pretty || config.environment === 'development') {
			pinoConfig.transport = {
				target: 'pino-pretty',
				options: {
					colorize: true,
					translateTime: 'HH:MM:ss',
					ignore: 'pid,hostname',
				},
			};
		}

		const pinoLogger = pino(pinoConfig);
		return new PinoLogger(pinoLogger);
	} catch (error) {
		// Fallback to no-op logger if there are any issues
		return new NoOpLogger();
	}
}

// Default configurations for different environments
export function getDefaultLogLevel(environment: string): LogLevel {
	switch (environment) {
		case 'production':
			return 'info';
		case 'test':
			return 'warn';
		case 'development':
		default:
			return 'debug';
	}
}

// Convenience function to create logger from environment
export function createLoggerFromEnv(service: string): ContextualLogger {
	try {
		// Safe environment access for browser contexts
		const getEnvVar = (name: string, defaultValue: string = ''): string => {
			if (typeof process !== 'undefined' && process.env) {
				return process.env[name] || defaultValue;
			}
			return defaultValue;
		};

		const environment = (getEnvVar('NODE_ENV') as 'development' | 'production' | 'test') || 'development';
		const level = (getEnvVar('LOG_LEVEL') as LogLevel) || getDefaultLogLevel(environment);
		const pretty = getEnvVar('LOG_PRETTY') === 'true' || environment === 'development';

		return createLogger({
			level,
			service,
			environment,
			pretty,
		});
	} catch (error) {
		// Fallback to no-op logger if there are any issues
		return new NoOpLogger();
	}
}

// Export common logger instance
export const logger = createLoggerFromEnv('resto-rate'); 