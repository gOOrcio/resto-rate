import { pgTable, text, integer, timestamp, boolean } from 'drizzle-orm/pg-core';

// =============================================================================
// CORE AUTHENTICATION TABLES
// =============================================================================

export const user = pgTable('user', {
	id: text('id').primaryKey(), // ULID
	googleId: text('google_id').unique(), // Google OAuth ID
	email: text('email').unique(), // Email from Google
	name: text('name'), // Full name from Google
	isAdmin: boolean('is_admin').default(false), // Future admin support
	username: text('username').unique(), // Optional for OAuth users
	passwordHash: text('password_hash'), // Optional for OAuth users
	age: integer('age'),
	createdAt: timestamp('created_at', { withTimezone: true, mode: 'date' }).defaultNow(),
	updatedAt: timestamp('updated_at', { withTimezone: true, mode: 'date' }).defaultNow(),
});

export const session = pgTable('session', {
	id: text('id').primaryKey(), // ULID
	userId: text('user_id')
		.notNull()
		.references(() => user.id, { onDelete: 'cascade' }),
	expiresAt: timestamp('expires_at', { withTimezone: true, mode: 'date' }).notNull(),
});

// =============================================================================
// BUSINESS DOMAIN TABLES
// =============================================================================

export const restaurant = pgTable('restaurant', {
	id: text('id').primaryKey(), // ULID
	name: text('name').notNull(),
	address: text('address'),
	rating: integer('rating'), // 1-5 scale
	comment: text('comment'),
	createdAt: timestamp('created_at', { withTimezone: true, mode: 'date' }).defaultNow(),
	updatedAt: timestamp('updated_at', { withTimezone: true, mode: 'date' }).defaultNow(),
});
