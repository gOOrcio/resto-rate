# Design: Flowbite → shadcn-svelte Migration

**Date:** 2026-02-24
**Branch:** dummy-auth (migration will proceed here)
**Approach:** Big bang — remove Flowbite entirely, rewrite all affected files in one pass

---

## Decisions

| Concern | Decision |
|---|---|
| Theme | Clean slate — shadcn defaults, remove custom oklch palette and Flowbite ThemeConfig |
| Dark mode | Not implemented in this migration |
| Animations | Replace CSS `transition-[max-width]` with Svelte `fly` transition |
| Icons | Lucide (`lucide-svelte`) — shadcn's standard icon set |

---

## Dependencies

### Remove
- `flowbite`
- `flowbite-svelte`
- `flowbite-svelte-icons`

### Add
- shadcn-svelte (via `bunx shadcn-svelte@latest init`)
- `lucide-svelte`
- `bits-ui` (installed automatically by shadcn)
- `clsx` + `tailwind-merge` (installed automatically by shadcn, provides `cn()` utility)

---

## CSS / Theme Changes (`app.css`)

**Remove:**
- `@plugin 'flowbite/plugin'`
- `@source "../node_modules/flowbite-svelte/dist"`
- `@source "../node_modules/flowbite-svelte-icons/dist"`
- Import of `resto-rate-theme.css`

**Keep:**
- `@import 'tailwindcss'`
- `@import 'tw-animate-css'`
- `@plugin '@tailwindcss/forms'`
- `@plugin '@tailwindcss/typography'`

**Add (via shadcn init):**
- CSS variable block: `--background`, `--foreground`, `--primary`, `--secondary`, `--muted`, `--accent`, `--destructive`, `--border`, `--ring`, etc.

**Delete:** `src/resto-rate-theme.css`, `src/lib/ui/theme/index.ts`

---

## Component Migration Map

### `+layout.svelte`
- Remove `ThemeProvider` import and wrapper
- No other changes

### `Footer.svelte`
- `Hr` → shadcn `Separator`

### `Header.svelte`
- `Navbar/NavBrand/NavUl/NavLi/NavHamburger` → semantic HTML (`<nav>`, `<a>`, `<ul>`, `<li>`) styled with Tailwind
- `Drawer` (mobile nav) → shadcn `Sheet`
- `Hr` → shadcn `Separator`
- `Button` → shadcn `Button`

### `LoginModal.svelte`
- Native `<dialog>` stays
- `Button` → shadcn `Button`
- `Input` → shadcn `Input`
- `Label` → shadcn `Label`
- `Helper` (error text) → plain `<p class="text-sm text-destructive">`
- `CloseOutline` icon → Lucide `X`

### `RestaurantSearchSv.svelte`
- `Input` → shadcn `Input`

### `RestaurantCard.svelte`
- `Spinner` → shadcn `Skeleton` or CSS animate-spin on a Lucide `Loader2` icon
- `Rating` / `Star` (star rating display) → hand-rolled: loop 1–5, Lucide `Star` icon filled/empty
- `RatingIconProps` type → removed
- `EditOutline` → Lucide `Pencil`
- `CheckOutline` → Lucide `Check`
- `CloseOutline` → Lucide `X`
- `MapPinAltOutline` → Lucide `MapPin`
- `ChevronRightOutline` → Lucide `ChevronRight`
- `ChevronLeftOutline` → Lucide `ChevronLeft`
- `PhoneOutline` → Lucide `Phone`
- `GlobeOutline` → Lucide `Globe`
- Slide animation: CSS `transition-[max-width,opacity]` → Svelte `fly` transition (`{ x: 50, duration: 300 }`)

---

## shadcn Components to Install

```
bunx shadcn-svelte@latest add button input label separator sheet
```

---

## File Structure After Migration

```
apps/web/src/
├── app.css                          # Tailwind + shadcn CSS vars, no flowbite
├── lib/
│   └── ui/
│       ├── components/
│       │   ├── ui/                  # shadcn-generated (button, input, label, separator, sheet)
│       │   ├── RestaurantCard.svelte
│       │   ├── RestaurantSearchSv.svelte
│       │   └── LoginModal.svelte
│       ├── navigation/
│       │   ├── Header.svelte
│       │   └── Footer.svelte
│       └── icons/
│           └── XIcon.svelte         # kept (custom SVG, not Flowbite)
│   (theme/index.ts — deleted)
└── routes/                          # unchanged
```

---

## Verification

After migration, confirm:
1. `cd apps/web && bun run check` passes (no TypeScript/Svelte errors)
2. `bun run dev:web` starts without errors
3. Header renders with working mobile drawer
4. Login modal opens/closes correctly
5. Restaurant search input works
6. Restaurant card renders with star rating and Lucide icons
7. Google details panel slides in with Svelte `fly` transition
8. No references to `flowbite`, `flowbite-svelte`, or `flowbite-svelte-icons` remain in source
