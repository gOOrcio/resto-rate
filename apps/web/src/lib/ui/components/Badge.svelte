<script lang="ts">
	import type { HTMLAttributes } from 'svelte/elements';

	interface $$Props extends HTMLAttributes<HTMLSpanElement> {
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

<span
	class="badge preset-{variant}-{color}-500 {size === 'sm'
		? 'badge-sm'
		: size === 'lg'
			? 'badge-lg'
			: ''} {className}"
	{...props}
>
	{@render children()}
</span>
