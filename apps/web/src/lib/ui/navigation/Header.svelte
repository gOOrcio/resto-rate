<script lang="ts">
	import { AppBar } from '@skeletonlabs/skeleton-svelte';
	import ListFilterPlus from '@lucide/svelte/icons/list-filter-plus';
	import Paperclip from '@lucide/svelte/icons/paperclip';
	import Calendar from '@lucide/svelte/icons/calendar';
	import CircleUser from '@lucide/svelte/icons/circle-user';

	let iconSize = 24;
</script>

<header>
	<AppBar
		headlineClasses="sm:hidden"
		centerClasses="hidden sm:block"
		background="bg-primary-900"
		base="--color-primary-contrast-900"
	>
		{#snippet lead()}
			<button class="btn space-x-4 sm:space-x-2">
				<ListFilterPlus size={iconSize} />
			</button>
		{/snippet}
		{#snippet trail()}
			<div class="flex space-x-4 sm:space-x-2">
				<button class="btn"><Paperclip size={iconSize} /></button>
				<button class="btn"><Calendar size={iconSize} /></button>
				<button class="btn"><CircleUser size={iconSize} /></button>
			</div>
		{/snippet}
	</AppBar>
</header>
