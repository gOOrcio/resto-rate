import type { FastifyPluginAsync } from 'fastify';
import { type CreateRestaurantRequest } from '@resto-rate/constants';
import * as restaurantService from '../services/restaurant.service';
import { handleRoute } from '../utils/route-helpers';

export const restaurantRoutes: FastifyPluginAsync = async (fastify) => {
	// Get all restaurants
	fastify.get('/', async (request, reply) => {
		return handleRoute(reply, async () => {
			const restaurants = await restaurantService.getRestaurants();
			return { restaurants };
		});
	});

	// Get restaurant by ID
	fastify.get('/:id', async (request, reply) => {
		return handleRoute(reply, async () => {
			const { id } = request.params as { id: string };
			const restaurant = await restaurantService.getRestaurantById(id);

			if (!restaurant) {
				throw new Error('Restaurant not found');
			}

			return { restaurant };
		});
	});

	// Create restaurant
	fastify.post<{ Body: CreateRestaurantRequest }>('/', async (request, reply) => {
		return handleRoute(reply, async () => {
			const restaurant = await restaurantService.createRestaurant(request.body);
			return { restaurant };
		});
	});

	// Update restaurant
	fastify.patch('/:id', async (request, reply) => {
		return handleRoute(reply, async () => {
			const { id } = request.params as { id: string };
			const restaurant = await restaurantService.updateRestaurant(
				id,
				request.body as Partial<{
					name: string;
					address: string;
					rating: number;
					comment: string;
				}>
			);
			return { restaurant };
		});
	});

	// Delete restaurant
	fastify.delete('/:id', async (request, reply) => {
		return handleRoute(reply, async () => {
			const { id } = request.params as { id: string };
			await restaurantService.deleteRestaurant(id);
			return { message: 'Restaurant deleted successfully' };
		});
	});
};
