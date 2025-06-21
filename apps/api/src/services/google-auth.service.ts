import { db } from '../db';
import { user } from '@resto-rate/database';
import { type GoogleUserInfo, type GoogleTokens, type User } from '@resto-rate/database';
import { eq } from 'drizzle-orm';
import { generateUserId } from '@resto-rate/ulid';
import { getAuthConfig } from '@resto-rate/config';

/**
 * Exchange authorization code for Google OAuth tokens
 */
export async function exchangeCodeForTokens(code: string): Promise<GoogleTokens> {
	const authConfig = getAuthConfig();
	
	const tokenResponse = await fetch('https://oauth2.googleapis.com/token', {
		method: 'POST',
		headers: {
			'Content-Type': 'application/x-www-form-urlencoded',
		},
		body: new URLSearchParams({
			client_id: authConfig.googleClientId,
			client_secret: authConfig.googleClientSecret,
			code,
			grant_type: 'authorization_code',
			redirect_uri: authConfig.googleRedirectUri,
		}),
	});

	if (!tokenResponse.ok) {
		const errorText = await tokenResponse.text();
		throw new Error(`Failed to exchange code for tokens: ${tokenResponse.status} ${errorText}`);
	}

	const tokens = await tokenResponse.json() as GoogleTokens;
	return tokens;
}

/**
 * Get user information from Google using access token
 */
export async function getGoogleUserInfo(accessToken: string): Promise<GoogleUserInfo> {
	const userInfoResponse = await fetch(
		`https://www.googleapis.com/oauth2/v2/userinfo?access_token=${accessToken}`
	);

	if (!userInfoResponse.ok) {
		const errorText = await userInfoResponse.text();
		throw new Error(`Failed to get user info: ${userInfoResponse.status} ${errorText}`);
	}

	const userInfo = await userInfoResponse.json() as GoogleUserInfo;
	return userInfo;
}

/**
 * Create or update user from Google OAuth information
 */
export async function createOrUpdateUserFromGoogle(googleUser: GoogleUserInfo): Promise<User> {
	// Check if user already exists by Google ID
	const existingUserByGoogleId = await db()
		.select()
		.from(user)
		.where(eq(user.googleId, googleUser.id))
		.limit(1);

	if (existingUserByGoogleId.length > 0) {
		// Update existing user
		const existingUser = existingUserByGoogleId[0];
		if (!existingUser) {
			throw new Error('User not found');
		}
		
		const [updatedUser] = await db()
			.update(user)
			.set({
				email: googleUser.email,
				name: googleUser.name,
				updatedAt: new Date(),
			})
			.where(eq(user.id, existingUser.id))
			.returning();

		return updatedUser!;
	}

	// Check if user exists by email (for linking existing accounts)
	const existingUserByEmail = await db()
		.select()
		.from(user)
		.where(eq(user.email, googleUser.email))
		.limit(1);

	if (existingUserByEmail.length > 0) {
		// Link existing account to Google
		const existingUser = existingUserByEmail[0];
		if (!existingUser) {
			throw new Error('User not found');
		}
		
		const [updatedUser] = await db()
			.update(user)
			.set({
				googleId: googleUser.id,
				name: googleUser.name,
				updatedAt: new Date(),
			})
			.where(eq(user.id, existingUser.id))
			.returning();

		return updatedUser!;
	}

	// Create new user
	const userId = generateUserId();
	const [newUser] = await db()
		.insert(user)
		.values({
			id: userId,
			googleId: googleUser.id,
			email: googleUser.email,
			name: googleUser.name,
			isAdmin: false, // Default to regular user
		})
		.returning();

	return newUser!;
}

/**
 * Generate Google OAuth authorization URL
 */
export function generateGoogleAuthUrl(): string {
	const authConfig = getAuthConfig();
	
	// Use frontend callback URL instead of backend
	const redirectUri = 'http://localhost:5173/auth/callback';
	
	const params = new URLSearchParams({
		client_id: authConfig.googleClientId,
		redirect_uri: redirectUri,
		response_type: 'code',
		scope: 'email profile',
		access_type: 'offline',
		prompt: 'consent',
	});

	return `https://accounts.google.com/o/oauth2/v2/auth?${params.toString()}`;
} 