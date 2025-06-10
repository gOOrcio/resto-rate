import type { Config } from 'drizzle-kit';
import { config } from 'dotenv';
import { resolve } from 'path';

config({ path: resolve(process.cwd(), '../../.env') });

const databaseUrl = process.env.DATABASE_URL || 'postgresql://postgres:password@localhost:5432/resto_rate';

export default {
	schema: '../../packages/database/schema.ts',
	out: './src/db/migrations',
	dialect: 'postgresql',
	dbCredentials: {
		url: databaseUrl,
	},
	verbose: true,
	strict: true,
} satisfies Config;
