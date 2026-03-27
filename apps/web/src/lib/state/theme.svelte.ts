let dark = $state(false);
let initialized = $state(false);

export const theme = {
	get dark() {
		return dark;
	},
	get initialized() {
		return initialized;
	},
	toggle() {
		dark = !dark;
	},
	init() {
		const stored = localStorage.getItem('theme');
		if (stored === 'dark') {
			dark = true;
		} else if (stored === 'light') {
			dark = false;
		} else {
			dark = window.matchMedia('(prefers-color-scheme: dark)').matches;
		}
		initialized = true;
	}
};
