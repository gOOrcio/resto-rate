import { paraglideVitePlugin } from '@inlang/paraglide-js';
import devtoolsJson from 'vite-plugin-devtools-json';
import tailwindcss from '@tailwindcss/vite';
import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';
import basicSsl from '@vitejs/plugin-basic-ssl';
import fs from 'node:fs';
import path from 'node:path';

export default defineConfig({
	plugins: [
		tailwindcss(),
		sveltekit(),
		devtoolsJson(),
		...(process.env.NODE_ENV === 'production' ? [basicSsl()] : []),
		paraglideVitePlugin({
			project: './project.inlang',
			outdir: './src/lib/paraglide'
		})
	],
	server: {
		port: Number.parseInt(process.env.VITE_PORT || '5173'),
		strictPort: true,
		...(process.env.NODE_ENV === 'production' ? {
			https: {
				key: fs.readFileSync(path.resolve(__dirname, 'key.pem')),
				cert: fs.readFileSync(path.resolve(__dirname, 'web-cert.pem')),
			}
		} : {}),
		host: true,
		hmr: {
			clientPort: 5173,
			host: process.env.VITE_HMR_HOST || 'localhost'
		}
	},
	preview: {
		// Only use HTTPS in production
		...(process.env.NODE_ENV === 'production' ? {
			https: {
				key: fs.readFileSync(path.resolve(__dirname, 'key.pem')),
				cert: fs.readFileSync(path.resolve(__dirname, 'web-cert.pem')),
			}
		} : {}),
		host: true
	},
	build: {
		outDir: '../../dist/apps/web'
	},
	optimizeDeps: {
		// Force pre-bundling for Safari compatibility
		include: ['svelte']
	},
	esbuild: {
		// Ensure Safari compatibility
		target: 'es2020'
	}
});
