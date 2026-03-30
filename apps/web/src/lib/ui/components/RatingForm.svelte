<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import { Star } from '@lucide/svelte';
	import client from '$lib/client/client';
	import TagPicker from './TagPicker.svelte';
	import type { ReviewProto } from '$lib/client/generated/reviews/v1/review_pb';
	import { PartySize, Occasion, WouldVisitAgain } from '$lib/client/generated/reviews/v1/review_pb';

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
		!!(existingReview?.visitedAt || existingReview?.partySize || existingReview?.occasion ||
		   existingReview?.wouldVisitAgain || existingReview?.dishHighlights)
	);

	// Visit detail fields
	let visitedAtStr = $state(
		existingReview?.visitedAt
			? new Date(Number(existingReview.visitedAt) * 1000).toISOString().slice(0, 10)
			: ''
	);
	let partySize = $state<PartySize>(existingReview?.partySize ?? PartySize.UNSPECIFIED);
	let occasion = $state<Occasion>(existingReview?.occasion ?? Occasion.UNSPECIFIED);
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
			error = 'Please select a star rating';
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
					partySize,
					occasion,
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
					partySize,
					occasion,
					wouldVisitAgain,
					dishHighlights
				});
				if (res.review) onSubmit(res.review);
			}
		} catch (e: unknown) {
			error = e instanceof Error ? e.message : 'Failed to save review';
		} finally {
			loading = false;
		}
	}
</script>

<div class="rounded-2xl bg-card p-6 shadow-xl">
	<!-- Header -->
	<div class="mb-5">
		<p class="text-xs font-medium uppercase tracking-wide text-muted-foreground">
			{isEdit ? 'Editing review' : 'Rate this place'}
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
				<span>Add comment</span>
				<span class="text-muted-foreground transition-transform {showComment ? 'rotate-90' : ''}">›</span>
			</button>
			{#if showComment}
				<div class="border-t border-border px-4 pb-4 pt-3">
					<textarea
						id="comment"
						bind:value={comment}
						rows="3"
						placeholder="What did you think?"
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
				<span>Add tags{tags.length > 0 ? ` (${tags.length})` : ''}</span>
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
				<span>Add visit details</span>
				<span class="text-muted-foreground transition-transform {showVisitDetails ? 'rotate-90' : ''}">›</span>
			</button>
			{#if showVisitDetails}
				<div class="space-y-4 border-t border-border px-4 pb-4 pt-3">
					<!-- Visit date + Would visit again -->
					<div class="grid grid-cols-2 gap-3">
						<div>
							<label for="visited-at" class="mb-1 block text-xs font-medium text-muted-foreground">Visit date</label>
							<input
								id="visited-at"
								type="date"
								bind:value={visitedAtStr}
								class="w-full rounded-md border border-border bg-background px-3 py-1.5 text-sm text-foreground focus:outline-none focus:ring-1 focus:ring-ring"
							/>
						</div>
						<div>
							<p class="mb-1 text-xs font-medium text-muted-foreground">Visit again?</p>
							<div class="flex gap-1">
								{#each [
									{ value: WouldVisitAgain.YES, label: 'Yes' },
									{ value: WouldVisitAgain.MAYBE, label: 'Maybe' },
									{ value: WouldVisitAgain.NO, label: 'No' }
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

					<!-- Party size + Occasion -->
					<div class="grid grid-cols-2 gap-3">
						<div>
							<label for="party-size" class="mb-1 block text-xs font-medium text-muted-foreground">Party size</label>
							<select
								id="party-size"
								bind:value={partySize}
								class="w-full rounded-md border border-border bg-card px-3 py-1.5 text-sm text-foreground focus:outline-none focus:ring-1 focus:ring-ring"
							>
								<option value={PartySize.UNSPECIFIED}>—</option>
								<option value={PartySize.SOLO}>Solo</option>
								<option value={PartySize.COUPLE}>Couple</option>
								<option value={PartySize.SMALL_GROUP}>Small group</option>
								<option value={PartySize.LARGE_GROUP}>Large group</option>
							</select>
						</div>
						<div>
							<label for="occasion" class="mb-1 block text-xs font-medium text-muted-foreground">Occasion</label>
							<select
								id="occasion"
								bind:value={occasion}
								class="w-full rounded-md border border-border bg-card px-3 py-1.5 text-sm text-foreground focus:outline-none focus:ring-1 focus:ring-ring"
							>
								<option value={Occasion.UNSPECIFIED}>—</option>
								<option value={Occasion.CASUAL}>Casual</option>
								<option value={Occasion.DATE_NIGHT}>Date night</option>
								<option value={Occasion.BUSINESS}>Business</option>
								<option value={Occasion.CELEBRATION}>Celebration</option>
								<option value={Occasion.QUICK_BITE}>Quick bite</option>
							</select>
						</div>
					</div>

					<!-- Dish highlights -->
					<div>
						<label for="dish-highlights" class="mb-1 block text-xs font-medium text-muted-foreground">Dish highlights</label>
						<textarea
							id="dish-highlights"
							bind:value={dishHighlights}
							rows="2"
							placeholder="Dishes you'd recommend…"
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
		{loading ? 'Saving…' : isEdit ? 'Update rating' : 'Save rating'}
	</Button>
</div>
