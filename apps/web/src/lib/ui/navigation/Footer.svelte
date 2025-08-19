<script lang="ts">
	import XIcon from '$lib/ui/icons/XIcon.svelte';
	import Instagram from '@lucide/svelte/icons/instagram';
	import Slack from '@lucide/svelte/icons/slack';
	import { Button } from 'flowbite-svelte';

	let iconSize = 26;
</script>

<footer class="bg-primary-400 text-primary-contrast-900 mt-10">
	<div class="mx-auto grid max-w-6xl gap-4 px-4 py-4 sm:grid-cols-3">
		<div>
			<h2 class="text-primary-contrast-900 text-2xl font-bold">Resto‑Rate</h2>
			<p class="mt-2 text-sm opacity-80">Discover, rate, and review the best restaurants.</p>
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

		<div class="mt-4 flex gap-2">
			<Button
				size="sm"
				color="rose"
				pill
				class="group p-2"
				href="https://x.com"
				target="_blank"
				rel="noopener noreferrer"
			>
				<span class="text-[#1DA1F2]">
					<!-- brand blue -->
					<XIcon size={iconSize} color="#000000" />
				</span>
			</Button>

			<Button
				size="md"
				class="group p-2"
				href="https://instagram.com"
				target="_blank"
				rel="noopener noreferrer"
			>
				<Instagram size={iconSize} class="block text-[#000000]" />
			</Button>

			<Button
				size="md"
				class="group p-2"
				href="https://slack.com"
				target="_blank"
				rel="noopener noreferrer"
			>
				<Slack size={iconSize} class="block text-[#000000]" />
			</Button>
		</div>
	</div>
	<hr class="border-primary-700" />
	<div class="mt-6 py-4 text-center text-sm opacity-80">
		© {new Date().getFullYear()} Resto‑Rate. All rights reserved.
	</div>
</footer>
