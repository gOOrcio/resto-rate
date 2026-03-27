<script lang="ts">
	import { onMount } from 'svelte';
	import client from '$lib/client/client';
	import { auth } from '$lib/state/auth.svelte';
	import { AuthProvider } from '$lib/client/generated/auth/v1/auth_service_pb';

	let { onSuccess }: { onSuccess?: () => void } = $props();

	let error = $state<string | null>(null);
	let buttonContainer: HTMLDivElement;

	const clientId = import.meta.env.VITE_GOOGLE_CLIENT_ID as string;

	const GIS_SRC = 'https://accounts.google.com/gsi/client';

	onMount(() => {
		if (!clientId) {
			console.warn('[SocialSignIn] VITE_GOOGLE_CLIENT_ID is not set');
			return;
		}

		// GIS already available (bootstrapped by the layout) — initialize and render directly.
		if (window.google?.accounts?.id) {
			initGIS();
			return;
		}

		// Script may already be in-flight (injected by the layout) — reuse it.
		const existing = document.querySelector<HTMLScriptElement>(`script[src="${GIS_SRC}"]`);
		if (existing) {
			existing.addEventListener('load', initGIS);
			return () => existing.removeEventListener('load', initGIS);
		}

		// No script yet — load it.
		const script = document.createElement('script');
		script.src = GIS_SRC;
		script.async = true;
		script.defer = true;
		script.addEventListener('load', initGIS);
		document.head.appendChild(script);

		return () => {
			script.removeEventListener('load', initGIS);
		};
	});

	function initGIS() {
		if (!window.google?.accounts?.id) return;

		window.google.accounts.id.initialize({
			client_id: clientId,
			callback: handleCredentialResponse,
		});

		if (buttonContainer) {
			window.google.accounts.id.renderButton(buttonContainer, {
				type: 'standard',
				theme: 'outline',
				size: 'large',
				width: Math.min(buttonContainer.offsetWidth || 300, 400),
			});
		}
	}

	async function handleCredentialResponse(response: { credential: string }) {
		error = null;
		try {
			const res = await client.auth.login({
				provider: AuthProvider.GOOGLE,
				idToken: response.credential,
			});
			if (res.user) {
				auth.setUser(res.user);
				onSuccess?.();
			}
		} catch (e: unknown) {
			error = e instanceof Error ? e.message : 'Sign in failed, please try again';
		}
	}

	async function devLogin() {
		error = null;
		try {
			const apiUrl = import.meta.env.VITE_API_URL ?? 'http://localhost:3001';
			const resp = await fetch(`${apiUrl}/dev/login`, {
				method: 'POST',
				credentials: 'include',
			});
			if (!resp.ok) throw new Error(`Dev login failed: ${resp.status}`);
			const res = await client.auth.getCurrentUser({});
			if (res.user) {
				auth.setUser(res.user);
				onSuccess?.();
			}
		} catch (e: unknown) {
			error = e instanceof Error ? e.message : 'Dev login failed';
		}
	}
</script>

<div class="flex flex-col items-center gap-3">
	<!-- Sign In With Google button (GIS renders into this div) -->
	<div bind:this={buttonContainer} class="w-full max-w-xs"></div>

	{#if import.meta.env.DEV}
		<button
			class="w-full max-w-xs rounded-md border border-dashed border-border px-4 py-2 text-sm text-muted-foreground transition-colors hover:border-primary hover:text-primary"
			onclick={devLogin}
		>
			Dev: sign in as test user
		</button>
	{/if}

	{#if error}
		<p class="text-sm text-red-600">{error}</p>
	{/if}

	<!-- TODO: Apple Sign-In button goes here -->
</div>
