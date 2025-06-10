import type { FastifyPluginAsync } from 'fastify';
import { db } from '../db';
import {
	restaurant,
	category,
	restaurantCategory,
	review,
	user,
	type CreateRestaurantRequest,
	type RestaurantResponse,
} from '@resto-rate/database';
import { eq, desc } from 'drizzle-orm';
import { requireAuth, optionalAuth } from '../middleware/auth';
import { generateRestaurantId, generateId } from '@resto-rate/ulid';
import { requireQueryResult } from '@resto-rate/validation';

export const restaurantRoutes: FastifyPluginAsync = async (fastify) => {
	fastify.get('/', { preHandler: [optionalAuth] }, async (request, reply) => {
		const { limit = 20, offset = 0 } = request.query as {
			limit?: number;
			offset?: number;
			category?: string;
		};

		const query = db()
			.select({
				id: restaurant.id,
				name: restaurant.name,
				description: restaurant.description,
				cuisineType: restaurant.cuisineType,
				address: restaurant.address,
				latitude: restaurant.latitude,
				longitude: restaurant.longitude,
				phone: restaurant.phone,
				website: restaurant.website,
				priceRange: restaurant.priceRange,
				averageRating: restaurant.averageRating,
				totalReviews: restaurant.totalReviews,
				isActive: restaurant.isActive,
				createdAt: restaurant.createdAt,
				updatedAt: restaurant.updatedAt,
			})
			.from(restaurant)
			.where(eq(restaurant.isActive, 1))
			.orderBy(desc(restaurant.createdAt))
			.limit(Number(limit))
			.offset(Number(offset));

		const restaurants = await query;

		reply.header('content-type', 'application/msgpack');
		return {
			restaurants,
			pagination: { limit: Number(limit), offset: Number(offset) },
			authenticated: !!request.user,
		};
	});

	fastify.get('/:id', { preHandler: [optionalAuth] }, async (request, reply) => {
		const { id } = request.params as { id: string };

		// Get restaurant details
		const restaurantResult = await db()
			.select()
			.from(restaurant)
			.where(eq(restaurant.id, id))
			.limit(1);

		const foundRestaurant = requireQueryResult(restaurantResult, 'Restaurant not found');

		// Get restaurant categories
		const categories = await db()
			.select({
				id: category.id,
				name: category.name,
				slug: category.slug,
				description: category.description,
				createdAt: category.createdAt,
				updatedAt: category.updatedAt,
			})
			.from(category)
			.innerJoin(restaurantCategory, eq(category.id, restaurantCategory.categoryId))
			.where(eq(restaurantCategory.restaurantId, id));

		const reviews = await db()
			.select({
				id: review.id,
				rating: review.rating,
				title: review.title,
				content: review.content,
				visitDate: review.visitDate,
				createdAt: review.createdAt,
				user: {
					id: user.id,
					username: user.username,
				},
			})
			.from(review)
			.innerJoin(user, eq(review.userId, user.id))
			.where(eq(review.restaurantId, id))
			.orderBy(desc(review.createdAt))
			.limit(10);

		const response: RestaurantResponse = {
			...foundRestaurant,
			categories,
			reviewStats: {
				averageRating: Number(foundRestaurant.averageRating) || 0,
				totalReviews: foundRestaurant.totalReviews || 0,
			},
		};

		reply.header('content-type', 'application/msgpack');
		return { restaurant: response, reviews, authenticated: !!request.user };
	});

	// Create restaurant (requires auth)
	fastify.post<{ Body: CreateRestaurantRequest }>(
		'/',
		{ preHandler: [requireAuth] },
		async (request, reply) => {
			const {
				name,
				description,
				cuisineType,
				address,
				latitude,
				longitude,
				phone,
				website,
				priceRange,
				categoryIds = [],
			} = request.body;

			if (!name) {
				return reply.status(400).send({ error: 'Restaurant name is required' });
			}

			if (!request.user) {
				return reply.status(401).send({ error: 'Authentication required' });
			}

			try {
				const restaurantId = generateRestaurantId();

				// Create restaurant
				const [newRestaurant] = await db()
					.insert(restaurant)
					.values({
						id: restaurantId,
						name,
						description,
						cuisineType,
						address,
						latitude: latitude?.toString(),
						longitude: longitude?.toString(),
						phone,
						website,
						priceRange,
						createdBy: request.user.id,
					})
					.returning();

				// Link categories if provided
				if (categoryIds.length > 0) {
					const categoryLinks = categoryIds.map((categoryId) => ({
						id: generateId(),
						restaurantId,
						categoryId,
					}));

					await db().insert(restaurantCategory).values(categoryLinks);
				}

				reply.header('content-type', 'application/msgpack');
				return { restaurant: newRestaurant };
			} catch (error) {
				if ((error as Error).message.includes('unique')) {
					return reply.status(409).send({ error: 'Restaurant with this name already exists' });
				}
				throw error;
			}
		}
	);

	// Update restaurant (requires auth, owner only)
	fastify.put<{ Body: Partial<CreateRestaurantRequest>; Params: { id: string } }>(
		'/:id',
		{ preHandler: [requireAuth] },
		async (request, reply) => {
			const { id } = request.params;
			const updateData = request.body;

			if (!request.user) {
				return reply.status(401).send({ error: 'Authentication required' });
			}

			// Check if restaurant exists and user owns it
			const existingRestaurant = await db()
				.select({ createdBy: restaurant.createdBy })
				.from(restaurant)
				.where(eq(restaurant.id, id))
				.limit(1);

			const foundRestaurant = requireQueryResult(existingRestaurant, 'Restaurant not found');

			if (foundRestaurant.createdBy !== request.user.id) {
				return reply.status(403).send({ error: 'You can only update restaurants you created' });
			}

			// Update restaurant
			const [updatedRestaurant] = await db()
				.update(restaurant)
				.set({
					...updateData,
					latitude: updateData.latitude?.toString(),
					longitude: updateData.longitude?.toString(),
					updatedAt: new Date(),
				})
				.where(eq(restaurant.id, id))
				.returning();

			reply.header('content-type', 'application/msgpack');
			return { restaurant: updatedRestaurant };
		}
	);

	// Delete restaurant (requires auth, owner only)
	fastify.delete('/:id', { preHandler: [requireAuth] }, async (request, reply) => {
		const { id } = request.params as { id: string };

		if (!request.user) {
			return reply.status(401).send({ error: 'Authentication required' });
		}

		// Check if restaurant exists and user owns it
		const existingRestaurant = await db()
			.select({ createdBy: restaurant.createdBy })
			.from(restaurant)
			.where(eq(restaurant.id, id))
			.limit(1);

		const foundRestaurant = requireQueryResult(existingRestaurant, 'Restaurant not found');

		if (foundRestaurant.createdBy !== request.user.id) {
			return reply.status(403).send({ error: 'You can only delete restaurants you created' });
		}

		// Soft delete (set isActive to false)
		await db()
			.update(restaurant)
			.set({ isActive: 0, updatedAt: new Date() })
			.where(eq(restaurant.id, id));

		reply.header('content-type', 'application/msgpack');
		return { message: 'Restaurant deleted successfully' };
	});
};
