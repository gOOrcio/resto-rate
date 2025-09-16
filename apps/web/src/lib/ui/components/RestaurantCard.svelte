<script lang="ts">
	import type { Place } from '$lib/client/generated/google_maps/v1/google_maps_service_pb';
	import { Card, Rating, Star, Dropdown, type RatingIconProps } from 'flowbite-svelte';
	import { MapPinAltOutline } from 'flowbite-svelte-icons';
	import { v4 as uuidv4 } from 'uuid';
	import { restaurantCardTheme, ratingTheme } from '$lib/ui/theme/components';

	type SizeVariant = 'desktop' | 'mobile';

	const { place, size = 'desktop' } = $props<{ place: Place; size?: SizeVariant }>();
	const wrapper = (props: RatingIconProps) => (anchor: any, _props: RatingIconProps) =>
		Star(anchor, { ..._props, ...props });

	// Size variants - Controls card dimensions for different screen contexts
	const sizeClasses: Record<SizeVariant, string> = {
		// Desktop: Auto height allows content to expand naturally, fixed width for grid layouts
		// Tip: Adjust w-120 (480px) to w-80 (320px) for tighter grids, or w-96 (384px) for medium spacing
		desktop: 'h-auto w-120',

		// Mobile: Auto height + full width with max constraint to prevent cards from being too wide
		// Tip: Adjust max-w-sm (384px) to max-w-md (448px) for wider mobile cards, or max-w-xs (320px) for narrower
		mobile: 'h-auto w-full max-w-sm'
	};
</script>

<!-- Main container with responsive sizing and z-index isolation -->
<!-- relative: Enables absolute positioning for dropdown -->
<!-- isolate: Creates new stacking context to prevent z-index conflicts -->
<!-- z-0: Base z-index level, dropdown will use higher z-index -->
<div class="relative isolate z-0 {size === 'desktop' ? sizeClasses.desktop : sizeClasses.mobile}">
	<!-- Card with theme styling and overflow control -->
	<!-- h-full: Takes full height of parent container -->
	<!-- overflow-hidden: Prevents content from spilling outside card boundaries -->
	<Card class={`${restaurantCardTheme.base} h-full overflow-hidden`}>
		<div class="relative h-full">
			<!-- Main content area with flexbox layout -->
			<!-- flex h-full flex-col: Vertical flexbox taking full height -->
			<!-- space-y-4: Consistent 16px spacing between child elements -->
			<!-- p-6: 24px padding on all sides for comfortable content spacing -->
			<div class="flex h-full flex-col space-y-4 p-6">
				<!-- Restaurant Name Header Section -->
				<!-- flex-shrink-0: Prevents header from shrinking when content overflows -->
				<!-- items-start: Aligns items to top (important for multi-line titles) -->
				<!-- justify-between: Spreads title and button to opposite ends -->
				<!-- gap-3: 12px horizontal gap between title and button -->
				<!-- space-y-2: 8px vertical spacing if title wraps to multiple lines -->
				<header class="flex flex-shrink-0 items-start justify-between gap-3 space-y-2">
					<!-- Restaurant title with responsive typography -->
					<!-- text-xl: 20px font size for prominence -->
					<!-- font-bold: Heavy weight for hierarchy -->
					<!-- leading-tight: Tight line height for compact display -->
					<!-- text-gray-900 dark:text-white: High contrast text for readability -->
					<h3 class="text-xl font-bold leading-tight text-gray-900 dark:text-white">
						{place.displayName?.text || place.name}
					</h3>

					<!-- Google Maps trigger button with comprehensive styling -->
					<!-- inline-flex: Inline flexbox for proper icon alignment -->
					<!-- items-center justify-center: Centers the Google Maps logo -->
					<!-- rounded: 4px border radius for modern look -->
					<!-- border border-gray-300: Light border for definition -->
					<!-- bg-white: Clean white background -->
					<!-- p-1: 4px padding for comfortable click target -->
					<!-- shadow-sm: Subtle shadow for depth -->
					<!-- transition-all duration-200: Smooth 200ms transitions for all properties -->
					<!-- hover:border-gray-400 hover:shadow-md: Enhanced border and shadow on hover -->
					<!-- focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-1: Accessibility focus ring -->
					<!-- dark: variants: Dark mode equivalents for all light mode styles -->
					<button
						id="place-details-button"
						type="button"
						class="inline-flex items-center justify-center rounded border border-gray-300 bg-white p-1 shadow-sm transition-all duration-200 hover:border-gray-400 hover:shadow-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-1 dark:border-gray-600 dark:bg-gray-800 dark:hover:border-gray-500"
						aria-label="Show Google Maps details"
					>
						<!-- Google Maps logo with constrained sizing -->
						<!-- h-4: 16px height for compact button -->
						<!-- w-auto: Maintains aspect ratio -->
						<!-- min-height/max-height: Ensures logo stays within button bounds -->
						<!-- Tip: Adjust h-4 to h-5 for larger logo, or h-3 for smaller -->
						<img
							src="/GoogleMaps_Logo_Gray.svg"
							alt="Google Maps"
							class="h-4 w-auto"
							style="min-height: 14px; max-height: 16px;"
						/>
					</button>

					<!-- Flowbite Dropdown with comprehensive theming and responsive sizing -->
					<!-- triggeredBy: Links dropdown to button by ID for proper positioning -->
					<!-- placement="bottom-end": Positions dropdown below and to the right of button -->
					<!-- w-80: Base width of 320px for mobile -->
					<!-- max-w-[90vw]: Prevents dropdown from exceeding 90% of viewport width on small screens -->
					<!-- bg-gray-900: Dark background for high contrast against light cards -->
					<!-- text-white: White text for readability on dark background -->
					<!-- sm:w-96: 384px width on small screens and up -->
					<!-- md:w-[28rem]: 448px width on medium screens and up -->
					<!-- Tip: Adjust w-80 to w-72 for narrower dropdowns, or w-96 for wider -->
					<!-- Tip: Change placement to "bottom-start" for left-aligned dropdown -->
					<Dropdown
						triggeredBy="#place-details-button"
						placement="bottom-end"
						class="w-80 max-w-[90vw] bg-gray-900 text-white sm:w-96 md:w-[28rem]"
					>
						<!-- Dropdown Content Container -->
						<!-- space-y-4: 16px vertical spacing between content sections -->
						<!-- p-4: 16px padding inside dropdown for comfortable content spacing -->
						<!-- Tip: Adjust p-4 to p-6 for more spacious content, or p-3 for tighter layout -->
						<div class="space-y-4 p-4">
							<!-- Header Section with Visual Separation -->
							<!-- mb-4: 16px bottom margin for spacing from content -->
							<!-- border-b border-gray-600: Bottom border for visual separation -->
							<!-- pb-3: 12px bottom padding for space between border and content -->
							<!-- Tip: Change border-gray-600 to border-gray-500 for lighter separator -->
							<div class="mb-4 border-b border-gray-600 pb-3">
								<!-- Restaurant name in dropdown header -->
								<!-- text-lg: 18px font size for prominence -->
								<!-- font-semibold: Semi-bold weight for hierarchy -->
								<!-- text-white: White text for contrast against dark background -->
								<h3 class="text-lg font-semibold text-white">
									{place.displayName?.text || place.name}
								</h3>
							</div>

							<!-- Rating Section with Star Display -->
							{#if place.rating}
								<!-- Horizontal layout for rating label and stars -->
								<!-- flex items-center: Horizontal alignment with vertical centering -->
								<!-- space-x-3: 12px horizontal spacing between label and rating component -->
								<div class="flex items-center space-x-3">
									<!-- Rating label with secondary text styling -->
									<!-- text-sm: 14px font size for secondary information -->
									<!-- font-medium: Medium weight for readability -->
									<!-- text-gray-300: Muted gray for secondary text hierarchy -->
									<h4 class="text-sm font-medium text-gray-300">Rating:</h4>

									<!-- Flowbite Rating component with custom styling -->
									<!-- id={uuidv4()}: Unique ID for each rating instance -->
									<!-- total={5}: 5-star rating system -->
									<!-- size={20}: 20px star size for compact display -->
									<!-- rating={place.rating}: Actual rating value -->
									<!-- fillColor: '#ffa200': Orange/yellow fill for active stars -->
									<!-- strokeColor: '#ffffff': White stroke for contrast on dark background -->
									<!-- class="flex items-center space-x-1": Horizontal layout with 4px spacing -->
									<!-- Tip: Adjust size={20} to size={16} for smaller stars, or size={24} for larger -->
									<!-- Tip: Change fillColor to '#fbbf24' for lighter orange, or '#f59e0b' for darker -->
									<Rating
										id={uuidv4()}
										total={5}
										size={20}
										rating={place.rating}
										icon={wrapper({ fillColor: '#ffa200', strokeColor: '#ffffff' })}
										class="flex items-center space-x-1"
									>
										{#snippet text()}
											<!-- Rating text with numerical value -->
											<!-- ml-2: 8px left margin for spacing from stars -->
											<!-- text-sm font-medium: 14px medium weight for readability -->
											<!-- text-white: White text for contrast -->
											<span class="ml-2 text-sm font-medium text-white">
												{place.rating.toFixed(1)}/5
											</span>
										{/snippet}
									</Rating>
								</div>
							{/if}

							<!-- Address Section with Icon and Text -->
							{#if place.formattedAddress}
								<!-- Vertical layout for address label and content -->
								<!-- space-y-2: 8px vertical spacing between label and address content -->
								<div class="space-y-2">
									<!-- Address label with consistent styling -->
									<!-- text-sm font-medium: 14px medium weight for section headers -->
									<!-- text-gray-300: Muted gray for secondary text hierarchy -->
									<h4 class="text-sm font-medium text-gray-300">Address:</h4>

									<!-- Address content with icon and text layout -->
									<!-- flex items-start: Horizontal layout with top alignment for multi-line addresses -->
									<!-- space-x-2: 8px horizontal spacing between icon and text -->
									<div class="flex items-start space-x-2">
										<!-- Map pin icon with positioning and sizing -->
										<!-- mt-0.5: 2px top margin to align with first line of text -->
										<!-- h-4 w-4: 16px icon size for compact display -->
										<!-- shrink-0: Prevents icon from shrinking in flex layout -->
										<!-- text-gray-400: Muted gray for icon color -->
										<!-- Tip: Adjust mt-0.5 to mt-1 for more spacing, or remove for tight alignment -->
										<MapPinAltOutline class="mt-0.5 h-4 w-4 shrink-0 text-gray-400" />

										<!-- Address text with proper line height -->
										<!-- text-sm: 14px font size for readability -->
										<!-- leading-relaxed: Relaxed line height for multi-line addresses -->
										<!-- text-gray-200: Light gray for good contrast on dark background -->
										<!-- Tip: Change leading-relaxed to leading-normal for tighter spacing -->
										<p class="text-sm leading-relaxed text-gray-200">
											{place.formattedAddress}
										</p>
									</div>
								</div>
							{/if}

							<!-- Google Maps Link Section -->
							{#if place.googleMapsUri}
								<!-- Vertical layout for link label and content -->
								<!-- space-y-2: 8px vertical spacing between label and link -->
								<div class="space-y-2">
									<!-- Google Maps label -->
									<!-- text-sm font-medium: 14px medium weight for section headers -->
									<!-- text-gray-300: Muted gray for secondary text hierarchy -->
									<h4 class="text-sm font-medium text-gray-300">Google Maps:</h4>

									<!-- External link with icon and hover effects -->
									<!-- inline-flex items-center: Horizontal layout with vertical centering -->
									<!-- space-x-2: 8px horizontal spacing between icon and text -->
									<!-- text-blue-400: Blue color for link visibility -->
									<!-- transition-colors: Smooth color transitions on hover -->
									<!-- hover:text-blue-300: Lighter blue on hover for interaction feedback -->
									<!-- target="_blank": Opens link in new tab -->
									<!-- rel="noopener noreferrer": Security attributes for external links -->
									<!-- Tip: Change text-blue-400 to text-blue-500 for darker blue, or text-cyan-400 for cyan -->
									<a
										href={place.googleMapsUri}
										target="_blank"
										rel="noopener noreferrer"
										class="inline-flex items-center space-x-2 text-blue-400 transition-colors hover:text-blue-300"
									>
										<!-- Map pin icon for visual consistency -->
										<!-- h-4 w-4: 16px icon size matching other icons -->
										<MapPinAltOutline class="h-4 w-4" />

										<!-- Link text with appropriate sizing -->
										<!-- text-sm: 14px font size for readability -->
										<span class="text-sm">Open in Maps</span>
									</a>
								</div>
							{/if}

							<!-- Price Level Section -->
							{#if place.priceLevel}
								<!-- Vertical layout for price level label and value -->
								<!-- space-y-2: 8px vertical spacing between label and value -->
								<div class="space-y-2">
									<!-- Price level label -->
									<!-- text-sm font-medium: 14px medium weight for section headers -->
									<!-- text-gray-300: Muted gray for secondary text hierarchy -->
									<h4 class="text-sm font-medium text-gray-300">Price Level:</h4>

									<!-- Price level value -->
									<!-- text-sm: 14px font size for readability -->
									<!-- text-gray-200: Light gray for good contrast on dark background -->
									<div class="text-sm text-gray-200">
										{place.priceLevel}
									</div>
								</div>
							{/if}

							<!-- Website Link Section -->
							{#if place.websiteUri}
								<!-- Vertical layout for website label and link -->
								<!-- space-y-2: 8px vertical spacing between label and link -->
								<div class="space-y-2">
									<!-- Website label -->
									<!-- text-sm font-medium: 14px medium weight for section headers -->
									<!-- text-gray-300: Muted gray for secondary text hierarchy -->
									<h4 class="text-sm font-medium text-gray-300">Website:</h4>

									<!-- External website link with hover effects -->
									<!-- text-sm: 14px font size for readability -->
									<!-- text-blue-400: Blue color for link visibility -->
									<!-- transition-colors: Smooth color transitions on hover -->
									<!-- hover:text-blue-300: Lighter blue on hover for interaction feedback -->
									<!-- target="_blank": Opens link in new tab -->
									<!-- rel="noopener noreferrer": Security attributes for external links -->
									<a
										href={place.websiteUri}
										target="_blank"
										rel="noopener noreferrer"
										class="text-sm text-blue-400 transition-colors hover:text-blue-300"
									>
										Visit Website
									</a>
								</div>
							{/if}

							<!-- Total Reviews Section -->
							{#if place.userRatingCount}
								<!-- Vertical layout for reviews label and count -->
								<!-- space-y-2: 8px vertical spacing between label and count -->
								<div class="space-y-2">
									<!-- Reviews label -->
									<!-- text-sm font-medium: 14px medium weight for section headers -->
									<!-- text-gray-300: Muted gray for secondary text hierarchy -->
									<h4 class="text-sm font-medium text-gray-300">Total Reviews:</h4>

									<!-- Review count with number formatting -->
									<!-- text-sm: 14px font size for readability -->
									<!-- text-gray-200: Light gray for good contrast on dark background -->
									<!-- toLocaleString(): Formats number with locale-appropriate separators -->
									<div class="text-sm text-gray-200">
										{place.userRatingCount.toLocaleString()} reviews
									</div>
								</div>
							{/if}

							<!-- Google Attribution Footer Section -->
							<!-- Required by Google Maps Platform policies for data attribution -->
							<!-- mt-6: 24px top margin for separation from content -->
							<!-- border-t border-gray-600: Top border for visual separation -->
							<!-- pt-4: 16px top padding for spacing between border and content -->
							<div class="mt-6 border-t border-gray-600 pt-4">
								<!-- Horizontal layout for logo and attribution text -->
								<!-- flex items-center: Horizontal alignment with vertical centering -->
								<!-- justify-between: Spreads logo and text to opposite ends -->
								<div class="flex items-center justify-between">
									<!-- Google Maps logo and branding -->
									<!-- flex items-center: Horizontal layout for logo and text -->
									<!-- space-x-2: 8px horizontal spacing between logo and text -->
									<div class="flex items-center space-x-2">
										<!-- Google Maps logo with proper sizing -->
										<!-- h-4: 16px height for compact display -->
										<!-- w-auto: Maintains aspect ratio -->
										<!-- min-height/max-height: Ensures logo stays within bounds -->
										<!-- Tip: Adjust h-4 to h-5 for larger logo, or h-3 for smaller -->
										<img
											src="/GoogleMaps_Logo_Gray.svg"
											alt="Google Maps"
											class="h-4 w-auto"
											style="min-height: 16px; max-height: 19px;"
										/>
									</div>

									<!-- Attribution text -->
									<!-- text-xs: 12px font size for footer text -->
									<!-- text-gray-500: Muted gray for subtle attribution -->
									<div class="text-xs text-gray-500">Data provided by Google</div>
								</div>
							</div>
						</div>
					</Dropdown>
				</header>

				<!-- Main Content Area -->
				<!-- flex flex-1 flex-col: Vertical flexbox that takes remaining space -->
				<!-- space-y-4: 16px vertical spacing between content sections -->
				<!-- flex-1: Takes remaining space after header, allowing address to expand -->
				<div class="flex flex-1 flex-col space-y-4">
					<!-- Rating Section with Enhanced Background -->
					{#if place.rating}
						<!-- Rating container with background styling -->
						<!-- inline-flex: Inline flexbox for compact display -->
						<!-- flex-shrink-0: Prevents rating from shrinking -->
						<!-- items-center justify-center: Centers rating content -->
						<!-- rounded-full: Fully rounded background -->
						<!-- font-medium: Medium font weight for rating text -->
						<div
							class="inline-flex flex-shrink-0 items-center justify-center rounded-full font-medium"
						>
							<div class="flex items-center justify-between">
								<div class="flex items-center space-x-3">
									<!-- Main card rating component -->
									<!-- id={uuidv4()}: Unique ID for each rating instance -->
									<!-- total={5}: 5-star rating system -->
									<!-- size={30}: 30px star size for prominent display -->
									<!-- rating={place.rating}: Actual rating value -->
									<!-- fillColor: '#ffa200': Orange/yellow fill for active stars -->
									<!-- strokeColor: '#000000': Black stroke for contrast on light background -->
									<!-- class="rating-primary": Custom rating styling -->
									<!-- Tip: Adjust size={30} to size={24} for smaller stars, or size={36} for larger -->
									<Rating
										id={uuidv4()}
										total={5}
										size={30}
										rating={place.rating}
										icon={wrapper({ fillColor: '#ffa200', strokeColor: '#000000' })}
										class="rating-primary flex items-center space-x-1"
									>
										{#snippet text()}
											<!-- Rating text container -->
											<!-- ml-3: 12px left margin for spacing from stars -->
											<div class="ml-3">
												<!-- Rating numerical value -->
												<!-- inline-flex items-center justify-center: Centers the rating number -->
												<!-- rounded-full: Fully rounded background for rating number -->
												<!-- font-medium: Medium font weight for readability -->
												<p class="inline-flex items-center justify-center rounded-full font-medium">
													{place.rating.toFixed(1)}/5
												</p>
											</div>
										{/snippet}
									</Rating>
								</div>
							</div>
						</div>
					{/if}

					<!-- Address Section -->
					{#if place.formattedAddress}
						<!-- Address container that expands to fill remaining space -->
						<!-- flex-1: Takes remaining vertical space after rating -->
						<!-- space-y-2: 8px vertical spacing between address elements -->
						<div class="flex-1 space-y-2">
							<!-- Address content with icon and text -->
							<!-- flex items-start: Horizontal layout with top alignment for multi-line addresses -->
							<!-- space-x-2: 8px horizontal spacing between icon and text -->
							<div class="flex items-start space-x-2">
								<!-- Map pin icon for address -->
								<!-- h-6 w-6: 24px icon size for prominent display -->
								<!-- shrink-0: Prevents icon from shrinking in flex layout -->
								<!-- text-gray-500 dark:text-gray-400: Muted gray for icon color -->
								<MapPinAltOutline class="h-6 w-6 shrink-0 text-gray-500 dark:text-gray-400" />

								<!-- Address text container -->
								<!-- min-w-0: Allows text to shrink and wrap properly -->
								<!-- flex-1: Takes remaining horizontal space -->
								<div class="min-w-0 flex-1">
									<!-- Address text with accessibility and styling -->
									<!-- text-sm: 14px font size for readability -->
									<!-- font-medium: Medium font weight for prominence -->
									<!-- text-gray-700 dark:text-gray-300: High contrast text for readability -->
									<!-- sr-only: Screen reader only text for accessibility -->
									<p class="text-sm font-medium text-gray-700 dark:text-gray-300">
										<strong class="sr-only">Address:</strong>
										{place.formattedAddress}
									</p>
								</div>
							</div>
						</div>
					{/if}
				</div>
			</div>
		</div>
	</Card>
</div>
