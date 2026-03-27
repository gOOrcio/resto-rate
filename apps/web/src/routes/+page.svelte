<script lang="ts">
	import { auth } from '$lib/state/auth.svelte';
	import SocialSignIn from '$lib/ui/components/SocialSignIn.svelte';
	import RestaurantSearchSv from '$lib/ui/components/RestaurantSearchSv.svelte';
</script>

{#if auth.isLoggedIn}
	<div class="container mx-auto max-w-3xl space-y-6 p-6">
		<section class="space-y-4">
			<h2 class="text-2xl font-semibold text-gray-800">Find a restaurant</h2>
			<RestaurantSearchSv />
		</section>
	</div>
{:else}
	<div class="container mx-auto max-w-4xl space-y-16 px-6 py-16">
		<!-- Hero -->
		<section class="flex flex-col items-center gap-8 text-center">
			<div class="space-y-4">
				<h1 class="text-4xl font-bold tracking-tight text-gray-900 sm:text-5xl">
					Your personal restaurant diary
				</h1>
				<p class="mx-auto max-w-xl text-lg text-gray-600">
					Track every meal, build your wishlist, and discover where your friends love to eat.
				</p>
			</div>

			<div class="w-full max-w-xs">
				<SocialSignIn />
			</div>
		</section>

		<!-- Feature highlights -->
		<section class="grid gap-6 sm:grid-cols-3">
			<div class="rounded-xl border border-gray-100 bg-white p-6 shadow-sm">
				<div class="mb-3 text-3xl">🍽️</div>
				<h3 class="mb-2 font-semibold text-gray-900">Rate &amp; Review</h3>
				<p class="text-sm text-gray-600">
					Leave personal ratings and notes for every restaurant you visit. Add tags like Romantic,
					Business lunch, or Great value.
				</p>
			</div>

			<div class="rounded-xl border border-gray-100 bg-white p-6 shadow-sm">
				<div class="mb-3 text-3xl">❤️</div>
				<h3 class="mb-2 font-semibold text-gray-900">Wishlist</h3>
				<p class="text-sm text-gray-600">
					Save restaurants you want to try. They're automatically removed when you leave a review.
				</p>
			</div>

			<div class="rounded-xl border border-gray-100 bg-white p-6 shadow-sm">
				<div class="mb-3 text-3xl">👥</div>
				<h3 class="mb-2 font-semibold text-gray-900">Friends</h3>
				<p class="text-sm text-gray-600">
					Connect with friends and see their reviews when you search. Discover hidden gems through
					people you trust.
				</p>
			</div>
		</section>

		<!-- Footer CTA -->
		<p class="text-center text-sm text-gray-400">Already have an account? Sign in above ↑</p>
	</div>
{/if}
