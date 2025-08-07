import { createConnectTransport } from '@connectrpc/connect-web';
import { createClient } from '@connectrpc/connect';
import { RestaurantsService } from '$lib/client/generated/restaurants/v1/restaurants_service_pb';
import { UsersService } from '$lib/client/generated/users/v1/users_service_pb';
import { GoogleMapsService } from '$lib/client/generated/google_maps/v1/google_maps_service_pb';

const transport = createConnectTransport({
  baseUrl: 'http://localhost:3001',
});

const restaurants = createClient(RestaurantsService, transport);
const users = createClient(UsersService, transport);
const googleMaps = createClient(GoogleMapsService, transport);

export default { restaurants, users, googleMaps };
