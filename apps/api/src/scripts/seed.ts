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
import { user, restaurant } from '@resto-rate/database';
import { generateUserId, generateRestaurantId } from '@resto-rate/ulid';

async function seed() {
	console.log('🌱 Seeding database...');

	try {
		// Create sample users
		const users = [
			{
				id: generateUserId(),
				username: 'admin',
				passwordHash: 'hashed_admin_password', // In real app, use proper hashing
				age: 30,
			},
			{
				id: generateUserId(),
				username: 'testuser',
				passwordHash: 'hashed_test_password',
				age: 25,
			},
			{
				id: generateUserId(),
				username: 'reviewer',
				passwordHash: 'hashed_reviewer_password',
				age: 35,
			},
		];

		console.log('👥 Creating users...');
		for (const userData of users) {
			await db().insert(user).values(userData).onConflictDoNothing();
		}

		// Create sample restaurants
		const restaurants = [
			{
				id: generateRestaurantId(),
				name: 'The Italian Corner',
				address: '123 Main St, City, State',
				rating: 5,
				comment: 'Excellent pasta and friendly service!',
			},
			{
				id: generateRestaurantId(),
				name: 'Burger Palace',
				address: '456 Oak Ave, City, State',
				rating: 4,
				comment: 'Great burgers, but can be crowded on weekends.',
			},
			{
				id: generateRestaurantId(),
				name: 'Sushi Zen',
				address: '789 Pine St, City, State',
				rating: 5,
				comment: 'Fresh sushi and beautiful presentation.',
			},
			{
				id: generateRestaurantId(),
				name: 'Pizza Express',
				address: '321 Elm St, City, State',
				rating: 3,
				comment: 'Quick delivery but pizza could be better.',
			},
			{
				id: generateRestaurantId(),
				name: 'Café Delight',
				address: '654 Maple Blvd, City, State',
				rating: 4,
				comment: 'Perfect spot for morning coffee and pastries.',
			},
		];

		console.log('🍽️  Creating restaurants...');
		for (const restaurantData of restaurants) {
			await db().insert(restaurant).values(restaurantData).onConflictDoNothing();
		}

		console.log('✅ Database seeded successfully!');
		console.log(`   - Created ${users.length} users`);
		console.log(`   - Created ${restaurants.length} restaurants`);
	} catch (error) {
		console.error('❌ Error seeding database:', error);
		process.exit(1);
	}
}

// Run the seed function if this script is executed directly
if (import.meta.url === `file://${process.argv[1]}`) {
	seed()
		.then(() => process.exit(0))
		.catch((error) => {
			console.error(error);
			process.exit(1);
		});
}

export { seed };
