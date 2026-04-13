<script lang="ts">
  import { connectionStore } from "$lib/stores/connections.svelte";
  import { tabStore } from "$lib/stores/tabs.svelte";
  import { schemaStore } from "$lib/stores/schema.svelte";
  import ConnectionDialog from "$lib/components/connections/ConnectionDialog.svelte";
  import TreeNode from "$lib/components/schema-browser/TreeNode.svelte";
  import { colors } from "$lib/colors";

  let showDialog = $state(false);

  async function selectConnection(id: string) {
    const tab = tabStore.active;
    if (tab) tabStore.setConnection(tab.id, id);
    connectionStore.activeId = id;

    const conn = connectionStore.list.find(c => c.id === id)
    if (conn && !conn.connected) {
      try {
        await connectionStore.connect(id)
      } catch {
        return
      }
    }

    if (!schemaStore.trees[id]) {
      schemaStore.loadDatabases(id);
    }
  }

  async function disconnect(id: string, e: MouseEvent) {
    e.stopPropagation();
    await connectionStore.disconnect(id);
  }

  function handleTableClick(schemaName: string, name: string, _type: 'table' | 'view') {
    if (!activeConn) return
    const driver = activeConn.config.DriverType
    let sql: string
    if (driver === 'mysql') {
      sql = `SELECT * FROM \`${schemaName}\`.\`${name}\` LIMIT 100`
    } else if (driver === 'mssql') {
      sql = `SELECT TOP 100 * FROM [${schemaName}].[${name}]`
    } else if (driver === 'sqlite') {
      sql = `SELECT * FROM "${name}" LIMIT 100`
    } else {
      sql = `SELECT * FROM "${schemaName}"."${name}" LIMIT 100`
    }
    tabStore.openDataTab(activeConn.id, name, sql)
  }

  async function refresh() {
    if (!activeConn) return
    await schemaStore.refresh(activeConn.id)
  }

  let activeConn = $derived(connectionStore.active);
  let tree = $derived(activeConn ? (schemaStore.trees[activeConn.id] ?? []) : []);
  let treeLoading = $derived(activeConn ? (schemaStore.loading[activeConn.id] ?? false) : false);
</script>

<aside class="w-60 flex flex-col shrink-0" style="background-color: {colors.background.secondary}; border-right: 1px solid {colors.border.primary}">
  <div class="flex items-center justify-between px-3 py-3" style="border-bottom: 1px solid {colors.border.primary}">
    <span class="text-xs font-semibold uppercase tracking-wider" style="color: {colors.text.muted}">Connections</span>
    <button
      onclick={() => showDialog = true}
      class="w-5 h-5 flex items-center justify-center rounded text-lg leading-none transition-colors cursor-pointer"
      style="color: {colors.text.muted}"
      onmouseenter={e => (e.currentTarget as HTMLElement).style.color = colors.accent.primary}
      onmouseleave={e => (e.currentTarget as HTMLElement).style.color = colors.text.muted}
      title="New Connection"
    >+</button>
  </div>

  <div class="py-1" style="border-bottom: 1px solid {colors.border.primary}">
    {#if connectionStore.list.length === 0}
      <p class="text-xs px-3 py-3 text-center" style="color: {colors.text.muted}">No connections yet</p>
    {:else}
      {#each connectionStore.list as conn (conn.id)}
        <button
          onclick={() => selectConnection(conn.id)}
          class="w-full flex items-center gap-2 px-3 py-2 text-left transition-colors group cursor-pointer"
          style="background-color: {connectionStore.activeId === conn.id ? colors.background.tertiary : 'transparent'}"
          onmouseenter={e => { if (connectionStore.activeId !== conn.id) (e.currentTarget as HTMLElement).style.backgroundColor = colors.background.tertiary }}
          onmouseleave={e => { if (connectionStore.activeId !== conn.id) (e.currentTarget as HTMLElement).style.backgroundColor = 'transparent' }}
        >
          <span class="w-2 h-2 rounded-full shrink-0" style="background-color: {conn.connected ? colors.accent.primary : colors.border.primary}"></span>
          <div class="flex-1 min-w-0">
            <p class="text-xs font-medium truncate" style="color: {colors.text.primary}">{conn.name}</p>
            <p class="text-[11px] truncate" style="color: {colors.text.muted}">{conn.config.DriverType}</p>
          </div>
          {#if conn.connected}
            <span
              role="button"
              tabindex="0"
              onclick={(e) => disconnect(conn.id, e)}
              onkeydown={e => { if (e.key === 'Enter') disconnect(conn.id, e as any) }}
              class="opacity-0 group-hover:opacity-100 text-xs px-1 transition-all cursor-pointer"
              style="color: {colors.text.muted}"
              onmouseenter={e => (e.currentTarget as HTMLElement).style.color = '#f87171'}
              onmouseleave={e => (e.currentTarget as HTMLElement).style.color = colors.text.muted}
              title="Disconnect"
            >✕</span>
          {/if}
        </button>
      {/each}
    {/if}
  </div>

  <div class="flex items-center justify-between px-3 py-1.5" style="border-bottom: 1px solid {colors.border.primary}">
    <span class="text-xs font-semibold uppercase tracking-wider" style="color: {colors.text.muted}">Schema</span>
    {#if activeConn}
      <button
        onclick={refresh}
        class="w-5 h-5 flex items-center justify-center rounded text-sm leading-none transition-colors cursor-pointer"
        style="color: {colors.text.muted}"
        onmouseenter={e => (e.currentTarget as HTMLElement).style.color = colors.accent.primary}
        onmouseleave={e => (e.currentTarget as HTMLElement).style.color = colors.text.muted}
        title="Refresh schema"
      >↺</button>
    {/if}
  </div>

  <div class="flex-1 overflow-y-auto py-1">
    {#if treeLoading}
      <p class="text-xs px-3 py-3 text-center" style="color: {colors.text.muted}">Loading…</p>
    {:else if !activeConn}
      <p class="text-xs px-3 py-3 text-center" style="color: {colors.text.muted}">Select a connection</p>
    {:else if tree.length === 0}
      <p class="text-xs px-3 py-3 text-center" style="color: {colors.text.muted}">No databases found</p>
    {:else}
      {#each tree as node (node.id)}
        <TreeNode
          {node}
          connectionId={activeConn.id}
          onTableClick={handleTableClick}
          onRefreshNode={(n) => schemaStore.refreshNode(activeConn.id, n)}
        />
      {/each}
    {/if}
  </div>
</aside>

{#if showDialog}
  <ConnectionDialog onclose={() => showDialog = false} />
{/if}
