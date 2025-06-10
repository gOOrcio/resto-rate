import type { FastifyPluginAsync } from 'fastify';
import { type CreateRestaurantRequest } from '@resto-rate/database';
import { requireAuth, optionalAuth } from '../middleware/auth';
import { restaurantService } from '../services/restaurant.service';

export const restaurantRoutes: FastifyPluginAsync = async (fastify) => {
	// Get all restaurants
	fastify.get('/', { preHandler: [optionalAuth] }, async (request, reply) => {
		try {
			const { limit = 20, offset = 0 } = request.query as {
				limit?: number;
				offset?: number;
			};

			const result = await restaurantService.getRestaurants({
				limit: Number(limit),
				offset: Number(offset),
			});

			reply.header('content-type', 'application/msgpack');
			return {
				...result,
				authenticated: !!request.user,
			};
		} catch (error) {
			return reply.status(500).send({ error: (error as Error).message });
		}
	});

	// Get restaurant by ID
	fastify.get('/:id', { preHandler: [optionalAuth] }, async (request, reply) => {
		try {
			const { id } = request.params as { id: string };

			const restaurant = await restaurantService.getRestaurantById(id);

			reply.header('content-type', 'application/msgpack');
			return { 
				restaurant: {
					...restaurant,
					reviewStats: restaurant.reviewStats,
				}, 
				reviews: restaurant.reviews, 
				authenticated: !!request.user 
			};
		} catch (error) {
			if ((error as Error).message === 'Restaurant not found') {
				return reply.status(404).send({ error: 'Restaurant not found' });
			}
			return reply.status(500).send({ error: (error as Error).message });
		}
	});

	// Create restaurant (requires auth)
	fastify.post<{ Body: CreateRestaurantRequest }>(
		'/',
		{ preHandler: [requireAuth] },
		async (request, reply) => {
			try {
				if (!request.user) {
					return reply.status(401).send({ error: 'Authentication required' });
				}

				const restaurant = await restaurantService.createRestaurant(
					request.body,
					request.user.id
				);

				reply.header('content-type', 'application/msgpack');
				return { restaurant };
			} catch (error) {
				if ((error as Error).message === 'Restaurant name is required') {
					return reply.status(400).send({ error: (error as Error).message });
				}
				if ((error as Error).message === 'Restaurant with this name already exists') {
					return reply.status(409).send({ error: (error as Error).message });
				}
				return reply.status(500).send({ error: (error as Error).message });
			}
		}
	);

	// Update restaurant (requires auth, owner only)
	fastify.put<{ Body: Partial<CreateRestaurantRequest>; Params: { id: string } }>(
		'/:id',
		{ preHandler: [requireAuth] },
		async (request, reply) => {
			try {
				const { id } = request.params;

				if (!request.user) {
					return reply.status(401).send({ error: 'Authentication required' });
				}

				const restaurant = await restaurantService.updateRestaurant(
					id,
					request.body,
					request.user.id
				);

				reply.header('content-type', 'application/msgpack');
				return { restaurant };
			} catch (error) {
				if ((error as Error).message === 'Restaurant not found') {
					return reply.status(404).send({ error: 'Restaurant not found' });
				}
				if ((error as Error).message === 'You can only update restaurants you created') {
					return reply.status(403).send({ error: (error as Error).message });
				}
				return reply.status(500).send({ error: (error as Error).message });
			}
		}
	);

	// Delete restaurant (requires auth, owner only)
	fastify.delete('/:id', { preHandler: [requireAuth] }, async (request, reply) => {
		try {
			const { id } = request.params as { id: string };

			if (!request.user) {
				return reply.status(401).send({ error: 'Authentication required' });
			}

			await restaurantService.deleteRestaurant(id, request.user.id);

			reply.header('content-type', 'application/msgpack');
			return { message: 'Restaurant deleted successfully' };
		} catch (error) {
			if ((error as Error).message === 'Restaurant not found') {
				return reply.status(404).send({ error: 'Restaurant not found' });
			}
			if ((error as Error).message === 'You can only delete restaurants you created') {
				return reply.status(403).send({ error: (error as Error).message });
			}
			return reply.status(500).send({ error: (error as Error).message });
		}
	});
};
