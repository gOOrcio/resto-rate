import type { FastifyRequest, FastifyReply } from 'fastify';
import { type User } from '@resto-rate/database';
import * as authService from '../services/auth.service';

declare module 'fastify' {
	interface FastifyRequest {
		user?: User;
		sessionId?: string;
	}
}

function extractSessionId(request: FastifyRequest): string | null {
	return (
		request.headers.authorization?.replace('Bearer ', '') ||
		(request.headers['x-session-id'] as string) ||
		null
	);
}

async function setUserFromSession(request: FastifyRequest, sessionId: string): Promise<boolean> {
	try {
		const authResult = await authService.verifySession(sessionId);
		request.user = authResult.user as User;
		request.sessionId = sessionId;
		return true;
	} catch {
		return false;
	}
}

export async function requireAuth(request: FastifyRequest, reply: FastifyReply) {
	const sessionId = extractSessionId(request);

	if (!sessionId) {
		return reply.status(401).send({ error: 'Authentication required' });
	}

	const success = await setUserFromSession(request, sessionId);
	if (!success) {
		return reply.status(401).send({ error: 'Invalid or expired session' });
	}
}

export async function optionalAuth(request: FastifyRequest, _reply: FastifyReply) {
	const sessionId = extractSessionId(request);
	if (sessionId) {
		await setUserFromSession(request, sessionId);
	}
}
