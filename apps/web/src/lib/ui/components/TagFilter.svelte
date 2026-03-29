<script lang="ts">
	import TagPicker from './TagPicker.svelte';

	let {
		selected = $bindable([]),
		mode = $bindable<'AND' | 'OR'>('OR'),
		onchange
	} = $props<{
		selected?: string[];
		mode?: 'AND' | 'OR';
		onchange?: (slugs: string[], mode: 'AND' | 'OR') => void;
	}>();

	function handleTagChange(slugs: string[]) {
		selected = slugs;
		onchange?.(slugs, mode);
	}

	function setMode(m: 'AND' | 'OR') {
		mode = m;
		onchange?.(selected, m);
	}
</script>

<div class="flex flex-col gap-3">
	<div class="flex items-center gap-2">
		<span class="text-xs font-medium text-muted-foreground">Match:</span>
		<div class="flex overflow-hidden rounded-md border border-border text-xs">
			<button
				type="button"
				onclick={() => setMode('OR')}
				class="px-3 py-1 transition-colors {mode === 'OR'
					? 'bg-primary text-primary-foreground'
					: 'bg-card text-muted-foreground hover:bg-muted'}"
			>
				Any
			</button>
			<button
				type="button"
				onclick={() => setMode('AND')}
				class="border-l border-border px-3 py-1 transition-colors {mode === 'AND'
					? 'bg-primary text-primary-foreground'
					: 'bg-card text-muted-foreground hover:bg-muted'}"
			>
				All
			</button>
		</div>
	</div>

	<TagPicker bind:selected onchange={handleTagChange} />
</div>
