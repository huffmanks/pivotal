<script lang="ts">
  import { PUBLIC_BASE_URL } from "$env/static/public";
  import { onMount } from "svelte";

  type Link = {
    id: number;
    slug: string;
    destination_url: string;
    is_custom: boolean;
    click_count: number;
    created_at: Date;
    expires_at?: Date;
  };

  let links = $state<Link[]>([]);
  let loading = $state(true);
  let error = $state<string | null>(null);

  onMount(async () => {
    try {
      const res = await fetch("/api/links");
      if (!res.ok) throw new Error(`HTTP error! status: ${res.status}`);
      links = (await res.json()) as Link[];
    } catch (e) {
      error = e instanceof Error ? e.message : "An unknown error occurred";
    } finally {
      loading = false;
    }
  });
</script>

{#if loading}
  <p class="text-sm text-neutral-500">Loading...</p>
{:else if error}
  <p class="text-sm text-red-500">Error: {error}</p>
{:else}
  <h1 class="mb-1 text-2xl font-bold tracking-tight text-neutral-900 dark:text-neutral-100">
    Links
  </h1>
  <p class="mb-6 text-sm text-neutral-500 dark:text-neutral-400">All my little links my guy.</p>

  <div class="w-full overflow-x-auto rounded-lg border border-neutral-200 dark:border-neutral-800">
    <table class="w-full text-left text-sm text-neutral-600 dark:text-neutral-400">
      <thead
        class="border-b border-neutral-200 bg-neutral-50/50 text-xs font-semibold tracking-wider text-neutral-500 uppercase dark:border-neutral-800 dark:bg-neutral-900/50 dark:text-neutral-400">
        <tr>
          <th scope="col" class="px-4 py-3">ID</th>
          <th scope="col" class="px-4 py-3">Slug</th>
          <th scope="col" class="px-4 py-3">Destination URL</th>
          <th scope="col" class="px-4 py-3 text-right">Clicks</th>
          <th scope="col" class="px-4 py-3">Created</th>
          <th scope="col" class="px-4 py-3">Expires</th>
        </tr>
      </thead>
      <tbody class="divide-y divide-neutral-200 dark:divide-neutral-800">
        {#each links as link (link.id)}
          <tr class="hover:bg-neutral-50/50 dark:hover:bg-neutral-900/50">
            <td class="px-4 py-3.5 font-mono text-xs whitespace-nowrap text-neutral-400">
              {link.id}
            </td>
            <td
              class="px-4 py-3.5 font-medium whitespace-nowrap text-neutral-900 dark:text-neutral-100">
              <a
                href={`${PUBLIC_BASE_URL}/${link.slug}`}
                target="_blank"
                rel="noreferrer"
                class="block truncate text-primary hover:text-primary/80 dark:text-primary dark:hover:text-primary/80">
                {link.slug}
              </a>
            </td>
            <td class="max-w-xs px-4 py-3.5 md:max-w-md">
              <a
                href={link.destination_url}
                target="_blank"
                rel="noreferrer"
                class="block truncate text-primary hover:text-primary/80 dark:text-primary dark:hover:text-primary/80">
                {link.destination_url}
              </a>
            </td>
            <td
              class="px-4 py-3.5 text-right font-mono whitespace-nowrap text-neutral-700 dark:text-neutral-300">
              {link.click_count}
            </td>
            <td class="px-4 py-3.5 text-xs whitespace-nowrap text-neutral-500">
              {link.created_at}
            </td>
            <td class="px-4 py-3.5 text-xs whitespace-nowrap text-neutral-500">
              {link.expires_at ?? "—"}
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
{/if}
