<script lang="ts">
	import XIcon from '$lib/ui/icons/XIcon.svelte';
	import Instagram from '@lucide/svelte/icons/instagram';
	import Slack from '@lucide/svelte/icons/slack';
	import { InputSv, ButtonSv } from '$lib/ui/components';

	let iconSize = 26;

	function openUrl(url: string) {
		window.open(url, '_blank');
	}
</script>

<footer class="bg-primary-400 text-primary-contrast-900 mt-10">
	<div class="mx-auto grid max-w-6xl gap-8 px-4 py-8 sm:grid-cols-3">
		<div>
			<h2 class="text-primary-contrast-900 text-2xl font-bold">Resto‑Rate</h2>
			<p class="mt-2 text-sm opacity-80">
				Discover, rate, and review the best restaurants around you.
			</p>
		</div>

		<div>
			<h3 class="text-primary-contrast-900 mb-3 text-lg font-semibold">Quick Links</h3>
			<ul class="space-y-2 text-sm">
				<li>
					<a
						href="/restaurants"
						class="text-primary-contrast-700 hover:text-primary-contrast-100 hover:underline"
					>
						Restaurants
					</a>
				</li>
				<li>
					<a
						href="/about"
						class="text-primary-contrast-700 hover:text-primary-contrast-100 hover:underline"
					>
						About Us
					</a>
				</li>
				<li>
					<a
						href="/contact"
						class="text-primary-contrast-700 hover:text-primary-contrast-100 hover:underline"
					>
						Contact
					</a>
				</li>
				<li>
					<a
						href="/privacy"
						class="text-primary-contrast-700 hover:text-primary-contrast-100 hover:underline"
					>
						Privacy Policy
					</a>
				</li>
			</ul>
		</div>

		<div>
			<h3 class="text-primary-contrast-900 mb-3 text-lg font-semibold">Stay Updated</h3>
			<form class="flex space-x-2">
				<InputSv
					type="email"
					placeholder="Your email"
					variant="outlined"
					color="primary"
					inputSize="md"
					bgColor="bg-surface-50 dark:bg-surface-950"
					class="text-primary-900 dark:text-primary-100 flex-1"
				/>
				<ButtonSv variant="filled" color="primary" size="md">Subscribe</ButtonSv>
			</form>
			<div class="mt-4 flex space-x-4">
				<ButtonSv
					variant="tonal"
					color="primary"
					size="md"
					class="p-2"
					onclick={() => openUrl('https://x.com')}
				>
					<XIcon size={iconSize} />
				</ButtonSv>
				<ButtonSv
					variant="tonal"
					color="primary"
					size="md"
					class="p-2"
					onclick={() => openUrl('https://instagram.com')}
				>
					<Instagram size={iconSize} />
				</ButtonSv>
				<ButtonSv
					variant="tonal"
					color="primary"
					size="md"
					class="p-2"
					onclick={() => openUrl('https://slack.com')}
				>
					<Slack size={iconSize} />
				</ButtonSv>
			</div>
		</div>
	</div>
	<hr class="border-primary-700" />
	<div class="mt-6 py-4 text-center text-sm opacity-80">
		© {new Date().getFullYear()} Resto‑Rate. All rights reserved.
	</div>
</footer>
