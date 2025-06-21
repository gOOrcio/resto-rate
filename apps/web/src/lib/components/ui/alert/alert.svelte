<script lang="ts">
	import { cn } from '$lib/utils';
	import { tv, type VariantProps } from 'tailwind-variants';

	const alertVariants = tv({
		base: 'relative w-full rounded-lg border p-4 [&>svg~*]:pl-7 [&>svg+div]:translate-y-[-3px] [&>svg]:absolute [&>svg]:left-4 [&>svg]:top-4 [&>svg]:text-foreground',
		variants: {
			variant: {
				default: 'bg-background text-foreground',
				destructive:
					'border-destructive/50 text-destructive dark:border-destructive [&>svg]:text-destructive',
			},
		},
		defaultVariants: {
			variant: 'default',
		},
	});

	type AlertVariant = VariantProps<typeof alertVariants>['variant'];

	const {
		variant = 'default',
		class: className = '',
		children,
	} = $props<{
		variant?: AlertVariant;
		class?: string;
		children?: import('svelte').Snippet;
	}>();
</script>

<div class={cn(alertVariants({ variant }), className)}>
	{@render children?.()}
</div>
