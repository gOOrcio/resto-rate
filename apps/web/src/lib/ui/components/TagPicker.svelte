<script lang="ts">
	import client from '$lib/client/client';
	import type { TagProto } from '$lib/client/generated/tags/v1/tag_pb';

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
	<p class="text-sm text-gray-400">Loading tags…</p>
{:else if loadError}
	<div class="flex items-center gap-2 text-sm text-red-500">
		<span>Failed to load tags.</span>
		<button type="button" onclick={loadTags} class="underline hover:no-underline">Retry</button>
	</div>
{:else if tags.length === 0}
	<p class="text-sm text-gray-400">No tags available.</p>
{:else}
	<div class="flex flex-col gap-3">
		{#each Object.entries(grouped) as [category, categoryTags]}
			<div>
				<p class="mb-1 text-xs font-semibold uppercase tracking-wide text-gray-400">{category}</p>
				<div class="flex flex-wrap gap-1.5">
					{#each categoryTags as tag}
						<button
							type="button"
							onclick={() => toggleTag(tag.slug)}
							class="rounded-full px-2.5 py-0.5 text-xs font-medium transition-colors
								{selected.includes(tag.slug)
								? 'bg-blue-600 text-white'
								: 'bg-gray-100 text-gray-600 hover:bg-gray-200'}"
						>
							{tag.label}
						</button>
					{/each}
				</div>
			</div>
		{/each}
	</div>
{/if}
