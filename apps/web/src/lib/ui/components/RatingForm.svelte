<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import { Star } from '@lucide/svelte';
	import client from '$lib/client/client';
	import TagPicker from './TagPicker.svelte';
	import type { ReviewProto } from '$lib/client/generated/reviews/v1/review_pb';
	import { WouldVisitAgain } from '$lib/client/generated/reviews/v1/review_pb';
	import * as m from '$lib/paraglide/messages';

	const {
		googlePlacesId,
		restaurantName,
		restaurantAddress,
		city = '',
		country = '',
		photoReference = '',
		existingReview,
		initialTags,
		onSubmit
	} = $props<{
		googlePlacesId: string;
		restaurantName: string;
		restaurantAddress: string;
		city?: string;
		country?: string;
		photoReference?: string;
		existingReview?: ReviewProto;
		initialTags?: string[];
		onSubmit: (review: ReviewProto) => void;
	}>();

	let rating = $state(existingReview?.rating ?? 0);
	let hoverRating = $state(0);
	let comment = $state(existingReview?.comment ?? '');
	let tags = $state<string[]>(existingReview?.tags ? [...existingReview.tags] : (initialTags ?? []));
	let loading = $state(false);
	let error = $state<string | null>(null);

	// Section visibility — auto-open if editing a review that already has the field set
	let showComment = $state(!!existingReview?.comment);
	let showTags = $state(!!(existingReview?.tags?.length || initialTags?.length));
	let showVisitDetails = $state(
		!!(existingReview?.visitedAt || existingReview?.wouldVisitAgain || existingReview?.dishHighlights)
	);

	// Visit detail fields
	let visitedAtStr = $state(
		existingReview?.visitedAt
			? new Date(Number(existingReview.visitedAt) * 1000).toISOString().slice(0, 10)
			: ''
	);
	let wouldVisitAgain = $state<WouldVisitAgain>(existingReview?.wouldVisitAgain ?? WouldVisitAgain.UNSPECIFIED);
	let dishHighlights = $state(existingReview?.dishHighlights ?? '');

	const isEdit = $derived(!!existingReview?.id);
	const displayRating = $derived(hoverRating || rating);

	function visitedAtTs(): bigint {
		if (!visitedAtStr) return 0n;
		const ms = new Date(visitedAtStr).getTime();
		return isNaN(ms) ? 0n : BigInt(Math.floor(ms / 1000));
	}

	async function handleSubmit() {
		if (rating < 1) {
			error = m.rating_form_select_star();
			return;
		}
		error = null;
		loading = true;
		try {
			if (isEdit && existingReview) {
				const res = await client.reviews.updateReview({
					id: existingReview.id,
					comment,
					rating,
					tags,
					visitedAt: visitedAtTs(),
					wouldVisitAgain,
					dishHighlights
				});
				if (res.review) onSubmit(res.review);
			} else {
				const res = await client.reviews.createReview({
					googlePlacesId,
					restaurantName,
					restaurantAddress,
					city,
					country,
					photoReference,
					comment,
					rating,
					tags,
					visitedAt: visitedAtTs(),
					wouldVisitAgain,
					dishHighlights
				});
				if (res.review) onSubmit(res.review);
			}
		} catch (e: unknown) {
			error = e instanceof Error ? e.message : m.rating_form_save_error();
		} finally {
			loading = false;
		}
	}
</script>

<div class="rounded-2xl bg-card p-6 shadow-xl">
	<!-- Header -->
	<div class="mb-5">
		<p class="text-xs font-medium uppercase tracking-wide text-muted-foreground">
			{isEdit ? m.rating_form_editing() : m.rating_form_rate()}
		</p>
		<h4 class="mt-0.5 text-base font-semibold text-foreground">{restaurantName}</h4>
	</div>

	<!-- Rating — always visible -->
	<div class="mb-5">
		<div class="flex gap-1">
			{#each Array(5) as _, i}
				<button
					type="button"
					onclick={() => (rating = i + 1)}
					onmouseenter={() => (hoverRating = i + 1)}
					onmouseleave={() => (hoverRating = 0)}
					class="transition-transform hover:scale-110"
					aria-label="Rate {i + 1} stars"
				>
					<Star
						class="h-8 w-8 {i < displayRating
							? 'fill-amber-400 text-amber-400'
							: 'fill-none text-gray-300 dark:text-gray-600'}"
					/>
				</button>
			{/each}
		</div>
		{#if error && rating < 1}
			<p class="mt-1 text-xs text-destructive">{error}</p>
		{/if}
	</div>

	<!-- Collapsible sections -->
	<div class="mb-5 space-y-2">
		<!-- Add comment -->
		<div class="rounded-lg border border-border">
			<button
				type="button"
				class="flex w-full items-center justify-between px-4 py-2.5 text-sm font-medium text-foreground"
				onclick={() => (showComment = !showComment)}
			>
				<span>{m.rating_form_add_comment()}</span>
				<span class="text-muted-foreground transition-transform {showComment ? 'rotate-90' : ''}">›</span>
			</button>
			{#if showComment}
				<div class="border-t border-border px-4 pb-4 pt-3">
					<textarea
						id="comment"
						bind:value={comment}
						rows="3"
						placeholder={m.rating_form_comment_placeholder()}
						class="w-full resize-none rounded-lg border border-input bg-background px-3 py-2 text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-ring"
					></textarea>
				</div>
			{/if}
		</div>

		<!-- Add tags -->
		<div class="rounded-lg border border-border">
			<button
				type="button"
				class="flex w-full items-center justify-between px-4 py-2.5 text-sm font-medium text-foreground"
				onclick={() => (showTags = !showTags)}
			>
				<span>{m.rating_form_add_tags({ count: String(tags.length) })}</span>
				<span class="text-muted-foreground transition-transform {showTags ? 'rotate-90' : ''}">›</span>
			</button>
			{#if showTags}
				<div class="border-t border-border px-4 pb-4 pt-3">
					<TagPicker bind:selected={tags} />
				</div>
			{/if}
		</div>

		<!-- Add visit details -->
		<div class="rounded-lg border border-border">
			<button
				type="button"
				class="flex w-full items-center justify-between px-4 py-2.5 text-sm font-medium text-foreground"
				onclick={() => (showVisitDetails = !showVisitDetails)}
			>
				<span>{m.rating_form_add_visit_details()}</span>
				<span class="text-muted-foreground transition-transform {showVisitDetails ? 'rotate-90' : ''}">›</span>
			</button>
			{#if showVisitDetails}
				<div class="space-y-4 border-t border-border px-4 pb-4 pt-3">
					<!-- Visit date + Would visit again -->
					<div class="grid grid-cols-2 gap-3">
						<div>
							<label for="visited-at" class="mb-1 block text-xs font-medium text-muted-foreground">{m.rating_form_visit_date()}</label>
							<input
								id="visited-at"
								type="date"
								bind:value={visitedAtStr}
								class="w-full rounded-md border border-border bg-background px-3 py-1.5 text-sm text-foreground focus:outline-none focus:ring-1 focus:ring-ring"
							/>
						</div>
						<div>
							<p class="mb-1 text-xs font-medium text-muted-foreground">{m.rating_form_visit_again()}</p>
							<div class="flex gap-1">
								{#each [
									{ value: WouldVisitAgain.YES, label: m.rating_form_visit_again_yes() },
									{ value: WouldVisitAgain.MAYBE, label: m.rating_form_visit_again_maybe() },
									{ value: WouldVisitAgain.NO, label: m.rating_form_visit_again_no() }
								] as opt}
									<button
										type="button"
										onclick={() => (wouldVisitAgain = wouldVisitAgain === opt.value ? WouldVisitAgain.UNSPECIFIED : opt.value)}
										class="flex-1 rounded-md border px-2 py-1.5 text-xs font-medium transition-colors {wouldVisitAgain === opt.value
											? 'border-primary bg-primary text-primary-foreground'
											: 'border-border bg-card text-muted-foreground hover:bg-muted'}"
									>
										{opt.label}
									</button>
								{/each}
							</div>
						</div>
					</div>

					<!-- Dish highlights -->
					<div>
						<label for="dish-highlights" class="mb-1 block text-xs font-medium text-muted-foreground">{m.rating_form_dish_highlights()}</label>
						<textarea
							id="dish-highlights"
							bind:value={dishHighlights}
							rows="2"
							placeholder={m.rating_form_dish_placeholder()}
							class="w-full resize-none rounded-lg border border-input bg-background px-3 py-2 text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-ring"
						></textarea>
					</div>
				</div>
			{/if}
		</div>
	</div>

	{#if error && rating >= 1}
		<p class="mb-3 text-sm text-destructive">{error}</p>
	{/if}

	<Button onclick={handleSubmit} disabled={loading || rating < 1} class="w-full">
		{loading ? m.common_saving() : isEdit ? m.rating_form_update() : m.rating_form_save()}
	</Button>
</div>
