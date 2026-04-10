<script lang="ts">
  import { tabStore } from "$lib/stores/tabs.svelte";
  import { colors } from "$lib/colors";
</script>

<div class="flex items-end shrink-0 overflow-x-auto" style="background-color: {colors.background.secondary}; border-bottom: 1px solid {colors.border.primary}">
  {#each tabStore.list as tab (tab.id)}
    <div
      role="tab"
      tabindex="0"
      aria-selected={tabStore.activeId === tab.id}
      class="flex items-center gap-2 px-4 py-2 cursor-pointer shrink-0 group transition-colors"
      style="background-color: {tabStore.activeId === tab.id ? colors.background.primary : 'transparent'}; color: {tabStore.activeId === tab.id ? colors.text.primary : colors.text.muted}; border-right: 1px solid {colors.border.primary}"
      onclick={() => tabStore.setActive(tab.id)}
      onkeydown={e => e.key === 'Enter' && tabStore.setActive(tab.id)}
      onmouseenter={e => { if (tabStore.activeId !== tab.id) (e.currentTarget as HTMLElement).style.color = colors.text.primary }}
      onmouseleave={e => { if (tabStore.activeId !== tab.id) (e.currentTarget as HTMLElement).style.color = colors.text.muted }}
    >
      <span class="text-xs">{tab.name}</span>
      {#if tabStore.list.length > 1}
        <span
          role="button"
          tabindex="0"
          onclick={(e) => { e.stopPropagation(); tabStore.close(tab.id) }}
          onkeydown={e => { if (e.key === 'Enter') { e.stopPropagation(); tabStore.close(tab.id) } }}
          class="opacity-0 group-hover:opacity-100 text-xs w-4 h-4 flex items-center justify-center transition-all cursor-pointer"
          style="color: {colors.text.muted}"
          onmouseenter={e => (e.currentTarget as HTMLElement).style.color = colors.text.primary}
          onmouseleave={e => (e.currentTarget as HTMLElement).style.color = colors.text.muted}
        >✕</span>
      {/if}
    </div>
  {/each}

  <button
    onclick={() => tabStore.add()}
    class="px-3 py-2 text-lg leading-none shrink-0 transition-colors cursor-pointer"
    style="color: {colors.text.muted}"
    onmouseenter={e => (e.currentTarget as HTMLElement).style.color = colors.accent.primary}
    onmouseleave={e => (e.currentTarget as HTMLElement).style.color = colors.text.muted}
    title="New Tab"
  >+</button>
</div>
