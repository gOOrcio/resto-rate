import globals from 'globals';
import sveltePlugin from 'eslint-plugin-svelte';
import baseConfig from './index.js';
import typescript from 'typescript-eslint';

export default [
	...baseConfig,
	...sveltePlugin.configs['flat/recommended'],
	{
		files: ['**/*.svelte'],
		languageOptions: {
			globals: {
				...globals.browser,
			},
			parserOptions: {
				parser: typescript.parser,
			},
		},
		rules: {
			'svelte/no-target-blank': 'error',
			'svelte/no-at-debug-tags': 'warn',
			'svelte/no-reactive-functions': 'error',
			'svelte/no-reactive-literals': 'error',
			'prefer-const': 'off',
		},
	},
];
