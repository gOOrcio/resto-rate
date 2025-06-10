import { db } from '../db';
import {
	restaurant,
	category,
	restaurantCategory,
	review,
	user,
	type CreateRestaurantRequest,
	type Restaurant,
} from '@resto-rate/database';
import { eq, desc, and } from 'drizzle-orm';
import { generateRestaurantId, generateId } from '@resto-rate/ulid';
import { requireQueryResult } from '@resto-rate/validation';

export interface RestaurantListOptions {
	limit?: number;
	offset?: number;
	category?: string;
	userId?: string; // for filtering by user's restaurants
}

export interface RestaurantWithDetails extends Restaurant {
	categories: Array<{
		id: string;
		name: string;
		slug: string;
		description: string | null;
		createdAt: Date | null;
		updatedAt: Date | null;
	}>;
	reviews: Array<{
		id: string;
		rating: number;
		title: string | null;
		content: string | null;
		visitDate: Date | null;
		createdAt: Date | null;
		user: {
			id: string;
			username: string;
		};
	}>;
	reviewStats: {
		averageRating: number;
		totalReviews: number;
	};
}

export class RestaurantService {
	async getRestaurants(options: RestaurantListOptions = {}): Promise<{
		restaurants: Restaurant[];
		pagination: { limit: number; offset: number };
	}> {
		const { limit = 20, offset = 0, userId } = options;

		// Build where conditions
		const conditions = [eq(restaurant.isActive, 1)];
		if (userId) {
			conditions.push(eq(restaurant.createdBy, userId));
		}

		const restaurants = await db()
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
				createdBy: restaurant.createdBy,
				createdAt: restaurant.createdAt,
				updatedAt: restaurant.updatedAt,
			})
			.from(restaurant)
			.where(and(...conditions))
			.orderBy(desc(restaurant.createdAt))
			.limit(Number(limit))
			.offset(Number(offset));

		return {
			restaurants,
			pagination: { limit: Number(limit), offset: Number(offset) },
		};
	}

	async getRestaurantById(id: string): Promise<RestaurantWithDetails> {
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

		// Get restaurant reviews
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

		return {
			...foundRestaurant,
			categories,
			reviews,
			reviewStats: {
				averageRating: Number(foundRestaurant.averageRating) || 0,
				totalReviews: foundRestaurant.totalReviews || 0,
			},
		};
	}

	async createRestaurant(
		data: CreateRestaurantRequest,
		createdBy: string
	): Promise<Restaurant> {
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
		} = data;

		if (!name) {
			throw new Error('Restaurant name is required');
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
					createdBy,
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

			return newRestaurant!;
		} catch (error) {
			if ((error as Error).message.includes('unique')) {
				throw new Error('Restaurant with this name already exists');
			}
			throw error;
		}
	}

	async updateRestaurant(
		id: string,
		data: Partial<CreateRestaurantRequest>,
		userId: string
	): Promise<Restaurant> {
		// Check if restaurant exists and user owns it
		const existingRestaurant = await db()
			.select({ createdBy: restaurant.createdBy })
			.from(restaurant)
			.where(eq(restaurant.id, id))
			.limit(1);

		const foundRestaurant = requireQueryResult(existingRestaurant, 'Restaurant not found');

		if (foundRestaurant.createdBy !== userId) {
			throw new Error('You can only update restaurants you created');
		}

		// Update restaurant
		const [updatedRestaurant] = await db()
			.update(restaurant)
			.set({
				...data,
				latitude: data.latitude?.toString(),
				longitude: data.longitude?.toString(),
				updatedAt: new Date(),
			})
			.where(eq(restaurant.id, id))
			.returning();

		return updatedRestaurant!;
	}

	async deleteRestaurant(id: string, userId: string): Promise<void> {
		// Check if restaurant exists and user owns it
		const existingRestaurant = await db()
			.select({ createdBy: restaurant.createdBy })
			.from(restaurant)
			.where(eq(restaurant.id, id))
			.limit(1);

		const foundRestaurant = requireQueryResult(existingRestaurant, 'Restaurant not found');

		if (foundRestaurant.createdBy !== userId) {
			throw new Error('You can only delete restaurants you created');
		}

		// Soft delete (set isActive to false)
		await db()
			.update(restaurant)
			.set({ isActive: 0, updatedAt: new Date() })
			.where(eq(restaurant.id, id));
	}

	async checkOwnership(restaurantId: string, userId: string): Promise<boolean> {
		const existingRestaurant = await db()
			.select({ createdBy: restaurant.createdBy })
			.from(restaurant)
			.where(eq(restaurant.id, restaurantId))
			.limit(1);

		const foundRestaurant = requireQueryResult(existingRestaurant, 'Restaurant not found');
		return foundRestaurant.createdBy === userId;
	}
}

export const restaurantService = new RestaurantService(); 