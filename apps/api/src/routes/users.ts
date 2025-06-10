import type { FastifyPluginAsync } from 'fastify';
import { db } from '../db';
import { user, type UserResponse, type CreateUserRequest } from '../db/schema';
import { eq } from 'drizzle-orm';
import { requireAuth, optionalAuth } from '../middleware/auth';
import { hash } from '@node-rs/argon2';
import { generateUserId } from '../lib/ulid';

export const userRoutes: FastifyPluginAsync = async (fastify) => {
  // Get all users (public endpoint with optional auth)
  fastify.get('/', { preHandler: [optionalAuth] }, async (request, reply) => {
    const users = await db()
      .select({
        id: user.id,
        username: user.username,
        age: user.age,
        createdAt: user.createdAt,
        updatedAt: user.updatedAt,
      })
      .from(user);

    reply.header('content-type', 'application/msgpack');
    return { users, authenticated: !!request.user };
  });

  // Get user by ID
  fastify.get('/:id', { preHandler: [optionalAuth] }, async (request, reply) => {
    const { id } = request.params as { id: string };

    const result = await db()
      .select({
        id: user.id,
        username: user.username,
        age: user.age,
        createdAt: user.createdAt,
        updatedAt: user.updatedAt,
      })
      .from(user)
      .where(eq(user.id, id))
      .limit(1);

    if (result.length === 0) {
      return reply.status(404).send({ error: 'User not found' });
    }

    reply.header('content-type', 'application/msgpack');
    return { user: result[0], authenticated: !!request.user };
  });

  // Create new user (public endpoint)
  fastify.post<{ Body: CreateUserRequest }>('/', async (request, reply) => {
    const { username, password, age } = request.body;

    if (!username || !password) {
      return reply.status(400).send({ error: 'Username and password are required' });
    }

    try {
      const userId = generateUserId();
      const passwordHash = await hash(password);

      const [newUser] = await db()
        .insert(user)
        .values({
          id: userId,
          username,
          passwordHash,
          age: age || null,
        })
        .returning({
          id: user.id,
          username: user.username,
          age: user.age,
          createdAt: user.createdAt,
          updatedAt: user.updatedAt,
        });

      reply.header('content-type', 'application/msgpack');
      return { user: newUser };
    } catch (error: any) {
      if (error.code === '23505') { // Unique constraint violation
        return reply.status(409).send({ error: 'Username already exists' });
      }
      throw error;
    }
  });

  // Update user (requires auth, users can only update themselves)
  fastify.put<{ Body: Partial<CreateUserRequest>; Params: { id: string } }>(
    '/:id',
    { preHandler: [requireAuth] },
    async (request, reply) => {
      const { id } = request.params;
      const { username, age } = request.body;

      if (!request.user) {
        return reply.status(401).send({ error: 'Authentication required' });
      }

      // Users can only update their own profile
      if (request.user.id !== id) {
        return reply.status(403).send({ error: 'Cannot update other users' });
      }

      const updateData: any = {};
      if (username) updateData.username = username;
      if (age !== undefined) updateData.age = age;

      if (Object.keys(updateData).length === 0) {
        return reply.status(400).send({ error: 'No valid fields to update' });
      }

      try {
        const [updatedUser] = await db()
          .update(user)
          .set(updateData)
          .where(eq(user.id, id))
          .returning({
            id: user.id,
            username: user.username,
            age: user.age,
            createdAt: user.createdAt,
            updatedAt: user.updatedAt,
          });

        reply.header('content-type', 'application/msgpack');
        return { user: updatedUser };
      } catch (error: any) {
        if (error.code === '23505') {
          return reply.status(409).send({ error: 'Username already exists' });
        }
        throw error;
      }
    }
  );

  // Delete user (requires auth, users can only delete themselves)
  fastify.delete('/:id', { preHandler: [requireAuth] }, async (request, reply) => {
    const { id } = request.params as { id: string };

    if (!request.user) {
      return reply.status(401).send({ error: 'Authentication required' });
    }

    // Users can only delete their own account
    if (request.user.id !== id) {
      return reply.status(403).send({ error: 'Cannot delete other users' });
    }

    await db().delete(user).where(eq(user.id, id));

    reply.header('content-type', 'application/msgpack');
    return { message: 'User deleted successfully' };
  });

  // Get current user profile
  fastify.get('/me/profile', { preHandler: [requireAuth] }, async (request, reply) => {
    if (!request.user) {
      return reply.status(401).send({ error: 'Authentication required' });
    }

    const userProfile: UserResponse = {
      id: request.user.id,
      username: request.user.username,
      age: request.user.age,
      createdAt: request.user.createdAt,
      updatedAt: request.user.updatedAt,
    };

    reply.header('content-type', 'application/msgpack');
    return { user: userProfile };
  });
}; 