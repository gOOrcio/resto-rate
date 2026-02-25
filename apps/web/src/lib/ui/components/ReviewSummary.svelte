<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import { Star } from '@lucide/svelte';
	import type { ReviewProto } from '$lib/client/generated/reviews/v1/review_pb';

	const { review, onEdit } = $props<{
		review: ReviewProto;
		onEdit: () => void;
	}>();
</script>

<div class="rounded-2xl bg-white p-6 shadow-xl">
	<div class="mb-3 flex items-center justify-between">
		<h4 class="text-base font-semibold text-gray-800">Your rating</h4>
		<Button size="sm" variant="outline" onclick={onEdit}>Edit</Button>
	</div>

	<!-- Stars (read-only) -->
	<div class="mb-3 flex items-center gap-0.5">
		{#each Array(5) as _, i}
			<Star
				class="h-5 w-5 {i < review.rating
					? 'fill-amber-400 text-amber-400'
					: 'fill-none text-gray-300'}"
			/>
		{/each}
		<span class="ml-2 text-sm font-semibold text-gray-800">{review.rating.toFixed(1)}</span>
	</div>

	{#if review.comment}
		<p class="mb-3 text-sm leading-relaxed text-gray-600">{review.comment}</p>
	{/if}

	{#if review.tags && review.tags.length > 0}
		<div class="flex flex-wrap gap-1.5">
			{#each review.tags as tag}
				<span class="rounded-full bg-blue-100 px-2.5 py-0.5 text-xs font-medium text-blue-700">
					{tag}
				</span>
			{/each}
		</div>
	{/if}
</div>
