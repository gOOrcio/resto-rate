import type { FastifyPluginAsync } from 'fastify';
import { type CreateUserRequest } from '@resto-rate/database';
import { requireAuth, optionalAuth } from '../middleware/auth';
import { userService } from '../services/user.service';

export const userRoutes: FastifyPluginAsync = async (fastify) => {
	// Get all users (public)
	fastify.get('/', { preHandler: [optionalAuth] }, async (request, reply) => {
		try {
			const users = await userService.getAllUsers();

			reply.header('content-type', 'application/msgpack');
			return { users };
		} catch (error) {
			return reply.status(500).send({ error: (error as Error).message });
		}
	});

	// Get user by ID (public)
	fastify.get('/:id', { preHandler: [optionalAuth] }, async (request, reply) => {
		try {
			const { id } = request.params as { id: string };

			const foundUser = await userService.getUserById(id);

			if (!foundUser) {
				return reply.status(404).send({ error: 'User not found' });
			}

			reply.header('content-type', 'application/msgpack');
			return { user: foundUser };
		} catch (error) {
			return reply.status(500).send({ error: (error as Error).message });
		}
	});

	// Create new user (public)
	fastify.post('/', async (request, reply) => {
		try {
			const user = await userService.createUser(request.body as CreateUserRequest);

			reply.header('content-type', 'application/msgpack');
			return { user };
		} catch (error) {
			if ((error as Error).message === 'Username and password are required') {
				return reply.status(400).send({ error: (error as Error).message });
			}
			if ((error as Error).message === 'Username already exists') {
				return reply.status(409).send({ error: (error as Error).message });
			}
			return reply.status(500).send({ error: (error as Error).message });
		}
	});

	// Update user (requires auth, own profile only)
	fastify.put(
		'/:id',
		{ preHandler: [requireAuth] },
		async (request, reply) => {
			try {
				const { id } = request.params as { id: string };
				const updateData = request.body as {
					username?: string;
					age?: number;
				};

				if (!request.user) {
					return reply.status(401).send({ error: 'Unauthorized' });
				}

				const user = await userService.updateUser(id, updateData, request.user.id);

				reply.header('content-type', 'application/msgpack');
				return { user };
			} catch (error) {
				if ((error as Error).message === 'Cannot update other users') {
					return reply.status(403).send({ error: (error as Error).message });
				}
				if ((error as Error).message === 'No update data provided') {
					return reply.status(400).send({ error: (error as Error).message });
				}
				if ((error as Error).message === 'User not found') {
					return reply.status(404).send({ error: (error as Error).message });
				}
				if ((error as Error).message === 'Username already exists') {
					return reply.status(409).send({ error: (error as Error).message });
				}
				return reply.status(500).send({ error: (error as Error).message });
			}
		},
	);

	// Delete user (requires auth, own profile only)
	fastify.delete(
		'/:id',
		{ preHandler: [requireAuth] },
		async (request, reply) => {
			try {
				const { id } = request.params as { id: string };

				if (!request.user) {
					return reply.status(401).send({ error: 'Unauthorized' });
				}

				await userService.deleteUser(id, request.user.id);

				reply.header('content-type', 'application/msgpack');
				return { message: 'User deleted successfully' };
			} catch (error) {
				if ((error as Error).message === 'Cannot delete other users') {
					return reply.status(403).send({ error: (error as Error).message });
				}
				return reply.status(500).send({ error: (error as Error).message });
			}
		},
	);

	// Get current user profile (requires auth)
	fastify.get(
		'/me/profile',
		{ preHandler: [requireAuth] },
		async (request, reply) => {
			if (!request.user) {
				return reply.status(401).send({ error: 'Unauthorized' });
			}

			reply.header('content-type', 'application/msgpack');
			return { user: request.user };
		},
	);
};
