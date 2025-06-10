import { drizzle } from 'drizzle-orm/postgres-js';
import postgres from 'postgres';
import * as schema from './schema';

let db: ReturnType<typeof drizzle> | null = null;

function getDb() {
  if (db) return db;
  
  // Import config after environment variables are loaded
  const { getDatabaseConfig } = require('@resto-rate/config');
  const dbConfig = getDatabaseConfig();

  if (!dbConfig.url) {
    throw new Error('DATABASE_URL is not set');
  }

  const client = postgres(dbConfig.url, {
    max: dbConfig.maxConnections,
    ssl: dbConfig.ssl
  });

  db = drizzle(client, { schema });
  return db;
}

// Export a proxy that lazy-loads the db
export { getDb as db }; 