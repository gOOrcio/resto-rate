<script lang="ts">
    import { cn } from "$lib/utils";
</script>

<td class={cn("p-4 align-middle [&:has([role=checkbox])]:pr-0")} {...$$restProps}>
    <slot />
</td> 