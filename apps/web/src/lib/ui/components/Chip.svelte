<script lang="ts">
	import type { HTMLAttributes } from 'svelte/elements';
	import { createEventDispatcher } from 'svelte';

	interface $$Props extends HTMLAttributes<HTMLDivElement> {
		variant?: 'filled' | 'outlined' | 'tonal' | 'ghost';
		color?: 'primary' | 'secondary' | 'tertiary' | 'success' | 'warning' | 'error' | 'surface';
		size?: 'sm' | 'md' | 'lg';
		removable?: boolean;
		class?: string;
		children?: any;
	}

	let {
		variant = 'filled',
		color = 'primary',
		size = 'md',
		removable = false,
		class: className = '',
		children,
		...props
	}: $$Props = $props();

	const dispatch = createEventDispatcher();

	function handleRemove() {
		dispatch('remove');
	}
</script>

<div
	class="chip preset-{variant}-{color}-500 {size === 'sm'
		? 'chip-sm'
		: size === 'lg'
			? 'chip-lg'
			: ''} {className}"
	{...props}
>
	{@render children()}
	{#if removable}
		<button class="chip-action" onclick={handleRemove} type="button" aria-label="Remove">
			<svg class="chip-action-icon" fill="currentColor" viewBox="0 0 20 20">
				<path
					fill-rule="evenodd"
					d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z"
					clip-rule="evenodd"
				/>
			</svg>
		</button>
	{/if}
</div>
