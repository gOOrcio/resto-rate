import svelteConfig from '@resto-rate/eslint-config/svelte';

export default [
	...svelteConfig,
	{
		ignores: ['build/', '.svelte-kit/', 'dist/'],
	},
];
