import type { Handle } from '@sveltejs/kit';

// Simple pass-through handler since all authentication is handled by the backend API
const handleAuth: Handle = async ({ event, resolve }) => {
	// All authentication is now handled by the backend API
	// The frontend is purely a client that communicates via API calls
	event.locals.user = null;
	event.locals.session = null;
	return resolve(event);
};

export const handle: Handle = handleAuth;
