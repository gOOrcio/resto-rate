<script lang="ts">
	import '../app.css';
	let { children } = $props();
</script>

<main class="justify-items-center pt-14">
	{@render children()}
</main>
