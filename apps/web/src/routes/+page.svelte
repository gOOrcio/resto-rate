<script lang="ts">
	import { auth } from '$lib/state/auth.svelte';
</script>

{#if auth.loading}
	<div class="flex min-h-[60vh] items-center justify-center">
		<div class="h-8 w-8 animate-spin rounded-full border-2 border-border border-t-primary"></div>
	</div>
{:else if auth.isLoggedIn}
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

			<div class="flex flex-wrap justify-center gap-3">
				<a
					href="/reviews"
					class="rounded-lg bg-primary px-6 py-3 text-base font-semibold text-primary-foreground transition-colors hover:bg-primary/90"
				>
					My Reviews
				</a>
				<a
					href="/wishlist"
					class="rounded-lg border border-border bg-card px-6 py-3 text-base font-semibold text-foreground transition-colors hover:bg-muted"
				>
					Wishlist
				</a>
				<a
					href="/friends"
					class="rounded-lg border border-border bg-card px-6 py-3 text-base font-semibold text-foreground transition-colors hover:bg-muted"
				>
					Friends
				</a>
				<a
					href="/profile"
					class="rounded-lg border border-border bg-card px-6 py-3 text-base font-semibold text-foreground transition-colors hover:bg-muted"
				>
					My Profile
				</a>
			</div>
		</section>

		<!-- Feature highlights -->
		<section class="grid gap-4 sm:grid-cols-3">
			<a href="/reviews" class="rounded-lg border border-border bg-card p-6 transition-colors hover:bg-muted">
				<h3 class="mb-2 font-semibold text-foreground">Rate &amp; Review</h3>
				<p class="text-sm text-muted-foreground">
					Leave personal ratings and notes for every restaurant you visit. Add tags like Romantic,
					Business lunch, or Great value.
				</p>
			</a>

			<a href="/wishlist" class="rounded-lg border border-border bg-card p-6 transition-colors hover:bg-muted">
				<h3 class="mb-2 font-semibold text-foreground">Wishlist</h3>
				<p class="text-sm text-muted-foreground">
					Save restaurants you want to try. They're automatically removed when you leave a review.
				</p>
			</a>

			<a href="/friends" class="rounded-lg border border-border bg-card p-6 transition-colors hover:bg-muted">
				<h3 class="mb-2 font-semibold text-foreground">Friends</h3>
				<p class="text-sm text-muted-foreground">
					Connect with friends and see their reviews when you search. Discover hidden gems through
					people you trust.
				</p>
			</a>
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

			<div class="flex flex-col items-center gap-3">
				<button
					onclick={() => auth.openLogin()}
					class="rounded-lg bg-primary px-8 py-3 text-base font-semibold text-primary-foreground transition-colors hover:bg-primary/90"
				>
					Get started — it's free
				</button>
				<p class="text-sm text-muted-foreground">
					Already have an account?
					<button
						onclick={() => auth.openLogin()}
						class="underline underline-offset-2 hover:text-foreground"
					>
						Sign in
					</button>
				</p>
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
	</div>
{/if}
