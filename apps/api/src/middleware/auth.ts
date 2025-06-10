import type { FastifyRequest, FastifyReply } from 'fastify';
import { db } from '../db';
import { session, user, type User } from '@resto-rate/database';
import { eq, and, gt } from 'drizzle-orm';
import { getFirstItem } from '@resto-rate/validation';

declare module 'fastify' {
	interface FastifyRequest {
		user?: User;
		sessionId?: string;
	}
}

export async function requireAuth(request: FastifyRequest, reply: FastifyReply) {
	try {
		const sessionId =
			request.headers.authorization?.replace('Bearer ', '') ||
			(request.headers['x-session-id'] as string);

		if (!sessionId) {
			return reply.status(401).send({ error: 'Authentication required' });
		}

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

		const authResult = getFirstItem(result);
		if (!authResult) {
			return reply.status(401).send({ error: 'Invalid or expired session' });
		}

		request.user = authResult.user as User;
		request.sessionId = sessionId;
	} catch (error) {
		request.log.error(error);
		return reply.status(500).send({ error: 'Authentication error' });
	}
}

export async function optionalAuth(request: FastifyRequest, _reply: FastifyReply) {
	try {
		const sessionId =
			request.headers.authorization?.replace('Bearer ', '') ||
			(request.headers['x-session-id'] as string);

		if (!sessionId) {
			return; // No auth required, continue
		}

		// Try to verify session
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

		const authResult = getFirstItem(result);
		if (authResult) {
			request.user = authResult.user as User;
			request.sessionId = sessionId;
		}
	} catch (error) {
		request.log.error(error);
	}
}
