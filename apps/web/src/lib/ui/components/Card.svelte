<script lang="ts">
	import type { HTMLAttributes } from 'svelte/elements';

	interface $$Props extends HTMLAttributes<HTMLDivElement> {
		variant?: 'filled' | 'outlined' | 'tonal';
		color?: 'primary' | 'secondary' | 'tertiary' | 'success' | 'warning' | 'error' | 'surface';
		padding?: 'none' | 'sm' | 'md' | 'lg';
		class?: string;
		children?: any;
	}

	let {
		variant = 'outlined',
		color = 'surface',
		padding = 'md',
		class: className = '',
		children,
		...props
	}: $$Props = $props();

	const paddingClasses = $derived(
		{
			none: '',
			sm: 'p-2',
			md: 'p-4',
			lg: 'p-6'
		}[padding]
	);
</script>

<div class="card preset-{variant}-{color}-200-800 {paddingClasses} {className}" {...props}>
	{@render children()}
</div>
