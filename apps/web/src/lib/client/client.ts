import { createConnectTransport } from '@connectrpc/connect-web';
import { createClient } from '@connectrpc/connect';
import { RestaurantService } from '$lib/client/generated/restaurants/v1/restaurants_service_pb';
import { UsersService } from '$lib/client/generated/users/v1/users_service_pb';


const transport = createConnectTransport({
  baseUrl: 'localhost:3001',
});

const restaurants = createClient(RestaurantService, transport);
const users = createClient(UsersService, transport);

export default { restaurants, users };
