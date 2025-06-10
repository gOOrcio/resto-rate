import type { FastifyPluginAsync } from 'fastify';
import { db } from '../db';
import { session, user, type AuthResponse } from '@resto-rate/database';
import { eq, and, gt } from 'drizzle-orm';
import { requireAuth } from '../middleware/auth';
import { requireQueryResult } from '@resto-rate/validation';

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
		const { sessionId } = request.params as { sessionId: string };

		const result = await db()
			.select({
				session: session,
				user: {
					id: user.id,
					username: user.username,
					age: user.age,
					createdAt: user.createdAt,
					updatedAt: user.updatedAt,
				},
			})
			.from(session)
			.innerJoin(user, eq(session.userId, user.id))
			.where(and(eq(session.id, sessionId), gt(session.expiresAt, new Date())))
			.limit(1);

		const sessionResult = requireQueryResult(result, 'Session not found or expired');

		const response: AuthResponse = {
			user: sessionResult.user,
			sessionId: sessionId,
		};

		reply.header('content-type', 'application/msgpack');
		return response;
	});

	fastify.delete('/logout', { preHandler: [requireAuth] }, async (request, reply) => {
		if (!request.sessionId) {
			return reply.status(400).send({ error: 'No session to logout' });
		}

		await db().delete(session).where(eq(session.id, request.sessionId));

		reply.header('content-type', 'application/msgpack');
		return { message: 'Logged out successfully' };
	});
};
