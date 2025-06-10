import { db } from '../db';
import { user, type CreateUserRequest, type User } from '@resto-rate/database';
import { eq } from 'drizzle-orm';
import { generateUserId } from '@resto-rate/ulid';
import { getFirstItem } from '@resto-rate/validation';

export type UserResponse = Omit<User, 'passwordHash'>;

export class UserService {
	async getAllUsers(): Promise<UserResponse[]> {
		const users = await db().select({
			id: user.id,
			username: user.username,
			age: user.age,
			createdAt: user.createdAt,
			updatedAt: user.updatedAt,
		}).from(user);

		return users;
	}

	async getUserById(id: string): Promise<UserResponse | null> {
		const users = await db().select({
			id: user.id,
			username: user.username,
			age: user.age,
			createdAt: user.createdAt,
			updatedAt: user.updatedAt,
		}).from(user).where(eq(user.id, id));

		return getFirstItem(users);
	}

	async createUser(data: CreateUserRequest): Promise<UserResponse> {
		const { username, password, age } = data;

		if (!username || !password) {
			throw new Error('Username and password are required');
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

			return newUser!;
		} catch (error) {
			if ((error as Error).message.includes('unique')) {
				throw new Error('Username already exists');
			}
			throw error;
		}
	}

	async updateUser(
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
			const [updatedUser] = await db().update(user)
				.set(data)
				.where(eq(user.id, id))
				.returning({
					id: user.id,
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

	async deleteUser(id: string, requestUserId: string): Promise<void> {
		if (requestUserId !== id) {
			throw new Error('Cannot delete other users');
		}

		await db().delete(user).where(eq(user.id, id));
	}
}

export const userService = new UserService(); 