<script lang="ts">
	import { page } from '$app/state';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Sheet from '$lib/components/ui/sheet/index.js';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
	import { auth } from '$lib/state/auth.svelte';
	import { mode, toggleMode } from '$lib/state/theme.svelte';
	import SocialSignIn from '$lib/ui/components/SocialSignIn.svelte';
	import client from '$lib/client/client';

	let loginOpen = $state(false);
	let activeUrl = $derived(page.url.pathname);

	function isActive(href: string) {
		return activeUrl === href;
	}

	function initDialog(el: HTMLDialogElement) {
		el.showModal();
		return {
			destroy() {
				if (el.open) el.close();
			}
		};
	}

	async function handleLogout() {
		try {
			await client.auth.logout({});
		} finally {
			auth.setUser(null);
		}
	}

	function getInitials(): string {
		if (auth.user?.username) return auth.user.username[0].toUpperCase();
		if (auth.user?.email) return auth.user.email[0].toUpperCase();
		return '?';
	}

	const authNavLinks = [
		{ href: '/reviews', label: 'My Reviews' },
		{ href: '/wishlist', label: 'Wishlist' },
		{ href: '/friends', label: 'Friends' }
	];

	$effect(() => {
		if (page.url.searchParams.get('login') === '1') {
			loginOpen = true;
		}
	});
</script>

<header class="sticky top-0 z-10 w-full border-b border-border bg-background/95 backdrop-blur-sm">
	<nav class="mx-auto flex max-w-6xl items-center justify-between px-4 py-3">
		<!-- Brand -->
		<a href="/" class="group flex items-center gap-2.5">
			<div
				class="flex h-7 w-7 items-center justify-center rounded-md bg-primary text-sm font-bold text-primary-foreground transition-opacity group-hover:opacity-80"
			>
				R
			</div>
			<span class="font-display text-lg font-semibold tracking-tight text-foreground">
				Restorate
			</span>
		</a>

		<!-- Desktop nav links (auth-gated) -->
		{#if auth.isLoggedIn}
			<ul class="hidden items-center gap-7 md:flex">
				{#each authNavLinks as link}
					<li>
						<a
							href={link.href}
							class="text-sm font-medium transition-colors {isActive(link.href)
								? 'border-b-2 border-primary pb-0.5 text-primary'
								: 'text-muted-foreground hover:text-foreground'}"
						>
							{link.label}
						</a>
					</li>
				{/each}
			</ul>
		{/if}

		<!-- Auth controls + mobile hamburger -->
		<div class="flex items-center gap-2">
			{#if auth.isLoggedIn}
				<!-- Avatar + Dropdown (desktop) -->
				<div class="hidden sm:block">
					<DropdownMenu.Root>
						<DropdownMenu.Trigger>
							{#snippet child({ props })}
								<button
									{...props}
									class="flex h-8 w-8 items-center justify-center rounded-full bg-primary text-xs font-semibold text-primary-foreground transition-opacity hover:opacity-80"
									aria-label="Account menu"
								>
									{getInitials()}
								</button>
							{/snippet}
						</DropdownMenu.Trigger>
						<DropdownMenu.Content align="end" class="w-48">
							<DropdownMenu.Label>Account</DropdownMenu.Label>
							<DropdownMenu.Item>
								<a href="/profile" class="w-full">My Profile</a>
							</DropdownMenu.Item>
							<DropdownMenu.Item onclick={handleLogout}>Sign out</DropdownMenu.Item>
							<DropdownMenu.Separator />
							<DropdownMenu.Label>Navigate</DropdownMenu.Label>
							<DropdownMenu.Item>
								<a href="/friends" class="w-full">Find a Friend</a>
							</DropdownMenu.Item>
							<DropdownMenu.Separator />
							<DropdownMenu.Label>Preferences</DropdownMenu.Label>
							<DropdownMenu.Item onclick={() => toggleMode()}>
								{mode.current === 'dark' ? 'Light mode' : 'Dark mode'}
							</DropdownMenu.Item>
						</DropdownMenu.Content>
					</DropdownMenu.Root>
				</div>
			{:else}
				<Button size="sm" onclick={() => (loginOpen = true)}>Sign in</Button>
			{/if}

			<!-- Mobile hamburger -->
			<Sheet.Root>
				<Sheet.Trigger>
					{#snippet child({ props })}
						<button
							{...props}
							class="ml-1 flex flex-col gap-1.5 rounded-md p-2 hover:bg-muted md:hidden"
							aria-label="Open menu"
						>
							<span class="block h-px w-5 bg-foreground"></span>
							<span class="block h-px w-5 bg-foreground"></span>
							<span class="block h-px w-5 bg-foreground"></span>
						</button>
					{/snippet}
				</Sheet.Trigger>
				<Sheet.Content side="right" class="w-64">
					<Sheet.Header>
						<div class="flex items-center gap-2.5 pb-4">
							<div
								class="flex h-7 w-7 items-center justify-center rounded-md bg-primary text-sm font-bold text-primary-foreground"
							>
								R
							</div>
							<span class="font-display font-semibold text-foreground">Restorate</span>
						</div>
					</Sheet.Header>
					<hr class="mb-4 border-border" />
					<nav class="flex flex-col gap-4 px-2">
						<a href="/" class="text-sm text-muted-foreground hover:text-foreground">Home</a>
						{#if auth.isLoggedIn}
							{#each authNavLinks as link}
								<a
									href={link.href}
									class="text-sm {isActive(link.href)
										? 'font-medium text-primary'
										: 'text-muted-foreground hover:text-foreground'}"
								>
									{link.label}
								</a>
							{/each}
							<hr class="border-border" />
							<button
								class="text-left text-sm text-muted-foreground hover:text-foreground"
								onclick={() => toggleMode()}
							>
								{mode.current === 'dark' ? 'Light mode' : 'Dark mode'}
							</button>
							<button
								class="text-left text-sm text-muted-foreground hover:text-foreground"
								onclick={handleLogout}
							>
								Sign out
							</button>
						{:else}
							<Button size="sm" onclick={() => (loginOpen = true)}>Sign in</Button>
						{/if}
					</nav>
				</Sheet.Content>
			</Sheet.Root>
		</div>
	</nav>
</header>

<!-- Sign in dialog -->
{#if loginOpen}
	<dialog
		use:initDialog
		oncancel={() => (loginOpen = false)}
		onclick={(e) => {
			const rect = (e.currentTarget as HTMLDialogElement).getBoundingClientRect();
			if (
				e.clientX < rect.left ||
				e.clientX > rect.right ||
				e.clientY < rect.top ||
				e.clientY > rect.bottom
			) {
				loginOpen = false;
			}
		}}
		class="m-auto w-full max-w-[calc(100%-2rem)] rounded-xl bg-card p-6 shadow-xl backdrop:bg-foreground/20 sm:max-w-sm"
	>
		<div class="flex flex-col gap-4">
			<h3 class="font-display text-xl font-semibold text-foreground">Sign in to Restorate</h3>
			<SocialSignIn onSuccess={() => (loginOpen = false)} />
		</div>
	</dialog>
{/if}
