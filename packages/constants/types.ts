// =============================================================================
// API REQUEST TYPES
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
// API RESPONSE TYPES
// =============================================================================

export type UserResponse = {
	id: string;
	googleId?: string | null;
	email?: string | null;
	name?: string | null;
	isAdmin?: boolean;
	username?: string | null;
	age?: number | null;
	createdAt: Date | null;
	updatedAt: Date | null;
};

export type RestaurantResponse = {
	id: string;
	name: string;
	address: string | null;
	rating: number | null;
	comment: string | null;
	createdAt: Date | null;
	updatedAt: Date | null;
};

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