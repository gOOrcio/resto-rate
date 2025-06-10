import js from '@eslint/js';
import typescript from 'typescript-eslint';
import prettier from 'eslint-config-prettier';

export default [
	js.configs.recommended,
	...typescript.configs.recommended,
	prettier,
	{
		languageOptions: {
			ecmaVersion: 'latest',
			sourceType: 'module',
		},
		rules: {
			'@typescript-eslint/no-unused-vars': ['error', { argsIgnorePattern: '^_' }],
			'@typescript-eslint/no-explicit-any': 'warn',
			'prefer-const': 'error',
			'no-var': 'error',
		},
	},
];
