<script lang="ts">
  import { tabStore } from "$lib/stores/tabs.svelte";
  import { connectionStore } from "$lib/stores/connections.svelte";
  import { colors } from "$lib/colors";

  type Props = { collapsed?: boolean; onToggleCollapse?: () => void }
  let { collapsed = false, onToggleCollapse }: Props = $props()

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
  <div class="flex items-center gap-2 px-3 py-2 shrink-0" style="background-color: {colors.background.secondary}; border-bottom: 1px solid {colors.border.primary}">
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

    <button
      onclick={onToggleCollapse}
      class="w-5 h-5 flex items-center justify-center rounded text-xs transition-colors cursor-pointer"
      style="color: {colors.text.muted}"
      onmouseenter={e => (e.currentTarget as HTMLElement).style.color = colors.text.primary}
      onmouseleave={e => (e.currentTarget as HTMLElement).style.color = colors.text.muted}
      title={collapsed ? 'Expand editor' : 'Collapse editor'}
    >{collapsed ? '▾' : '▴'}</button>
  </div>

  <textarea
    class="flex-1 w-full font-mono text-sm p-4 resize-none outline-none leading-relaxed"
    style="background-color: {colors.background.primary}; color: {colors.text.primary}; display: {collapsed ? 'none' : 'block'}"
    placeholder="SELECT * FROM ..."
    value={tab?.sql ?? ''}
    oninput={(e) => tab && tabStore.updateSQL(tab.id, (e.target as HTMLTextAreaElement).value)}
    onkeydown={onKeydown}
    spellcheck={false}
  ></textarea>
</div>
