import { db } from '../db';
import { session, user, type AuthResponse } from '@resto-rate/database';
import { eq, and, gt } from 'drizzle-orm';
import { requireQueryResult } from '@resto-rate/validation';
import { generateUserId, generateSessionId } from '@resto-rate/ulid';
import { hash, verify } from '@node-rs/argon2';
import { sha256 } from '@oslojs/crypto/sha2';
import { encodeHexLowerCase } from '@oslojs/encoding';

const DAY_IN_MS = 1000 * 60 * 60 * 24;

export async function login(username: string, password: string): Promise<AuthResponse> {
	// Find user by username
	const users = await db().select().from(user).where(eq(user.username, username)).limit(1);
	const existingUser = users[0];
	
	if (!existingUser) {
		throw new Error('Invalid username or password');
	}

	// Verify password
	const validPassword = await verify(existingUser.passwordHash, password, {
		memoryCost: 19456,
		timeCost: 2,
		outputLen: 32,
		parallelism: 1,
	});

	if (!validPassword) {
		throw new Error('Invalid username or password');
	}

	// Create session
	const sessionToken = generateSessionId();
	const sessionId = encodeHexLowerCase(sha256(new TextEncoder().encode(sessionToken)));
	const expiresAt = new Date(Date.now() + DAY_IN_MS * 30);

	const newSession = {
		id: sessionId,
		userId: existingUser.id,
		expiresAt,
	};

	await db().insert(session).values(newSession);

	return {
		user: {
			id: existingUser.id,
			username: existingUser.username,
			age: existingUser.age,
			createdAt: existingUser.createdAt,
			updatedAt: existingUser.updatedAt,
		},
		sessionId: sessionToken, // Return the token, not the hashed ID
	};
}

export async function register(username: string, password: string, age?: number): Promise<AuthResponse> {
	// Basic validation
	if (username.length < 3 || username.length > 31) {
		throw new Error('Username must be between 3 and 31 characters');
	}
	if (password.length < 6 || password.length > 255) {
		throw new Error('Password must be between 6 and 255 characters');
	}

	// Hash password
	const passwordHash = await hash(password, {
		memoryCost: 19456,
		timeCost: 2,
		outputLen: 32,
		parallelism: 1,
	});

	// Create user
	const userId = generateUserId();
	const newUser = {
		id: userId,
		username,
		passwordHash,
		age: age || null,
	};

	try {
		await db().insert(user).values(newUser);
	} catch (error) {
		if ((error as Error).message.includes('unique')) {
			throw new Error('Username already exists');
		}
		throw error;
	}

	// Create session
	const sessionToken = generateSessionId();
	const sessionId = encodeHexLowerCase(sha256(new TextEncoder().encode(sessionToken)));
	const expiresAt = new Date(Date.now() + DAY_IN_MS * 30);

	const newSession = {
		id: sessionId,
		userId,
		expiresAt,
	};

	await db().insert(session).values(newSession);

	return {
		user: {
			id: userId,
			username,
			age: age || null,
			createdAt: new Date(),
			updatedAt: new Date(),
		},
		sessionId: sessionToken, // Return the token, not the hashed ID
	};
}

export async function verifySession(sessionId: string): Promise<AuthResponse> {
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

export async function getSession(sessionId: string): Promise<AuthResponse> {
	return verifySession(sessionId);
}

export async function invalidateSession(sessionId: string): Promise<void> {
	await db().delete(session).where(eq(session.id, sessionId));
}
