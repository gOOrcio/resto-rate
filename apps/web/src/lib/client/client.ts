import { createConnectTransport } from '@connectrpc/connect-web';
import { createClient } from '@connectrpc/connect';
import { RestaurantsService } from '$lib/client/generated/restaurants/v1/restaurants_service_pb';
import { UsersService } from '$lib/client/generated/users/v1/users_service_pb';


const transport = createConnectTransport({
  baseUrl: 'http://localhost:3001',
});

const restaurants = createClient(RestaurantsService, transport);
const users = createClient(UsersService, transport);

export default { restaurants, users };
