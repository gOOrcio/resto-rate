import { browser } from '$app/environment';
import { writable } from 'svelte/store';

type Theme = 'light' | 'dark';

function createThemeStore() {
    const stored = browser ? localStorage.getItem('theme') as Theme : 'light';
    const { subscribe, set } = writable<Theme>(stored);

    return {
        subscribe,
        set: (theme: Theme) => {
            if (browser) {
                localStorage.setItem('theme', theme);
                document.documentElement.classList.toggle('dark', theme === 'dark');
            }
            set(theme);
        },
        toggle: () => {
            subscribe(theme => {
                set(theme === 'dark' ? 'light' : 'dark');
            })();
        }
    };
}

export const theme = createThemeStore(); 