<script lang="ts">
  import { getColumns } from '$lib/api/schema'
  import { execute } from '$lib/api/query'
  import { connectionStore } from '$lib/stores/connections.svelte'
  import { tabStore } from '$lib/stores/tabs.svelte'
  import { schemaStore } from '$lib/stores/schema.svelte'
  import { colors } from '$lib/colors'
  import { generateDDL, type ColumnEdit } from '$lib/utils/ddl'
  import type { DriverType } from '$lib/utils/ddl'

  let tab = $derived(tabStore.active)
  let conn = $derived(connectionStore.active)
  let meta = $derived(tab?.schemaTab)

  let originalTableName = $state('')
  let editedTableName = $state('')
  let edits = $state<ColumnEdit[]>([])
  let loading = $state(false)
  let loadError = $state<string | null>(null)

  let showApply = $state(false)
  let ddlStatements = $state<string[]>([])
  let ddlWarnings = $state<string[]>([])
  let applying = $state(false)
  let applyError = $state<string | null>(null)
  let applySuccess = $state(false)

  // Reload when tab changes
  let prevTabId = $state<string | undefined>(undefined)
  $effect(() => {
    const id = tab?.id
    if (id !== prevTabId) {
      prevTabId = id
      load()
    }
  })

  async function load() {
    if (!tab?.connectionId || !meta) return
    loading = true
    loadError = null
    try {
      const cols = await getColumns(tab.connectionId, meta.schemaName, meta.tableName)
      originalTableName = meta.tableName
      editedTableName = meta.tableName
      edits = cols.map(c => ({
        id: crypto.randomUUID(),
        original: c,
        name: c.Name,
        type: c.Type,
        nullable: c.Nullable,
        default: c.Default ?? '',
        deleted: false,
      }))
    } catch (e: any) {
      loadError = e?.message ?? String(e)
    } finally {
      loading = false
    }
  }

  function addColumn() {
    edits = [...edits, {
      id: crypto.randomUUID(),
      original: null,
      name: 'new_column',
      type: 'VARCHAR(255)',
      nullable: true,
      default: '',
      deleted: false,
    }]
  }

  function toggleDelete(id: string) {
    edits = edits.map(e => e.id === id ? { ...e, deleted: !e.deleted } : e)
  }

  function updateEdit(id: string, field: keyof ColumnEdit, value: any) {
    edits = edits.map(e => e.id === id ? { ...e, [field]: value } : e)
  }

  function hasChanges(): boolean {
    if (editedTableName !== originalTableName) return true
    return edits.some(e => {
      if (!e.original) return !e.deleted
      if (e.deleted) return true
      return e.name !== e.original.Name ||
             e.type !== e.original.Type ||
             e.nullable !== e.original.Nullable ||
             e.default !== (e.original.Default ?? '')
    })
  }

  function isChanged(e: ColumnEdit): boolean {
    if (!e.original) return true
    return e.name !== e.original.Name ||
           e.type !== e.original.Type ||
           e.nullable !== e.original.Nullable ||
           e.default !== (e.original.Default ?? '')
  }

  function openApply() {
    if (!conn || !meta) return
    const result = generateDDL(
      conn.config.DriverType as DriverType,
      meta.schemaName,
      originalTableName,
      editedTableName,
      edits
    )
    ddlStatements = result.statements
    ddlWarnings = result.warnings
    applyError = null
    showApply = true
  }

  async function confirmApply() {
    if (!tab?.connectionId) return
    applying = true
    applyError = null
    try {
      for (const stmt of ddlStatements) {
        await execute(tab.connectionId, stmt)
      }
      if (conn) await schemaStore.refresh(tab.connectionId)
      showApply = false
      await load()
      applySuccess = true
      setTimeout(() => applySuccess = false, 3000)
    } catch (e: any) {
      applyError = e?.message ?? String(e)
    } finally {
      applying = false
    }
  }

  function discard() {
    load()
  }

  function rowBg(e: ColumnEdit): string {
    if (e.deleted) return '#7f1d1d22'
    if (!e.original) return '#14532d22'
    if (isChanged(e)) return '#78350f22'
    return 'transparent'
  }

  function rowColor(e: ColumnEdit): string {
    if (e.deleted) return '#fca5a5'
    if (!e.original) return '#86efac'
    if (isChanged(e)) return '#fbbf24'
    return colors.text.primary
  }

  const inputStyle = `background: transparent; outline: none; width: 100%; color: inherit`
</script>

<div class="flex flex-col h-full" style="background-color: {colors.background.primary}">

  <!-- Header -->
  <div class="flex items-center gap-3 px-4 py-2.5 shrink-0 flex-wrap" style="background-color: {colors.background.secondary}; border-bottom: 1px solid {colors.border.primary}">
    <span class="text-xs font-semibold uppercase tracking-wider" style="color: {colors.text.muted}">Table</span>

    <input
      type="text"
      value={editedTableName}
      oninput={e => editedTableName = (e.target as HTMLInputElement).value}
      class="text-sm font-medium px-1 rounded"
      style="background: transparent; color: {editedTableName !== originalTableName ? '#fbbf24' : colors.text.primary}; border: 1px solid {editedTableName !== originalTableName ? '#fbbf24' : 'transparent'}; outline: none; min-width: 120px"
    />

    {#if meta?.schemaName}
      <span class="text-xs px-1.5 py-0.5 rounded" style="background-color: {colors.background.tertiary}; color: {colors.text.muted}">{meta.schemaName}</span>
    {/if}

    <div class="ml-auto flex items-center gap-2">
      {#if hasChanges()}
        <button
          onclick={discard}
          class="px-2 py-1 rounded text-xs cursor-pointer transition-colors"
          style="color: {colors.text.muted}; border: 1px solid {colors.border.primary}"
          onmouseenter={e => (e.currentTarget as HTMLElement).style.color = colors.text.primary}
          onmouseleave={e => (e.currentTarget as HTMLElement).style.color = colors.text.muted}
        >Discard</button>
        <button
          onclick={openApply}
          class="px-2 py-1 rounded text-xs font-medium cursor-pointer"
          style="background-color: {colors.accent.primary}; color: #fff; border: none"
        >Apply Changes</button>
      {/if}
    </div>
  </div>

  <!-- Column grid -->
  <div class="flex-1 overflow-auto">
    {#if loading}
      <div class="flex items-center justify-center h-full text-sm" style="color: {colors.text.muted}">Loading…</div>
    {:else if loadError}
      <div class="p-4 text-sm font-mono" style="color: #f87171">{loadError}</div>
    {:else}
      <table class="w-full text-xs border-collapse">
        <thead class="sticky top-0 z-10" style="background-color: {colors.background.secondary}">
          <tr>
            <th class="text-left px-4 py-2.5 font-medium w-8" style="color: {colors.text.muted}; border-bottom: 1px solid {colors.border.primary}">#</th>
            <th class="text-left px-3 py-2.5 font-medium" style="color: {colors.text.muted}; border-bottom: 1px solid {colors.border.primary}; min-width: 160px">Name</th>
            <th class="text-left px-3 py-2.5 font-medium" style="color: {colors.text.muted}; border-bottom: 1px solid {colors.border.primary}; min-width: 160px">Type</th>
            <th class="text-center px-3 py-2.5 font-medium w-24" style="color: {colors.text.muted}; border-bottom: 1px solid {colors.border.primary}">Nullable</th>
            <th class="text-left px-3 py-2.5 font-medium" style="color: {colors.text.muted}; border-bottom: 1px solid {colors.border.primary}; min-width: 140px">Default</th>
            <th class="w-10" style="border-bottom: 1px solid {colors.border.primary}"></th>
          </tr>
        </thead>
        <tbody>
          {#each edits as edit, i (edit.id)}
            <tr style="background-color: {rowBg(edit)}; color: {rowColor(edit)}; opacity: {edit.deleted ? 0.5 : 1}; text-decoration: {edit.deleted ? 'line-through' : 'none'}">
              <td class="px-4 py-2" style="border-bottom: 1px solid {colors.border.primary}; color: {colors.text.muted}">{i + 1}</td>
              <td class="px-3 py-2 font-mono" style="border-bottom: 1px solid {colors.border.primary}">
                {#if !edit.deleted}
                  <input
                    type="text"
                    value={edit.name}
                    oninput={e => updateEdit(edit.id, 'name', (e.target as HTMLInputElement).value)}
                    style={inputStyle}
                    class="font-mono text-xs"
                  />
                {:else}
                  {edit.name}
                {/if}
              </td>
              <td class="px-3 py-2 font-mono" style="border-bottom: 1px solid {colors.border.primary}">
                {#if !edit.deleted}
                  <input
                    type="text"
                    value={edit.type}
                    oninput={e => updateEdit(edit.id, 'type', (e.target as HTMLInputElement).value)}
                    style={inputStyle}
                    class="font-mono text-xs"
                  />
                {:else}
                  {edit.type}
                {/if}
              </td>
              <td class="px-3 py-2 text-center" style="border-bottom: 1px solid {colors.border.primary}">
                <input
                  type="checkbox"
                  checked={edit.nullable}
                  disabled={edit.deleted}
                  onchange={e => updateEdit(edit.id, 'nullable', (e.target as HTMLInputElement).checked)}
                  class="cursor-pointer"
                />
              </td>
              <td class="px-3 py-2 font-mono" style="border-bottom: 1px solid {colors.border.primary}">
                {#if !edit.deleted}
                  <input
                    type="text"
                    value={edit.default}
                    placeholder="—"
                    oninput={e => updateEdit(edit.id, 'default', (e.target as HTMLInputElement).value)}
                    style={inputStyle}
                    class="font-mono text-xs"
                  />
                {:else}
                  {edit.default || '—'}
                {/if}
              </td>
              <td class="px-2 py-2 text-center" style="border-bottom: 1px solid {colors.border.primary}">
                <button
                  onclick={() => toggleDelete(edit.id)}
                  class="w-5 h-5 flex items-center justify-center rounded text-xs cursor-pointer transition-colors mx-auto"
                  style="color: {edit.deleted ? colors.accent.primary : colors.text.muted}"
                  title={edit.deleted ? 'Restore column' : 'Delete column'}
                  onmouseenter={e => (e.currentTarget as HTMLElement).style.color = edit.deleted ? colors.accent.primary : '#f87171'}
                  onmouseleave={e => (e.currentTarget as HTMLElement).style.color = edit.deleted ? colors.accent.primary : colors.text.muted}
                >{edit.deleted ? '↩' : '×'}</button>
              </td>
            </tr>
          {/each}
        </tbody>
      </table>

      <!-- Add column -->
      <button
        onclick={addColumn}
        class="flex items-center gap-2 px-4 py-2.5 w-full text-left text-xs cursor-pointer transition-colors"
        style="color: {colors.text.muted}"
        onmouseenter={e => (e.currentTarget as HTMLElement).style.color = colors.accent.primary}
        onmouseleave={e => (e.currentTarget as HTMLElement).style.color = colors.text.muted}
      >
        <span style="color: {colors.accent.primary}">+</span> Add column
      </button>
    {/if}
  </div>

  <!-- Legend -->
  {#if edits.some(e => e.deleted || !e.original || isChanged(e))}
    <div class="flex items-center gap-4 px-4 py-1.5 text-xs shrink-0" style="border-top: 1px solid {colors.border.primary}; color: {colors.text.muted}">
      {#if edits.some(e => !e.original && !e.deleted)}<span><span style="color:#86efac">■</span> New</span>{/if}
      {#if edits.some(e => e.original && isChanged(e) && !e.deleted)}<span><span style="color:#fbbf24">■</span> Modified</span>{/if}
      {#if edits.some(e => e.deleted)}<span><span style="color:#fca5a5">■</span> Deleted</span>{/if}
    </div>
  {/if}

  <!-- Success toast -->
  {#if applySuccess}
    <div class="absolute bottom-4 right-4 z-50 px-3 py-2 rounded text-xs" style="background-color: #14532d; color: #86efac; border: 1px solid #22c55e">
      Changes applied successfully
    </div>
  {/if}

</div>

<!-- Apply confirmation modal -->
{#if showApply}
  <div class="fixed inset-0 z-50 flex items-center justify-center" style="background: rgba(0,0,0,0.6)">
    <div class="rounded-lg shadow-2xl w-[520px] max-h-[80vh] flex flex-col" style="background-color: {colors.background.tertiary}; border: 1px solid {colors.border.primary}">
      <div class="px-5 py-4 shrink-0" style="border-bottom: 1px solid {colors.border.primary}">
        <p class="text-sm font-semibold" style="color: {colors.text.primary}">Apply Changes</p>
        <p class="text-xs mt-0.5" style="color: {colors.text.muted}">The following SQL will be executed:</p>
      </div>

      <div class="px-5 py-3 flex-1 overflow-auto">
        {#if ddlStatements.length === 0}
          <p class="text-xs" style="color: {colors.text.muted}">No changes to apply.</p>
        {:else}
          <pre class="text-xs font-mono whitespace-pre-wrap" style="color: {colors.text.primary}">{ddlStatements.join(';\n') + ';'}</pre>
        {/if}

        {#if ddlWarnings.length > 0}
          <div class="mt-3 flex flex-col gap-1">
            {#each ddlWarnings as w}
              <p class="text-xs" style="color: #fbbf24">⚠ {w}</p>
            {/each}
          </div>
        {/if}

        {#if applyError}
          <p class="text-xs mt-3 font-mono" style="color: #f87171">{applyError}</p>
        {/if}
      </div>

      <div class="px-5 py-3 flex justify-end gap-2 shrink-0" style="border-top: 1px solid {colors.border.primary}">
        <button
          onclick={() => showApply = false}
          class="px-3 py-1.5 rounded text-xs cursor-pointer transition-colors"
          style="color: {colors.text.muted}; border: 1px solid {colors.border.primary}"
          onmouseenter={e => (e.currentTarget as HTMLElement).style.color = colors.text.primary}
          onmouseleave={e => (e.currentTarget as HTMLElement).style.color = colors.text.muted}
        >Cancel</button>
        <button
          onclick={confirmApply}
          disabled={applying || ddlStatements.length === 0}
          class="px-3 py-1.5 rounded text-xs font-medium cursor-pointer"
          style="background-color: {colors.accent.primary}; color: #fff; opacity: {applying || ddlStatements.length === 0 ? '0.6' : '1'}"
        >{applying ? 'Executing…' : 'Execute'}</button>
      </div>
    </div>
  </div>
{/if}
