<script lang="ts">
  import { colors } from '$lib/colors'

  type MenuItem = {
    label: string
    action: () => void
    danger?: boolean
  }

  type Props = {
    x: number
    y: number
    items: MenuItem[]
    onclose: () => void
  }

  let { x, y, items, onclose }: Props = $props()

  function onKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') onclose()
  }
</script>

<svelte:window onkeydown={onKeydown} />

<div class="fixed inset-0 z-40" onclick={onclose} role="presentation"></div>

<div
  class="fixed z-50 rounded-md shadow-xl py-1 min-w-[160px]"
  style="left: {x}px; top: {y}px; background-color: {colors.background.tertiary}; border: 1px solid {colors.border.primary}"
>
  {#each items as item}
    <button
      onclick={() => { item.action(); onclose() }}
      class="w-full text-left px-3 py-1.5 text-xs cursor-pointer transition-colors"
      style="color: {item.danger ? '#f87171' : colors.text.primary}"
      onmouseenter={e => (e.currentTarget as HTMLElement).style.backgroundColor = colors.background.secondary}
      onmouseleave={e => (e.currentTarget as HTMLElement).style.backgroundColor = 'transparent'}
    >
      {item.label}
    </button>
  {/each}
</div>
