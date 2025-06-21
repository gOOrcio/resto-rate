/**
 * Essential validation utilities
 */

// Basic input validation functions
export function validateUsername(username: unknown): username is string {
	return typeof username === 'string' && username.length >= 3 && username.length <= 31;
}

export function validatePassword(password: unknown): password is string {
	return typeof password === 'string' && password.length >= 6 && password.length <= 255;
}

export function validateEmail(email: unknown): email is string {
	return typeof email === 'string' && /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email);
}

export function validateAge(age: unknown): age is number {
	return typeof age === 'number' && age >= 13 && age <= 120;
}

export function validateRating(rating: unknown): rating is number {
	return typeof rating === 'number' && rating >= 1 && rating <= 5;
}

export function validatePriceRange(priceRange: unknown): priceRange is number {
	return typeof priceRange === 'number' && priceRange >= 1 && priceRange <= 4;
}

/**
 * Basic URL validation
 */
export function validateUrl(url: unknown): url is string {
	if (typeof url !== 'string') return false;
	try {
		new URL(url);
		return true;
	} catch {
		return false;
	}
}

/**
 * Basic phone number validation (very loose)
 */
export function validatePhoneNumber(phone: unknown): phone is string {
	return typeof phone === 'string' && /^[\d\s\-\+\(\)]+$/.test(phone) && phone.length >= 10;
}

/**
 * Validation error type
 */
export type ValidationError = {
	field: string;
	message: string;
};

/**
 * Validation result type
 */
export type ValidationResult = {
	valid: boolean;
	errors: ValidationError[];
};

/**
 * Create a validation error
 */
export function createValidationError(field: string, message: string): ValidationError {
	return { field, message };
}

/**
 * Combine multiple validation results
 */
export function combineValidationResults(...results: ValidationResult[]): ValidationResult {
	const errors = results.flatMap((r) => r.errors);
	return { valid: errors.length === 0, errors };
}

// Array utility functions for safe access
export function getFirstItem<T>(array: T[]): T | null {
	return array[0] || null;
}

export function requireFirstItem<T>(array: T[], errorMessage: string = 'Item not found'): T {
	const item = getFirstItem(array);
	if (!item) {
		throw new Error(errorMessage);
	}
	return item;
}

// Database query result utilities
export type QueryResult<T> = {
	found: boolean;
	item: T | null;
};

export function toQueryResult<T>(array: T[]): QueryResult<T> {
	const item = getFirstItem(array);
	return { found: item !== null, item };
}

export function requireQueryResult<T>(
	array: T[],
	notFoundMessage: string = 'Resource not found'
): T {
	const item = getFirstItem(array);
	if (!item) {
		throw new Error(notFoundMessage);
	}
	return item;
}

// =============================================================================
// AUTH VALIDATION
// =============================================================================

export * from './auth';
