<script lang="ts">
	import { auth } from '$lib/state/auth.svelte';
	import SocialSignIn from '$lib/ui/components/SocialSignIn.svelte';
	import RestaurantSearchSv from '$lib/ui/components/RestaurantSearchSv.svelte';
</script>

{#if auth.loading}
	<div class="flex min-h-[60vh] items-center justify-center">
		<div class="h-8 w-8 animate-spin rounded-full border-2 border-border border-t-primary"></div>
	</div>
{:else if auth.isLoggedIn}
	<div class="mx-auto max-w-3xl space-y-6 px-4 py-8 sm:px-6">
		<section class="space-y-4">
			<h2 class="font-display text-2xl font-semibold text-foreground">Find a restaurant</h2>
			<RestaurantSearchSv />
		</section>
	</div>
{:else}
	<div class="mx-auto max-w-4xl space-y-16 px-6 py-16">
		<!-- Hero -->
		<section class="flex flex-col items-center gap-8 text-center">
			<div class="space-y-4">
				<h1 class="font-display text-4xl font-bold tracking-tight text-foreground sm:text-5xl">
					Your personal restaurant diary
				</h1>
				<p class="mx-auto max-w-xl text-lg text-muted-foreground">
					Track every meal, build your wishlist, and discover where your friends love to eat.
				</p>
			</div>

			<div class="w-full max-w-xs">
				<SocialSignIn />
			</div>
		</section>

		<!-- Feature highlights -->
		<section class="grid gap-4 sm:grid-cols-3">
			<div class="rounded-lg border border-border bg-card p-6">
				<h3 class="mb-2 font-semibold text-foreground">Rate &amp; Review</h3>
				<p class="text-sm text-muted-foreground">
					Leave personal ratings and notes for every restaurant you visit. Add tags like Romantic,
					Business lunch, or Great value.
				</p>
			</div>

			<div class="rounded-lg border border-border bg-card p-6">
				<h3 class="mb-2 font-semibold text-foreground">Wishlist</h3>
				<p class="text-sm text-muted-foreground">
					Save restaurants you want to try. They're automatically removed when you leave a review.
				</p>
			</div>

			<div class="rounded-lg border border-border bg-card p-6">
				<h3 class="mb-2 font-semibold text-foreground">Friends</h3>
				<p class="text-sm text-muted-foreground">
					Connect with friends and see their reviews when you search. Discover hidden gems through
					people you trust.
				</p>
			</div>
		</section>

		<p class="text-center text-sm text-muted-foreground">Already have an account? Sign in above ↑</p>
	</div>
{/if}
