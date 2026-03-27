<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { auth } from '$lib/state/auth.svelte';
	import client from '$lib/client/client';
	import type { FriendRequestProto, FriendProto } from '$lib/client/generated/friendship/v1/friendship_pb';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';

	let friends = $state<FriendProto[]>([]);
	let pendingRequests = $state<FriendRequestProto[]>([]);
	let loading = $state(true);

	let addEmail = $state('');
	let addLoading = $state(false);
	let addError = $state('');
	let addSuccess = $state('');

	let accepting = $state<Set<string>>(new Set());
	let declining = $state<Set<string>>(new Set());
	let removing = $state<Set<string>>(new Set());

	async function loadData() {
		try {
			const [friendsRes, pendingRes] = await Promise.all([
				client.friendship.listFriends({}),
				client.friendship.listPendingRequests({}),
			]);
			friends = friendsRes.friends;
			pendingRequests = pendingRes.requests;
		} catch (e) {
			console.error('Failed to load friends:', e);
		} finally {
			loading = false;
		}
	}

	async function sendRequest() {
		if (!addEmail.trim()) return;
		addLoading = true;
		addError = '';
		addSuccess = '';
		try {
			await client.friendship.sendFriendRequest({ receiverEmail: addEmail.trim() });
			addSuccess = `Friend request sent to ${addEmail.trim()}`;
			addEmail = '';
		} catch (e: unknown) {
			addError = (e as Error).message || 'Failed to send request';
		} finally {
			addLoading = false;
		}
	}

	async function accept(requestId: string) {
		accepting = new Set([...accepting, requestId]);
		try {
			await client.friendship.acceptFriendRequest({ requestId });
			pendingRequests = pendingRequests.filter((r) => r.id !== requestId);
			await loadData();
		} catch (e) {
			console.error('Failed to accept request:', e);
		} finally {
			accepting.delete(requestId);
			accepting = new Set(accepting);
		}
	}

	async function decline(requestId: string) {
		const removed = pendingRequests.find((r) => r.id === requestId);
		if (!removed) return;
		declining = new Set([...declining, requestId]);
		pendingRequests = pendingRequests.filter((r) => r.id !== requestId);
		try {
			await client.friendship.declineFriendRequest({ requestId });
		} catch (e) {
			console.error('Failed to decline request:', e);
			pendingRequests = [...pendingRequests, removed];
		} finally {
			declining.delete(requestId);
			declining = new Set(declining);
		}
	}

	async function removeFriend(friendUserId: string) {
		const removed = friends.find((f) => f.userId === friendUserId);
		if (!removed) return;
		removing = new Set([...removing, friendUserId]);
		friends = friends.filter((f) => f.userId !== friendUserId);
		try {
			await client.friendship.removeFriend({ friendUserId });
		} catch (e) {
			console.error('Failed to remove friend:', e);
			friends = [...friends, removed];
		} finally {
			removing.delete(friendUserId);
			removing = new Set(removing);
		}
	}

	onMount(() => {
		if (!auth.isLoggedIn) {
			goto('/?login=1');
			return;
		}
		loadData();
	});
</script>

<div class="container mx-auto max-w-3xl space-y-8 p-6">
	<h2 class="text-2xl font-semibold text-blue-800">Friends</h2>

	{#if loading}
		<div class="flex items-center gap-2 text-sm text-gray-500">
			<div class="h-4 w-4 animate-spin rounded-full border-2 border-gray-300 border-t-blue-500"></div>
			Loading…
		</div>
	{:else}
		<!-- Add Friend -->
		<section>
			<h3 class="mb-3 text-lg font-medium text-gray-800">Add a friend</h3>
			<form
				class="flex gap-2"
				onsubmit={(e) => {
					e.preventDefault();
					sendRequest();
				}}
			>
				<Input
					type="email"
					placeholder="Their email address"
					bind:value={addEmail}
					class="flex-1"
					disabled={addLoading}
				/>
				<Button type="submit" disabled={addLoading || !addEmail.trim()}>
					{addLoading ? 'Sending…' : 'Send Request'}
				</Button>
			</form>
			{#if addError}
				<p class="mt-2 text-sm text-red-600">{addError}</p>
			{/if}
			{#if addSuccess}
				<p class="mt-2 text-sm text-green-600">{addSuccess}</p>
			{/if}
		</section>

		<!-- Pending Requests -->
		{#if pendingRequests.length > 0}
			<section>
				<h3 class="mb-3 text-lg font-medium text-gray-800">
					Pending requests <span class="ml-1 rounded-full bg-blue-100 px-2 py-0.5 text-xs font-semibold text-blue-700">{pendingRequests.length}</span>
				</h3>
				<ul class="space-y-2">
					{#each pendingRequests as req (req.id)}
						<li class="flex items-center justify-between rounded-lg border border-gray-200 bg-white p-3 shadow-sm">
							<div>
								<p class="font-medium text-gray-900">{req.senderName}</p>
								<p class="text-sm text-gray-500">{req.senderEmail}</p>
							</div>
							<div class="flex gap-2">
								<Button
									size="sm"
									disabled={accepting.has(req.id)}
									onclick={() => accept(req.id)}
								>
									{accepting.has(req.id) ? 'Accepting…' : 'Accept'}
								</Button>
								<Button
									variant="outline"
									size="sm"
									disabled={declining.has(req.id)}
									onclick={() => decline(req.id)}
									class="text-red-600 hover:border-red-300 hover:text-red-700"
								>
									{declining.has(req.id) ? 'Declining…' : 'Decline'}
								</Button>
							</div>
						</li>
					{/each}
				</ul>
			</section>
		{/if}

		<!-- Friends List -->
		<section>
			<h3 class="mb-3 text-lg font-medium text-gray-800">My Friends ({friends.length})</h3>
			{#if friends.length === 0}
				<p class="text-sm text-gray-500">No friends yet. Send a request above to get started.</p>
			{:else}
				<ul class="space-y-2">
					{#each friends as friend (friend.userId)}
						<li class="flex items-center justify-between rounded-lg border border-gray-200 bg-white p-3 shadow-sm">
							<div>
								<p class="font-medium text-gray-900">{friend.name}</p>
								<p class="text-sm text-gray-500">{friend.email}</p>
							</div>
							<div class="flex gap-2">
								<Button
									variant="outline"
									size="sm"
									href="/friends/{friend.userId}"
								>
									View Profile
								</Button>
								<Button
									variant="outline"
									size="sm"
									disabled={removing.has(friend.userId)}
									onclick={() => removeFriend(friend.userId)}
									class="text-red-600 hover:border-red-300 hover:text-red-700"
								>
									{removing.has(friend.userId) ? 'Removing…' : 'Remove'}
								</Button>
							</div>
						</li>
					{/each}
				</ul>
			{/if}
		</section>
	{/if}
</div>
