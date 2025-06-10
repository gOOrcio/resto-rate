/**
 * Essential validation utilities
 */

// Basic input validation functions
export function validateUsername(username: unknown): username is string {
	return (
		typeof username === 'string' &&
		username.length >= 3 &&
		username.length <= 31 &&
		/^[a-z0-9_-]+$/.test(username)
	);
}

export function validatePassword(password: unknown): password is string {
	return typeof password === 'string' && password.length >= 6 && password.length <= 255;
}

export function validateEmail(email: unknown): email is string {
	const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
	return typeof email === 'string' && emailRegex.test(email);
}

export function validateAge(age: unknown): age is number {
	return typeof age === 'number' && age >= 13 && age <= 120;
}

export function validateRating(rating: unknown): rating is number {
	return typeof rating === 'number' && rating >= 1 && rating <= 5 && Number.isInteger(rating);
}

export function validatePriceRange(priceRange: unknown): priceRange is number {
	return (
		typeof priceRange === 'number' &&
		priceRange >= 1 &&
		priceRange <= 4 &&
		Number.isInteger(priceRange)
	);
}

export function validateUrl(url: unknown): url is string {
	if (typeof url !== 'string') return false;
	try {
		new URL(url);
		return true;
	} catch {
		return false;
	}
}

export function validatePhoneNumber(phone: unknown): phone is string {
	// Basic phone validation - adjust regex as needed for your use case
	const phoneRegex = /^[\+]?[1-9][\d]{0,15}$/;
	return typeof phone === 'string' && phoneRegex.test(phone.replace(/[\s\-\(\)]/g, ''));
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
	const errors = results.flatMap((result) => result.errors);
	return {
		valid: errors.length === 0,
		errors,
	};
}

// Array utility functions for safe access
export function getFirstItem<T>(array: T[]): T | null {
	return array.length > 0 ? array[0]! : null;
}

export function requireFirstItem<T>(array: T[], errorMessage: string = 'Item not found'): T {
	if (array.length === 0) {
		throw new Error(errorMessage);
	}
	return array[0]!;
}

// Database query result utilities
export type QueryResult<T> = {
	found: boolean;
	item: T | null;
};

export function toQueryResult<T>(array: T[]): QueryResult<T> {
	return {
		found: array.length > 0,
		item: array.length > 0 ? array[0]! : null,
	};
}

export function requireQueryResult<T>(
	array: T[],
	notFoundMessage: string = 'Resource not found'
): T {
	if (array.length === 0) {
		throw new Error(notFoundMessage);
	}
	return array[0]!;
}
