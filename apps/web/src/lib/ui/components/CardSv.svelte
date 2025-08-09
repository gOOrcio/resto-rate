<script lang="ts">
	import type { HTMLAttributes } from 'svelte/elements';
	import clsx from 'clsx';
	import { twMerge } from 'tailwind-merge';

	type Variant = 'filled' | 'outlined' | 'tonal';
	type Color = 'primary' | 'secondary' | 'tertiary' | 'success' | 'warning' | 'error' | 'surface';
	type Padding = 'none' | 'sm' | 'md' | 'lg';

	interface $$Props extends HTMLAttributes<HTMLDivElement> {
		variant?: Variant;
		color?: Color;
		padding?: Padding;
		class?: string;
		children?: any;
	}

	let {
		variant = 'outlined',
		color = 'surface',
		padding = 'md',
		class: className = '',
		children,
		...rest
	}: $$Props = $props();

	const paddingClasses = $derived({ none: '', sm: 'p-2', md: 'p-4', lg: 'p-6' }[padding] ?? '');

	const classes = $derived(
		twMerge(clsx('card', `preset-${variant}-${color}-200-800`, paddingClasses, className))
	);
</script>

<div class={classes} {...rest}>
	{@render children()}
</div>
