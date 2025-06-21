<script lang="ts">
	import {
		Card,
		CardContent,
		CardDescription,
		CardHeader,
		CardTitle,
	} from '$lib/components/ui/card';
	import { Button } from '$lib/components/ui/button';
</script>

<div class="container mx-auto max-w-4xl p-6">
	<h1 class="mb-6 text-3xl font-bold tracking-tight">Users</h1>

	<Card>
		<CardHeader>
			<CardTitle>User Management</CardTitle>
			<CardDescription>Manage user accounts and permissions</CardDescription>
		</CardHeader>
		<CardContent>
			<div class="space-y-4">
				<h2 class="text-xl font-semibold">Features to be implemented:</h2>
				<ul class="list-disc list-inside space-y-2 text-muted-foreground">
					<li>View all registered users</li>
					<li>User profiles and details</li>
					<li>User management features</li>
					<li>Search and filter users</li>
				</ul>
			</div>

			<div class="mt-6">
				<Button variant="link" href="/" class="p-0">← Back to Home</Button>
			</div>
		</CardContent>
	</Card>
</div>
