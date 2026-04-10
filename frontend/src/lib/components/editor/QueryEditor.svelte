<script lang="ts">
  import { tabStore } from "$lib/stores/tabs.svelte";
  import { connectionStore } from "$lib/stores/connections.svelte";
  import { colors } from "$lib/colors";

  let tab = $derived(tabStore.active);
  let activeConn = $derived(connectionStore.active);

  function onKeydown(e: KeyboardEvent) {
    if ((e.metaKey || e.ctrlKey) && e.key === 'Enter') {
      e.preventDefault();
      if (tab) tabStore.execute(tab.id);
    }
  }
</script>

<div class="flex flex-col h-full">
  <div class="flex items-center gap-2 px-3 py-2" style="background-color: {colors.background.secondary}; border-bottom: 1px solid {colors.border.primary}">
    <button
      onclick={() => tab && tabStore.execute(tab.id)}
      disabled={!tab?.connectionId || tab?.isExecuting}
      class="flex items-center gap-1.5 px-3 py-1.5 rounded text-xs font-medium transition-colors disabled:opacity-40 disabled:cursor-not-allowed cursor-pointer"
      style="background-color: {colors.accent.primary}; color: #fff"
      onmouseenter={e => { if (tab?.connectionId && !tab?.isExecuting) (e.currentTarget as HTMLElement).style.backgroundColor = colors.accent.hover }}
      onmouseleave={e => (e.currentTarget as HTMLElement).style.backgroundColor = colors.accent.primary}
    >
      {#if tab?.isExecuting}
        <span class="animate-spin">⟳</span> Running…
      {:else}
        ▶ Run
      {/if}
    </button>

    <span style="color: {colors.border.primary}">|</span>

    {#if activeConn}
      <span class="text-xs" style="color: {colors.accent.primary}">● {activeConn.name}</span>
    {:else}
      <span class="text-xs" style="color: {colors.text.muted}">No connection selected</span>
    {/if}

    <span class="ml-auto text-[10px]" style="color: {colors.text.muted}">⌘ + Enter to run</span>
  </div>

  <textarea
    class="flex-1 w-full font-mono text-sm p-4 resize-none outline-none leading-relaxed"
    style="background-color: {colors.background.primary}; color: {colors.text.primary}"
    placeholder="SELECT * FROM ..."
    value={tab?.sql ?? ''}
    oninput={(e) => tab && tabStore.updateSQL(tab.id, (e.target as HTMLTextAreaElement).value)}
    onkeydown={onKeydown}
    spellcheck={false}
  ></textarea>
</div>
