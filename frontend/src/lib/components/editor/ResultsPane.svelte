<script lang="ts">
  import { tabStore } from "$lib/stores/tabs.svelte";
  import { colors } from "$lib/colors";

  let tab = $derived(tabStore.active);
</script>

<div class="flex flex-col h-full" style="background-color: {colors.background.primary}">
  <div class="flex items-center gap-3 px-3 py-2" style="background-color: {colors.background.secondary}; border-bottom: 1px solid {colors.border.primary}">
    <span class="text-xs font-semibold uppercase tracking-wider" style="color: {colors.text.muted}">Results</span>
    {#if tab?.result}
      <span class="text-xs" style="color: {colors.text.muted}">{tab.result.Rows.length} row{tab.result.Rows.length !== 1 ? 's' : ''}</span>
    {/if}
    {#if tab?.executionTime != null}
      <span class="text-xs" style="color: {colors.text.muted}">· {tab.executionTime}ms</span>
    {/if}
  </div>

  <div class="flex-1 overflow-auto">
    {#if tab?.isExecuting}
      <div class="flex items-center justify-center h-full gap-2 text-sm" style="color: {colors.text.muted}">
        <span class="animate-spin">⟳</span> Executing…
      </div>

    {:else if tab?.error}
      <div class="p-4">
        <p class="text-sm font-mono whitespace-pre-wrap" style="color: #f87171">{tab.error}</p>
      </div>

    {:else if tab?.result}
      <table class="w-full text-xs border-collapse">
        <thead class="sticky top-0" style="background-color: {colors.background.secondary}">
          <tr>
            {#each tab.result.Columns as col}
              <th class="text-left px-3 py-2 font-medium whitespace-nowrap" style="color: {colors.text.muted}; border-bottom: 1px solid {colors.border.primary}">
                {col}
              </th>
            {/each}
          </tr>
        </thead>
        <tbody>
          {#each tab.result.Rows as row, i}
            <tr style="background-color: {i % 2 === 0 ? colors.background.primary : colors.background.secondary}">
              {#each row as cell}
                <td class="px-3 py-1.5 whitespace-nowrap max-w-xs truncate font-mono" style="color: {colors.text.primary}; border-bottom: 1px solid {colors.border.primary}">
                  {#if cell === null}
                    <span class="italic" style="color: {colors.text.muted}">NULL</span>
                  {:else}
                    {String(cell)}
                  {/if}
                </td>
              {/each}
            </tr>
          {/each}
        </tbody>
      </table>

    {:else}
      <div class="flex items-center justify-center h-full text-sm" style="color: {colors.text.muted}">
        Run a query to see results
      </div>
    {/if}
  </div>
</div>
