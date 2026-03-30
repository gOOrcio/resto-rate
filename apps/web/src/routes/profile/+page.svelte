<script lang="ts">
	import { goto } from '$app/navigation';
	import { auth } from '$lib/state/auth.svelte';
	import { mode, setMode } from '$lib/state/theme.svelte';
	import client from '$lib/client/client';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';

	// ── Stats ────────────────────────────────────────────────────────────────
	let stats = $state<{ reviewCount: number; wishlistCount: number; friendCount: number } | null>(null);
	let statsLoading = $state(true);

	// ── Identity ─────────────────────────────────────────────────────────────
	let usernameInput = $state('');
	let usernameError = $state('');
	let usernameSaving = $state(false);
	let usernameSuccess = $state(false);

	// ── Dark mode ─────────────────────────────────────────────────────────────
	let darkModeSaving = $state(false);

	// ── Danger zone ──────────────────────────────────────────────────────────
	let deleteConfirm = $state('');
	let deleteError = $state('');
	let deleteBusy = $state(false);
	let signOutAllBusy = $state(false);

	// ── Formatters ────────────────────────────────────────────────────────────
	function memberSince(): string {
		if (!auth.user) return '';
		return new Date(Number(auth.user.createdAt) * 1000).toLocaleDateString(undefined, {
			year: 'numeric',
			month: 'long'
		});
	}

	// ── Username ──────────────────────────────────────────────────────────────
	async function saveUsername() {
		usernameError = '';
		usernameSuccess = false;
		const val = usernameInput.trim().toLowerCase();
		if (!val) return;
		if (!/^[a-z0-9_]{3,30}$/.test(val)) {
			usernameError = 'Username must be 3–30 chars: lowercase letters, digits, underscores only.';
			return;
		}
		usernameSaving = true;
		try {
			const res = await client.auth.updateMyProfile({ username: val });
			auth.setUser(res.user!);
			usernameSuccess = true;
			usernameInput = '';
		} catch (e: unknown) {
			usernameError = (e as Error).message || 'Failed to save username';
		} finally {
			usernameSaving = false;
		}
	}

	// ── Dark mode toggle ──────────────────────────────────────────────────────
	async function toggleDarkMode() {
		darkModeSaving = true;
		const next = mode.current !== 'dark';
		setMode(next ? 'dark' : 'light');
		try {
			const res = await client.auth.updateMyProfile({
				setIsDarkModeEnabled: true,
				isDarkModeEnabled: next
			});
			auth.setUser(res.user!);
		} catch {
			// revert on failure
			setMode(next ? 'light' : 'dark');
		} finally {
			darkModeSaving = false;
		}
	}

	// ── Sign out all devices ──────────────────────────────────────────────────
	async function signOutAll() {
		signOutAllBusy = true;
		try {
			await client.auth.signOutAllDevices({});
			auth.setUser(null);
			goto('/');
		} catch {
			signOutAllBusy = false;
		}
	}

	// ── Delete account ────────────────────────────────────────────────────────
	async function deleteAccount() {
		if (deleteConfirm !== 'DELETE') {
			deleteError = 'Type DELETE to confirm.';
			return;
		}
		deleteBusy = true;
		deleteError = '';
		try {
			await client.auth.deleteMyAccount({});
			auth.setUser(null);
			goto('/');
		} catch (e: unknown) {
			deleteError = (e as Error).message || 'Failed to delete account.';
			deleteBusy = false;
		}
	}

	let initialized = $state(false);

	async function initPage() {
		usernameInput = auth.user?.username ?? '';
		try {
			const res = await client.auth.getMyStats({});
			stats = {
				reviewCount: res.reviewCount,
				wishlistCount: res.wishlistCount,
				friendCount: res.friendCount
			};
		} catch {
			// silent
		} finally {
			statsLoading = false;
		}
	}

	$effect(() => {
		if (auth.loading || initialized) return;
		if (!auth.isLoggedIn) {
			goto('/?login=1');
			return;
		}
		initialized = true;
		void initPage();
	});
</script>

<div class="mx-auto max-w-2xl space-y-8 px-4 py-8 sm:px-6">
	<h2 class="font-display text-3xl font-semibold text-foreground">My Profile</h2>

	<!-- Identity -->
	<section class="rounded-xl border border-border bg-card p-6 space-y-4">
		<h3 class="text-lg font-medium text-foreground">Identity</h3>

		<div class="grid gap-1">
			<p class="text-sm text-muted-foreground">Email</p>
			<p class="font-medium text-foreground">{auth.user?.email ?? '—'}</p>
		</div>

		<div class="grid gap-1">
			<p class="text-sm text-muted-foreground">Member since</p>
			<p class="font-medium text-foreground">{memberSince()}</p>
		</div>

		<div class="grid gap-2">
			<label class="text-sm text-muted-foreground" for="username-input">
				Username / handle
				{#if auth.user?.username}
					<span class="ml-1 text-primary">@{auth.user.username}</span>
				{:else}
					<span class="ml-1 text-destructive text-xs">Not set</span>
				{/if}
			</label>
			<div class="flex gap-2">
				<Input
					id="username-input"
					type="text"
					placeholder="e.g. jane_eats"
					bind:value={usernameInput}
					disabled={usernameSaving}
					class="flex-1 font-mono"
				/>
				<Button
					size="sm"
					disabled={usernameSaving || !usernameInput.trim()}
					onclick={saveUsername}
				>
					{usernameSaving ? 'Saving…' : 'Save'}
				</Button>
			</div>
			{#if usernameError}
				<p class="text-sm text-destructive">{usernameError}</p>
			{/if}
			{#if usernameSuccess}
				<p class="text-sm text-primary">Username saved!</p>
			{/if}
			<p class="text-xs text-muted-foreground">3–30 characters · lowercase letters, digits, underscores</p>
		</div>
	</section>

	<!-- Stats -->
	<section class="rounded-xl border border-border bg-card p-6">
		<h3 class="mb-4 text-lg font-medium text-foreground">Activity</h3>
		{#if statsLoading}
			<div class="flex items-center gap-2 text-sm text-muted-foreground">
				<div class="h-4 w-4 animate-spin rounded-full border-2 border-border border-t-primary"></div>
				Loading…
			</div>
		{:else if stats}
			<div class="grid grid-cols-3 gap-4 text-center">
				<a href="/reviews" class="group rounded-lg border border-border p-4 hover:border-primary/50 transition-colors">
					<p class="text-2xl font-bold text-foreground group-hover:text-primary transition-colors">{stats.reviewCount}</p>
					<p class="text-xs text-muted-foreground mt-1">Reviews</p>
				</a>
				<a href="/wishlist" class="group rounded-lg border border-border p-4 hover:border-primary/50 transition-colors">
					<p class="text-2xl font-bold text-foreground group-hover:text-primary transition-colors">{stats.wishlistCount}</p>
					<p class="text-xs text-muted-foreground mt-1">Wishlist</p>
				</a>
				<a href="/friends" class="group rounded-lg border border-border p-4 hover:border-primary/50 transition-colors">
					<p class="text-2xl font-bold text-foreground group-hover:text-primary transition-colors">{stats.friendCount}</p>
					<p class="text-xs text-muted-foreground mt-1">Friends</p>
				</a>
			</div>
		{/if}
	</section>

	<!-- Preferences -->
	<section class="rounded-xl border border-border bg-card p-6 space-y-5">
		<h3 class="text-lg font-medium text-foreground">Preferences</h3>

		<!-- Dark mode -->
		<div class="flex items-center justify-between">
			<div>
				<p class="font-medium text-foreground">Dark mode</p>
				<p class="text-sm text-muted-foreground">Synced across all your devices</p>
			</div>
			<button
				class="relative inline-flex h-6 w-11 items-center rounded-full transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring {mode.current === 'dark'
					? 'bg-primary'
					: 'bg-muted'}"
				role="switch"
				aria-label="Dark mode"
				aria-checked={mode.current === 'dark'}
				disabled={darkModeSaving}
				onclick={toggleDarkMode}
			>
				<span
					class="inline-block h-4 w-4 transform rounded-full bg-white shadow transition-transform {mode.current === 'dark'
						? 'translate-x-6'
						: 'translate-x-1'}"
				></span>
			</button>
		</div>

	</section>

	<!-- Danger zone -->
	<section class="rounded-xl border border-destructive/40 bg-card p-6 space-y-5">
		<h3 class="text-lg font-medium text-destructive">Danger zone</h3>

		<!-- Sign out all devices -->
		<div class="flex items-start justify-between gap-4">
			<div>
				<p class="font-medium text-foreground">Sign out all devices</p>
				<p class="text-sm text-muted-foreground">Invalidates all active sessions including this one</p>
			</div>
			<Button
				variant="outline"
				size="sm"
				class="shrink-0 text-destructive hover:border-destructive/50 hover:text-destructive"
				disabled={signOutAllBusy}
				onclick={signOutAll}
			>
				{signOutAllBusy ? 'Signing out…' : 'Sign out all'}
			</Button>
		</div>

		<hr class="border-border" />

		<!-- Delete account -->
		<div class="space-y-3">
			<div>
				<p class="font-medium text-foreground">Delete account</p>
				<p class="text-sm text-muted-foreground">
					Permanently deletes your account, reviews, and wishlist. This cannot be undone.
				</p>
			</div>
			<div class="flex gap-2">
				<Input
					type="text"
					placeholder='Type "DELETE" to confirm'
					bind:value={deleteConfirm}
					disabled={deleteBusy}
					class="flex-1"
				/>
				<Button
					variant="destructive"
					size="sm"
					class="shrink-0"
					disabled={deleteBusy || deleteConfirm !== 'DELETE'}
					onclick={deleteAccount}
				>
					{deleteBusy ? 'Deleting…' : 'Delete'}
				</Button>
			</div>
			{#if deleteError}
				<p class="text-sm text-destructive">{deleteError}</p>
			{/if}
		</div>
	</section>
</div>
