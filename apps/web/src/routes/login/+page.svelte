<div class="container mx-auto max-w-md p-6">
	<h1 class="mb-6 text-3xl font-bold text-center">Login</h1>

	<div class="bg-white rounded-lg shadow p-6">
		<p class="text-gray-600 text-center mb-6">This is a placeholder for the login functionality.</p>

		<form class="space-y-4">
			<div>
				<label for="username" class="block text-sm font-medium text-gray-700 mb-1">
					Username
				</label>
				<input
					type="text"
					id="username"
					class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
					placeholder="Enter your username"
				/>
			</div>

			<div>
				<label for="password" class="block text-sm font-medium text-gray-700 mb-1">
					Password
				</label>
				<input
					type="password"
					id="password"
					class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
					placeholder="Enter your password"
				/>
			</div>

			<button
				type="button"
				class="w-full bg-blue-500 text-white py-2 px-4 rounded-md hover:bg-blue-700 transition disabled:opacity-50"
				disabled
			>
				Login (Coming Soon)
			</button>
		</form>

		<div class="mt-6 text-center">
			<a href="/" class="text-blue-500 hover:text-blue-700">← Back to Home</a>
		</div>
	</div>
</div>
