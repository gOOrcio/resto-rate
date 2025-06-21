import { db } from '../db';
import { user } from '@resto-rate/database';
import { type CreateUserRequest } from '@resto-rate/constants';
import { eq } from 'drizzle-orm';
import { generateUserId } from '@resto-rate/ulid';

export type UserResponse = Omit<typeof user.$inferSelect, 'passwordHash'>;

export async function getAllUsers(): Promise<UserResponse[]> {
	return db()
		.select({
			id: user.id,
			googleId: user.googleId,
			email: user.email,
			name: user.name,
			isAdmin: user.isAdmin,
			username: user.username,
			age: user.age,
			createdAt: user.createdAt,
			updatedAt: user.updatedAt,
		})
		.from(user);
}

export async function getUserById(id: string): Promise<UserResponse | null> {
	const users = await db()
		.select({
			id: user.id,
			googleId: user.googleId,
			email: user.email,
			name: user.name,
			isAdmin: user.isAdmin,
			username: user.username,
			age: user.age,
			createdAt: user.createdAt,
			updatedAt: user.updatedAt,
		})
		.from(user)
		.where(eq(user.id, id))
		.limit(1);

	return users[0] || null;
}

export async function createUser(data: CreateUserRequest): Promise<UserResponse> {
	const { username, password, age } = data;

	if (!username || !password) {
		throw new Error('Username and password are required');
	}

	try {
		const userId = generateUserId();
		const [newUser] = await db()
			.insert(user)
			.values({
				id: userId,
				username,
				passwordHash: password, // In real app, hash this!
				age,
			})
			.returning({
				id: user.id,
				googleId: user.googleId,
				email: user.email,
				name: user.name,
				isAdmin: user.isAdmin,
				username: user.username,
				age: user.age,
				createdAt: user.createdAt,
				updatedAt: user.updatedAt,
			});

		return newUser!;
	} catch (error) {
		if ((error as Error).message.includes('unique')) {
			throw new Error('Username already exists');
		}
		throw error;
	}
}

export async function updateUser(
	id: string,
	data: { username?: string; age?: number },
	requestUserId: string
): Promise<UserResponse> {
	if (requestUserId !== id) {
		throw new Error('Cannot update other users');
	}

	if (Object.keys(data).length === 0) {
		throw new Error('No update data provided');
	}

	try {
		const [updatedUser] = await db().update(user).set(data).where(eq(user.id, id)).returning({
			id: user.id,
			googleId: user.googleId,
			email: user.email,
			name: user.name,
			isAdmin: user.isAdmin,
			username: user.username,
			age: user.age,
			createdAt: user.createdAt,
			updatedAt: user.updatedAt,
		});

		if (!updatedUser) {
			throw new Error('User not found');
		}

		return updatedUser;
	} catch (error) {
		if ((error as Error).message.includes('unique')) {
			throw new Error('Username already exists');
		}
		throw error;
	}
}

export async function deleteUser(id: string, requestUserId: string): Promise<void> {
	if (requestUserId !== id) {
		throw new Error('Cannot delete other users');
	}

	await db().delete(user).where(eq(user.id, id));
}
