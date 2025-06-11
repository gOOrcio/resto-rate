// Frontend environment variables
// The frontend should only need basic environment variables
// All database access is handled by the backend API

export const env = {
	NODE_ENV: typeof window === 'undefined' ? process.env.NODE_ENV : undefined,
};
