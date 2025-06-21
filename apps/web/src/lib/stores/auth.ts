import { writable, type Writable } from 'svelte/store';
import { browser } from '$app/environment';
import type { UserResponse } from '@resto-rate/constants';

interface AuthState {
	user: UserResponse | null;
	sessionId: string | null;
	isLoading: boolean;
	isAuthenticated: boolean;
}

interface AuthStore extends Writable<AuthState> {
	login: (user: UserResponse, sessionId: string) => void;
	logout: () => void;
	setLoading: (loading: boolean) => void;
	updateUser: (user: UserResponse) => void;
}

function createAuthStore(): AuthStore {
	// Initialize from localStorage if available
	const storedSessionId = browser ? localStorage.getItem('sessionId') : null;
	const storedUser = browser ? localStorage.getItem('user') : null;
	
	const initialState: AuthState = {
		user: storedUser ? JSON.parse(storedUser) : null,
		sessionId: storedSessionId,
		isLoading: false,
		isAuthenticated: !!storedSessionId && !!storedUser,
	};

	const { subscribe, set, update }: Writable<AuthState> = writable(initialState);

	return {
		subscribe,
		set,
		update,
		// Custom methods for auth operations
		login: (user: UserResponse, sessionId: string) => {
			if (browser) {
				localStorage.setItem('sessionId', sessionId);
				localStorage.setItem('user', JSON.stringify(user));
			}
			set({
				user,
				sessionId,
				isLoading: false,
				isAuthenticated: true,
			});
		},
		logout: () => {
			if (browser) {
				localStorage.removeItem('sessionId');
				localStorage.removeItem('user');
			}
			set({
				user: null,
				sessionId: null,
				isLoading: false,
				isAuthenticated: false,
			});
		},
		setLoading: (loading: boolean) => {
			update(state => ({ ...state, isLoading: loading }));
		},
		updateUser: (user: UserResponse) => {
			if (browser) {
				localStorage.setItem('user', JSON.stringify(user));
			}
			update(state => ({ ...state, user }));
		},
	};
}

export const authStore = createAuthStore(); 