import { ulid } from 'ulid';

/**
 * Generate a new ULID (Universally Unique Lexicographically Sortable Identifier)
 * ULIDs are:
 * - 26 characters long
 * - Lexicographically sortable
 * - Canonically encoded as a 26 character string
 * - Uses Crockford's base32 for better efficiency and readability
 * - Case insensitive
 * - No special characters (URL safe)
 * - Monotonic sort order (correctly detects and handles the same millisecond)
 */
export function generateId(): string {
  return ulid();
}

/**
 * Generate a ULID with a specific timestamp
 * Useful for testing or when you need consistent ordering
 */
export function generateIdWithTime(timestamp: number): string {
  return ulid(timestamp);
}

/**
 * Extract timestamp from a ULID
 * Returns the timestamp in milliseconds since Unix epoch
 */
export function getTimeFromId(id: string): number {
  // ULID timestamp is first 10 characters (48 bits)
  const timestamp = id.substring(0, 10);
  return parseInt(timestamp, 32);
}

/**
 * Check if a string is a valid ULID format
 */
export function isValidId(id: string): boolean {
  // ULID is 26 characters long and uses Crockford's base32
  const ULID_REGEX = /^[0123456789ABCDEFGHJKMNPQRSTVWXYZ]{26}$/i;
  return ULID_REGEX.test(id);
}

/**
 * Generate a user ID (alias for generateId for clarity)
 */
export function generateUserId(): string {
  return generateId();
}

/**
 * Generate a session ID (alias for generateId for clarity)
 */
export function generateSessionId(): string {
  return generateId();
} 