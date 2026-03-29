<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
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
		existingReview,
		initialTags,
		onSubmit
	} = $props<{
		googlePlacesId: string;
		restaurantName: string;
		restaurantAddress: string;
		city?: string;
		country?: string;
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
	let showDetails = $state(
		!!(existingReview?.visitedAt || existingReview?.partySize || existingReview?.occasion ||
		   existingReview?.pricePaidPerPerson || existingReview?.wouldVisitAgain || existingReview?.dishHighlights)
	);

	// Extra fields
	let visitedAtStr = $state(
		existingReview?.visitedAt
			? new Date(Number(existingReview.visitedAt) * 1000).toISOString().slice(0, 10)
			: ''
	);
	let partySize = $state<PartySize>(existingReview?.partySize ?? PartySize.UNSPECIFIED);
	let occasion = $state<Occasion>(existingReview?.occasion ?? Occasion.UNSPECIFIED);
	let pricePaidPerPerson = $state(existingReview?.pricePaidPerPerson ?? 0);
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
				// Always send optional extra fields so the server can update/clear them.
				const res = await client.reviews.updateReview({
					id: existingReview.id,
					comment,
					rating,
					tags,
					visitedAt: visitedAtTs(),
					partySize,
					occasion,
					pricePaidPerPerson: pricePaidPerPerson || 0,
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
					comment,
					rating,
					tags,
					visitedAt: visitedAtTs(),
					partySize,
					occasion,
					pricePaidPerPerson,
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
	<h4 class="mb-4 text-base font-semibold text-foreground">
		{isEdit ? 'Edit your rating' : 'Rate this place'}
	</h4>

	<!-- Star picker -->
	<div class="mb-4">
		<Label class="mb-1 block text-sm">Rating *</Label>
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
						class="h-7 w-7 {i < displayRating
							? 'fill-amber-400 text-amber-400'
							: 'fill-none text-gray-300 dark:text-gray-600'}"
					/>
				</button>
			{/each}
		</div>
	</div>

	<!-- Comment -->
	<div class="mb-4">
		<Label for="comment" class="mb-1 block text-sm">Comment (optional)</Label>
		<textarea
			id="comment"
			bind:value={comment}
			rows="3"
			placeholder="What did you think?"
			class="w-full resize-none rounded-lg border border-input bg-background px-3 py-2 text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-ring"
		></textarea>
	</div>

	<!-- Tags -->
	<div class="mb-4">
		<Label class="mb-1 block text-sm">Tags (optional)</Label>
		<TagPicker bind:selected={tags} />
	</div>

	<!-- More details toggle -->
	<button
		type="button"
		class="mb-4 flex items-center gap-1.5 text-sm text-muted-foreground hover:text-foreground"
		onclick={() => (showDetails = !showDetails)}
	>
		<span class="transition-transform {showDetails ? 'rotate-90' : ''}">›</span>
		{showDetails ? 'Hide details' : 'Add more details'}
	</button>

	{#if showDetails}
		<div class="mb-5 space-y-4 rounded-lg border border-border bg-muted/30 p-4">
			<!-- Visit date + Would visit again -->
			<div class="grid grid-cols-2 gap-3">
				<div>
					<Label for="visited-at" class="mb-1 block text-sm">Visit date</Label>
					<input
						id="visited-at"
						type="date"
						bind:value={visitedAtStr}
						class="w-full rounded-md border border-border bg-card px-3 py-1.5 text-sm text-foreground focus:outline-none focus:ring-1 focus:ring-ring"
					/>
				</div>
				<div>
					<Label class="mb-1 block text-sm">Visit again?</Label>
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
					<Label for="party-size" class="mb-1 block text-sm">Party size</Label>
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
					<Label for="occasion" class="mb-1 block text-sm">Occasion</Label>
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

			<!-- Price paid per person -->
			<div>
				<Label for="price-paid" class="mb-1 block text-sm">Price paid per person</Label>
				<div class="relative w-36">
					<span class="pointer-events-none absolute inset-y-0 left-3 flex items-center text-sm text-muted-foreground">$</span>
					<input
						id="price-paid"
						type="number"
						min="0"
						value={pricePaidPerPerson || ''}
						oninput={(e) => {
							const n = (e.currentTarget as HTMLInputElement).valueAsNumber;
							pricePaidPerPerson = Number.isNaN(n) || n < 0 ? 0 : Math.floor(n);
						}}
						placeholder="0"
						class="w-full rounded-md border border-border bg-card py-1.5 pl-7 pr-3 text-sm text-foreground focus:outline-none focus:ring-1 focus:ring-ring"
					/>
				</div>
			</div>

			<!-- Dish highlights -->
			<div>
				<Label for="dish-highlights" class="mb-1 block text-sm">Dish highlights</Label>
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

	{#if error}
		<p class="mb-3 text-sm text-destructive">{error}</p>
	{/if}

	<Button onclick={handleSubmit} disabled={loading || rating < 1} class="w-full">
		{loading ? 'Saving…' : isEdit ? 'Update rating' : 'Save rating'}
	</Button>
</div>
