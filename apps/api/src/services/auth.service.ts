import { db } from '../db';
import { session, user, type AuthResponse } from '@resto-rate/database';
import { eq, and, gt } from 'drizzle-orm';
import { requireQueryResult } from '@resto-rate/validation';

export class AuthService {
	async verifySession(sessionId: string): Promise<AuthResponse> {
		const result = await db()
			.select({
				session: session,
				user: {
					id: user.id,
					username: user.username,
					age: user.age,
					createdAt: user.createdAt,
					updatedAt: user.updatedAt,
				},
			})
			.from(session)
			.innerJoin(user, eq(session.userId, user.id))
			.where(and(eq(session.id, sessionId), gt(session.expiresAt, new Date())))
			.limit(1);

		const sessionResult = requireQueryResult(result, 'Session not found or expired');

		return {
			user: sessionResult.user,
			sessionId: sessionId,
		};
	}

	async getSession(sessionId: string): Promise<AuthResponse> {
		return this.verifySession(sessionId);
	}

	async invalidateSession(sessionId: string): Promise<void> {
		await db().delete(session).where(eq(session.id, sessionId));
	}
}

export const authService = new AuthService(); 