// =============================================================================
// AUTH VALIDATION TYPES
// =============================================================================

export type SessionValidationResult = {
	session: {
		id: string;
		userId: string;
		expiresAt: Date;
	} | null;
	user: {
		id: string;
		username: string;
	} | null;
};

// =============================================================================
// AUTH VALIDATION UTILITIES
// =============================================================================

export function createSessionValidationResult(
	session: SessionValidationResult['session'],
	user: SessionValidationResult['user']
): SessionValidationResult {
	return { session, user };
}

export function isValidSession(result: SessionValidationResult): result is {
	session: NonNullable<SessionValidationResult['session']>;
	user: NonNullable<SessionValidationResult['user']>;
} {
	return result.session !== null && result.user !== null;
}
