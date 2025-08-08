<script lang="ts">
	import type { HTMLAttributes } from 'svelte/elements';
	import clsx from 'clsx';
	import { twMerge } from 'tailwind-merge';

	type Variant = 'filled' | 'outlined' | 'tonal' | 'ghost';
	type Color = 'primary' | 'secondary' | 'tertiary' | 'success' | 'warning' | 'error' | 'surface';
	type Size = 'sm' | 'md' | 'lg';

	interface $$Props extends Omit<HTMLAttributes<HTMLSpanElement>, 'class' | 'color' | 'size'> {
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
		class: className = '',
		children,
		...rest
	}: $$Props = $props();

	const sizeClass = $derived(size === 'sm' ? 'chip-sm' : size === 'lg' ? 'chip-lg' : '');
	const preset = $derived(`preset-${variant}-${color}-500`);
	const classes = $derived(twMerge(clsx('chip', preset, sizeClass, className)));
</script>

<span class={classes} {...rest}>
	{@render children?.()}
</span>
