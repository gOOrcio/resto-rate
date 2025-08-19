<script lang="ts">
	import { page } from '$app/state';
	import { Navbar, NavBrand, NavLi, NavUl, NavHamburger, Drawer, Hr } from 'flowbite-svelte';
	import { sineIn } from 'svelte/easing';

	let openDrawer = $state(false);
	let activeUrl = $derived(page.url.pathname);
	let activeClass =
		'text-white bg-green-700 md:bg-transparent md:text-red-700 md:dark:text-white dark:bg-green-600 md:dark:bg-transparent';
	let nonActiveClass =
		'text-gray-700 hover:bg-gray-100 md:hover:bg-transparent md:border-0 md:hover:text-red-700 dark:text-gray-400 md:dark:hover:text-white dark:hover:bg-gray-700 dark:hover:text-white md:dark:hover:bg-transparent';
	let transitionParamsRight = {
		x: 80,
		duration: 200,
		easing: sineIn
	};
</script>

<header>
	<Navbar class="bg-primary-200 dark:bg-primary-900 sm:px- sticky top-0 z-10 w-full p-2">
		<div class="flex w-full items-center justify-between align-middle">
			<NavBrand href="/">
				<img src="/resto-rate-logo.svg" class="h-6 sm:h-9" alt="App Logo" />
				<span class="self-center whitespace-nowrap text-xl font-semibold dark:text-white"
					>Restorate</span
				>
			</NavBrand>

			<div class="flex items-center">
				<div class="hidden md:block">
					<NavUl {activeUrl} classes={{ active: activeClass, nonActive: nonActiveClass }}>
						<div class="flex flex-row items-center space-x-4">
							<NavLi href="/">Home</NavLi>
							<NavLi href="/about">About</NavLi>
							<NavLi href="/pricing">Pricing</NavLi>
							<NavLi href="/contact">Contact</NavLi>
						</div>
					</NavUl>
				</div>

				<NavHamburger onclick={() => (openDrawer = true)} name="" class="ml-3 md:hidden" />
			</div>
		</div>
	</Navbar>
</header>

<Drawer
	bind:open={openDrawer}
	placement="right"
	transitionParams={transitionParamsRight}
	class="!w-27 bg-surface-50 dark:bg-surface-900 !fixed !bottom-0 !left-auto
         !right-0 !top-0 !z-50 !m-0 !h-screen
         !min-h-screen !max-w-none overflow-hidden !rounded-none !p-0"
>
	<div class="flex w-full items-center justify-center p-4">
		<img src="/resto-rate-logo.svg" class="h-6 sm:h-9" alt="App Logo" />
	</div>
	<Hr class="mx-auto my-4 h-1 w-10 rounded-sm md:my-10" />
	<div class="flex flex-col items-center gap-4 px-4">
		<a href="/" onclick={() => (openDrawer = false)}>Home</a>
		<a href="/" onclick={() => (openDrawer = false)}>About</a>
		<a href="/" onclick={() => (openDrawer = false)}>Navbar</a>
		<a href="/" onclick={() => (openDrawer = false)}>Pricing</a>
		<a href="/" onclick={() => (openDrawer = false)}>Contact</a>
	</div>
</Drawer>
