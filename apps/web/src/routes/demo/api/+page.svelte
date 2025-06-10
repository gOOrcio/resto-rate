<script lang="ts">
	import { onMount } from 'svelte';
	import { apiClient } from '$lib/api';

	let healthStatus: any = null;
	let users: any = null;
	let error: string | null = null;
	let loading = false;

	async function testHealthCheck() {
		loading = true;
		error = null;
		try {
			healthStatus = await apiClient.healthCheck();
		} catch (err) {
			error = `Health check failed: ${err}`;
		} finally {
			loading = false;
		}
	}

	async function testGetUsers() {
		loading = true;
		error = null;
		try {
			users = await apiClient.getUsers();
		} catch (err) {
			error = `Get users failed: ${err}`;
		} finally {
			loading = false;
		}
	}

	async function testCreateUser() {
		loading = true;
		error = null;
		try {
			const userData = {
				username: `testuser_${Date.now()}`,
				password: 'testpassword123',
				age: 25,
			};
			const result = await apiClient.createUser(userData);
			// User created successfully, refresh users list
			await testGetUsers();
		} catch (err) {
			error = `Create user failed: ${err}`;
		} finally {
			loading = false;
		}
	}

	onMount(() => {
		testHealthCheck();
	});
</script>

<div class="container mx-auto max-w-4xl p-6">
	<h1 class="mb-6 text-3xl font-bold">API Communication Test</h1>

	<p class="mb-6 text-gray-600">
		This page demonstrates MessagePack communication between the SvelteKit frontend and Fastify API.
	</p>

	{#if error}
		<div class="mb-4 rounded border border-red-400 bg-red-100 px-4 py-3 text-red-700">
			{error}
		</div>
	{/if}

	<div class="space-y-6">
		<!-- Health Check -->
		<div class="rounded-lg bg-white p-6 shadow">
			<h2 class="mb-4 text-xl font-semibold">Health Check</h2>
			<button
				on:click={testHealthCheck}
				disabled={loading}
				class="rounded bg-blue-500 px-4 py-2 font-bold text-white hover:bg-blue-700 disabled:opacity-50"
			>
				{loading ? 'Loading...' : 'Test Health Check'}
			</button>

			{#if healthStatus}
				<pre class="mt-4 overflow-auto rounded bg-gray-100 p-4">{JSON.stringify(
						healthStatus,
						null,
						2
					)}</pre>
			{/if}
		</div>

		<!-- Users -->
		<div class="rounded-lg bg-white p-6 shadow">
			<h2 class="mb-4 text-xl font-semibold">Users API</h2>
			<div class="mb-4 space-x-2">
				<button
					on:click={testGetUsers}
					disabled={loading}
					class="rounded bg-green-500 px-4 py-2 font-bold text-white hover:bg-green-700 disabled:opacity-50"
				>
					{loading ? 'Loading...' : 'Get Users'}
				</button>

				<button
					on:click={testCreateUser}
					disabled={loading}
					class="rounded bg-purple-500 px-4 py-2 font-bold text-white hover:bg-purple-700 disabled:opacity-50"
				>
					{loading ? 'Loading...' : 'Create Test User'}
				</button>
			</div>

			{#if users}
				<pre class="overflow-auto rounded bg-gray-100 p-4">{JSON.stringify(users, null, 2)}</pre>
			{/if}
		</div>

		<!-- Instructions -->
		<div class="rounded-lg bg-blue-50 p-6">
			<h3 class="mb-2 text-lg font-semibold">Setup Instructions</h3>
			<ol class="list-inside list-decimal space-y-2 text-sm">
				<li>Make sure PostgreSQL is running and configured in your .env file</li>
				<li>
					Start the API server: <code class="rounded bg-gray-200 px-2 py-1"
						>cd apps/api && bun run dev</code
					>
				</li>
				<li>
					The API will run on <code class="rounded bg-gray-200 px-2 py-1"
						>http://localhost:3001</code
					>
				</li>
				<li>All communication uses MessagePack for efficient binary serialization</li>
				<li>Authentication shares the same session system as the web app</li>
			</ol>
		</div>
	</div>
</div>
