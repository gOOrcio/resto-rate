import * as schema from './schema';

// =============================================================================
// DATABASE TABLE TYPES
// =============================================================================
export type Session = typeof schema.session.$inferSelect;
export type User = typeof schema.user.$inferSelect;
export type Restaurant = typeof schema.restaurant.$inferInsert;

// =============================================================================
// DATABASE INSERT TYPES
// =============================================================================
export type InsertUser = typeof schema.user.$inferInsert;
export type InsertSession = typeof schema.session.$inferInsert;
export type InsertRestaurant = typeof schema.restaurant.$inferInsert;

// =============================================================================
// GOOGLE OAUTH TYPES
// =============================================================================

export interface GoogleUserInfo {
	id: string;
	email: string;
	verified_email: boolean;
	name: string;
	given_name: string;
	family_name: string;
	locale: string;
}

export interface GoogleTokens {
	access_token: string;
	id_token: string;
	expires_in: number;
	refresh_token?: string;
	token_type: string;
}

// =============================================================================
// REQUEST TYPES
// =============================================================================

export type CreateUserRequest = {
	username: string;
	password: string;
	age?: number;
};

export type CreateRestaurantRequest = {
	name: string;
	address?: string;
	rating?: number; // 1-5 scale
	comment?: string;
};

// =============================================================================
// RESPONSE TYPES
// =============================================================================

export type UserResponse = Omit<User, 'passwordHash'>;

export type RestaurantResponse = Restaurant;

export type AuthResponse = {
	user: UserResponse;
	sessionId: string;
};

// =============================================================================
// ERROR TYPES
// =============================================================================

export type ValidationError = {
	field: string;
	message: string;
};

export type ApiError = {
	error: string;
	details?: ValidationError[];
};

// =============================================================================
// AUTH TYPES
// =============================================================================

export type SessionValidationResult = {
	session: Session | null;
	user: Pick<User, 'id' | 'username' | 'email' | 'name'> | null;
};
