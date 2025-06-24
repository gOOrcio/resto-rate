import { getGoogleConfig } from '@resto-rate/config';
import type { Place } from '@googlemaps/google-maps-services-js';

const BASE_URL = 'https://maps.googleapis.com/maps/api/place';

type PlacesResponse = {
	results: Partial<Place>[];
	status: string;
};

export async function searchPlaces(query: string): Promise<Partial<Place>[]> {
	const config = getGoogleConfig();
	if (!config.apiKey) {
		throw new Error('Google API key is not configured.');
	}

	const url = `${BASE_URL}/textsearch/json?query=${encodeURIComponent(
		query
	)}&type=restaurant&key=${config.apiKey}`;

	const response = await fetch(url);
	const data = (await response.json()) as PlacesResponse;

	if (data.status !== 'OK' && data.status !== 'ZERO_RESULTS') {
		throw new Error(`Google Places API error: ${data.status}`);
	}

	return data.results;
}
