import { db } from '../db';
import { restaurant, type CreateRestaurantRequest, type Restaurant } from '@resto-rate/database';
import { eq, desc } from 'drizzle-orm';
import { generateRestaurantId } from '@resto-rate/ulid';

export async function getRestaurants(): Promise<Restaurant[]> {
	return db().select().from(restaurant).orderBy(desc(restaurant.createdAt));
}

export async function getRestaurantById(id: string): Promise<Restaurant | null> {
	const restaurants = await db().select().from(restaurant).where(eq(restaurant.id, id)).limit(1);
	return restaurants[0] || null;
}

export async function createRestaurant(data: CreateRestaurantRequest): Promise<Restaurant> {
	const { name, address, rating, comment } = data;

	if (!name) {
		throw new Error('Restaurant name is required');
	}

	if (rating && (rating < 1 || rating > 5)) {
		throw new Error('Rating must be between 1 and 5');
	}

	const restaurantId = generateRestaurantId();

	const [newRestaurant] = await db()
		.insert(restaurant)
		.values({
			id: restaurantId,
			name,
			address,
			rating,
			comment,
		})
		.returning();

	return newRestaurant!;
}

export async function deleteRestaurant(id: string): Promise<void> {
	await db().delete(restaurant).where(eq(restaurant.id, id));
}
