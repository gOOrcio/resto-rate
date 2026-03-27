<script lang="ts">
	import clients from '$lib/client/client';
	import type {
		Place,
		Suggestion
	} from '$lib/client/generated/google_maps/v1/google_maps_service_pb';
	import type { ReviewProto } from '$lib/client/generated/reviews/v1/review_pb';
	import { onMount, onDestroy } from 'svelte';
	import { Input } from '$lib/components/ui/input/index.js';
	import { v4 as uuidv4 } from 'uuid';
	import { auth } from '$lib/state/auth.svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import PlacePreviewCard from './PlacePreviewCard.svelte';
	import RatingForm from './RatingForm.svelte';
	import ReviewSummary from './ReviewSummary.svelte';

	function randomUUID(): string {
		return uuidv4();
	}

	let autocompleteSessionToken = randomUUID();
	let input = $state('');
	let suggestions = $state<Suggestion[]>([]);
	let isLoading = $state(false);
	let selectedIndex = $state(-1);
	let showSuggestions = $state(false);
	let queryPrediction = $state('');

	let selectedPlace = $state<Place | null>(null);
	let isCheckingReview = $state(false);
	let currentReview = $state<ReviewProto | null>(null);
	let isEditingReview = $state(false);
	let isWishlisted = $state(false);
	let wishlistLoading = $state(false);

	let debounceTimer: ReturnType<typeof setTimeout> | null = null;

	let { onPlaceSelected } = $props<{
		onPlaceSelected?: (place: Place) => void;
	}>();

	function debouncedAutocomplete(input: string) {
		if (debounceTimer) clearTimeout(debounceTimer);

		if (input.length < 2) {
			suggestions = [];
			showSuggestions = false;
			queryPrediction = '';
			autocompleteSessionToken = randomUUID();
			return;
		}

		if (input.length >= 2 && !isLoading) isLoading = true;

		debounceTimer = setTimeout(() => {
			if (input.length >= 2) performAutocomplete(input);
		}, 300);
	}

	async function performAutocomplete(input: string, regionCode: string = 'pl') {
		if (input.length < 2) return;
		isLoading = true;
		try {
			const response = await clients.googleMaps.autocompletePlaces({
				input,
				languageCode: 'pl',
				includedRegionCodes: [regionCode],
				sessionToken: autocompleteSessionToken,
				includeQueryPrediction: true
			});

			suggestions = response.suggestions || [];
			showSuggestions = suggestions.length > 0;

			const querySuggestion = suggestions.find((s) => s.queryPrediction);
			queryPrediction = querySuggestion?.queryPrediction?.text?.text ?? '';
			selectedIndex = -1;
		} catch (error) {
			console.error('Autocomplete error:', error);
			suggestions = [];
			showSuggestions = false;
		} finally {
			isLoading = false;
		}
	}

	async function getPlaceDetails(name: string) {
		try {
			const place = await clients.googleMaps.getRestaurantDetails({
				name,
				languageCode: 'pl',
				regionCode: 'pl',
				sessionToken: autocompleteSessionToken
			});

			selectedPlace = place;
			currentReview = null;
			isEditingReview = false;
			isWishlisted = false;
			suggestions = [];
			showSuggestions = false;
			queryPrediction = '';
			input = place.displayName?.text || place.name || '';
			autocompleteSessionToken = randomUUID();

			if (onPlaceSelected) onPlaceSelected(place);

			// Check if current user already has a review for this place
			isCheckingReview = true;
			try {
				const res = await clients.reviews.listReviews({ googlePlacesId: place.name || '' });
				currentReview = res.reviews?.[0] ?? null;
			} catch {
				currentReview = null;
			} finally {
				isCheckingReview = false;
			}

			// Check if current user has this place wishlisted (only when no review exists)
			if (!currentReview) {
				try {
					const wRes = await clients.wishlist.listWishlist({ googlePlacesId: place.name || '' });
					isWishlisted = (wRes.items?.length ?? 0) > 0;
				} catch {
					isWishlisted = false;
				}
			}
		} catch (error) {
			console.error('Get place details error:', error);
		}
	}

	function handleInputChange(event: Event) {
		const target = event.target as HTMLInputElement;
		input = target.value;

		if (!input.trim()) {
			suggestions = [];
			showSuggestions = false;
			queryPrediction = '';
			isLoading = false;
			return;
		}

		debouncedAutocomplete(input);
	}

	function handleKeyDown(event: KeyboardEvent) {
		if (!showSuggestions) return;
		switch (event.key) {
			case 'ArrowDown':
				event.preventDefault();
				selectedIndex = Math.min(selectedIndex + 1, suggestions.length - 1);
				break;
			case 'ArrowUp':
				event.preventDefault();
				selectedIndex = Math.max(selectedIndex - 1, -1);
				break;
			case 'Enter':
				event.preventDefault();
				if (selectedIndex >= 0 && selectedIndex < suggestions.length) {
					selectSuggestion(suggestions[selectedIndex]);
				}
				break;
			case 'Escape':
				showSuggestions = false;
				selectedIndex = -1;
				break;
		}
	}

	async function toggleWishlist() {
		if (!selectedPlace) return;
		wishlistLoading = true;
		try {
			if (isWishlisted) {
				await clients.wishlist.removeFromWishlist({
					googlePlacesId: selectedPlace.name || ''
				});
				isWishlisted = false;
			} else {
				await clients.wishlist.addToWishlist({
					googlePlacesId: selectedPlace.name || '',
					restaurantName: selectedPlace.displayName?.text || selectedPlace.name || '',
					restaurantAddress: selectedPlace.formattedAddress || '',
					city: selectedPlace.postalAddress?.locality ?? '',
					country: selectedPlace.postalAddress?.country ?? ''
				});
				isWishlisted = true;
			}
		} catch (e) {
			console.error('Wishlist toggle error:', e);
		} finally {
			wishlistLoading = false;
		}
	}

	function selectSuggestion(suggestion: Suggestion) {
		if (suggestion.placePrediction?.place) {
			getPlaceDetails(suggestion.placePrediction.place);
		}
	}

	function getSuggestionText(suggestion: Suggestion): string {
		return (
			suggestion.placePrediction?.structuredFormat?.mainText?.text ||
			suggestion.placePrediction?.text?.text ||
			''
		);
	}

	function getSuggestionSubtext(suggestion: Suggestion): string {
		return suggestion.placePrediction?.structuredFormat?.secondaryText?.text || '';
	}

	onMount(() => {
		autocompleteSessionToken = randomUUID();
	});

	onDestroy(() => {
		if (debounceTimer) {
			clearTimeout(debounceTimer);
			debounceTimer = null;
		}
	});
</script>

{#if !auth.isLoggedIn}
	<p class="text-sm text-gray-500">Please log in to search restaurants.</p>
{:else}
	<div class="relative w-full max-w-md">
		<div class="relative flex items-center">
			<Input
				type="text"
				bind:value={input}
				oninput={handleInputChange}
				onkeydown={handleKeyDown}
				placeholder="Search for restaurants..."
				class="w-full bg-[url('/GoogleMaps_Logo_Gray.svg')] bg-[length:60px_60px] bg-[position:calc(100%-2.25rem)_50%] bg-no-repeat pr-10"
			/>
			{#if isLoading}
				<div class="absolute right-3 flex items-center">
					<div class="border-t-primary-500 h-4 w-4 animate-spin rounded-full border-2 border-gray-300"></div>
				</div>
			{/if}
		</div>

		{#if showSuggestions && suggestions.length > 0}
			<div
				class="absolute left-0 right-0 top-full z-50 max-h-80 overflow-y-auto rounded-b-lg border-2 border-t-0 border-gray-200 bg-white shadow-lg"
			>
				{#each suggestions as suggestion, index}
					{#if suggestion.placePrediction}
						<div
							class="cursor-pointer border-b border-gray-100 p-3 transition-colors duration-200 last:border-b-0 hover:bg-gray-50 {index === selectedIndex ? 'bg-gray-50' : ''}"
							onclick={() => selectSuggestion(suggestion)}
							onkeydown={(e) => e.key === 'Enter' && selectSuggestion(suggestion)}
							onmouseenter={() => (selectedIndex = index)}
							tabindex="0"
							role="button"
							aria-label="Select {getSuggestionText(suggestion)}"
						>
							<div class="mb-1 font-medium text-gray-900">{getSuggestionText(suggestion)}</div>
							{#if getSuggestionSubtext(suggestion)}
								<div class="text-sm text-gray-500">{getSuggestionSubtext(suggestion)}</div>
							{/if}
						</div>
					{/if}
				{/each}
			</div>
		{/if}

		{#if queryPrediction && input.length > 0}
			<div
				class="pointer-events-none absolute left-4 top-1/2 z-10 -translate-y-1/2 transform text-base text-gray-500 {showSuggestions ? 'hidden' : ''}"
			>
				<span class="text-transparent">{input}</span>
				<span class="text-gray-500 opacity-60">{queryPrediction.substring(input.length)}</span>
			</div>
		{/if}

		{#if input.length > 0 && input.length < 2}
			<div class="absolute left-0 right-0 top-full mt-1 rounded border border-gray-200 bg-gray-50 p-2 text-sm text-gray-500">
				Type at least 2 characters to search...
			</div>
		{/if}
	</div>

	{#if selectedPlace}
		<div class="mt-6 space-y-4">
			<PlacePreviewCard place={selectedPlace} />

			{#if !isCheckingReview && !currentReview}
				<div>
					<Button
						variant={isWishlisted ? 'outline' : 'secondary'}
						onclick={toggleWishlist}
						disabled={wishlistLoading}
						class="gap-2"
					>
						{#if wishlistLoading}
							<div class="h-4 w-4 animate-spin rounded-full border-2 border-current border-t-transparent"></div>
						{:else if isWishlisted}
							★ Wishlisted — click to remove
						{:else}
							☆ Save to wishlist
						{/if}
					</Button>
				</div>
			{/if}

			{#if isCheckingReview}
				<div class="flex items-center gap-2 text-sm text-gray-500">
					<div class="h-4 w-4 animate-spin rounded-full border-2 border-gray-300 border-t-blue-500"></div>
					Checking your review…
				</div>
			{:else if currentReview && !isEditingReview}
				<ReviewSummary
					review={currentReview}
					onEdit={() => (isEditingReview = true)}
				/>
			{:else}
				<RatingForm
					googlePlacesId={selectedPlace.name || ''}
					restaurantName={selectedPlace.displayName?.text || selectedPlace.name || ''}
					restaurantAddress={selectedPlace.formattedAddress || ''}
					existingReview={isEditingReview ? currentReview ?? undefined : undefined}
					onSubmit={(review) => {
						currentReview = review;
						isEditingReview = false;
					}}
				/>
			{/if}
		</div>
	{/if}
{/if}
