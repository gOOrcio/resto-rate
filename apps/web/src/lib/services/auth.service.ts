import { authStore } from '$lib/stores/auth';
import { apiClient } from '$lib/api';
import { goto } from '$app/navigation';
import { browser } from '$app/environment';
import type { UserResponse } from '@resto-rate/constants';
import { get } from 'svelte/store';

export class AuthService {
	/**
	 * Initiate Google OAuth login
	 */
	async initiateGoogleLogin(): Promise<void> {
		if (!browser) return;
		
		try {
			authStore.setLoading(true);
			
			// Get Google OAuth URL from backend using apiClient
			const response = await apiClient.getGoogleAuthUrl();
			
			// Redirect to Google OAuth
			window.location.href = (response as { authUrl: string }).authUrl;
		} catch (error) {
			authStore.setLoading(false);
			throw new Error(`Failed to initiate Google login: ${error}`);
		}
	}

	/**
	 * Handle Google OAuth callback
	 */
	async handleGoogleCallback(code: string): Promise<void> {
		if (!browser) return;
		
		try {
			authStore.setLoading(true);
			
			// Exchange code for session via backend using apiClient
			const response = await apiClient.handleGoogleCallback(code);
			
			// Update auth store with the session
			const { user, sessionId } = response as { user: UserResponse; sessionId: string };
			authStore.login(user, sessionId);
		} catch (error) {
			authStore.setLoading(false);
			throw new Error(`Failed to complete Google authentication: ${error}`);
		}
	}

	/**
	 * Verify current session
	 */
	async verifySession(): Promise<void> {
		if (!browser) return;
		
		const state = get(authStore);
		const sessionId = state.sessionId;
		
		if (!sessionId) {
			authStore.logout();
			return;
		}

		try {
			authStore.setLoading(true);
			
			const response = await apiClient.verifySession(sessionId) as { user: UserResponse };
			
			// Update store with fresh user data
			authStore.updateUser(response.user);
		} catch (error) {
			// Session is invalid, logout user
			authStore.logout();
			throw new Error(`Session verification failed: ${error}`);
		} finally {
			authStore.setLoading(false);
		}
	}

	/**
	 * Logout user
	 */
	async logout(): Promise<void> {
		if (!browser) return;
		
		try {
			const state = get(authStore);
			const sessionId = state.sessionId;
			
			if (sessionId) {
				// Call backend to invalidate session
				await apiClient.logout(sessionId);
			}
		} catch (error) {
			// Continue with logout even if backend call fails
			console.warn('Failed to invalidate session on backend:', error);
		} finally {
			// Clear local state
			authStore.logout();
			
			// Redirect to home page
			goto('/');
		}
	}

	/**
	 * Refresh session if needed
	 */
	async refreshSession(): Promise<void> {
		// For now, just verify the current session
		// In the future, this could implement token refresh logic
		await this.verifySession();
	}

	/**
	 * Check if user is authenticated
	 */
	isAuthenticated(): boolean {
		const state = get(authStore);
		return state.isAuthenticated;
	}

	/**
	 * Get current user
	 */
	getCurrentUser(): UserResponse | null {
		const state = get(authStore);
		return state.user;
	}
}

export const authService = new AuthService(); 