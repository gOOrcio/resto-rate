import { createConnectTransport } from '@connectrpc/connect-web';
import { createClient } from '@connectrpc/connect';
import { RestaurantsService } from '$lib/client/generated/restaurants/v1/restaurants_service_pb';
import { UsersService } from '$lib/client/generated/users/v1/users_service_pb';
import { GoogleMapsService } from '$lib/client/generated/google_maps/v1/google_maps_service_pb';
import { AuthService } from '$lib/client/generated/auth/v1/auth_service_pb';
import { ReviewsService } from '$lib/client/generated/reviews/v1/reviews_service_pb';

const baseUrl = import.meta.env.VITE_API_URL || 'http://localhost:3001';
const transport = createConnectTransport({
  baseUrl: baseUrl,
  useHttpGet: false,
  fetch: (input, init) => globalThis.fetch(input, { ...init, credentials: 'include' }),
  interceptors: []
});

const restaurants = createClient(RestaurantsService, transport);
const users = createClient(UsersService, transport);
const googleMaps = createClient(GoogleMapsService, transport);
const auth = createClient(AuthService, transport);
const reviews = createClient(ReviewsService, transport);

export default { restaurants, users, googleMaps, auth, reviews };
