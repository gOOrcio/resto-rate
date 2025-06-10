import tailwindcss from '@tailwindcss/vite';
import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';
import { resolve } from 'path';
import dotenv from 'dotenv';

// Load environment variables from project root before anything else
const envPath = resolve(process.cwd(), '../..', '.env');
console.log('Loading environment from:', envPath);
dotenv.config({ path: envPath });

export default defineConfig({
	plugins: [tailwindcss(), sveltekit()],
});
