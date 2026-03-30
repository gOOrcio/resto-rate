<script lang="ts">
	import { auth } from '$lib/state/auth.svelte';
	import * as m from '$lib/paraglide/messages';
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
					{m.home_hero_title()}
				</h1>
				<p class="mx-auto max-w-xl text-lg text-muted-foreground">
					{m.home_hero_subtitle()}
				</p>
			</div>

			<div class="flex flex-wrap justify-center gap-3">
				<a
					href="/reviews"
					class="rounded-lg bg-primary px-6 py-3 text-base font-semibold text-primary-foreground transition-colors hover:bg-primary/90"
				>
					{m.nav_my_reviews()}
				</a>
				<a
					href="/wishlist"
					class="rounded-lg border border-border bg-card px-6 py-3 text-base font-semibold text-foreground transition-colors hover:bg-muted"
				>
					{m.nav_wishlist()}
				</a>
				<a
					href="/friends"
					class="rounded-lg border border-border bg-card px-6 py-3 text-base font-semibold text-foreground transition-colors hover:bg-muted"
				>
					{m.nav_friends()}
				</a>
				<a
					href="/profile"
					class="rounded-lg border border-border bg-card px-6 py-3 text-base font-semibold text-foreground transition-colors hover:bg-muted"
				>
					{m.nav_my_profile()}
				</a>
			</div>
		</section>

		<!-- Feature highlights -->
		<section class="grid gap-4 sm:grid-cols-3">
			<a href="/reviews" class="rounded-lg border border-border bg-card p-6 transition-colors hover:bg-muted">
				<h3 class="mb-2 font-semibold text-foreground">{m.home_feature_review_title()}</h3>
				<p class="text-sm text-muted-foreground">
					{m.home_feature_review_desc()}
				</p>
			</a>

			<a href="/wishlist" class="rounded-lg border border-border bg-card p-6 transition-colors hover:bg-muted">
				<h3 class="mb-2 font-semibold text-foreground">{m.home_feature_wishlist_title()}</h3>
				<p class="text-sm text-muted-foreground">
					{m.home_feature_wishlist_desc()}
				</p>
			</a>

			<a href="/friends" class="rounded-lg border border-border bg-card p-6 transition-colors hover:bg-muted">
				<h3 class="mb-2 font-semibold text-foreground">{m.home_feature_friends_title()}</h3>
				<p class="text-sm text-muted-foreground">
					{m.home_feature_friends_desc()}
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
					{m.home_hero_title()}
				</h1>
				<p class="mx-auto max-w-xl text-lg text-muted-foreground">
					{m.home_hero_subtitle()}
				</p>
			</div>

			<div class="flex flex-col items-center gap-3">
				<button
					onclick={() => auth.openLogin()}
					class="rounded-lg bg-primary px-8 py-3 text-base font-semibold text-primary-foreground transition-colors hover:bg-primary/90"
				>
					{m.home_get_started()}
				</button>
				<p class="text-sm text-muted-foreground">
					{m.home_already_account()}
					<button
						onclick={() => auth.openLogin()}
						class="underline underline-offset-2 hover:text-foreground"
					>
						{m.nav_sign_in()}
					</button>
				</p>
			</div>
		</section>

		<!-- Feature highlights -->
		<section class="grid gap-4 sm:grid-cols-3">
			<div class="rounded-lg border border-border bg-card p-6">
				<h3 class="mb-2 font-semibold text-foreground">{m.home_feature_review_title()}</h3>
				<p class="text-sm text-muted-foreground">
					{m.home_feature_review_desc()}
				</p>
			</div>

			<div class="rounded-lg border border-border bg-card p-6">
				<h3 class="mb-2 font-semibold text-foreground">{m.home_feature_wishlist_title()}</h3>
				<p class="text-sm text-muted-foreground">
					{m.home_feature_wishlist_desc()}
				</p>
			</div>

			<div class="rounded-lg border border-border bg-card p-6">
				<h3 class="mb-2 font-semibold text-foreground">{m.home_feature_friends_title()}</h3>
				<p class="text-sm text-muted-foreground">
					{m.home_feature_friends_desc()}
				</p>
			</div>
		</section>
	</div>
{/if}
