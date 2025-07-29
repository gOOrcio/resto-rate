<script lang="ts">
	import '../app.css';
	let { children } = $props();
</script>


<main class="pt-14">
	{@render children()}
</main>
