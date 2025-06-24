import { config } from 'dotenv';
import { resolve, dirname } from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

config({ path: resolve(__dirname, '../../../.env') });

import Fastify from 'fastify';
import cors from '@fastify/cors';
import helmet from '@fastify/helmet';
import { encode, decode } from '@msgpack/msgpack';
import { getApiConfig, getDatabaseConfig, getLoggingConfig } from '@resto-rate/config';
import { createServerLogger } from '@resto-rate/logger';
import { userRoutes } from './routes/users';
import { authRoutes } from './routes/auth';
import { restaurantRoutes } from './routes/restaurants';
import { googleRoutes } from './routes/google';
import { db } from './db';

const loggingConfig = getLoggingConfig();
const logger = createServerLogger({
	level: loggingConfig.level,
	service: 'api',
	environment: (process.env.NODE_ENV as 'development' | 'production' | 'test') || 'development',
	pretty: loggingConfig.pretty,
});

const server = Fastify({
	logger: {
		level: loggingConfig.level,
		transport: loggingConfig.pretty
			? {
					target: 'pino-pretty',
					options: {
						colorize: true,
						translateTime: 'HH:MM:ss',
						ignore: 'pid,hostname',
					},
				}
			: undefined,
		serializers: {
			req: (request) => ({
				method: request.method,
				url: request.url,
				host: request.headers?.host,
				remoteAddress: request.ip,
				remotePort: request.socket?.remotePort,
			}),
			res: (response) => ({
				statusCode: response.statusCode,
			}),
		},
	},
	// Disable default request logging and implement our own at DEBUG level
	disableRequestLogging: true,
});

async function startServer() {
	try {
		const apiConfig = getApiConfig();
		const dbConfig = getDatabaseConfig();

		// Register security plugins
		await server.register(helmet);
		await server.register(cors, {
			origin: apiConfig.corsOrigin,
			credentials: true,
			methods: ['GET', 'POST', 'PUT', 'DELETE', 'PATCH', 'OPTIONS'],
			allowedHeaders: ['Content-Type', 'Authorization', 'X-Session-Id'],
		});

		// Add custom request logging at DEBUG level
		server.addHook('onRequest', async (request) => {
			logger.debug('Incoming request', {
				method: request.method,
				url: request.url,
				host: request.headers?.host,
				remoteAddress: request.ip,
				userAgent: request.headers?.['user-agent'],
			});
		});

		server.addHook('onResponse', async (request, reply) => {
			const responseTime = reply.elapsedTime;
			logger.debug('Request completed', {
				method: request.method,
				url: request.url,
				statusCode: reply.statusCode,
				responseTime: `${responseTime}ms`,
			});
		});

		server.addContentTypeParser('application/msgpack', { parseAs: 'buffer' }, (req, body, done) => {
			try {
				// Handle empty bodies gracefully
				if (!body || (body as Buffer).length === 0) {
					done(null, null);
					return;
				}

				const parsed = decode(body as Uint8Array);
				done(null, parsed);
			} catch (err) {
				logger.error('MessagePack parsing error', { error: err });
				done(err as Error);
			}
		});

		server.addHook('onSend', async (request, reply, payload) => {
			// Always use MessagePack for API responses
			if (reply.getHeader('content-type') === 'application/msgpack') {
				try {
					let data;
					if (typeof payload === 'string') {
						data = JSON.parse(payload);
					} else if (Buffer.isBuffer(payload)) {
						data = JSON.parse(payload.toString());
					} else {
						data = payload;
					}

					const encoded = encode(data);
					return Buffer.from(encoded);
				} catch (error) {
					logger.error('Critical MessagePack encoding error', { error });
					// If MessagePack encoding fails, this is a server error
					reply.status(500);
					reply.header('content-type', 'application/msgpack');
					return Buffer.from(encode({ error: 'Internal server error during response encoding' }));
				}
			}
			return payload;
		});

		server.get('/health', async (request, reply) => {
			let dbConnected: boolean;
			try {
				// Perform a simple DB query to check connectivity
				await db().execute('SELECT 1');
				dbConnected = true;
			} catch (err) {
				logger.error('Database health check failed', { error: err });
				dbConnected = false;
			}
			reply.header('content-type', 'application/json');
			return {
				status: dbConnected ? 'ok' : 'degraded',
				timestamp: new Date().toISOString(),
				environment: apiConfig.nodeEnv,
				database: {
					connected: dbConnected,
					ssl: dbConfig.ssl,
				},
			};
		});

		await server.register(authRoutes, { prefix: '/api/auth' });
		await server.register(userRoutes, { prefix: '/api/users' });
		await server.register(restaurantRoutes, { prefix: '/api/restaurants' });
		await server.register(googleRoutes, { prefix: '/api/google' });

		await server.listen({
			port: apiConfig.port,
			host: apiConfig.host,
		});

		logger.info('API Server started', {
			url: `http://${apiConfig.host}:${apiConfig.port}`,
			environment: apiConfig.nodeEnv,
			database: dbConfig.url.replace(/\/\/.*@/, '//*****@'),
			logLevel: loggingConfig.level,
		});
	} catch (err) {
		server.log.error(err);
		process.exit(1);
	}
}

startServer().catch((err) => {
	logger.error(err);
});
