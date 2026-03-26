<script lang="ts">
	import '../app.css';
	import Footer from '$lib/ui/navigation/Footer.svelte';
	import Header from '$lib/ui/navigation/Header.svelte';
	import { onMount } from 'svelte';
	import client from '$lib/client/client';
	import { auth } from '$lib/state/auth.svelte';
	import { AuthProvider } from '$lib/client/generated/auth/v1/auth_service_pb';

	let { children } = $props();

	const clientId = import.meta.env.VITE_GOOGLE_CLIENT_ID as string;

	async function handleCredentialResponse(response: { credential: string }) {
		try {
			const res = await client.auth.login({
				provider: AuthProvider.GOOGLE,
				idToken: response.credential,
			});
			if (res.user) {
				auth.setUser(res.user);
			}
		} catch (e) {
			console.error('[GIS] One Tap sign-in failed', e);
		}
	}

	onMount(async () => {
		let isAuthenticated = false;
		try {
			const res = await client.auth.getCurrentUser({});
			if (res.user) {
				auth.setUser(res.user);
				isAuthenticated = true;
			}
		} catch {
			// Not authenticated — fall through to GIS bootstrapping
		}

		if (isAuthenticated || !clientId) return;

		// Bootstrap GIS: load script, initialize, then show One Tap for unauthenticated users.
		const GIS_SRC = 'https://accounts.google.com/gsi/client';
		const existing = document.querySelector<HTMLScriptElement>(`script[src="${GIS_SRC}"]`);
		if (existing) {
			// Script already injected (e.g. by SocialSignIn) — GIS is either ready or still loading.
			if (window.google?.accounts?.id) {
				window.google.accounts.id.initialize({ client_id: clientId, callback: handleCredentialResponse });
				window.google.accounts.id.prompt();
			} else {
				existing.addEventListener('load', () => {
					if (!window.google?.accounts?.id) return;
					window.google.accounts.id.initialize({ client_id: clientId, callback: handleCredentialResponse });
					window.google.accounts.id.prompt();
				}, { once: true });
			}
			return;
		}

		const script = document.createElement('script');
		script.src = GIS_SRC;
		script.async = true;
		script.defer = true;
		script.onload = () => {
			if (!window.google?.accounts?.id) return;
			window.google.accounts.id.initialize({
				client_id: clientId,
				callback: handleCredentialResponse,
			});
			window.google.accounts.id.prompt();
		};
		document.head.appendChild(script);
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
