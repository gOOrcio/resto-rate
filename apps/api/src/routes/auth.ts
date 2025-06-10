import type { FastifyPluginAsync } from 'fastify';
import { type AuthResponse } from '@resto-rate/database';
import { requireAuth } from '../middleware/auth';
import { authService } from '../services/auth.service';

export const authRoutes: FastifyPluginAsync = async (fastify) => {
	fastify.get('/verify', { preHandler: [requireAuth] }, async (request, reply) => {
		if (!request.user || !request.sessionId) {
			return reply.status(401).send({ error: 'Authentication failed' });
		}

		const response: AuthResponse = {
			user: {
				id: request.user.id,
				username: request.user.username,
				age: request.user.age,
				createdAt: request.user.createdAt,
				updatedAt: request.user.updatedAt,
			},
			sessionId: request.sessionId,
		};

		reply.header('content-type', 'application/msgpack');
		return response;
	});

	fastify.get('/session/:sessionId', async (request, reply) => {
		try {
			const { sessionId } = request.params as { sessionId: string };

			const response = await authService.getSession(sessionId);

			reply.header('content-type', 'application/msgpack');
			return response;
		} catch (error) {
			if ((error as Error).message === 'Session not found or expired') {
				return reply.status(404).send({ error: 'Session not found or expired' });
			}
			return reply.status(500).send({ error: (error as Error).message });
		}
	});

	fastify.delete('/logout', { preHandler: [requireAuth] }, async (request, reply) => {
		try {
			if (!request.sessionId) {
				return reply.status(400).send({ error: 'No session to logout' });
			}

			await authService.invalidateSession(request.sessionId);

			reply.header('content-type', 'application/msgpack');
			return { message: 'Logged out successfully' };
		} catch (error) {
			return reply.status(500).send({ error: (error as Error).message });
		}
	});
};
