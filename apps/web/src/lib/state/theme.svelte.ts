let dark = $state(false);

export const theme = {
	get dark() {
		return dark;
	},
	toggle() {
		dark = !dark;
	},
	init() {
		const stored = typeof localStorage !== 'undefined' ? localStorage.getItem('theme') : null;
		if (stored === 'dark') {
			dark = true;
		} else if (stored === 'light') {
			dark = false;
		} else {
			dark = typeof window !== 'undefined' && window.matchMedia('(prefers-color-scheme: dark)').matches;
		}
	}
};
