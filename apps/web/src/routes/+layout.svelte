<script lang="ts">
	import '../app.css';
	import Footer from '$lib/ui/navigation/Footer.svelte';
	import Header from '$lib/ui/navigation/Header.svelte';
	import { ThemeProvider } from 'flowbite-svelte';
	import { appTheme } from '$lib/ui/theme';

	let { children } = $props();
</script>

<ThemeProvider theme={appTheme}>
	<div
		class="flex min-h-screen flex-col bg-gradient-to-br from-blue-50 via-indigo-50 to-purple-50 dark:from-gray-900 dark:via-gray-800 dark:to-gray-900"
	>
		<Header />

		<main class="flex-grow">
			{@render children()}
		</main>

		<Footer />
	</div>
</ThemeProvider>
