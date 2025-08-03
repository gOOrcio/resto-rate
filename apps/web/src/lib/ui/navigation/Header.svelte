<script lang="ts">
	import { AppBar } from '@skeletonlabs/skeleton-svelte';
	import SquareMenu from '@lucide/svelte/icons/square-menu';
	import Paperclip from '@lucide/svelte/icons/paperclip';
	import Calendar from '@lucide/svelte/icons/calendar';
	import CircleUser from '@lucide/svelte/icons/circle-user';

	let iconSize = 26;
</script>

<header>
	<AppBar
		headlineClasses="sm:hidden"
		centerClasses="hidden sm:block"
		background="bg-primary-900"
		base="--color-primary-contrast-900"
	>
		{#snippet lead()}
			<div class="space-x-4 sm:space-x-10">
				<SquareMenu size={iconSize} />
			</div>
		{/snippet}
		{#snippet trail()}
			<div class="flex space-x-4 sm:space-x-10">
				<Paperclip size={iconSize} />
				<Calendar size={iconSize} />
				<CircleUser size={iconSize} />
			</div>
		{/snippet}
	</AppBar>
</header>
