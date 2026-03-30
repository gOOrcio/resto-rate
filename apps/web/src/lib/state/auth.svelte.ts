import type { UserProto } from '$lib/client/generated/users/v1/user_pb';

let currentUser = $state<UserProto | null>(null);
let authLoading = $state(true);
let loginDialogOpen = $state(false);

export const auth = {
	get user() {
		return currentUser;
	},
	get isLoggedIn() {
		return currentUser !== null;
	},
	get loading() {
		return authLoading;
	},
	get loginOpen() {
		// Auto-close whenever the user becomes logged in, regardless of call path.
		return loginDialogOpen && currentUser === null;
	},
	setUser(u: UserProto | null) {
		currentUser = u;
	},
	setLoaded() {
		authLoading = false;
	},
	openLogin() {
		loginDialogOpen = true;
	},
	closeLogin() {
		loginDialogOpen = false;
	}
};
