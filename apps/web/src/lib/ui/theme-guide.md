# Resto Rate Theme Guide

This guide explains how to use the unified theming system for the Resto Rate application.

## Theme Overview

The application uses a custom theme called `resto-rate-theme` that extends Skeleton UI with custom colors and styling. The theme is defined in `src/resto-rate-theme.css` and applied globally.

## Color Palette

### Primary Colors (Purple/Blue)
- **Primary-50**: Light purple background
- **Primary-500**: Main brand color
- **Primary-900**: Dark purple for headers and navigation
- **Primary-contrast-900**: Light text on dark backgrounds

### Secondary Colors (Green)
- **Secondary-50**: Light green background
- **Secondary-500**: Accent green color
- **Secondary-900**: Dark green

### Surface Colors (Neutral)
- **Surface-50**: Light background
- **Surface-200**: Card backgrounds
- **Surface-600**: Text color
- **Surface-950**: Dark background

### Semantic Colors
- **Success**: Green for success states
- **Warning**: Orange for warnings
- **Error**: Red for errors

## Component Usage

### Buttons

```svelte
<!-- Primary button -->
<Button variant="filled" color="primary" size="md">
  Primary Action
</Button>

<!-- Secondary button -->
<Button variant="outlined" color="secondary" size="md">
  Secondary Action
</Button>

<!-- Tonal button -->
<Button variant="tonal" color="surface" size="md">
  Tonal Action
</Button>
```

### Cards

```svelte
<!-- Surface card -->
<Card variant="outlined" color="surface" padding="md">
  <h3>Card Title</h3>
  <p>Card content</p>
</Card>

<!-- Primary card -->
<Card variant="filled" color="primary" padding="lg">
  <h3>Primary Card</h3>
  <p>Content</p>
</Card>
```

### Inputs

```svelte
<!-- Basic surface input -->
<Input 
  type="text" 
  variant="outlined" 
  color="surface" 
  placeholder="Enter text"
/>

<!-- Primary input -->
<Input 
  type="email" 
  variant="filled" 
  color="primary" 
  placeholder="Email"
/>

<!-- Custom background and placeholder colors -->
<Input 
  type="text"
  variant="outlined"
  color="surface"
  bgColor="bg-primary-50"
  placeholderColor="placeholder:text-primary-400"
  placeholder="Custom styled input"
/>

<!-- Dark mode aware custom colors -->
<Input 
  type="text"
  variant="outlined"
  color="surface"
  bgColor="bg-surface-100 dark:bg-surface-800"
  placeholderColor="placeholder:text-surface-400 dark:placeholder:text-surface-500"
  placeholder="Dark mode aware"
/>
```

### Badges

```svelte
<!-- Primary badge -->
<Badge variant="filled" color="primary" size="sm">
  New
</Badge>

<!-- Secondary badge -->
<Badge variant="outlined" color="secondary" size="md">
  Popular
</Badge>
```

## Tailwind Classes

### Text Colors
- `text-primary-900` - Dark primary text
- `text-primary-100` - Light primary text
- `text-surface-600` - Body text
- `text-surface-400` - Muted text

### Background Colors
- `bg-primary-900` - Dark primary background
- `bg-surface-50` - Light surface background
- `bg-surface-950` - Dark surface background

### Border Colors
- `border-primary-200` - Light primary border
- `border-surface-200` - Light surface border

### Input Customization
- `bg-{color}-{shade}` - Custom background colors
- `placeholder:text-{color}-{shade}` - Custom placeholder colors
- `dark:bg-{color}-{shade}` - Dark mode background colors
- `dark:placeholder:text-{color}-{shade}` - Dark mode placeholder colors

## Dark Mode Support

All components automatically support dark mode using the `dark:` prefix:

```svelte
<div class="bg-surface-50 dark:bg-surface-950">
  <h1 class="text-primary-900 dark:text-primary-100">
    Title
  </h1>
</div>
```

## Skeleton UI Components

### Available Components
- **Accordion**: Collapsible content sections
- **AppBar**: Top navigation bar
- **Avatar**: User profile pictures
- **Navigation**: Side navigation
- **Pagination**: Page navigation
- **Progress**: Progress indicators
- **ProgressRing**: Circular progress
- **Rating**: Star ratings
- **Segment Control**: Toggle buttons
- **Slider**: Range sliders
- **Switch**: Toggle switches
- **Tabs**: Tabbed content
- **Tags Input**: Tag input field
- **Toast**: Notifications

### Usage Example

```svelte
<script>
  import { AppBar, Avatar, ProgressRing } from '@skeletonlabs/skeleton-svelte';
</script>

<AppBar background="bg-primary-900" base="text-primary-contrast-900">
  <Avatar src="/user.jpg" name="User" />
</AppBar>

<ProgressRing value={75} />
```

## Best Practices

1. **Use semantic colors**: Use `primary`, `secondary`, `success`, `warning`, `error` for their intended purposes
2. **Consistent spacing**: Use the spacing scale defined in the theme
3. **Dark mode**: Always include dark mode variants
4. **Accessibility**: Ensure sufficient contrast ratios
5. **Component composition**: Use the provided UI components for consistency

## Theme Customization

To modify the theme, edit `src/resto-rate-theme.css`. The theme uses CSS custom properties that can be overridden:

```css
[data-theme='resto-rate-theme'] {
  --color-primary-500: oklch(48.65% 0.3 279.02deg);
  --color-secondary-500: oklch(89.46% 0.16 171.47deg);
  /* Add more customizations */
}
``` 