<script lang="ts">
  import { connectionStore } from "$lib/stores/connections.svelte";
  import { tabStore } from "$lib/stores/tabs.svelte";
  import { colors } from "$lib/colors";

  let conn = $derived(connectionStore.active);
  let tab = $derived(tabStore.active);
</script>

<footer class="flex items-center gap-4 px-4 h-7 shrink-0 text-[11px]" style="background-color: {colors.background.secondary}; border-top: 1px solid {colors.border.primary}; color: {colors.text.muted}">
  {#if conn?.connected}
    <span class="flex items-center gap-1.5">
      <span class="w-1.5 h-1.5 rounded-full" style="background-color: {colors.accent.primary}"></span>
      {conn.name} · {conn.config.DriverType}
    </span>
  {:else}
    <span class="flex items-center gap-1.5">
      <span class="w-1.5 h-1.5 rounded-full" style="background-color: {colors.border.primary}"></span>
      Not connected
    </span>
  {/if}

  {#if tab?.result}
    <span class="ml-auto">{tab.result.Rows.length} rows</span>
    {#if tab.executionTime != null}
      <span>{tab.executionTime}ms</span>
    {/if}
  {/if}
</footer>
