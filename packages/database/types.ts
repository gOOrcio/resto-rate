import * as schema from './schema';

// =============================================================================
// TABLE TYPES
// =============================================================================
export type Session = typeof schema.session.$inferSelect;
export type User = typeof schema.user.$inferSelect;
export type Restaurant = typeof schema.restaurant.$inferSelect;

// =============================================================================
// INSERT TYPES
// =============================================================================
export type InsertUser = typeof schema.user.$inferInsert;
export type InsertSession = typeof schema.session.$inferInsert;
export type InsertRestaurant = typeof schema.restaurant.$inferInsert;

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
	user: Pick<User, 'id' | 'username'> | null;
};
