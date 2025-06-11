/**
 * Shared constants across frontend and backend
 */

// =============================================================================
// TIME CONSTANTS
// =============================================================================

export const TIME_CONSTANTS = {
	DAY_IN_MS: 1000 * 60 * 60 * 24,
	SESSION_LIFETIME_DAYS: 30,
	SESSION_RENEWAL_THRESHOLD_DAYS: 15,
} as const;

// =============================================================================
// API ENDPOINTS
// =============================================================================

export const API_ENDPOINTS = {
	AUTH: {
		VERIFY: '/api/auth/verify',
		SESSION: '/api/auth/session',
		LOGOUT: '/api/auth/logout',
	},
	USERS: {
		BASE: '/api/users',
		ME: '/api/users/me/profile',
		BY_ID: (id: string) => `/api/users/${id}`,
	},
	RESTAURANTS: {
		BASE: '/api/restaurants',
		BY_ID: (id: string) => `/api/restaurants/${id}`,
	},
	REVIEWS: {
		BASE: '/api/reviews',
		BY_ID: (id: string) => `/api/reviews/${id}`,
		BY_RESTAURANT: (restaurantId: string) => `/api/restaurants/${restaurantId}/reviews`,
	},
	HEALTH: '/health',
} as const;

// =============================================================================
// HTTP STATUS CODES
// =============================================================================

export const HTTP_STATUS = {
	OK: 200,
	CREATED: 201,
	NO_CONTENT: 204,
	BAD_REQUEST: 400,
	UNAUTHORIZED: 401,
	FORBIDDEN: 403,
	NOT_FOUND: 404,
	CONFLICT: 409,
	UNPROCESSABLE_ENTITY: 422,
	INTERNAL_SERVER_ERROR: 500,
} as const;

// =============================================================================
// VALIDATION CONSTANTS
// =============================================================================

export const VALIDATION_LIMITS = {
	USERNAME: {
		MIN_LENGTH: 3,
		MAX_LENGTH: 31,
	},
	PASSWORD: {
		MIN_LENGTH: 6,
		MAX_LENGTH: 255,
	},
	AGE: {
		MIN: 13,
		MAX: 120,
	},
	RATING: {
		MIN: 1,
		MAX: 5,
	},
	PRICE_RANGE: {
		MIN: 1,
		MAX: 4,
	},
	RESTAURANT_NAME: {
		MIN_LENGTH: 1,
		MAX_LENGTH: 255,
	},
	REVIEW_TITLE: {
		MAX_LENGTH: 255,
	},
	REVIEW_CONTENT: {
		MAX_LENGTH: 2000,
	},
} as const;

// =============================================================================
// ERROR MESSAGES
// =============================================================================

export const ERROR_MESSAGES = {
	VALIDATION: {
		REQUIRED_FIELD: (field: string) => `${field} is required`,
		INVALID_USERNAME: `Username must be 3-31 characters, alphanumeric with underscores and hyphens only`,
		INVALID_PASSWORD: `Password must be 6-255 characters`,
		INVALID_EMAIL: 'Please enter a valid email address',
		INVALID_AGE: 'Age must be between 13 and 120',
		INVALID_RATING: 'Rating must be between 1 and 5',
		INVALID_PRICE_RANGE: 'Price range must be between 1 and 4',
	},
	AUTH: {
		INVALID_CREDENTIALS: 'Invalid username or password',
		SESSION_EXPIRED: 'Your session has expired. Please log in again.',
		UNAUTHORIZED: 'You must be logged in to access this resource',
		FORBIDDEN: 'You do not have permission to access this resource',
	},
	GENERAL: {
		NOT_FOUND: 'The requested resource was not found',
		SERVER_ERROR: 'An internal server error occurred',
		NETWORK_ERROR: 'A network error occurred. Please try again.',
	},
} as const;

// =============================================================================
// REGEX PATTERNS
// =============================================================================

export const REGEX_PATTERNS = {
	USERNAME: /^[a-z0-9_-]+$/,
	EMAIL: /^[^\s@]+@[^\s@]+\.[^\s@]+$/,
	PHONE: /^[\+]?[1-9][\d]{0,15}$/,
	ULID: /^[0123456789ABCDEFGHJKMNPQRSTVWXYZ]{26}$/i,
} as const;

// =============================================================================
// CONTENT TYPES
// =============================================================================

export const CONTENT_TYPES = {
	JSON: 'application/json',
	MSGPACK: 'application/msgpack',
	FORM_DATA: 'multipart/form-data',
	URL_ENCODED: 'application/x-www-form-urlencoded',
} as const;

// =============================================================================
// SESSION CONSTANTS
// =============================================================================

export const SESSION_CONSTANTS = {
	COOKIE_NAME: 'auth-session',
	HEADER_NAME: 'X-Session-Id',
	BCRYPT_ROUNDS: 12,
} as const;

export * from './types';
