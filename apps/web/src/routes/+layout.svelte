<script lang="ts">
	import '../app.css';
	import Footer from '$lib/ui/navigation/Footer.svelte';
	import Header from '$lib/ui/navigation/Header.svelte';
	import { onMount } from 'svelte';
	import client from '$lib/client/client';
	import { auth } from '$lib/state/auth.svelte';

	let { children } = $props();

	onMount(async () => {
		try {
			const res = await client.auth.getCurrentUser({});
			if (res.user) {
				auth.setUser(res.user);
				// Already logged in — do not show One Tap
				return;
			}
		} catch {
			// Not authenticated — fall through to One Tap
		}

		// Trigger One Tap for unauthenticated users.
		// GIS silently skips this in WebViews and unsupported browsers.
		window.google?.accounts?.id?.prompt();
	});
</script>

<div
	class="flex min-h-screen flex-col bg-gradient-to-br from-blue-50 via-indigo-50 to-purple-50"
>
	<Header />

	<main class="flex-grow">
		{@render children()}
	</main>

	<Footer />
</div>
