import { drizzle } from 'drizzle-orm/postgres-js';
import postgres from 'postgres';
import * as schema from '@resto-rate/database';
import { getDatabaseConfig } from '@resto-rate/config';

let db: ReturnType<typeof drizzle> | null = null;

function getDb() {
	if (db) return db;

	const dbConfig = getDatabaseConfig();

	if (!dbConfig.url) {
		throw new Error('DATABASE_URL is not set');
	}

	const client = postgres(dbConfig.url, {
		max: dbConfig.maxConnections,
		ssl: dbConfig.ssl,
	});

	db = drizzle(client, { schema });
	return db;
}

export { getDb as db };
