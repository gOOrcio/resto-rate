<script lang="ts">
	import { cn } from '$lib/utils';
</script>

<label
	class={cn(
		'text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70'
	)}
	{...$$restProps}
>
	<slot />
</label>
