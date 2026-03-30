import { describe, expect, test } from 'vitest';
import en from '../../../messages/en.json';
import pl from '../../../messages/pl.json';

describe('message completeness', () => {
	test('every en key exists in pl with a non-empty value', () => {
		for (const key of Object.keys(en)) {
			if (key === '$schema') continue;
			expect(pl, `pl is missing key: ${key}`).toHaveProperty(key);
			expect(
				(pl as Record<string, string>)[key],
				`pl["${key}"] is empty`
			).toBeTruthy();
		}
	});

	test('pl has no extra keys not present in en', () => {
		for (const key of Object.keys(pl)) {
			if (key === '$schema') continue;
			expect(en, `en is missing key: ${key}`).toHaveProperty(key);
		}
	});
});
