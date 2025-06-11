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

export function getConfigFromEnv(service: string): LoggerConfig {
	const getEnvVar = (name: string, defaultValue: string = ''): string => {
		if (typeof process !== 'undefined' && process.env) {
			return process.env[name] || defaultValue;
		}
		return defaultValue;
	};

	const environment = (getEnvVar('NODE_ENV') as 'development' | 'production' | 'test') || 'development';
	const level = (getEnvVar('LOG_LEVEL') as LogLevel) || getDefaultLogLevel(environment);
	const pretty = getEnvVar('LOG_PRETTY') === 'true' || environment === 'development';

	return { level, service, environment, pretty };
} 