import globals from 'globals';
import baseConfig from './index.js';

export default [
	...baseConfig,
	{
		languageOptions: {
			globals: {
				...globals.node,
			},
		},
		rules: {
			'no-console': 'off', // Allow console in Node.js
			'@typescript-eslint/no-require-imports': 'off',
		},
	},
];
