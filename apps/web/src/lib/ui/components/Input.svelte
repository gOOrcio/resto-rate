<script lang="ts">
	import type { HTMLInputAttributes } from 'svelte/elements';

	interface $$Props extends HTMLInputAttributes {
		variant?: 'filled' | 'outlined' | 'tonal';
		color?: 'primary' | 'secondary' | 'tertiary' | 'success' | 'warning' | 'error' | 'surface';
		inputSize?: 'sm' | 'md' | 'lg';
		bgColor?: string;
		placeholderColor?: string;
		class?: string;
	}

	let {
		variant = 'outlined',
		color = 'surface',
		inputSize = 'md',
		bgColor = '',
		placeholderColor = '',
		class: className = '',
		...props
	}: $$Props = $props();
</script>

<input
	class="input preset-{variant}-{color}-200-800 {inputSize === 'sm'
		? 'input-sm'
		: inputSize === 'lg'
			? 'input-lg'
			: ''} {bgColor} {placeholderColor} {className}"
	{...props}
/>
