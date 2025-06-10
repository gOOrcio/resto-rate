import type { FastifyPluginAsync } from 'fastify';
import { db } from '../db';
import { user, type CreateUserRequest } from '@resto-rate/database';
import { eq } from 'drizzle-orm';
import { requireAuth, optionalAuth } from '../middleware/auth';
import { generateUserId } from '@resto-rate/ulid';

export const userRoutes: FastifyPluginAsync = async (fastify) => {
	// Get all users (public)
	fastify.get('/', { preHandler: [optionalAuth] }, async (request, reply) => {
		const users = await db().select({
			id: user.id,
			username: user.username,
			age: user.age,
			createdAt: user.createdAt,
			updatedAt: user.updatedAt,
		}).from(user);

		reply.header('content-type', 'application/msgpack');
		return { users };
	});

	// Get user by ID (public)
	fastify.get('/:id', { preHandler: [optionalAuth] }, async (request, reply) => {
		const { id } = request.params as { id: string };

		const [foundUser] = await db().select({
			id: user.id,
			username: user.username,
			age: user.age,
			createdAt: user.createdAt,
			updatedAt: user.updatedAt,
		}).from(user).where(eq(user.id, id));

		if (!foundUser) {
			return reply.status(404).send({ error: 'User not found' });
		}

		reply.header('content-type', 'application/msgpack');
		return { user: foundUser };
	});

	// Create new user (public)
	fastify.post('/', async (request, reply) => {
		const { username, password, age } = request.body as CreateUserRequest;

		if (!username || !password) {
			return reply.status(400).send({ error: 'Username and password are required' });
		}

		try {
			const userId = generateUserId();
			const [newUser] = await db().insert(user).values({
				id: userId,
				username,
				passwordHash: password, // In real app, hash this!
				age,
			}).returning({
				id: user.id,
				username: user.username,
				age: user.age,
				createdAt: user.createdAt,
				updatedAt: user.updatedAt,
			});

			reply.header('content-type', 'application/msgpack');
			return { user: newUser };
		} catch (error) {
			if ((error as Error).message.includes('unique')) {
				return reply.status(409).send({ error: 'Username already exists' });
			}
			throw error;
		}
	});

	// Update user (requires auth, own profile only)
	fastify.put(
		'/:id',
		{ preHandler: [requireAuth] },
		async (request, reply) => {
			const { id } = request.params as { id: string };
			const updateData = request.body as {
				username?: string;
				age?: number;
			};

			if (!request.user) {
				return reply.status(401).send({ error: 'Unauthorized' });
			}

			if (request.user.id !== id) {
				return reply.status(403).send({ error: 'Cannot update other users' });
			}

			if (Object.keys(updateData).length === 0) {
				return reply.status(400).send({ error: 'No update data provided' });
			}

			try {
				const [updatedUser] = await db().update(user)
					.set(updateData)
					.where(eq(user.id, id))
					.returning({
						id: user.id,
						username: user.username,
						age: user.age,
						createdAt: user.createdAt,
						updatedAt: user.updatedAt,
					});

				if (!updatedUser) {
					return reply.status(404).send({ error: 'User not found' });
				}

				reply.header('content-type', 'application/msgpack');
				return { user: updatedUser };
			} catch (error) {
				if ((error as Error).message.includes('unique')) {
					return reply.status(409).send({ error: 'Username already exists' });
				}
				throw error;
			}
		},
	);

	// Delete user (requires auth, own profile only)
	fastify.delete(
		'/:id',
		{ preHandler: [requireAuth] },
		async (request, reply) => {
			const { id } = request.params as { id: string };

			if (!request.user) {
				return reply.status(401).send({ error: 'Unauthorized' });
			}

			if (request.user.id !== id) {
				return reply.status(403).send({ error: 'Cannot delete other users' });
			}

			await db().delete(user).where(eq(user.id, id));

			reply.header('content-type', 'application/msgpack');
			return { message: 'User deleted successfully' };
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
