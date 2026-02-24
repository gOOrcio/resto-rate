<script lang="ts">
	import { Modal, Button, Input, Label, Helper } from 'flowbite-svelte';
	import client from '$lib/client/client';
	import { auth } from '$lib/state/auth.svelte';

	let { open = $bindable(false) } = $props<{ open: boolean }>();

	let username = $state('');
	let error = $state<string | null>(null);
	let loading = $state(false);

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
		if (e.key === 'Enter') {
			handleLogin();
		}
	}
</script>

<Modal bind:open title="Sign in" size="sm" autoclose={false}>
	<div class="flex flex-col gap-4">
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
				<Helper class="mt-1" color="red">{error}</Helper>
			{/if}
		</div>
		<Button onclick={handleLogin} disabled={loading} class="w-full">
			{loading ? 'Signing in…' : 'Sign in'}
		</Button>
	</div>
</Modal>
