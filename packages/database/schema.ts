import { pgTable, text, integer, timestamp } from 'drizzle-orm/pg-core';

// =============================================================================
// CORE AUTHENTICATION TABLES
// =============================================================================

export const user = pgTable('user', {
	id: text('id').primaryKey(), // ULID
	age: integer('age'),
	username: text('username').notNull().unique(),
	passwordHash: text('password_hash').notNull(),
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
