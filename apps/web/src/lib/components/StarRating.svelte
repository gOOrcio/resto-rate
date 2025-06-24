<script lang="ts">
	import { createEventDispatcher } from 'svelte';

	export let rating: number | null = 0;
	export let maxRating = 5;
	export let editable = false;

	const dispatch = createEventDispatcher();

	function setRating(newRating: number) {
		if (!editable) return;
		rating = newRating;
		dispatch('rate', { rating });
	}
</script>

<div class="flex items-center">
	{#each Array(maxRating) as _, i (i)}
		<button
			type="button"
			class="focus:outline-none"
			on:click={() => setRating(i + 1)}
			disabled={!editable}
			aria-label={`Set rating to ${i + 1} star${i === 0 ? '' : 's'}`}
		>
			<svg
				class="w-5 h-5"
				fill={i < (rating || 0) ? 'currentColor' : 'none'}
				stroke="currentColor"
				viewBox="0 0 24 24"
				xmlns="http://www.w3.org/2000/svg"
			>
				<path
					stroke-linecap="round"
					stroke-linejoin="round"
					stroke-width="2"
					d="M11.049 2.927c.3-.921 1.603-.921 1.902 0l1.519 4.674a1 1 0 00.95.69h4.915c.969 0 1.371 1.24.588 1.81l-3.976 2.888a1 1 0 00-.363 1.118l1.52 4.674c.3.921-.755 1.688-1.539 1.118l-3.976-2.888a1 1 0 00-1.176 0l-3.976 2.888c-.783.57-1.838-.197-1.539-1.118l1.52-4.674a1 1 0 00-.363-1.118L2.04 10.1c-.783-.57-.38-1.81.588-1.81h4.914a1 1 0 00.95-.69L11.049 2.927z"
				></path>
			</svg>
		</button>
	{/each}
</div>
