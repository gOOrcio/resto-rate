# Validation Package

This package provides validation utilities and safe array access functions for the Resto Rate application.

## Array Utilities

### Safe Array Access

When working with database queries that return arrays, TypeScript often complains about potentially undefined array access. These utilities provide safe ways to handle such scenarios:

```typescript
import { getFirstItem, requireQueryResult, toQueryResult } from '@resto-rate/validation';

// Example: Database query that returns an array
const restaurants = await db()
  .select()
  .from(restaurant)
  .where(eq(restaurant.id, id))
  .limit(1);

// ❌ Problem: TypeScript error "Object is possibly 'undefined'"
// const found = restaurants[0]; // Error!

// ✅ Solution 1: Safe access with null return
const found = getFirstItem(restaurants);
if (!found) {
  return reply.status(404).send({ error: 'Not found' });
}
// Now `found` is guaranteed to be defined

// ✅ Solution 2: Throw error if not found
try {
  const found = requireQueryResult(restaurants, 'Restaurant not found');
  // `found` is guaranteed to be defined here
} catch (error) {
  // Handle the error (will be thrown with your custom message)
}

// ✅ Solution 3: Query result with metadata
const result = toQueryResult(restaurants);
if (!result.found) {
  return reply.status(404).send({ error: 'Not found' });
}
// result.item is guaranteed to be defined when result.found is true
```

### Function Reference

#### `getFirstItem<T>(array: T[]): T | null`
Returns the first item of an array, or `null` if the array is empty.

#### `requireFirstItem<T>(array: T[], errorMessage?: string): T`
Returns the first item of an array, or throws an error if the array is empty.

#### `requireQueryResult<T>(array: T[], notFoundMessage?: string): T`  
Similar to `requireFirstItem` but specifically designed for database query results.

#### `toQueryResult<T>(array: T[]): QueryResult<T>`
Returns an object with `found: boolean` and `item: T | null` properties.

## Validation Functions

The package also includes various validation functions for common data types:

- `validateUsername(username: unknown): username is string`
- `validatePassword(password: unknown): password is string`
- `validateEmail(email: unknown): email is string`
- `validateAge(age: unknown): age is number`
- `validateRating(rating: unknown): rating is number`
- `validatePriceRange(priceRange: unknown): priceRange is number`
- `validateUrl(url: unknown): url is string`
- `validatePhoneNumber(phone: unknown): phone is string`

## Error Handling

```typescript
import { createValidationError, combineValidationResults, type ValidationResult } from '@resto-rate/validation';

const result1: ValidationResult = { valid: true, errors: [] };
const result2: ValidationResult = { 
  valid: false, 
  errors: [createValidationError('username', 'Username is required')] 
};

const combined = combineValidationResults(result1, result2);
// combined.valid will be false if any result is invalid
``` 