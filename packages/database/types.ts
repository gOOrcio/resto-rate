import * as schema from './schema';

// =============================================================================
// INFERRED TYPES FROM SCHEMA
// =============================================================================

export type Session = typeof schema.session.$inferSelect;
export type User = typeof schema.user.$inferSelect;
export type Category = typeof schema.category.$inferSelect;
export type Restaurant = typeof schema.restaurant.$inferSelect;
export type RestaurantCategory = typeof schema.restaurantCategory.$inferSelect;
export type Review = typeof schema.review.$inferSelect;
export type ReviewPhoto = typeof schema.reviewPhoto.$inferSelect;
export type ReviewHelpful = typeof schema.reviewHelpful.$inferSelect;

// Insert types
export type InsertUser = typeof schema.user.$inferInsert;
export type InsertSession = typeof schema.session.$inferInsert;
export type InsertRestaurant = typeof schema.restaurant.$inferInsert;
export type InsertReview = typeof schema.review.$inferInsert;
export type InsertCategory = typeof schema.category.$inferInsert;
export type InsertRestaurantCategory = typeof schema.restaurantCategory.$inferInsert;
export type InsertReviewPhoto = typeof schema.reviewPhoto.$inferInsert;
export type InsertReviewHelpful = typeof schema.reviewHelpful.$inferInsert;

// =============================================================================
// API REQUEST/RESPONSE TYPES
// =============================================================================

export type CreateUserRequest = {
	username: string;
	password: string;
	age?: number;
};

export type CreateRestaurantRequest = {
	name: string;
	description?: string;
	cuisineType?: string;
	address?: string;
	latitude?: number;
	longitude?: number;
	phone?: string;
	website?: string;
	priceRange?: number;
	categoryIds?: string[];
};

export type CreateReviewRequest = {
	restaurantId: string;
	rating: number;
	title?: string;
	content?: string;
	visitDate?: string;
};

export type UserResponse = Omit<User, 'passwordHash'>;

export type RestaurantResponse = Restaurant & {
	categories?: Category[];
	reviewStats?: {
		averageRating: number;
		totalReviews: number;
	};
};

export type ReviewResponse = Review & {
	user: UserResponse;
	restaurant: Pick<Restaurant, 'id' | 'name'>;
	photos?: ReviewPhoto[];
	helpfulCount?: number;
	userHelpfulVote?: boolean;
};

export type AuthResponse = {
	user: UserResponse;
	sessionId: string;
};

// =============================================================================
// VALIDATION HELPERS
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
// SESSION TYPES
// =============================================================================

export type SessionValidationResult = {
	session: Session | null;
	user: Pick<User, 'id' | 'username'> | null;
}; 