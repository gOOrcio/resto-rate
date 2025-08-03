<script lang="ts">
	import XIcon from '$lib/ui/icons/XIcon.svelte';
	import Instagram from '@lucide/svelte/icons/instagram';
	import Slack from '@lucide/svelte/icons/slack';

	let iconSize = 26;
</script>

<footer
	class="mt-10"
	style="
		background-color: var(--color-primary-900);
		color: var(--color-primary-contrast-900);
	"
>
	<div class="mx-auto grid max-w-6xl gap-8 px-4 py-8 sm:grid-cols-3">
		<div>
			<h2 class="heading-font-color-dark heading-font-family text-2xl font-bold">Resto‑Rate</h2>
			<p class="mt-2 text-sm opacity-80">
				Discover, rate, and review the best restaurants around you.
			</p>
		</div>

		<div>
			<h3 class="heading-font-color-dark mb-3 text-lg font-semibold">Quick Links</h3>
			<ul class="space-y-2 text-sm">
				<li>
					<a href="/restaurants" class="anchor-font-color-light) hover:underline"> Restaurants </a>
				</li>
				<li>
					<a href="/about" class="anchor-font-color-light hover:underline"> About Us </a>
				</li>
				<li>
					<a href="/contact" class="anchor-font-color-light hover:underline"> Contact </a>
				</li>
				<li>
					<a href="/privacy" class="anchor-font-color-light hover:underline"> Privacy Policy </a>
				</li>
			</ul>
		</div>

		<div>
			<h3 class="heading-font-color-dark mb-3 text-lg font-semibold">Stay Updated</h3>
			<form class="flex space-x-2">
				<input
					type="email"
					placeholder="Your email"
					class="color-surface-50 base-font-color w-full rounded-lg px-3 py-2 focus:outline-none"
				/>
				<button class="btn preset-filled-primary-500"> Subscribe </button>
			</form>
			<div class="mt-4 flex space-x-4">
				<a href="https://x.com" aria-label="Visit our X profile">
					<XIcon size={iconSize} />
				</a>
				<a href="https://instagram.com" aria-label="Visit our Instagram profile">
					<Instagram size={iconSize} />
				</a>
				<a href="https://slack.com" aria-label="Visit our Slack channel">
					<Slack size={iconSize} />
				</a>
			</div>
		</div>
	</div>
	<hr class="hr" />
	<div class="mt-6 py-4 text-center text-sm opacity-80">
		© {new Date().getFullYear()} Resto‑Rate. All rights reserved.
	</div>
</footer>
