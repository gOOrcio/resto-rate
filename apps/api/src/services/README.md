# Services Layer

This directory contains the business logic layer for the API. Services handle database operations, business rules, and data transformations, keeping the routes clean and focused on HTTP concerns.

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│     ROUTES      │    │    SERVICES     │    │    DATABASE     │
│                 │    │                 │    │                 │
│ • HTTP handling │───►│ • Business      │───►│ • Data storage  │
│ • Validation    │    │   logic         │    │ • Queries       │
│ • Error codes   │    │ • Data trans-   │    │ • Transactions  │
│ • Response      │    │   formations    │    │                 │
│   formatting    │    │ • Permissions   │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## Services

### RestaurantService (`restaurant.service.ts`)
Handles all restaurant-related operations:
- ✅ Get restaurants with filtering and pagination
- ✅ Get restaurant details with categories and reviews
- ✅ Create restaurants with category linking
- ✅ Update restaurants (owner-only)
- ✅ Delete restaurants (soft delete, owner-only)
- ✅ Check ownership permissions

### UserService (`user.service.ts`)
Manages user operations:
- ✅ Get all users
- ✅ Get user by ID
- ✅ Create new users
- ✅ Update user profiles (own profile only)
- ✅ Delete users (own profile only)

### AuthService (`auth.service.ts`)
Authentication and session management:
- ✅ Verify session tokens
- ✅ Get session information
- ✅ Invalidate sessions (logout)

## Benefits

### ✅ Separation of Concerns
- **Routes**: Handle HTTP specifics (status codes, headers, request/response)
- **Services**: Handle business logic and data operations
- **Database**: Handle data persistence

### ✅ Reusability
Services can be used across multiple routes or even different interfaces (REST, GraphQL, etc.)

### ✅ Testability
Business logic can be unit tested independently of HTTP layer

### ✅ Error Handling
Services throw descriptive errors that routes can map to appropriate HTTP status codes

### ✅ Type Safety
Full TypeScript support with proper input/output types

## Usage Examples

### In Routes
```typescript
import { restaurantService } from '../services/restaurant.service';

// Simple service call
const restaurants = await restaurantService.getRestaurants({ limit: 10 });

// Error handling
try {
  const restaurant = await restaurantService.getRestaurantById(id);
  return { restaurant };
} catch (error) {
  if (error.message === 'Restaurant not found') {
    return reply.status(404).send({ error: 'Restaurant not found' });
  }
  return reply.status(500).send({ error: error.message });
}
```

### Service Implementation Pattern
```typescript
export class ExampleService {
  async getItem(id: string): Promise<Item> {
    // 1. Validate input
    if (!id) throw new Error('ID is required');
    
    // 2. Database operation
    const items = await db().select().from(table).where(eq(table.id, id));
    
    // 3. Handle not found
    const item = requireQueryResult(items, 'Item not found');
    
    // 4. Transform/return data
    return item;
  }
}
```

## Error Handling Strategy

Services throw errors with descriptive messages that routes map to HTTP status codes:

| Service Error | HTTP Status | Description |
|---------------|-------------|-------------|
| "Item not found" | 404 | Resource doesn't exist |
| "Permission denied" | 403 | User lacks permission |
| "Invalid input" | 400 | Bad request data |
| "Already exists" | 409 | Conflict/duplicate |
| Generic errors | 500 | Server error |

## Future Extensions

Potential service additions:
- **ReviewService**: Handle review CRUD and statistics
- **CategoryService**: Manage restaurant categories
- **SearchService**: Advanced search and filtering
- **NotificationService**: User notifications
- **FileService**: Image uploads and management 