<script lang="ts">
	import TagPicker from './TagPicker.svelte';
	import * as m from '$lib/paraglide/messages';

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

	function setMode(m_: 'AND' | 'OR') {
		mode = m_;
		onchange?.(selected, m_);
	}
</script>

<div class="flex flex-col gap-3">
	<div class="flex items-center gap-2">
		<span class="text-xs font-medium text-muted-foreground">{m.common_filter_tag_mode_match()}</span>
		<div class="flex overflow-hidden rounded-md border border-border text-xs">
			<button
				type="button"
				onclick={() => setMode('OR')}
				class="px-3 py-1 transition-colors {mode === 'OR'
					? 'bg-primary text-primary-foreground'
					: 'bg-card text-muted-foreground hover:bg-muted'}"
			>
				{m.common_filter_tag_mode_any()}
			</button>
			<button
				type="button"
				onclick={() => setMode('AND')}
				class="border-l border-border px-3 py-1 transition-colors {mode === 'AND'
					? 'bg-primary text-primary-foreground'
					: 'bg-card text-muted-foreground hover:bg-muted'}"
			>
				{m.common_filter_tag_mode_all()}
			</button>
		</div>
	</div>

	<TagPicker bind:selected onchange={handleTagChange} />
</div>
