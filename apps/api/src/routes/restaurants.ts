import type { FastifyPluginAsync } from 'fastify';
import { type CreateRestaurantRequest } from '@resto-rate/database';
import { requireAuth, optionalAuth } from '../middleware/auth';
import * as restaurantService from '../services/restaurant.service';
import { handleRoute, successMessage, requireUser } from '../utils/route-helpers';

export const restaurantRoutes: FastifyPluginAsync = async (fastify) => {
	// Get all restaurants
	fastify.get('/', { preHandler: [optionalAuth] }, async (request, reply) => {
		return handleRoute(reply, async () => {
			const { limit = 20, offset = 0 } = request.query as {
				limit?: number;
				offset?: number;
			};

			const result = await restaurantService.getRestaurants({
				limit: Number(limit),
				offset: Number(offset),
			});

			return {
				...result,
				authenticated: !!request.user,
			};
		});
	});

	// Get restaurant by ID
	fastify.get('/:id', { preHandler: [optionalAuth] }, async (request, reply) => {
		return handleRoute(reply, async () => {
			const { id } = request.params as { id: string };
			const restaurant = await restaurantService.getRestaurantById(id);

			return {
				restaurant: {
					...restaurant,
					reviewStats: restaurant.reviewStats,
				},
				reviews: restaurant.reviews,
				authenticated: !!request.user,
			};
		});
	});

	// Create restaurant (requires auth)
	fastify.post<{ Body: CreateRestaurantRequest }>(
		'/',
		{ preHandler: [requireAuth] },
		async (request, reply) => {
			return handleRoute(reply, async () => {
				const userId = requireUser(request.user?.id);
				const restaurant = await restaurantService.createRestaurant(request.body, userId);
				return { restaurant };
			});
		}
	);

	// Update restaurant (requires auth, owner only)
	fastify.put<{ Body: Partial<CreateRestaurantRequest>; Params: { id: string } }>(
		'/:id',
		{ preHandler: [requireAuth] },
		async (request, reply) => {
			return handleRoute(reply, async () => {
				const { id } = request.params;
				const userId = requireUser(request.user?.id);

				const restaurant = await restaurantService.updateRestaurant(id, request.body, userId);
				return { restaurant };
			});
		}
	);

	// Delete restaurant (requires auth, owner only)
	fastify.delete('/:id', { preHandler: [requireAuth] }, async (request, reply) => {
		return handleRoute(reply, async () => {
			const { id } = request.params as { id: string };
			const userId = requireUser(request.user?.id);

			await restaurantService.deleteRestaurant(id, userId);
			return successMessage('Restaurant deleted successfully');
		});
	});
};
