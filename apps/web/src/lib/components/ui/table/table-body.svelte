<script lang="ts">
	import { cn } from '$lib/utils';
</script>

<tbody class={cn('[&_tr:last-child]:border-0')} {...$$restProps}>
	<slot />
</tbody>
