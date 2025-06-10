import { pgTable, text, integer, timestamp, decimal, uuid, unique } from 'drizzle-orm/pg-core';

// Users table with ULID
export const user = pgTable('user', {
  id: text('id').primaryKey(), // Will use ULID
  age: integer('age'),
  username: text('username').notNull().unique(),
  passwordHash: text('password_hash').notNull(),
  createdAt: timestamp('created_at', { withTimezone: true, mode: 'date' }).defaultNow(),
  updatedAt: timestamp('updated_at', { withTimezone: true, mode: 'date' }).defaultNow(),
});

// Sessions table with ULID
export const session = pgTable('session', {
  id: text('id').primaryKey(), // Will use ULID
  userId: text('user_id')
    .notNull()
    .references(() => user.id, { onDelete: 'cascade' }),
  expiresAt: timestamp('expires_at', { withTimezone: true, mode: 'date' }).notNull(),
});

// Categories table
export const category = pgTable('category', {
  id: text('id').primaryKey(), // ULID
  name: text('name').notNull().unique(),
  slug: text('slug').notNull().unique(),
  description: text('description'),
  createdAt: timestamp('created_at', { withTimezone: true, mode: 'date' }).defaultNow(),
  updatedAt: timestamp('updated_at', { withTimezone: true, mode: 'date' }).defaultNow(),
});

// Restaurants table
export const restaurant = pgTable('restaurant', {
  id: text('id').primaryKey(), // ULID
  name: text('name').notNull(),
  description: text('description'),
  cuisineType: text('cuisine_type'),
  address: text('address'),
  latitude: decimal('latitude', { precision: 10, scale: 8 }),
  longitude: decimal('longitude', { precision: 11, scale: 8 }),
  phone: text('phone'),
  website: text('website'),
  priceRange: integer('price_range'), // 1-4 scale ($, $$, $$$, $$$$)
  averageRating: decimal('average_rating', { precision: 3, scale: 2 }),
  totalReviews: integer('total_reviews').default(0),
  isActive: integer('is_active').default(1), // 1 = true, 0 = false
  createdBy: text('created_by').references(() => user.id),
  createdAt: timestamp('created_at', { withTimezone: true, mode: 'date' }).defaultNow(),
  updatedAt: timestamp('updated_at', { withTimezone: true, mode: 'date' }).defaultNow(),
});

// Restaurant categories junction table
export const restaurantCategory = pgTable('restaurant_category', {
  id: text('id').primaryKey(), // ULID
  restaurantId: text('restaurant_id')
    .notNull()
    .references(() => restaurant.id, { onDelete: 'cascade' }),
  categoryId: text('category_id')
    .notNull()
    .references(() => category.id, { onDelete: 'cascade' }),
  createdAt: timestamp('created_at', { withTimezone: true, mode: 'date' }).defaultNow(),
}, (table) => ({
  uniqueRestaurantCategory: unique().on(table.restaurantId, table.categoryId),
}));

// Reviews table
export const review = pgTable('review', {
  id: text('id').primaryKey(), // ULID
  restaurantId: text('restaurant_id')
    .notNull()
    .references(() => restaurant.id, { onDelete: 'cascade' }),
  userId: text('user_id')
    .notNull()
    .references(() => user.id, { onDelete: 'cascade' }),
  rating: integer('rating').notNull(), // 1-5 stars
  title: text('title'),
  content: text('content'),
  visitDate: timestamp('visit_date', { withTimezone: true, mode: 'date' }),
  isVerified: integer('is_verified').default(0), // 1 = true, 0 = false
  helpfulCount: integer('helpful_count').default(0),
  createdAt: timestamp('created_at', { withTimezone: true, mode: 'date' }).defaultNow(),
  updatedAt: timestamp('updated_at', { withTimezone: true, mode: 'date' }).defaultNow(),
}, (table) => ({
  uniqueUserRestaurant: unique().on(table.restaurantId, table.userId),
}));

// Review photos table
export const reviewPhoto = pgTable('review_photo', {
  id: text('id').primaryKey(), // ULID
  reviewId: text('review_id')
    .notNull()
    .references(() => review.id, { onDelete: 'cascade' }),
  url: text('url').notNull(),
  caption: text('caption'),
  orderIndex: integer('order_index').default(0),
  createdAt: timestamp('created_at', { withTimezone: true, mode: 'date' }).defaultNow(),
});

// Review helpful votes table
export const reviewHelpful = pgTable('review_helpful', {
  id: text('id').primaryKey(), // ULID
  reviewId: text('review_id')
    .notNull()
    .references(() => review.id, { onDelete: 'cascade' }),
  userId: text('user_id')
    .notNull()
    .references(() => user.id, { onDelete: 'cascade' }),
  isHelpful: integer('is_helpful').notNull(), // 1 = helpful, 0 = not helpful
  createdAt: timestamp('created_at', { withTimezone: true, mode: 'date' }).defaultNow(),
}, (table) => ({
  uniqueUserReview: unique().on(table.reviewId, table.userId),
}));

// Type exports
export type Session = typeof session.$inferSelect;
export type User = typeof user.$inferSelect;
export type Category = typeof category.$inferSelect;
export type Restaurant = typeof restaurant.$inferSelect;
export type RestaurantCategory = typeof restaurantCategory.$inferSelect;
export type Review = typeof review.$inferSelect;
export type ReviewPhoto = typeof reviewPhoto.$inferSelect;
export type ReviewHelpful = typeof reviewHelpful.$inferSelect;

// Insert types
export type InsertUser = typeof user.$inferInsert;
export type InsertRestaurant = typeof restaurant.$inferInsert;
export type InsertReview = typeof review.$inferInsert;
export type InsertCategory = typeof category.$inferInsert;

// API request/response types
export type CreateUserRequest = {
  username: string;
  password: string;
  age?: number;
};

export type CreateRestaurantRequest = {
  name: string;
  description?: string;
  cuisineType?: string;
  address?: string;
  latitude?: number;
  longitude?: number;
  phone?: string;
  website?: string;
  priceRange?: number;
  categoryIds?: string[];
};

export type CreateReviewRequest = {
  restaurantId: string;
  rating: number;
  title?: string;
  content?: string;
  visitDate?: string;
};

export type UserResponse = Omit<User, 'passwordHash'>;
export type RestaurantResponse = Restaurant & {
  categories?: Category[];
  reviewStats?: {
    averageRating: number;
    totalReviews: number;
  };
};

export type ReviewResponse = Review & {
  user: UserResponse;
  restaurant: Pick<Restaurant, 'id' | 'name'>;
  photos?: ReviewPhoto[];
  helpfulCount?: number;
  userHelpfulVote?: boolean;
};

export type AuthResponse = {
  user: UserResponse;
  sessionId: string;
}; 