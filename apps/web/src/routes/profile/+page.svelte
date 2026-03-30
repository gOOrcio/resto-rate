<script lang="ts">
	import { goto } from '$app/navigation';
	import { auth } from '$lib/state/auth.svelte';
	import { locale } from '$lib/state/locale.svelte';
	import { mode, setMode } from '$lib/state/theme.svelte';
	import client from '$lib/client/client';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import * as m from '$lib/paraglide/messages';

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

	// ── Locale switcher ───────────────────────────────────────────────────────
	let localeSaving = $state(false);

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
			usernameError = m.profile_username_error();
			return;
		}
		usernameSaving = true;
		try {
			const res = await client.auth.updateMyProfile({ username: val });
			auth.setUser(res.user!);
			usernameSuccess = true;
			usernameInput = '';
		} catch (e: unknown) {
			usernameError = (e as Error).message || m.profile_username_save_error();
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
			setMode(next ? 'light' : 'dark');
		} finally {
			darkModeSaving = false;
		}
	}

	// ── Locale switch ─────────────────────────────────────────────────────────
	async function switchLocale(l: 'en' | 'pl') {
		if (l === locale.current) return;
		const previous = locale.current;
		localeSaving = true;
		locale.set(l);
		try {
			const res = await client.auth.updateMyProfile({ defaultLanguage: l });
			auth.setUser(res.user!);
		} catch {
			locale.set(previous);
		} finally {
			localeSaving = false;
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
			deleteError = m.profile_delete_confirm_error();
			return;
		}
		deleteBusy = true;
		deleteError = '';
		try {
			await client.auth.deleteMyAccount({});
			auth.setUser(null);
			goto('/');
		} catch (e: unknown) {
			deleteError = (e as Error).message || m.profile_delete_failed();
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
	<h2 class="font-display text-3xl font-semibold text-foreground">{m.profile_title()}</h2>

	<!-- Identity -->
	<section class="rounded-xl border border-border bg-card p-6 space-y-4">
		<h3 class="text-lg font-medium text-foreground">{m.profile_section_identity()}</h3>

		<div class="grid gap-1">
			<p class="text-sm text-muted-foreground">{m.profile_email_label()}</p>
			<p class="font-medium text-foreground">{auth.user?.email ?? '—'}</p>
		</div>

		<div class="grid gap-1">
			<p class="text-sm text-muted-foreground">{m.profile_member_since()}</p>
			<p class="font-medium text-foreground">{memberSince()}</p>
		</div>

		<div class="grid gap-2">
			<label class="text-sm text-muted-foreground" for="username-input">
				{m.profile_username_label()}
				{#if auth.user?.username}
					<span class="ml-1 text-primary">@{auth.user.username}</span>
				{:else}
					<span class="ml-1 text-destructive text-xs">{m.profile_username_not_set()}</span>
				{/if}
			</label>
			<div class="flex gap-2">
				<Input
					id="username-input"
					type="text"
					placeholder={m.profile_username_placeholder()}
					bind:value={usernameInput}
					disabled={usernameSaving}
					class="flex-1 font-mono"
				/>
				<Button
					size="sm"
					disabled={usernameSaving || !usernameInput.trim()}
					onclick={saveUsername}
				>
					{usernameSaving ? m.common_saving() : m.common_save()}
				</Button>
			</div>
			{#if usernameError}
				<p class="text-sm text-destructive">{usernameError}</p>
			{/if}
			{#if usernameSuccess}
				<p class="text-sm text-primary">{m.profile_username_saved()}</p>
			{/if}
			<p class="text-xs text-muted-foreground">{m.profile_username_hint()}</p>
		</div>
	</section>

	<!-- Stats -->
	<section class="rounded-xl border border-border bg-card p-6">
		<h3 class="mb-4 text-lg font-medium text-foreground">{m.profile_section_activity()}</h3>
		{#if statsLoading}
			<div class="flex items-center gap-2 text-sm text-muted-foreground">
				<div class="h-4 w-4 animate-spin rounded-full border-2 border-border border-t-primary"></div>
				{m.common_loading()}
			</div>
		{:else if stats}
			<div class="grid grid-cols-3 gap-4 text-center">
				<a href="/reviews" class="group rounded-lg border border-border p-4 hover:border-primary/50 transition-colors">
					<p class="text-2xl font-bold text-foreground group-hover:text-primary transition-colors">{stats.reviewCount}</p>
					<p class="text-xs text-muted-foreground mt-1">{m.nav_my_reviews()}</p>
				</a>
				<a href="/wishlist" class="group rounded-lg border border-border p-4 hover:border-primary/50 transition-colors">
					<p class="text-2xl font-bold text-foreground group-hover:text-primary transition-colors">{stats.wishlistCount}</p>
					<p class="text-xs text-muted-foreground mt-1">{m.nav_wishlist()}</p>
				</a>
				<a href="/friends" class="group rounded-lg border border-border p-4 hover:border-primary/50 transition-colors">
					<p class="text-2xl font-bold text-foreground group-hover:text-primary transition-colors">{stats.friendCount}</p>
					<p class="text-xs text-muted-foreground mt-1">{m.nav_friends()}</p>
				</a>
			</div>
		{/if}
	</section>

	<!-- Preferences -->
	<section class="rounded-xl border border-border bg-card p-6 space-y-5">
		<h3 class="text-lg font-medium text-foreground">{m.profile_section_preferences()}</h3>

		<!-- Dark mode -->
		<div class="flex items-center justify-between">
			<div>
				<p class="font-medium text-foreground">{m.profile_dark_mode_label()}</p>
				<p class="text-sm text-muted-foreground">{m.profile_dark_mode_desc()}</p>
			</div>
			<button
				class="relative inline-flex h-6 w-11 items-center rounded-full transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring {mode.current === 'dark'
					? 'bg-primary'
					: 'bg-muted'}"
				role="switch"
				aria-label={m.profile_dark_mode_label()}
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

		<!-- Language -->
		<div class="flex items-center justify-between">
			<p class="font-medium text-foreground">{m.profile_locale_label()}</p>
			<div class="flex overflow-hidden rounded-md border border-border text-sm" aria-label={m.profile_locale_label()}>
				<button
					type="button"
					onclick={() => switchLocale('en')}
					disabled={localeSaving}
					class="px-4 py-1.5 transition-colors {locale.current === 'en'
						? 'bg-primary text-primary-foreground'
						: 'bg-card text-muted-foreground hover:bg-muted'}"
				>
					{m.profile_locale_en()}
				</button>
				<button
					type="button"
					onclick={() => switchLocale('pl')}
					disabled={localeSaving}
					class="border-l border-border px-4 py-1.5 transition-colors {locale.current === 'pl'
						? 'bg-primary text-primary-foreground'
						: 'bg-card text-muted-foreground hover:bg-muted'}"
				>
					{m.profile_locale_pl()}
				</button>
			</div>
		</div>
	</section>

	<!-- Danger zone -->
	<section class="rounded-xl border border-destructive/40 bg-card p-6 space-y-5">
		<h3 class="text-lg font-medium text-destructive">{m.profile_section_danger()}</h3>

		<!-- Sign out all devices -->
		<div class="flex items-start justify-between gap-4">
			<div>
				<p class="font-medium text-foreground">{m.profile_sign_out_all_label()}</p>
				<p class="text-sm text-muted-foreground">{m.profile_sign_out_all_desc()}</p>
			</div>
			<Button
				variant="outline"
				size="sm"
				class="shrink-0 text-destructive hover:border-destructive/50 hover:text-destructive"
				disabled={signOutAllBusy}
				onclick={signOutAll}
			>
				{signOutAllBusy ? m.profile_sign_out_all_busy() : m.profile_sign_out_all_btn()}
			</Button>
		</div>

		<hr class="border-border" />

		<!-- Delete account -->
		<div class="space-y-3">
			<div>
				<p class="font-medium text-foreground">{m.profile_delete_label()}</p>
				<p class="text-sm text-muted-foreground">{m.profile_delete_desc()}</p>
			</div>
			<div class="flex gap-2">
				<Input
					type="text"
					placeholder={m.profile_delete_placeholder()}
					bind:value={deleteConfirm}
					disabled={deleteBusy}
					class="flex-1 font-mono"
				/>
				<Button
					variant="destructive"
					size="sm"
					class="shrink-0"
					disabled={deleteBusy || deleteConfirm !== 'DELETE'}
					onclick={deleteAccount}
				>
					{deleteBusy ? m.common_deleting() : m.profile_delete_label()}
				</Button>
			</div>
			{#if deleteError}
				<p class="text-sm text-destructive">{deleteError}</p>
			{/if}
		</div>
	</section>
</div>
