# Database Architecture Strategy

## 🎯 Recommended Approach: **Hybrid Architecture**

### Frontend Database Usage

- ✅ **Sessions Only** - Direct DB access for Lucia authentication
- ❌ **No Business Logic** - All other data operations via API

### Backend Database Usage

- ✅ **All Business Data** - Users, restaurants, reviews, ratings, etc.
- ✅ **Session Validation** - Read sessions for API authentication
- ✅ **Data Integrity** - Business logic, validations, relations

## 🔄 Data Flow Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   FRONTEND      │    │      API        │    │   DATABASE      │
│   (SvelteKit)   │    │   (Fastify)     │    │ (PostgreSQL)    │
├─────────────────┤    ├─────────────────┤    ├─────────────────┤
│                 │    │                 │    │                 │
│ 🔐 AUTH:        │◄──►│ 🔐 AUTH:        │◄──►│ 📋 TABLES:      │
│  • Login/Logout │    │  • Verify Token │    │  • users        │
│  • Session Mgmt │    │  • Middleware   │    │  • sessions     │
│                 │    │                 │    │  • restaurants  │
│ 📱 UI/UX:       │    │ 🏢 BUSINESS:    │    │  • reviews      │
│  • Forms        │◄──►│  • CRUD Ops     │◄──►│  • ratings      │
│  • State Mgmt   │    │  • Validation   │    │  • categories   │
│  • Caching      │    │  • Relations    │    │  • ...more      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                        │
         └────── MessagePack ─────┘
```

## 📊 Database Schema Design

### Core Authentication (Shared)

```sql
-- Managed by Frontend (Lucia)
TABLE users (
  id TEXT PRIMARY KEY,
  username TEXT UNIQUE NOT NULL,
  password_hash TEXT NOT NULL,
  created_at TIMESTAMP DEFAULT NOW()
);

TABLE sessions (
  id TEXT PRIMARY KEY,
  user_id TEXT REFERENCES users(id),
  expires_at TIMESTAMP NOT NULL
);
```

### Business Data (API Only)

```sql
-- Restaurants & Reviews Domain
TABLE restaurants (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name TEXT NOT NULL,
  description TEXT,
  cuisine_type TEXT,
  address TEXT,
  latitude DECIMAL,
  longitude DECIMAL,
  created_by TEXT REFERENCES users(id),
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);

TABLE reviews (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  restaurant_id UUID REFERENCES restaurants(id),
  user_id TEXT REFERENCES users(id),
  rating INTEGER CHECK (rating >= 1 AND rating <= 5),
  title TEXT,
  content TEXT,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW(),
  UNIQUE(restaurant_id, user_id) -- One review per user per restaurant
);

TABLE categories (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name TEXT UNIQUE NOT NULL,
  slug TEXT UNIQUE NOT NULL,
  description TEXT
);

TABLE restaurant_categories (
  restaurant_id UUID REFERENCES restaurants(id),
  category_id UUID REFERENCES categories(id),
  PRIMARY KEY (restaurant_id, category_id)
);
```

## 🔒 Security & Access Control

### Frontend (Direct DB Access)

```typescript
// ✅ ALLOWED: Session management only
import { lucia } from '$lib/server/auth';
import { db } from '$lib/server/db';

// Login, logout, session validation
const session = await lucia.createSession(userId, {});
```

### API (Full DB Access)

```typescript
// ✅ ALLOWED: All business operations
import { db } from './db';

// Restaurant operations
const restaurants = await db.select().from(restaurantsTable);
const reviews = await db.insert(reviewsTable).values(newReview);
```

## 🚀 Implementation Strategy

### 1. Extend Database Schema

```typescript
// apps/api/src/db/schema.ts
export const restaurants = pgTable('restaurants', {
	id: uuid('id').defaultRandom().primaryKey(),
	name: text('name').notNull(),
	description: text('description'),
	cuisineType: text('cuisine_type'),
	address: text('address'),
	latitude: decimal('latitude'),
	longitude: decimal('longitude'),
	createdBy: text('created_by').references(() => user.id),
	createdAt: timestamp('created_at').defaultNow(),
	updatedAt: timestamp('updated_at').defaultNow(),
});

export const reviews = pgTable(
	'reviews',
	{
		id: uuid('id').defaultRandom().primaryKey(),
		restaurantId: uuid('restaurant_id').references(() => restaurants.id),
		userId: text('user_id').references(() => user.id),
		rating: integer('rating').notNull(),
		title: text('title'),
		content: text('content'),
		createdAt: timestamp('created_at').defaultNow(),
		updatedAt: timestamp('updated_at').defaultNow(),
	},
	(table) => ({
		uniqueUserRestaurant: unique().on(table.restaurantId, table.userId),
	})
);
```

### 2. Frontend Data Layer

```typescript
// apps/web/src/lib/stores/restaurants.ts
import { writable } from 'svelte/store';
import { apiClient } from '$lib/api';

export const restaurants = writable([]);
export const loading = writable(false);

export async function loadRestaurants() {
	loading.set(true);
	try {
		const data = await apiClient.getRestaurants();
		restaurants.set(data.restaurants);
	} finally {
		loading.set(false);
	}
}
```

### 3. API Business Logic

```typescript
// apps/api/src/routes/restaurants.ts
export const restaurantRoutes: FastifyPluginAsync = async (fastify) => {
	// Get all restaurants
	fastify.get('/', async (request, reply) => {
		const restaurants = await db.select().from(restaurantsTable);
		reply.header('content-type', 'application/msgpack');
		return { restaurants };
	});

	// Create restaurant (auth required)
	fastify.post('/', { preHandler: [requireAuth] }, async (request, reply) => {
		const { name, description, cuisineType, address } = request.body;

		const [restaurant] = await db
			.insert(restaurantsTable)
			.values({
				name,
				description,
				cuisineType,
				address,
				createdBy: request.user!.id,
			})
			.returning();

		reply.header('content-type', 'application/msgpack');
		return { restaurant };
	});
};
```

## 📈 Benefits of This Architecture

### ✅ Advantages

- **Clear Separation**: Frontend focuses on UI/UX, API handles business logic
- **Security**: Centralized data validation and access control in API
- **Scalability**: API can serve multiple clients (web, mobile, etc.)
- **Performance**: Efficient MessagePack serialization
- **Maintainability**: Single source of truth for business logic
- **Caching**: Easy to implement API-level caching strategies

### ⚠️ Considerations

- **Network Latency**: Extra HTTP calls vs direct DB access
- **Complexity**: More moving parts than direct DB access
- **Development**: Need to maintain API contracts

## 🛠️ Migration Path

### Phase 1: Current State (Keep as-is)

- Frontend: Direct DB for sessions (Lucia)
- API: Sessions + Users management

### Phase 2: Extend Business Domain

- API: Add restaurants, reviews, categories
- Frontend: Use API client for all business data
- Keep: Direct session access for auth

### Phase 3: Optimization

- Add caching layers
- Implement real-time features (WebSockets)
- Add advanced search/filtering

## 📝 Recommendation

**Stick with the hybrid approach** because:

1. **Lucia requires direct DB access** for optimal performance
2. **Business logic belongs in the API** for better separation
3. **MessagePack provides efficiency** for data transfer
4. **Scalable architecture** for future requirements

This gives you the best of both worlds: fast authentication and clean business logic separation.
