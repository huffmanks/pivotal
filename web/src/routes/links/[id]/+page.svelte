<script lang="ts">
  import { page } from "$app/state";
  import { onMount } from "svelte";
  import type { LinkClicks } from "../../../types";

  let links = $state<LinkClicks[]>([]);
  let loading = $state(true);
  let error = $state<string | null>(null);

  onMount(async () => {
    try {
      const res = await fetch(`/api/links/${page.params.id}/clicks`);
      if (!res.ok) throw new Error(`HTTP error! status: ${res.status}`);
      links = (await res.json()) as LinkClicks[];
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
    Link click events
  </h1>
  <p class="mb-6 text-sm text-neutral-500 dark:text-neutral-400">All my little clicks my guy.</p>

  <div class="w-full overflow-x-auto rounded-lg border border-neutral-200 dark:border-neutral-800">
    <table class="w-full text-left text-sm text-neutral-600 dark:text-neutral-400">
      <thead
        class="border-b border-neutral-200 bg-neutral-50/50 text-xs font-semibold tracking-wider text-neutral-500 uppercase dark:border-neutral-800 dark:bg-neutral-900/50 dark:text-neutral-400">
        <tr>
          <th scope="col" class="px-4 py-3">ID</th>
          <th scope="col" class="px-4 py-3">Link ID</th>
          <th scope="col" class="px-4 py-3">Referer</th>
          <th scope="col" class="px-4 py-3">User agent</th>
          <th scope="col" class="px-4 py-3">Browser</th>
          <th scope="col" class="px-4 py-3">OS</th>
          <th scope="col" class="px-4 py-3">Device</th>
          <th scope="col" class="px-4 py-3">Country</th>
          <th scope="col" class="px-4 py-3">Region</th>
          <th scope="col" class="px-4 py-3">City</th>
          <th scope="col" class="px-4 py-3">UTM params</th>
          <th scope="col" class="px-4 py-3">QR scan</th>
          <th scope="col" class="px-4 py-3">Date</th>
        </tr>
      </thead>
      <tbody class="divide-y divide-neutral-200 dark:divide-neutral-800">
        {#each links as link (link.id)}
          <tr class="hover:bg-neutral-50/50 dark:hover:bg-neutral-900/50">
            <td class="px-4 py-3.5 font-mono text-xs whitespace-nowrap text-neutral-400">
              {link.id}
            </td>
            <td class="px-4 py-3.5 font-mono text-xs whitespace-nowrap text-neutral-400">
              {link.link_id}
            </td>
            <td
              class="px-4 py-3.5 font-mono whitespace-nowrap text-neutral-700 dark:text-neutral-300">
              {#if !link.referer}
                <span class="text-neutral-500">Empty</span>
              {:else}
                {link.referer}
              {/if}
            </td>
            <td
              class="px-4 py-3.5 font-mono whitespace-nowrap text-neutral-700 dark:text-neutral-300">
              {link.user_agent}
            </td>
            <td
              class="px-4 py-3.5 font-mono whitespace-nowrap text-neutral-700 dark:text-neutral-300">
              {link.browser}
            </td>
            <td
              class="px-4 py-3.5 font-mono whitespace-nowrap text-neutral-700 dark:text-neutral-300">
              {link.os}
            </td>
            <td
              class="px-4 py-3.5 font-mono whitespace-nowrap text-neutral-700 dark:text-neutral-300">
              {link.device}
            </td>
            <td
              class="px-4 py-3.5 font-mono whitespace-nowrap text-neutral-700 dark:text-neutral-300">
              {link.country}
            </td>
            <td
              class="px-4 py-3.5 font-mono whitespace-nowrap text-neutral-700 dark:text-neutral-300">
              {link.region}
            </td>
            <td
              class="px-4 py-3.5 font-mono whitespace-nowrap text-neutral-700 dark:text-neutral-300">
              {link.city}
            </td>
            <td
              class="px-4 py-3.5 font-mono whitespace-nowrap text-neutral-700 dark:text-neutral-300">
              {#if Object.keys(link.utm_params ?? {}).length === 0}
                <span class="text-neutral-500">Empty</span>
              {:else}
                {Object.entries(link.utm_params)
                  .map(([key, val]) => `${key}: ${val}`)
                  .join(", ")}
              {/if}
            </td>
            <td
              class="px-4 py-3.5 font-mono whitespace-nowrap text-neutral-700 dark:text-neutral-300">
              {link.qr_scan}
            </td>

            <td class="px-4 py-3.5 text-xs whitespace-nowrap text-neutral-500">
              {link.clicked_at}
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
{/if}
