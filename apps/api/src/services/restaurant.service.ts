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
	userId?: string;
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

export async function getRestaurants(options: RestaurantListOptions = {}) {
	const { limit = 20, offset = 0, userId } = options;

	const conditions = [eq(restaurant.isActive, 1)];
	if (userId) {
		conditions.push(eq(restaurant.createdBy, userId));
	}

	const restaurants = await db()
		.select()
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

export async function getRestaurantById(id: string): Promise<RestaurantWithDetails> {
	const restaurantResult = await db()
		.select()
		.from(restaurant)
		.where(eq(restaurant.id, id))
		.limit(1);

	const foundRestaurant = requireQueryResult(restaurantResult, 'Restaurant not found');

	const [categories, reviews] = await Promise.all([
		db()
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
			.where(eq(restaurantCategory.restaurantId, id)),

		db()
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
			.limit(10),
	]);

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

export async function createRestaurant(
	data: CreateRestaurantRequest,
	createdBy: string
): Promise<Restaurant> {
	const { name, categoryIds = [], ...rest } = data;

	if (!name) {
		throw new Error('Restaurant name is required');
	}

	try {
		const restaurantId = generateRestaurantId();

		const [newRestaurant] = await db()
			.insert(restaurant)
			.values({
				id: restaurantId,
				name,
				...rest,
				latitude: rest.latitude?.toString(),
				longitude: rest.longitude?.toString(),
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

export async function updateRestaurant(
	id: string,
	data: Partial<CreateRestaurantRequest>,
	userId: string
): Promise<Restaurant> {
	await checkOwnership(id, userId);

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

	if (!updatedRestaurant) {
		throw new Error('Restaurant not found');
	}

	return updatedRestaurant;
}

export async function deleteRestaurant(id: string, userId: string): Promise<void> {
	await checkOwnership(id, userId);

	await db()
		.update(restaurant)
		.set({ isActive: 0, updatedAt: new Date() })
		.where(eq(restaurant.id, id));
}

async function checkOwnership(restaurantId: string, userId: string): Promise<void> {
	const result = await db()
		.select({ createdBy: restaurant.createdBy })
		.from(restaurant)
		.where(eq(restaurant.id, restaurantId))
		.limit(1);

	const foundRestaurant = requireQueryResult(result, 'Restaurant not found');

	if (foundRestaurant.createdBy !== userId) {
		throw new Error('You can only modify restaurants you created');
	}
}
