<script lang="ts">
	import type { HTMLButtonAttributes } from 'svelte/elements';
	import clsx from 'clsx';
	import { twMerge } from 'tailwind-merge';

	type Variant = 'filled' | 'outlined' | 'tonal' | 'ghost';
	type Color = 'primary' | 'secondary' | 'tertiary' | 'success' | 'warning' | 'error' | 'surface';
	type Size = 'sm' | 'md' | 'lg';

	interface $$Props extends HTMLButtonAttributes {
		variant?: Variant;
		color?: Color;
		size?: Size;
		class?: string;
		children?: any;
	}

	let {
		variant = 'filled',
		color = 'primary',
		size = 'md',
		type = 'button',
		class: className = '',
		children,
		...rest
	}: $$Props = $props();

	const sizeClass = $derived({ sm: 'btn-sm', md: '', lg: 'btn-lg' }[size] ?? '');

	const classes = $derived(
		twMerge(clsx('btn', `preset-${variant}-${color}-500`, sizeClass, className))
	);
</script>

<button class={classes} {type} {...rest}>
	{@render children()}
</button>
