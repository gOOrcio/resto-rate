import type { Place } from '$lib/client/generated/google_maps/v1/google_maps_service_pb';

/** Returns the best available city name from a Place, falling back through available address components. */
export function extractCity(place: Place): string {
	return place.postalAddress?.locality || place.postalAddress?.administrativeArea || '';
}
