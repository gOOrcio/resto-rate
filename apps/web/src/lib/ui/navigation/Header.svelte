<script lang="ts">
	import { page } from '$app/state';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Sheet from '$lib/components/ui/sheet/index.js';
	import { auth } from '$lib/state/auth.svelte';
	import LoginModal from '$lib/ui/components/LoginModal.svelte';
	import client from '$lib/client/client';

	let loginOpen = $state(false);
	let activeUrl = $derived(page.url.pathname);

	function isActive(href: string) {
		return activeUrl === href;
	}

	async function handleLogout() {
		try {
			await client.auth.logout({});
		} finally {
			auth.setUser(null);
		}
	}
</script>

<header class="sticky top-0 z-10 w-full bg-blue-200 p-2 shadow-sm">
	<nav class="flex w-full items-center justify-between px-2">
		<!-- Brand -->
		<a href="/" class="flex items-center gap-2">
			<img src="/resto-rate-logo.svg" class="h-5 w-5" alt="App Logo" />
			<span class="self-center whitespace-nowrap text-xl font-semibold text-blue-800">
				Restorate
			</span>
		</a>

		<!-- Desktop nav links -->
		<ul class="hidden items-center gap-6 md:flex">
			{#each [{ href: '/', label: 'Home' }, { href: '/about', label: 'About' }, { href: '/pricing', label: 'Pricing' }, { href: '/contact', label: 'Contact' }] as link}
				<li>
					<a
						href={link.href}
						class="text-sm font-medium transition-colors {isActive(link.href)
							? 'text-blue-700 underline underline-offset-4'
							: 'text-gray-700 hover:text-blue-700'}"
					>
						{link.label}
					</a>
				</li>
			{/each}
		</ul>

		<!-- Auth controls + mobile hamburger -->
		<div class="flex items-center gap-2">
			{#if auth.isLoggedIn}
				<span class="hidden text-sm text-gray-700 sm:inline">{auth.user?.username}</span>
				<Button size="sm" variant="outline" onclick={handleLogout}>Logout</Button>
			{:else}
				<Button size="sm" onclick={() => (loginOpen = true)}>Login</Button>
			{/if}

			<!-- Mobile hamburger (Sheet trigger) -->
			<Sheet.Root>
				<Sheet.Trigger>
					{#snippet child({ props })}
						<button
							{...props}
							class="ml-2 flex flex-col gap-1 rounded p-2 hover:bg-blue-100 md:hidden"
							aria-label="Open menu"
						>
							<span class="block h-0.5 w-5 bg-gray-700"></span>
							<span class="block h-0.5 w-5 bg-gray-700"></span>
							<span class="block h-0.5 w-5 bg-gray-700"></span>
						</button>
					{/snippet}
				</Sheet.Trigger>
				<Sheet.Content side="right" class="w-64">
					<Sheet.Header>
						<div class="flex items-center gap-2 pb-4">
							<img src="/resto-rate-logo.svg" class="h-6 w-6" alt="App Logo" />
							<span class="font-semibold text-blue-800">Restorate</span>
						</div>
					</Sheet.Header>
					<hr class="mb-4 border-gray-200" />
					<nav class="flex flex-col gap-4 px-2">
						<a href="/" class="text-gray-700 hover:text-blue-700">Home</a>
						<a href="/about" class="text-gray-700 hover:text-blue-700">About</a>
						<a href="/pricing" class="text-gray-700 hover:text-blue-700">Pricing</a>
						<a href="/contact" class="text-gray-700 hover:text-blue-700">Contact</a>
					</nav>
				</Sheet.Content>
			</Sheet.Root>
		</div>
	</nav>
</header>

<LoginModal bind:open={loginOpen} />
