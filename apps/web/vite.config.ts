import tailwindcss from '@tailwindcss/vite';
import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';
import { resolve } from 'path';
import dotenv from 'dotenv';

const envPath = resolve(process.cwd(), '../..', '.env');
console.log('Loading environment from:', envPath);
dotenv.config({ path: envPath });

export default defineConfig({
	plugins: [
		// @ts-expect-error. Tailwind is not typed.
		tailwindcss(),
		sveltekit(), 
	],
	optimizeDeps: {
		include: ['@msgpack/msgpack']
	}
});
