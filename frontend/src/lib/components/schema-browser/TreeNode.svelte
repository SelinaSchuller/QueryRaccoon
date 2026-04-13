<script lang="ts">
  import TreeNode from './TreeNode.svelte'
  import { schemaStore } from '$lib/stores/schema.svelte'
  import { colors } from '$lib/colors'
  import type { TreeNode as TreeNodeData } from '$lib/stores/schema.svelte'

  type Props = {
    node: TreeNodeData
    connectionId: string
    depth?: number
    onTableClick?: (schemaName: string, name: string, _type: 'table' | 'view') => void
    onRefreshNode?: (node: TreeNodeData) => void
  }
  let { node, connectionId, depth = 0, onTableClick, onRefreshNode }: Props = $props()

  let hovered = $state(false)

  const icons: Record<string, string> = {
    database: '⊞',
    schema: '◫',
    folder: '▤',
    table: '▤',
    view: '◱',
    column: '◦',
  }

  const typeColors: Record<string, string> = {
    database: '#60a5fa',
    schema: '#a78bfa',
    folder: '#a1a1aa',
    table: '#34d399',
    view: '#fb923c',
    column: '#a1a1aa',
  }

  async function toggle() {
    if (node.type === 'column') return
    if ((node.type === 'table' || node.type === 'view') && onTableClick) {
      onTableClick(node.schemaName!, node.label, node.type)
      return
    }
    await schemaStore.expand(connectionId, node)
  }

  function handleRefresh(e: MouseEvent) {
    e.stopPropagation()
    onRefreshNode?.(node)
  }
</script>

<div>
  <div
    role="presentation"
    class="flex items-center rounded transition-colors"
    style="background-color: {hovered ? colors.background.tertiary : 'transparent'}"
    onmouseenter={() => hovered = true}
    onmouseleave={() => hovered = false}
  >
    <button
      onclick={toggle}
      class="flex-1 flex items-center gap-1.5 py-0.5 text-sm cursor-pointer text-left min-w-0"
      style="padding-left: {8 + depth * 12}px"
    >
      {#if node.type !== 'column'}
        <span class="text-[10px] shrink-0" style="color: {colors.text.muted}">
          {node.expanded ? '▾' : '▸'}
        </span>
      {:else}
        <span class="w-3 shrink-0"></span>
      {/if}
      <span class="shrink-0" style="color: {typeColors[node.type]}">{icons[node.type]}</span>
      <span class="truncate" style="color: {node.type === 'column' ? colors.text.muted : colors.text.primary}">
        {node.label}
      </span>
      {#if node.meta}
        <span class="ml-auto shrink-0 text-[11px] pl-2" style="color: {colors.text.muted}">{node.meta.Type}</span>
      {/if}
    </button>

    {#if node.type !== 'column' && onRefreshNode}
      <button
        type="button"
        class="shrink-0 text-[11px] px-1.5 py-0.5 rounded cursor-pointer transition-opacity"
        style="color: {colors.text.muted}; opacity: {hovered ? 1 : 0}"
        onclick={handleRefresh}
        title="Refresh"
      >↺</button>
    {/if}
  </div>

  {#if node.expanded && node.children.length > 0}
    {#each node.children as child (child.id)}
      <TreeNode node={child} {connectionId} depth={depth + 1} {onTableClick} {onRefreshNode} />
    {/each}
  {/if}
</div>
