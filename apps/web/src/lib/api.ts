import { encode, decode } from '@msgpack/msgpack';
import { browser } from '$app/environment';
import { getWebConfig } from '@resto-rate/config';

const webConfig = getWebConfig();
const API_BASE_URL = browser ? webConfig.apiUrl : webConfig.apiUrl;

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

		const headers: Record<string, string> = {
			'Content-Type': 'application/msgpack',
		};

		// Merge existing headers
		if (options.headers) {
			Object.entries(options.headers).forEach(([key, value]) => {
				if (typeof value === 'string') {
					headers[key] = value;
				}
			});
		}

		// Add session ID if provided
		if (sessionId) {
			headers['X-Session-Id'] = sessionId;
		}

		// Encode body if it exists
		let body: BodyInit | undefined;
		if (
			options.body &&
			typeof options.body === 'object' &&
			!(options.body instanceof FormData) &&
			!(options.body instanceof URLSearchParams)
		) {
			body = encode(options.body);
		} else {
			body = options.body as BodyInit;
		}

		const response = await fetch(url, {
			...options,
			headers,
			body,
			credentials: 'include',
		});

		if (!response.ok) {
			const errorText = await response.text();
			throw new Error(`API Error: ${response.status} ${errorText}`);
		}

		const contentType = response.headers.get('content-type');
		if (contentType?.includes('application/msgpack')) {
			const buffer = await response.arrayBuffer();
			return decode(new Uint8Array(buffer)) as T;
		}

		return response.json();
	}

	// Auth endpoints
	async verifySession(sessionId: string) {
		return this.request('/api/auth/verify', { method: 'GET' }, sessionId);
	}

	async getSession(sessionId: string) {
		return this.request(`/api/auth/session/${sessionId}`, { method: 'GET' });
	}

	async logout(sessionId: string) {
		return this.request('/api/auth/logout', { method: 'DELETE' }, sessionId);
	}

	// User endpoints
	async getUsers(sessionId?: string) {
		return this.request('/api/users', { method: 'GET' }, sessionId);
	}

	async getUser(id: string, sessionId?: string) {
		return this.request(`/api/users/${id}`, { method: 'GET' }, sessionId);
	}

	async createUser(userData: { username: string; password: string; age?: number }) {
		return this.request('/api/users', {
			method: 'POST',
			body: userData,
		});
	}

	async updateUser(id: string, userData: { username?: string; age?: number }, sessionId: string) {
		return this.request(
			`/api/users/${id}`,
			{
				method: 'PUT',
				body: userData,
			},
			sessionId
		);
	}

	async deleteUser(id: string, sessionId: string) {
		return this.request(`/api/users/${id}`, { method: 'DELETE' }, sessionId);
	}

	async getCurrentUser(sessionId: string) {
		return this.request('/api/users/me/profile', { method: 'GET' }, sessionId);
	}

	// Restaurant endpoints
	async getRestaurants(sessionId?: string) {
		return this.request('/api/restaurants', { method: 'GET' }, sessionId);
	}

	async getRestaurant(id: string, sessionId?: string) {
		return this.request(`/api/restaurants/${id}`, { method: 'GET' }, sessionId);
	}

	async createRestaurant(
		restaurantData: {
			name: string;
			description?: string;
			cuisineType?: string;
			address?: string;
			latitude?: number;
			longitude?: number;
			phone?: string;
			website?: string;
			priceRange?: number;
			categoryIds?: string[];
		},
		sessionId: string
	) {
		return this.request(
			'/api/restaurants',
			{
				method: 'POST',
				body: restaurantData,
			},
			sessionId
		);
	}

	// Health check
	async healthCheck() {
		return this.request('/health', { method: 'GET' });
	}
}

export const apiClient = new ApiClient();
