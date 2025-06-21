<script lang="ts">
	import { cn } from '$lib/utils';
</script>

<div class="relative w-full overflow-auto">
	<table class={cn('w-full caption-bottom text-sm')} {...$$restProps}>
		<slot />
	</table>
</div>
