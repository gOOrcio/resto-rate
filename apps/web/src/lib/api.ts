import { encode, decode } from '@msgpack/msgpack';

// Client-side API URL - use environment variable or fallback to localhost in dev
const API_BASE_URL =
	typeof window !== 'undefined'
		? window.location.hostname === 'localhost'
			? 'http://localhost:3001/api'
			: '/api'
		: 'http://localhost:3001/api';

type CreateRestaurantRequest = {
	name: string;
	address?: string;
	rating?: number; // 1-5 scale
	comment?: string;
};

export class ApiClient {
	private baseUrl: string;

	constructor(baseUrl: string = API_BASE_URL) {
		this.baseUrl = baseUrl;
	}

	private async request<T>(
		endpoint: string,
		options: Omit<RequestInit, 'body'> & { body?: unknown } = {},
		sessionId?: string
	): Promise<T> {
		const url = `${this.baseUrl}${endpoint}`;
		const headers: HeadersInit = {
			'Accept': 'application/msgpack', // Always request MessagePack responses
		};

		// Only set Content-Type for requests with a body
		if (options.body !== undefined) {
			headers['Content-Type'] = 'application/msgpack';
		}

		if (sessionId) {
			headers['x-session-id'] = sessionId;
		}

		let body: Uint8Array | undefined;
		if (options.body !== undefined) {
			body = encode(options.body);
		}

		try {
			console.log(`🌐 API Request: ${options.method || 'GET'} ${url}`);
			
			const response = await fetch(url, {
				...options,
				headers: {
					...headers,
					...options.headers,
				},
				body,
			});

			console.log(`📡 Response: ${response.status} ${response.statusText}`);

			if (!response.ok) {
				// Even error responses should be MessagePack
				try {
					const arrayBuffer = await response.arrayBuffer();
					const errorData = decode(new Uint8Array(arrayBuffer)) as { error: string };
					throw new Error(errorData.error || `HTTP ${response.status}`);
				} catch {
					// Last resort fallback if error response isn't MessagePack
					const text = await response.text();
					console.warn('⚠️ Error response not in MessagePack format:', text);
					throw new Error(`HTTP ${response.status}: ${text}`);
				}
			}

			// Always expect MessagePack responses
			const arrayBuffer = await response.arrayBuffer();
			const data = decode(new Uint8Array(arrayBuffer)) as T;

			console.log(`✅ API Success:`, data);
			return data;
		} catch (error) {
			console.error('❌ API request failed:', { url, error });
			throw error;
		}
	}

	// Auth methods
	async verifySession(sessionId: string) {
		return this.request('/auth/verify', {}, sessionId);
	}

	async getSession(sessionId: string) {
		return this.request(`/auth/session/${sessionId}`);
	}

	async logout(sessionId: string) {
		return this.request('/auth/logout', { method: 'DELETE' }, sessionId);
	}

	// User methods
	async getUsers(sessionId?: string) {
		return this.request('/users', {}, sessionId);
	}

	async getUser(id: string, sessionId?: string) {
		return this.request(`/users/${id}`, {}, sessionId);
	}

	async createUser(userData: { username: string; password: string; age?: number }) {
		return this.request('/users', {
			method: 'POST',
			body: userData,
		});
	}

	async updateUser(id: string, userData: { username?: string; age?: number }, sessionId: string) {
		return this.request(
			`/users/${id}`,
			{
				method: 'PUT',
				body: userData,
			},
			sessionId
		);
	}

	async deleteUser(id: string, sessionId: string) {
		return this.request(`/users/${id}`, { method: 'DELETE' }, sessionId);
	}

	async getCurrentUser(sessionId: string) {
		return this.request('/users/me/profile', {}, sessionId);
	}

	// Restaurant methods
	async getRestaurants() {
		return this.request('/restaurants');
	}

	async getRestaurant(id: string) {
		return this.request(`/restaurants/${id}`);
	}

	async createRestaurant(restaurantData: CreateRestaurantRequest) {
		return this.request('/restaurants', {
			method: 'POST',
			body: restaurantData,
		});
	}

	async deleteRestaurant(id: string) {
		return this.request(`/restaurants/${id}`, { method: 'DELETE' });
	}

	// Health check (not under /api prefix) - Uses JSON for monitoring compatibility
	async healthCheck() {
		const healthUrl = this.baseUrl.replace('/api', '/health');
		const headers: HeadersInit = {
			'Accept': 'application/json', // Health check uses JSON for monitoring systems
		};

		try {
			console.log(`🌐 Health Check: GET ${healthUrl}`);
			
			const response = await fetch(healthUrl, { headers });

			console.log(`📡 Response: ${response.status} ${response.statusText}`);

			if (!response.ok) {
				throw new Error(`HTTP ${response.status}`);
			}

			const data = await response.json(); // Parse as JSON for health checks

			console.log(`✅ Health Check Success:`, data);
			return data;
		} catch (error) {
			console.error('❌ Health check failed:', { url: healthUrl, error });
			throw error;
		}
	}
}

export const apiClient = new ApiClient();
export default apiClient;
