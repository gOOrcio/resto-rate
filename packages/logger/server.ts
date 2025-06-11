import pino, { type Logger } from 'pino';
import { type LoggerConfig, type ContextualLogger, getConfigFromEnv } from './shared';

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

export function createServerLogger(config: LoggerConfig): ContextualLogger {
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
}

export function createServerLoggerFromEnv(service: string): ContextualLogger {
	return createServerLogger(getConfigFromEnv(service));
} 