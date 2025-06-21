// Load environment variables from root .env file
import { config } from 'dotenv';
import { resolve, dirname } from 'path';
import { fileURLToPath } from 'url';

// Get __dirname equivalent for ES modules
const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

// Load .env from project root (three levels up from scripts/)
const envPath = resolve(__dirname, '../../../../.env');
console.log('Loading environment from:', envPath);
const result = config({ path: envPath });
if (result.error) {
	console.error('Failed to load .env file:', result.error);
} else {
	console.log('✅ Environment loaded successfully');
}

import { db } from '../db/index.js';
import { createServerLoggerFromEnv } from '@resto-rate/logger';

const logger = createServerLoggerFromEnv('migration');

async function migrateGoogleOAuth() {
	console.log('🔄 Starting Google OAuth migration...');

	try {
		// Add Google OAuth columns to user table
		console.log('📝 Adding Google OAuth columns to user table...');
		
		await db().execute(`
			ALTER TABLE "user" 
			ADD COLUMN IF NOT EXISTS "google_id" text UNIQUE,
			ADD COLUMN IF NOT EXISTS "email" text UNIQUE,
			ADD COLUMN IF NOT EXISTS "name" text,
			ADD COLUMN IF NOT EXISTS "is_admin" boolean DEFAULT false;
		`);

		// Make username and password_hash optional for OAuth users
		console.log('🔧 Making username and password_hash optional...');
		
		await db().execute(`
			ALTER TABLE "user" 
			ALTER COLUMN "username" DROP NOT NULL,
			ALTER COLUMN "password_hash" DROP NOT NULL;
		`);

		// Add indexes for performance
		console.log('📊 Adding performance indexes...');
		
		await db().execute(`
			CREATE INDEX IF NOT EXISTS idx_user_google_id ON "user"("google_id");
			CREATE INDEX IF NOT EXISTS idx_user_email ON "user"("email");
			CREATE INDEX IF NOT EXISTS idx_user_is_admin ON "user"("is_admin");
		`);

		console.log('✅ Google OAuth migration completed successfully!');
		console.log('   - Added google_id, email, name, is_admin columns');
		console.log('   - Made username and password_hash optional');
		console.log('   - Added performance indexes');
		
	} catch (error) {
		console.error('❌ Migration failed:', error);
		logger.error('Google OAuth migration failed', { error });
		process.exit(1);
	}
}

// Run the migration if this script is executed directly
if (import.meta.url === `file://${process.argv[1]}`) {
	migrateGoogleOAuth()
		.then(() => process.exit(0))
		.catch((error) => {
			console.error(error);
			process.exit(1);
		});
}

export { migrateGoogleOAuth }; 