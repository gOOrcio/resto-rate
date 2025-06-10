import { type SessionValidationResult, type Session, type User } from './types';

export function createSessionValidationResult(
	session: Session | null,
	user: Pick<User, 'id' | 'username'> | null
): SessionValidationResult {
	return { session, user };
}

export function isValidSession(result: SessionValidationResult): result is { 
	session: Session; 
	user: Pick<User, 'id' | 'username'> 
} {
	return result.session !== null && result.user !== null;
} 