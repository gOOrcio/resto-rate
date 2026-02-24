<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { X } from '@lucide/svelte';
	import client from '$lib/client/client';
	import { auth } from '$lib/state/auth.svelte';

	let { open = $bindable(false) } = $props<{ open: boolean }>();

	let username = $state('');
	let error = $state<string | null>(null);
	let loading = $state(false);

	function initDialog(el: HTMLDialogElement) {
		el.showModal();
		return {
			destroy() {
				if (el.open) el.close();
			}
		};
	}

	function handleBackdropClick(e: MouseEvent & { currentTarget: HTMLDialogElement }) {
		const rect = e.currentTarget.getBoundingClientRect();
		const inside =
			e.clientX >= rect.left &&
			e.clientX <= rect.right &&
			e.clientY >= rect.top &&
			e.clientY <= rect.bottom;
		if (!inside) open = false;
	}

	async function handleLogin() {
		if (!username.trim()) {
			error = 'Username is required';
			return;
		}
		error = null;
		loading = true;
		try {
			const res = await client.auth.login({ username: username.trim() });
			if (res.user) {
				auth.setUser(res.user);
			}
			username = '';
			open = false;
		} catch (e: unknown) {
			error = e instanceof Error ? e.message : 'Login failed';
		} finally {
			loading = false;
		}
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Enter') handleLogin();
	}
</script>

{#if open}
	<dialog
		use:initDialog
		oncancel={() => (open = false)}
		onclick={handleBackdropClick}
		class="m-auto w-full max-w-[calc(100%-2rem)] rounded-lg bg-white p-0 shadow-xl backdrop:bg-gray-900/50 sm:max-w-sm"
	>
		<div class="flex flex-col gap-4 p-6">
			<div class="flex items-center justify-between">
				<h3 class="text-xl font-semibold text-gray-900">Sign in</h3>
				<button
					type="button"
					onclick={() => (open = false)}
					class="rounded-lg p-1.5 text-gray-400 hover:bg-gray-100 hover:text-gray-900"
					aria-label="Close"
				>
					<X class="h-5 w-5" />
				</button>
			</div>
			<div>
				<Label for="username" class="mb-2">Username</Label>
				<Input
					id="username"
					bind:value={username}
					placeholder="Enter your username"
					onkeydown={handleKeydown}
					disabled={loading}
				/>
				{#if error}
					<p class="mt-1 text-sm text-red-600">{error}</p>
				{/if}
			</div>
			<Button onclick={handleLogin} disabled={loading} class="w-full">
				{loading ? 'Signing in…' : 'Sign in'}
			</Button>
		</div>
	</dialog>
{/if}
