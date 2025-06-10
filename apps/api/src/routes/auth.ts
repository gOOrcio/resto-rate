import type { FastifyPluginAsync } from 'fastify';
import { type AuthResponse } from '@resto-rate/database';
import { requireAuth } from '../middleware/auth';
import * as authService from '../services/auth.service';
import { handleRoute, successMessage, requireUser } from '../utils/route-helpers';

export const authRoutes: FastifyPluginAsync = async (fastify) => {
	fastify.get('/verify', { preHandler: [requireAuth] }, async (request, reply) => {
		return handleRoute(reply, async () => {
			requireUser(request.user?.id);
			const sessionId = requireUser(request.sessionId);

			const response: AuthResponse = {
				user: {
					id: request.user!.id,
					username: request.user!.username,
					age: request.user!.age,
					createdAt: request.user!.createdAt,
					updatedAt: request.user!.updatedAt,
				},
				sessionId,
			};

			return response;
		});
	});

	fastify.get('/session/:sessionId', async (request, reply) => {
		return handleRoute(reply, async () => {
			const { sessionId } = request.params as { sessionId: string };
			return authService.getSession(sessionId);
		});
	});

	fastify.delete('/logout', { preHandler: [requireAuth] }, async (request, reply) => {
		return handleRoute(reply, async () => {
			const sessionId = requireUser(request.sessionId);
			await authService.invalidateSession(sessionId);
			return successMessage('Logged out successfully');
		});
	});
};
