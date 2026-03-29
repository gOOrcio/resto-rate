<script lang="ts">
	import '../app.css';
	import Footer from '$lib/ui/navigation/Footer.svelte';
	import Header from '$lib/ui/navigation/Header.svelte';
	import { onMount } from 'svelte';
	import { fly } from 'svelte/transition';
	import { page } from '$app/state';
	import client from '$lib/client/client';
	import { auth } from '$lib/state/auth.svelte';
	import { setMode } from '$lib/state/theme.svelte';
	import { ModeWatcher } from 'mode-watcher';
	import { AuthProvider } from '$lib/client/generated/auth/v1/auth_service_pb';

	let { children } = $props();

	const clientId = import.meta.env.VITE_GOOGLE_CLIENT_ID as string;
	let reducedMotion = $state(false);

	async function handleCredentialResponse(response: { credential: string }) {
		try {
			const res = await client.auth.login({
				provider: AuthProvider.GOOGLE,
				idToken: response.credential,
			});
			if (res.user) {
				auth.setUser(res.user);
				// Sync dark mode preference from backend on first login
				setMode(res.user.isDarkModeEnabled ? 'dark' : 'light');
			}
		} catch (e) {
			console.error('[GIS] One Tap sign-in failed', e);
		}
	}

	onMount(async () => {
		reducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches;

		let isAuthenticated = false;
		try {
			const res = await client.auth.getCurrentUser({});
			if (res.user) {
				auth.setUser(res.user);
				// Sync dark mode preference from backend
				setMode(res.user.isDarkModeEnabled ? 'dark' : 'light');
				isAuthenticated = true;
			}
		} catch {
			// Not authenticated — fall through to GIS bootstrapping
		} finally {
			auth.setLoaded();
		}

		if (isAuthenticated || !clientId) return;

		const GIS_SRC = 'https://accounts.google.com/gsi/client';
		const existing = document.querySelector<HTMLScriptElement>(`script[src="${GIS_SRC}"]`);
		if (existing) {
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

<ModeWatcher defaultMode="system" />

<svelte:head>
	<link rel="preconnect" href="https://fonts.googleapis.com" />
	<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin="anonymous" />
	<link
		href="https://fonts.googleapis.com/css2?family=DM+Sans:opsz,wght@9..40,300;9..40,400;9..40,500;9..40,600&family=Playfair+Display:wght@500;600;700&display=swap"
		rel="stylesheet"
	/>
</svelte:head>

<div class="flex min-h-screen flex-col bg-background">
	<Header />

	{#key page.url.pathname}
		<main
			class="flex-grow"
			in:fly={reducedMotion ? { duration: 0 } : { y: 8, duration: 200, delay: 140 }}
			out:fly={reducedMotion ? { duration: 0 } : { y: -6, duration: 140 }}
		>
			{@render children()}
		</main>
	{/key}

	<Footer />
</div>
