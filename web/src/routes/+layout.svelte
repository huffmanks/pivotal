<script lang="ts">
  import favicon from "$lib/assets/favicon.svg";
  import "./layout.css";

  let { children } = $props();
</script>

<svelte:head><link rel="icon" href={favicon} /></svelte:head>
<main class="flex min-h-dvh w-full flex-col items-center justify-center py-6 pl-3">
  {@render children()}
</main>
