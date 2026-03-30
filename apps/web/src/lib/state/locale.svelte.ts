import { setLocale, getLocale } from '$lib/paraglide/runtime';

export type Locale = 'en' | 'pl';

const VALID: Locale[] = ['en', 'pl'];

function createLocale() {
	let current = $state<Locale>(getLocale() as Locale);
	return {
		get current() {
			return current;
		},
		set(l: Locale) {
			if (!VALID.includes(l)) return;
			setLocale(l);
			current = l;
		}
	};
}

export const locale = createLocale();
