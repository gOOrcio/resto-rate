<script lang="ts">
	import { onMount } from 'svelte';
	import client from '$lib/client/client';
	import { auth } from '$lib/state/auth.svelte';
	import { AuthProvider } from '$lib/client/generated/auth/v1/auth_service_pb';

	let { onSuccess }: { onSuccess?: () => void } = $props();

	let error = $state<string | null>(null);
	let buttonContainer: HTMLDivElement;

	const clientId = import.meta.env.VITE_GOOGLE_CLIENT_ID as string;

	onMount(() => {
		if (!clientId) {
			console.warn('[SocialSignIn] VITE_GOOGLE_CLIENT_ID is not set');
			return;
		}

		const script = document.createElement('script');
		script.src = 'https://accounts.google.com/gsi/client';
		script.async = true;
		script.defer = true;
		script.onload = initGIS;
		document.head.appendChild(script);

		return () => {
			document.head.removeChild(script);
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
</script>

<div class="flex flex-col items-center gap-3">
	<!-- Sign In With Google button (GIS renders into this div) -->
	<div bind:this={buttonContainer} class="w-full max-w-xs"></div>

	{#if error}
		<p class="text-sm text-red-600">{error}</p>
	{/if}

	<!-- TODO: Apple Sign-In button goes here -->
</div>
