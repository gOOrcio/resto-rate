import { drizzle } from 'drizzle-orm/postgres-js';
import postgres from 'postgres';
import * as schema from '@resto-rate/database';
import { env } from '$lib/env';

const client = postgres(env.DATABASE_URL, {
	max: env.DATABASE_MAX_CONNECTIONS ? parseInt(env.DATABASE_MAX_CONNECTIONS) : 20,
	ssl: env.DATABASE_SSL === 'true' || env.NODE_ENV === 'production',
});

export const db = drizzle(client, { schema });
