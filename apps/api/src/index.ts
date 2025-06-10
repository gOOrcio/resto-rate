// Load environment variables from root .env file
import { config } from 'dotenv';
import { resolve, dirname } from 'path';
import { fileURLToPath } from 'url';

// Get __dirname equivalent for ES modules
const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

// Load .env from project root (two levels up from src/)
config({ path: resolve(__dirname, '../../../.env') });

import Fastify from 'fastify';
import cors from '@fastify/cors';
import helmet from '@fastify/helmet';
import { encode, decode } from '@msgpack/msgpack';
import { getApiConfig, getDatabaseConfig } from '@resto-rate/config';
import { userRoutes } from './routes/users';
import { authRoutes } from './routes/auth';
import { restaurantRoutes } from './routes/restaurants';

const server = Fastify({
	logger: true,
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
				console.error('❌ MessagePack parsing error:', err);
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
					console.error('❌ Critical MessagePack encoding error:', error);
					// If MessagePack encoding fails, this is a server error
					reply.status(500);
					reply.header('content-type', 'application/msgpack');
					return Buffer.from(encode({ error: 'Internal server error during response encoding' }));
				}
			}
			return payload;
		});

		server.get('/health', async (request, reply) => {
			reply.header('content-type', 'application/json');
			return {
				status: 'ok',
				timestamp: new Date().toISOString(),
				environment: apiConfig.nodeEnv,
				database: {
					connected: true, // You could add actual DB health check here
					ssl: dbConfig.ssl,
				},
			};
		});

		await server.register(authRoutes, { prefix: '/api/auth' });
		await server.register(userRoutes, { prefix: '/api/users' });
		await server.register(restaurantRoutes, { prefix: '/api/restaurants' });

		await server.listen({
			port: apiConfig.port,
			host: apiConfig.host,
		});

		console.log(`🚀 API Server ready at http://${apiConfig.host}:${apiConfig.port}`);
		console.log(`📊 Environment: ${apiConfig.nodeEnv}`);
		console.log(`🗄️  Database: ${dbConfig.url.replace(/\/\/.*@/, '//*****@')}`);
	} catch (err) {
		server.log.error(err);
		process.exit(1);
	}
}

startServer();
