<script lang="ts">
	import { Navbar, NavBrand, NavLi, NavUl, NavHamburger, Drawer } from "flowbite-svelte";
	import { sineIn } from "svelte/easing";
	let open6 = $state(false)
	let transitionParamsRight = {
		x: 80,
		duration: 200,
		easing: sineIn
	};
</script>

	<header>
		<Navbar class="bg-primary-200 dark:bg-primary-900 px-4">
			<div class="flex items-center justify-between w-full">
					<NavBrand href="/">
						<img src="/favicon.svg" class="me-3 h-6 sm:h-9" alt="App Logo" />
						<span class="self-center text-xl font-semibold whitespace-nowrap dark:text-white">Restorate</span>
					</NavBrand>

				<div class="flex items-center">
					<div class="hidden md:block">
						<NavUl>
							<div class="flex flex-row space-x-2 items-center">
								<NavLi href="/">Home</NavLi>
								<NavLi href="/about">About</NavLi>
								<NavLi href="/docs/components/navbar">Navbar</NavLi>
								<NavLi href="/pricing">Pricing</NavLi>
								<NavLi href="/contact">Contact</NavLi>
							</div>
						</NavUl>
					</div>

					<NavHamburger onclick={() => (open6 = true)} name="" class="md:hidden ml-3" />
				</div>
			</div>
		</Navbar>
</header>

<Drawer
	bind:open={open6}
	placement="right"
	transitionParams={transitionParamsRight}
	backdropClass="bg-black/50"
	class="!fixed !top-0 !bottom-0 !right-0 !left-auto !m-0
         !h-[100dvh] !w-25 !max-w-none !rounded-none p-0
         overflow-y-auto bg-surface-50 dark:bg-surface-900"
	formClass="!h-full !rounded-none !p-0"
>
	<div class="h-full p-4 items-center flex flex-col gap-1">
		<a href="/" on:click={() => (open6 = false)}>Home</a>
		<a href="/about" on:click={() => (open6 = false)}>About</a>
		<a href="/docs/components/navbar" on:click={() => (open6 = false)}>Navbar</a>
		<a href="/pricing" on:click={() => (open6 = false)}>Pricing</a>
		<a href="/contact" on:click={() => (open6 = false)}>Contact</a>
	</div>
</Drawer>
