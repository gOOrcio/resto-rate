import type { UserProto } from '$lib/client/generated/users/v1/user_pb';

let currentUser = $state<UserProto | null>(null);

export const auth = {
	get user() {
		return currentUser;
	},
	get isLoggedIn() {
		return currentUser !== null;
	},
	setUser(u: UserProto | null) {
		currentUser = u;
	}
};
