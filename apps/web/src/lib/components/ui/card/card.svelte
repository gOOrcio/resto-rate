<script lang="ts">
    import { cn } from "$lib/utils";
</script>

<div class={cn("rounded-lg border bg-card text-card-foreground shadow-sm")} {...$$restProps}>
    <slot />
</div> 