import type { FastifyPluginAsync } from 'fastify';
import { type CreateUserRequest } from '@resto-rate/constants';
import { requireAuth, optionalAuth } from '../middleware/auth';
import * as userService from '../services/user.service';
import { handleRoute, successMessage, requireUser } from '../utils/route-helpers';

export const userRoutes: FastifyPluginAsync = async (fastify) => {
	// Get all users (public)
	fastify.get('/', { preHandler: [optionalAuth] }, async (request, reply) => {
		return handleRoute(reply, async () => {
			const users = await userService.getAllUsers();
			return { users };
		});
	});

	// Get user by ID (public)
	fastify.get('/:id', { preHandler: [optionalAuth] }, async (request, reply) => {
		return handleRoute(reply, async () => {
			const { id } = request.params as { id: string };
			const user = await userService.getUserById(id);

			if (!user) {
				throw new Error('User not found');
			}

			return { user };
		});
	});

	// Create new user (public)
	fastify.post('/', async (request, reply) => {
		return handleRoute(reply, async () => {
			const user = await userService.createUser(request.body as CreateUserRequest);
			return { user };
		});
	});

	// Update user (requires auth, own profile only)
	fastify.put('/:id', { preHandler: [requireAuth] }, async (request, reply) => {
		return handleRoute(reply, async () => {
			const { id } = request.params as { id: string };
			const updateData = request.body as { username?: string; age?: number };
			const userId = requireUser(request.user?.id);

			const user = await userService.updateUser(id, updateData, userId);
			return { user };
		});
	});

	// Delete user (requires auth, own profile only)
	fastify.delete('/:id', { preHandler: [requireAuth] }, async (request, reply) => {
		return handleRoute(reply, async () => {
			const { id } = request.params as { id: string };
			const userId = requireUser(request.user?.id);

			await userService.deleteUser(id, userId);
			return successMessage('User deleted successfully');
		});
	});

	// Get current user profile (requires auth)
	fastify.get('/me/profile', { preHandler: [requireAuth] }, async (request, reply) => {
		return handleRoute(reply, async () => {
			requireUser(request.user?.id);
			return { user: request.user };
		});
	});
};
