import { browser } from '$app/environment';
import { writable } from 'svelte/store';

type Theme = 'dark' | 'light';

function createThemeStore() {
	const storedTheme = browser ? (localStorage.getItem('theme') as Theme | null) : null;
	const prefersDark = browser ? window.matchMedia('(prefers-color-scheme: dark)').matches : false;
	const initialTheme = storedTheme ?? (prefersDark ? 'dark' : 'light');

	const { subscribe, set, update } = writable<Theme>(initialTheme);

	if (browser) {
		document.documentElement.classList.toggle('dark', initialTheme === 'dark');
	}

	return {
		subscribe,
		set: (value: Theme) => {
			if (browser) {
				document.documentElement.classList.toggle('dark', value === 'dark');
				localStorage.setItem('theme', value);
			}
			set(value);
		},
		toggle: () => {
			update((current) => {
				const newTheme = current === 'dark' ? 'light' : 'dark';
				if (browser) {
					document.documentElement.classList.toggle('dark', newTheme === 'dark');
					localStorage.setItem('theme', newTheme);
				}
				return newTheme;
			});
		},
	};
}

export const theme = createThemeStore();
