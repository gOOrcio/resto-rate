import { resolve } from 'path';
import dotenv from 'dotenv';

// Load environment variables from project root
const envPath = resolve(process.cwd(), '../..', '.env');
dotenv.config({ path: envPath });

// Verify critical environment variables are loaded
if (!process.env.DATABASE_URL) {
	console.error('Environment variables loaded from:', envPath);
	console.error(
		'Available DATABASE_* variables:',
		Object.keys(process.env).filter((key) => key.startsWith('DATABASE_'))
	);
	throw new Error('DATABASE_URL is not set. Please check your .env file in the project root.');
}

export const env = {
	DATABASE_URL: process.env.DATABASE_URL,
	DATABASE_MAX_CONNECTIONS: process.env.DATABASE_MAX_CONNECTIONS,
	DATABASE_SSL: process.env.DATABASE_SSL,
	NODE_ENV: process.env.NODE_ENV,
};
