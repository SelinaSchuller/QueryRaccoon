<script lang="ts">
  import { getColumns } from '$lib/api/schema'
  import { execute } from '$lib/api/query'
  import { colors } from '$lib/colors'
  import {
    getStrategy,
    generateRows,
    buildInsertSQL,
    isFKColumn,
    guessRefTable,
    type FakerStrategy,
  } from '$lib/utils/seedFaker'

  type Props = {
    connectionId: string
    driver: string
    schemaName: string
    tableName: string
    onclose: () => void
  }
  let { connectionId, driver, schemaName, tableName, onclose }: Props = $props()

  type ColState = {
    name: string
    type: string
    nullable: boolean
    strategy: FakerStrategy
    fkMin: number
    fkMax: number
    enabled: boolean
  }

  let rowCount = $state(10)
  let useLorem = $state(false)
  let loading = $state(true)
  let loadError = $state<string | null>(null)
  let cols = $state<ColState[]>([])
  let executing = $state(false)
  let execError = $state<string | null>(null)
  let execSuccess = $state(false)

  async function load() {
    loading = true
    loadError = null
    try {
      const rawCols = await getColumns(connectionId, schemaName, tableName)
      cols = rawCols.map(c => {
        const fkMin = 1
        const fkMax = 100
        const strategy = getStrategy(c.Name, c.Type, tableName, useLorem, { min: fkMin, max: fkMax })
        return {
          name: c.Name,
          type: c.Type,
          nullable: c.Nullable,
          strategy,
          fkMin,
          fkMax,
          enabled: strategy.kind !== 'skip',
        }
      })
    } catch (e: any) {
      loadError = e?.message ?? String(e)
    } finally {
      loading = false
    }
  }

  // Recompute strategies when lorem toggle changes
  function refreshStrategies() {
    cols = cols.map(c => ({
      ...c,
      strategy: getStrategy(c.name, c.type, tableName, useLorem, { min: c.fkMin, max: c.fkMax }),
    }))
  }

  function updateFKRange(name: string, field: 'fkMin' | 'fkMax', val: number) {
    cols = cols.map(c => {
      if (c.name !== name) return c
      const updated = { ...c, [field]: val }
      updated.strategy = getStrategy(c.name, c.type, tableName, useLorem, { min: updated.fkMin, max: updated.fkMax })
      return updated
    })
  }

  // Preview rows (first 3)
  let previewRows = $derived.by(() => {
    if (cols.length === 0) return []
    const enabledCols = cols.filter(c => c.enabled)
    if (enabledCols.length === 0) return []
    const fkConfigs: Record<string, { min: number; max: number }> = {}
    for (const c of cols) {
      if (isFKColumn(c.name)) fkConfigs[c.name] = { min: c.fkMin, max: c.fkMax }
    }
    return generateRows(
      enabledCols.map(c => ({ name: c.name, type: c.type, nullable: c.nullable })),
      tableName,
      3,
      useLorem,
      fkConfigs
    )
  })

  let previewCols = $derived(cols.filter(c => c.enabled).map(c => c.name))

  async function execute_() {
    executing = true
    execError = null
    execSuccess = false
    try {
      const enabledCols = cols.filter(c => c.enabled)
      const fkConfigs: Record<string, { min: number; max: number }> = {}
      for (const c of cols) {
        if (isFKColumn(c.name)) fkConfigs[c.name] = { min: c.fkMin, max: c.fkMax }
      }
      const rows = generateRows(
        enabledCols.map(c => ({ name: c.name, type: c.type, nullable: c.nullable })),
        tableName,
        rowCount,
        useLorem,
        fkConfigs
      )
      const sql = buildInsertSQL(driver, schemaName, tableName, rows)
      await execute(connectionId, sql)
      execSuccess = true
    } catch (e: any) {
      execError = e?.message ?? String(e)
    } finally {
      executing = false
    }
  }

  load()
</script>

<div class="fixed inset-0 z-50 flex items-center justify-center" style="background: rgba(0,0,0,0.6)">
  <div
    class="rounded-lg shadow-2xl flex flex-col"
    style="width: 640px; max-height: 85vh; background-color: {colors.background.tertiary}; border: 1px solid {colors.border.primary}"
  >
    <!-- Header -->
    <div class="px-5 py-4 shrink-0 flex items-center justify-between" style="border-bottom: 1px solid {colors.border.primary}">
      <div>
        <p class="text-sm font-semibold" style="color: {colors.text.primary}">Seed Data</p>
        <p class="text-xs mt-0.5" style="color: {colors.text.muted}">{schemaName ? `${schemaName}.` : ''}{tableName}</p>
      </div>
      <button
        onclick={onclose}
        class="text-lg leading-none cursor-pointer px-1"
        style="color: {colors.text.muted}"
        onmouseenter={e => (e.currentTarget as HTMLElement).style.color = colors.text.primary}
        onmouseleave={e => (e.currentTarget as HTMLElement).style.color = colors.text.muted}
      >×</button>
    </div>

    <div class="flex-1 overflow-y-auto px-5 py-4 flex flex-col gap-4">
      {#if loading}
        <p class="text-xs text-center py-8" style="color: {colors.text.muted}">Loading columns…</p>
      {:else if loadError}
        <p class="text-xs font-mono" style="color: #f87171">{loadError}</p>
      {:else}

        <!-- Options row -->
        <div class="flex items-center gap-6">
          <div class="flex items-center gap-2">
            <label class="text-xs" style="color: {colors.text.muted}">Rows</label>
            <input
              type="number"
              min="1"
              max="10000"
              bind:value={rowCount}
              class="w-20 px-2 py-1 rounded text-xs text-center"
              style="background-color: {colors.background.secondary}; border: 1px solid {colors.border.primary}; color: {colors.text.primary}; outline: none"
            />
          </div>

          <div class="flex items-center gap-2">
            <label class="text-xs" style="color: {colors.text.muted}">Text style</label>
            <div class="flex rounded overflow-hidden" style="border: 1px solid {colors.border.primary}">
              <button
                onclick={() => { useLorem = false; refreshStrategies() }}
                class="px-2.5 py-1 text-xs cursor-pointer transition-colors"
                style="background-color: {!useLorem ? colors.accent.primary : colors.background.secondary}; color: {!useLorem ? '#fff' : colors.text.muted}"
              >Realistic</button>
              <button
                onclick={() => { useLorem = true; refreshStrategies() }}
                class="px-2.5 py-1 text-xs cursor-pointer transition-colors"
                style="background-color: {useLorem ? colors.accent.primary : colors.background.secondary}; color: {useLorem ? '#fff' : colors.text.muted}"
              >Lorem ipsum</button>
            </div>
          </div>
        </div>

        <!-- Column mapping -->
        <div>
          <p class="text-xs font-medium mb-2" style="color: {colors.text.muted}">Column mapping</p>
          <div class="rounded overflow-hidden" style="border: 1px solid {colors.border.primary}">
            <table class="w-full text-xs">
              <thead>
                <tr style="background-color: {colors.background.secondary}">
                  <th class="text-left px-3 py-2 font-medium w-6" style="color: {colors.text.muted}"></th>
                  <th class="text-left px-3 py-2 font-medium" style="color: {colors.text.muted}">Column</th>
                  <th class="text-left px-3 py-2 font-medium" style="color: {colors.text.muted}">Type</th>
                  <th class="text-left px-3 py-2 font-medium" style="color: {colors.text.muted}">Generates</th>
                  <th class="text-left px-3 py-2 font-medium" style="color: {colors.text.muted}">FK range</th>
                </tr>
              </thead>
              <tbody>
                {#each cols as col (col.name)}
                  <tr style="border-top: 1px solid {colors.border.primary}; opacity: {col.enabled ? 1 : 0.4}">
                    <td class="px-3 py-2 text-center">
                      <input
                        type="checkbox"
                        checked={col.enabled}
                        disabled={col.strategy.kind === 'skip'}
                        onchange={e => {
                          col.enabled = (e.target as HTMLInputElement).checked
                          cols = [...cols]
                        }}
                        class="cursor-pointer"
                      />
                    </td>
                    <td class="px-3 py-2 font-mono" style="color: {colors.text.primary}">{col.name}</td>
                    <td class="px-3 py-2 font-mono text-[11px]" style="color: {colors.text.muted}">{col.type}</td>
                    <td class="px-3 py-2">
                      <span
                        class="px-1.5 py-0.5 rounded text-[11px]"
                        style="
                          background-color: {col.strategy.kind === 'skip' ? colors.background.secondary :
                                             col.strategy.kind === 'uuid' ? '#1e3a5f' :
                                             col.strategy.kind === 'fk'   ? '#3b1f6e' :
                                             '#1a3a2a'};
                          color: {col.strategy.kind === 'skip' ? colors.text.muted :
                                  col.strategy.kind === 'uuid' ? '#60a5fa' :
                                  col.strategy.kind === 'fk'   ? '#a78bfa' :
                                  '#34d399'}
                        "
                      >{col.strategy.label}</span>
                    </td>
                    <td class="px-3 py-2">
                      {#if col.strategy.kind === 'fk'}
                        <div class="flex items-center gap-1">
                          <input
                            type="number"
                            min="1"
                            value={col.fkMin}
                            onchange={e => updateFKRange(col.name, 'fkMin', parseInt((e.target as HTMLInputElement).value) || 1)}
                            class="w-14 px-1.5 py-0.5 rounded text-[11px] text-center"
                            style="background-color: {colors.background.secondary}; border: 1px solid {colors.border.primary}; color: {colors.text.primary}; outline: none"
                          />
                          <span style="color: {colors.text.muted}">–</span>
                          <input
                            type="number"
                            min="1"
                            value={col.fkMax}
                            onchange={e => updateFKRange(col.name, 'fkMax', parseInt((e.target as HTMLInputElement).value) || 100)}
                            class="w-14 px-1.5 py-0.5 rounded text-[11px] text-center"
                            style="background-color: {colors.background.secondary}; border: 1px solid {colors.border.primary}; color: {colors.text.primary}; outline: none"
                          />
                        </div>
                      {/if}
                    </td>
                  </tr>
                {/each}
              </tbody>
            </table>
          </div>
        </div>

        <!-- Preview -->
        {#if previewRows.length > 0}
          <div>
            <p class="text-xs font-medium mb-2" style="color: {colors.text.muted}">Preview (first 3 rows)</p>
            <div class="rounded overflow-auto" style="border: 1px solid {colors.border.primary}; max-height: 140px">
              <table class="text-xs w-full">
                <thead class="sticky top-0" style="background-color: {colors.background.secondary}">
                  <tr>
                    {#each previewCols as col}
                      <th class="text-left px-3 py-1.5 font-medium whitespace-nowrap" style="color: {colors.text.muted}; border-bottom: 1px solid {colors.border.primary}">{col}</th>
                    {/each}
                  </tr>
                </thead>
                <tbody>
                  {#each previewRows as row, i}
                    <tr style="border-top: {i > 0 ? `1px solid ${colors.border.primary}` : 'none'}">
                      {#each previewCols as col}
                        <td class="px-3 py-1.5 font-mono whitespace-nowrap max-w-[160px] truncate" style="color: {colors.text.primary}">
                          {row[col] === null || row[col] === undefined ? 'NULL' : String(row[col])}
                        </td>
                      {/each}
                    </tr>
                  {/each}
                </tbody>
              </table>
            </div>
          </div>
        {/if}

        {#if execError}
          <p class="text-xs font-mono" style="color: #f87171">{execError}</p>
        {/if}

        {#if execSuccess}
          <p class="text-xs" style="color: #34d399">
            Inserted {rowCount} {rowCount === 1 ? 'row' : 'rows'} into {tableName}.
          </p>
        {/if}

      {/if}
    </div>

    <!-- Footer -->
    <div class="px-5 py-3 flex justify-end gap-2 shrink-0" style="border-top: 1px solid {colors.border.primary}">
      <button
        onclick={onclose}
        class="px-3 py-1.5 rounded text-xs cursor-pointer transition-colors"
        style="color: {colors.text.muted}; border: 1px solid {colors.border.primary}"
        onmouseenter={e => (e.currentTarget as HTMLElement).style.color = colors.text.primary}
        onmouseleave={e => (e.currentTarget as HTMLElement).style.color = colors.text.muted}
      >Cancel</button>
      <button
        onclick={execute_}
        disabled={executing || loading || cols.filter(c => c.enabled).length === 0}
        class="px-3 py-1.5 rounded text-xs font-medium cursor-pointer"
        style="background-color: {colors.accent.primary}; color: #fff; opacity: {executing || loading ? '0.6' : '1'}"
      >{executing ? 'Inserting…' : `Insert ${rowCount} rows`}</button>
    </div>
  </div>
</div>
