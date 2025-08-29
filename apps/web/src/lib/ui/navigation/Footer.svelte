<script lang="ts">
	import { Hr } from 'flowbite-svelte';
</script>

<footer class="bg-primary-200 dark:bg-primary-900 py-8">
	<div class="container mx-auto max-w-6xl px-6">
		<div class="flex flex-col items-center space-y-4 text-center">
			<!-- Logo and Brand -->
			<div class="flex items-center space-x-3">
				<img src="/resto-rate-logo.svg" class="h-8 w-8" alt="App Logo" />
				<span class="text-primary-800 dark:text-primary-200 text-xl font-bold"> Restorate </span>
			</div>

			<Hr class="bg-primary-400 dark:bg-primary-600 mx-auto h-0.5 w-16 rounded-sm" />

			<!-- Footer Links -->
			<nav class="flex flex-wrap justify-center gap-6 text-sm">
				<a
					href="/"
					class="text-primary-700 hover:text-primary-800 dark:text-primary-300 dark:hover:text-primary-200 transition-colors"
				>
					Home
				</a>
				<a
					href="/about"
					class="text-primary-700 hover:text-primary-800 dark:text-primary-300 dark:hover:text-primary-200 transition-colors"
				>
					About
				</a>
				<a
					href="/pricing"
					class="text-primary-700 hover:text-primary-800 dark:text-primary-300 dark:hover:text-primary-200 transition-colors"
				>
					Pricing
				</a>
				<a
					href="/contact"
					class="text-primary-700 hover:text-primary-800 dark:text-primary-300 dark:hover:text-primary-200 transition-colors"
				>
					Contact
				</a>
			</nav>

			<!-- Copyright -->
			<p class="text-primary-600 dark:text-primary-400 text-sm">
				© 2024 Restorate. All rights reserved.
			</p>
		</div>
	</div>
</footer>
