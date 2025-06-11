import type { FastifyPluginAsync } from 'fastify';
import { type AuthResponse } from '@resto-rate/constants';
import { requireAuth } from '../middleware/auth';
import * as authService from '../services/auth.service';
import { handleRoute, successMessage, requireUser } from '../utils/route-helpers';

export const authRoutes: FastifyPluginAsync = async (fastify) => {
	// Login endpoint
	fastify.post('/login', async (request, reply) => {
		return handleRoute(reply, async () => {
			const { username, password } = request.body as { username: string; password: string };
			
			if (!username || !password) {
				throw new Error('Username and password are required');
			}

			const result = await authService.login(username, password);
			return result;
		});
	});

	// Register endpoint
	fastify.post('/register', async (request, reply) => {
		return handleRoute(reply, async () => {
			const { username, password, age } = request.body as { username: string; password: string; age?: number };
			
			if (!username || !password) {
				throw new Error('Username and password are required');
			}

			const result = await authService.register(username, password, age);
			return result;
		});
	});

	// Verify session endpoint
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

	// Get session by ID endpoint
	fastify.get('/session/:sessionId', async (request, reply) => {
		return handleRoute(reply, async () => {
			const { sessionId } = request.params as { sessionId: string };
			return authService.getSession(sessionId);
		});
	});

	// Logout endpoint
	fastify.delete('/logout', { preHandler: [requireAuth] }, async (request, reply) => {
		return handleRoute(reply, async () => {
			const sessionId = requireUser(request.sessionId);
			await authService.invalidateSession(sessionId);
			return successMessage('Logged out successfully');
		});
	});
};
