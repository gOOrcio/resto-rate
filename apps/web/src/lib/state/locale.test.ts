import { beforeEach, describe, expect, test, vi } from 'vitest';

const { mockSetLocale, mockGetLocale } = vi.hoisted(() => ({
	mockSetLocale: vi.fn(),
	mockGetLocale: vi.fn(() => 'en')
}));

vi.mock('$lib/paraglide/runtime', () => ({
	setLocale: mockSetLocale,
	getLocale: mockGetLocale
}));

import { locale } from './locale.svelte';

describe('locale', () => {
	beforeEach(() => {
		mockSetLocale.mockClear();
	});

	test('set("pl") calls setLocale with pl', () => {
		locale.set('pl');
		expect(mockSetLocale).toHaveBeenCalledWith('pl');
	});

	test('set("en") calls setLocale with en', () => {
		locale.set('en');
		expect(mockSetLocale).toHaveBeenCalledWith('en');
	});

	test('set with unsupported locale does not call setLocale', () => {
		locale.set('de' as 'en');
		expect(mockSetLocale).not.toHaveBeenCalled();
	});
});
