import type { FastifyPluginAsync } from 'fastify';
import { z } from 'zod';
import { searchPlaces } from '../services/google-places.service';
import { handleRoute } from '../utils/route-helpers';

const searchSchema = z.object({
	q: z.string().min(1),
});

export const googleRoutes: FastifyPluginAsync = async (fastify) => {
	fastify.get('/places', async (request, reply) => {
		return handleRoute(reply, async () => {
			const { q } = searchSchema.parse(request.query);
			return await searchPlaces(q);
		});
	});
};
