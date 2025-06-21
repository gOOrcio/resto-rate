<script lang="ts">
	import {
		Card,
		CardContent,
		CardDescription,
		CardHeader,
		CardTitle,
	} from '$lib/components/ui/card';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { Button } from '$lib/components/ui/button';
</script>

<div class="container mx-auto max-w-md p-6">
	<h1 class="mb-6 text-3xl font-bold text-center tracking-tight">Login</h1>

	<Card>
		<CardHeader>
			<CardTitle>Welcome back</CardTitle>
			<CardDescription>Sign in to your account to continue</CardDescription>
		</CardHeader>
		<CardContent>
			<form class="space-y-4">
				<div class="space-y-2">
					<Label for="username">Username</Label>
					<Input type="text" id="username" placeholder="Enter your username" />
				</div>

				<div class="space-y-2">
					<Label for="password">Password</Label>
					<Input type="password" id="password" placeholder="Enter your password" />
				</div>

				<Button type="button" class="w-full" disabled>Login (Coming Soon)</Button>
			</form>

			<div class="mt-6 text-center">
				<Button variant="link" href="/" class="p-0">← Back to Home</Button>
			</div>
		</CardContent>
	</Card>
</div>
