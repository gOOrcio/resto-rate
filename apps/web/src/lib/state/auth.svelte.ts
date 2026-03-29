import type { UserProto } from '$lib/client/generated/users/v1/user_pb';

let currentUser = $state<UserProto | null>(null);
let authLoading = $state(true);

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
	setUser(u: UserProto | null) {
		currentUser = u;
	},
	setLoaded() {
		authLoading = false;
	}
};
