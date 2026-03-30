import * as m from '$lib/paraglide/messages';

type MessageFn = () => string;

export function tagLabel(slug: string): string {
	const key = `tag_${slug.replace(/-/g, '_')}` as keyof typeof m;
	return typeof m[key] === 'function' ? (m[key] as MessageFn)() : slug;
}

export function tagCategoryLabel(category: string): string {
	const key = `tag_category_${category.toLowerCase()}` as keyof typeof m;
	return typeof m[key] === 'function' ? (m[key] as MessageFn)() : category;
}
