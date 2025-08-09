<script lang="ts">
	import { AppBar } from '@skeletonlabs/skeleton-svelte';
	import ListFilterPlus from '@lucide/svelte/icons/list-filter-plus';
	import Paperclip from '@lucide/svelte/icons/paperclip';
	import Calendar from '@lucide/svelte/icons/calendar';
	import CircleUser from '@lucide/svelte/icons/circle-user';
	import { ButtonSv } from '$lib/ui/components';

	let iconSize = 24;
</script>

<header>
	<AppBar
		headlineClasses="sm:hidden"
		centerClasses="hidden sm:block"
		background="bg-primary-900"
		base="text-primary-contrast-900"
	>
		{#snippet lead()}
			<ButtonSv variant="tonal" color="primary" size="md" class="space-x-4 sm:space-x-2">
				<ListFilterPlus size={iconSize} />
			</ButtonSv>
		{/snippet}
		{#snippet trail()}
			<div class="flex space-x-4 sm:space-x-2">
				<ButtonSv variant="tonal" color="primary" size="md">
					<Paperclip size={iconSize} />
				</ButtonSv>
				<ButtonSv variant="tonal" color="primary" size="md">
					<Calendar size={iconSize} />
				</ButtonSv>
				<ButtonSv variant="tonal" color="primary" size="md">
					<CircleUser size={iconSize} />
				</ButtonSv>
			</div>
		{/snippet}
	</AppBar>
</header>
