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
		tailwindcss(),
		sveltekit(), 
	],
	build: {
		rollupOptions: {
			onwarn(warning, warn) {
				// Suppress some warnings that might cause issues
				if (warning.code === 'CIRCULAR_DEPENDENCY') return;
				warn(warning);
			}
		}
	},
	optimizeDeps: {
		include: ['@msgpack/msgpack']
	}
});
