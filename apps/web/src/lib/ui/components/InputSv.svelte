<script lang="ts">
	import type { HTMLInputAttributes } from 'svelte/elements';
	import clsx from 'clsx';
	import { twMerge } from 'tailwind-merge';

	type Variant = 'filled' | 'outlined' | 'tonal';
	type Color = 'primary' | 'secondary' | 'tertiary' | 'success' | 'warning' | 'error' | 'surface';
	type InputSize = 'sm' | 'md' | 'lg';

	interface $$Props extends Omit<HTMLInputAttributes, 'size' | 'color'> {
		variant?: Variant;
		color?: Color;
		inputSize?: InputSize;
		bgColor?: string;
		placeholderColor?: string;
		class?: string;
	}

	let {
		value = $bindable(''),
		variant = 'outlined',
		color = 'surface',
		inputSize = 'md',
		bgColor = '',
		placeholderColor = '',
		class: className = '',
		...rest
	}: $$Props = $props();
</script>

<input
	class={twMerge(
		clsx(
			'input',
			`preset-${variant}-${color}-200-800`,
			inputSize === 'sm' && 'input-sm',
			inputSize === 'lg' && 'input-lg',
			bgColor,
			placeholderColor,
			className
		)
	)}
	bind:value
	{...rest}
/>
