import { page } from '@vitest/browser/context';
import { describe, expect, it } from 'vitest';
import { render } from 'vitest-browser-svelte';
import Page from './+page.svelte';
import { auth } from '$lib/state/auth.svelte';

describe('/+page.svelte', () => {
	it('should render h1', async () => {
		// Clear the loading state so the page renders its content instead of the spinner.
		auth.setLoaded();

		render(Page);

		const heading = page.getByRole('heading', { level: 1 });
		await expect.element(heading).toBeInTheDocument();
	});
});
