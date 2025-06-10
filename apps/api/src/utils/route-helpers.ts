import type { FastifyReply } from 'fastify';

// Standard error mapping
const ERROR_STATUS_MAP: Record<string, number> = {
	'not found': 404,
	'user not found': 404,
	'restaurant not found': 404,
	'session not found': 404,
	'already exists': 409,
	'username already exists': 409,
	'restaurant with this name already exists': 409,
	required: 400,
	invalid: 400,
	'no update data provided': 400,
	'cannot update': 403,
	'cannot delete': 403,
	permission: 403,
	'you can only': 403,
	'authentication required': 401,
	unauthorized: 401,
};

function getStatusFromError(message: string): number {
	const lowerMessage = message.toLowerCase();
	for (const [key, status] of Object.entries(ERROR_STATUS_MAP)) {
		if (lowerMessage.includes(key)) return status;
	}
	return 500;
}

/**
 * Standard response handler that automatically sets msgpack headers and handles errors
 */
export async function handleRoute<T>(
	reply: FastifyReply,
	operation: () => Promise<T>
): Promise<T | void> {
	try {
		const result = await operation();
		reply.header('content-type', 'application/msgpack');
		return result;
	} catch (error) {
		const message = (error as Error).message;
		const status = getStatusFromError(message);
		return reply.status(status).send({ error: message });
	}
}

/**
 * Simple success response with message
 */
export function successMessage(message: string) {
	return { message };
}

/**
 * Check if user is authenticated (for protected routes)
 */
export function requireUser(userId?: string): string {
	if (!userId) {
		throw new Error('Authentication required');
	}
	return userId;
}
