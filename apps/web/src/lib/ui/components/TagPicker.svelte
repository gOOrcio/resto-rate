<script lang="ts">
	import client from '$lib/client/client';
	import type { TagProto } from '$lib/client/generated/tags/v1/tag_pb';
	import { tagLabel, tagCategoryLabel } from '$lib/i18n/tags';
	import * as m from '$lib/paraglide/messages';

	let { selected = $bindable([]), onchange } = $props<{
		selected?: string[];
		onchange?: (slugs: string[]) => void;
	}>();

	let tags = $state<TagProto[]>([]);
	let loading = $state(true);
	let loadError = $state(false);

	async function loadTags() {
		loading = true;
		loadError = false;
		try {
			const res = await client.tags.listTags({});
			tags = res.tags;
		} catch {
			loadError = true;
		} finally {
			loading = false;
		}
	}

	$effect(() => {
		loadTags();
	});

	function toggleTag(slug: string) {
		const next = selected.includes(slug)
			? selected.filter((s: string) => s !== slug)
			: [...selected, slug];
		selected = next;
		onchange?.(next);
	}

	// Group tags by category, preserving server order
	const grouped = $derived(
		tags.reduce(
			(acc, tag) => {
				if (!acc[tag.category]) acc[tag.category] = [];
				acc[tag.category].push(tag);
				return acc;
			},
			{} as Record<string, TagProto[]>
		)
	);
</script>

{#if loading}
	<p class="text-sm text-muted-foreground">{m.tag_picker_loading()}</p>
{:else if loadError}
	<div class="flex items-center gap-2 text-sm text-destructive">
		<span>{m.tag_picker_error()}</span>
		<button type="button" onclick={loadTags} class="underline hover:no-underline">{m.common_retry()}</button>
	</div>
{:else if tags.length === 0}
	<p class="text-sm text-muted-foreground">{m.tag_picker_empty()}</p>
{:else}
	<div class="flex flex-col gap-3">
		{#each Object.entries(grouped) as [category, categoryTags]}
			<div>
				<p class="mb-1 text-xs font-semibold uppercase tracking-wide text-muted-foreground">{tagCategoryLabel(category)}</p>
				<div class="flex flex-wrap gap-1.5">
					{#each categoryTags as tag}
						<button
							type="button"
							onclick={() => toggleTag(tag.slug)}
							class="rounded-full px-2.5 py-0.5 text-xs font-medium transition-colors
								{selected.includes(tag.slug)
								? 'bg-primary text-primary-foreground'
								: 'bg-muted text-muted-foreground hover:bg-muted/70'}"
						>
							{tagLabel(tag.slug)}
						</button>
					{/each}
				</div>
			</div>
		{/each}
	</div>
{/if}
