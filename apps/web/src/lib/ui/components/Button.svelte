<script lang="ts">
	import type { HTMLButtonAttributes } from 'svelte/elements';

	interface $$Props extends HTMLButtonAttributes {
		variant?: 'filled' | 'outlined' | 'tonal' | 'ghost';
		color?: 'primary' | 'secondary' | 'tertiary' | 'success' | 'warning' | 'error' | 'surface';
		size?: 'sm' | 'md' | 'lg';
		class?: string;
		children?: any;
	}

	let {
		variant = 'filled',
		color = 'primary',
		size = 'md',
		class: className = '',
		children,
		...props
	}: $$Props = $props();
</script>

<button
	class="btn preset-{variant}-{color}-500 {size === 'sm'
		? 'btn-sm'
		: size === 'lg'
			? 'btn-lg'
			: ''} {className}"
	{...props}
>
	{@render children()}
</button>
