<script lang="ts">
  import { tabStore } from "$lib/stores/tabs.svelte";
  import { connectionStore } from "$lib/stores/connections.svelte";
  import { colors } from "$lib/colors";
  import { SaveFile, SaveXLSX } from '$wailsjs/go/bindings/ExportService';

  type Props = { collapsed?: boolean; onToggleCollapse?: () => void }
  let { collapsed = false, onToggleCollapse }: Props = $props()

  let tab = $derived(tabStore.active);
  let activeConn = $derived(connectionStore.active);
  let result = $derived(tab?.result);
  let allRows = $derived(result?.Rows ?? []);

  // --- Row selection state ---
  let selectedRows = $state(new Set<number>());
  let lastClickedRow = $state<number | null>(null);

  // Reset selection on tab change or new result
  let prevTabId = $state<string | undefined>(undefined);
  let prevResult = $state<object | null>(null);
  $effect(() => {
    const id = tab?.id;
    if (id !== prevTabId) {
      prevTabId = id;
      selectedRows = new Set();
      lastClickedRow = null;
    } else if (result !== prevResult) {
      prevResult = result ?? null;
      selectedRows = new Set();
      lastClickedRow = null;
    }
  });

  // --- Row click handling ---
  function handleRowClick(i: number, e: MouseEvent) {
    const meta = e.metaKey || e.ctrlKey;
    const shift = e.shiftKey;

    if (shift && lastClickedRow !== null) {
      const lo = Math.min(lastClickedRow, i);
      const hi = Math.max(lastClickedRow, i);
      const next = new Set(selectedRows);
      for (let r = lo; r <= hi; r++) next.add(r);
      selectedRows = next;
    } else if (meta) {
      const next = new Set(selectedRows);
      next.has(i) ? next.delete(i) : next.add(i);
      selectedRows = next;
      lastClickedRow = i;
    } else {
      if (selectedRows.size === 1 && selectedRows.has(i)) {
        selectedRows = new Set();
        lastClickedRow = null;
      } else {
        selectedRows = new Set([i]);
        lastClickedRow = i;
      }
    }
  }

  function handleKeydown(e: KeyboardEvent) {
    if (!result) return;
    if ((e.metaKey || e.ctrlKey) && e.key === 'a') {
      e.preventDefault();
      selectedRows = new Set(allRows.map((_, i) => i));
    }
    if (e.key === 'Escape') {
      selectedRows = new Set();
      lastClickedRow = null;
    }
  }

  // --- Status label ---
  let statusLabel = $derived(
    !result ? '' :
    selectedRows.size > 0
      ? `${selectedRows.size} of ${allRows.length} row${allRows.length !== 1 ? 's' : ''} selected`
      : `${allRows.length} row${allRows.length !== 1 ? 's' : ''}`
  );

  // --- Export dropdown ---
  let showExportMenu = $state(false);

  // --- Export modal state ---
  type ExportFormat = 'csv' | 'json' | 'sql' | 'xlsx';
  type ExportModal = {
    format: ExportFormat;
    filename: string;
    includedColumns: Set<string>;
    selectedRowsOnly: boolean;
  };
  let exportModal = $state<ExportModal | null>(null);

  function defaultFilename(ext: string): string {
    const db = activeConn?.config.Database;
    const name = tab?.name;
    if (db && name && name !== 'Query') return `${db}.${name}.${ext}`;
    if (db) return `${db}.${ext}`;
    if (name && name !== 'Query') return `${name}.${ext}`;
    return `results.${ext}`;
  }

  const FORMAT_EXT: Record<ExportFormat, string> = { csv: 'csv', json: 'json', sql: 'sql', xlsx: 'xlsx' };
  const FORMAT_LABELS: Record<ExportFormat, string> = { csv: 'CSV', json: 'JSON', sql: 'SQL INSERT', xlsx: 'XLSX' };

  function openExportModal(format: ExportFormat) {
    showExportMenu = false;
    exportModal = {
      format,
      filename: defaultFilename(FORMAT_EXT[format]),
      includedColumns: new Set(result?.Columns ?? []),
      selectedRowsOnly: selectedRows.size > 0,
    };
  }

  function toggleModalColumn(col: string) {
    if (!exportModal) return;
    const next = new Set(exportModal.includedColumns);
    next.has(col) ? next.delete(col) : next.add(col);
    exportModal = { ...exportModal, includedColumns: next };
  }

  function toggleAllModalColumns(all: boolean) {
    if (!exportModal || !result) return;
    exportModal = { ...exportModal, includedColumns: all ? new Set(result.Columns) : new Set() };
  }

  async function confirmExport() {
    if (!exportModal || !result) return;
    const { format, filename, includedColumns, selectedRowsOnly } = exportModal;
    exportModal = null;

    const cols = result.Columns.filter(c => includedColumns.has(c));
    const colIndices = cols.map(c => result!.Columns.indexOf(c));
    const rows = selectedRowsOnly && selectedRows.size > 0
      ? [...selectedRows].sort((a, b) => a - b).map(i => allRows[i])
      : allRows;
    const data = rows.map(row => colIndices.map(i => row[i]));

    if (data.length > 100_000) {
      if (!confirm(`You're about to export ${data.length.toLocaleString()} rows. Continue?`)) return;
    }

    if (format === 'xlsx') {
      await SaveXLSX(cols, data, filename);
      return;
    }

    let content = '';
    if (format === 'csv') {
      const esc = (v: any) => { const s = v === null ? '' : String(v); return s.includes(',') || s.includes('"') || s.includes('\n') ? `"${s.replace(/"/g, '""')}"` : s; };
      content = [cols.map(esc).join(','), ...data.map(r => r.map(esc).join(','))].join('\n');
    } else if (format === 'json') {
      content = JSON.stringify(data.map(row => Object.fromEntries(cols.map((c, i) => [c, row[i]]))), null, 2);
    } else {
      const tableName = tab?.name && tab.name !== 'Query' ? tab.name : 'table_name';
      const escVal = (v: any) => v === null ? 'NULL' : (typeof v === 'number' || typeof v === 'boolean') ? String(v) : `'${String(v).replace(/'/g, "''")}'`;
      content = data.length === 0
        ? `-- Table: ${tableName}\n-- No rows`
        : `-- Table: ${tableName}\nINSERT INTO "${tableName}" (${cols.map(c => `"${c}"`).join(', ')}) VALUES\n${data.map(row => `(${row.map(escVal).join(', ')})`).join(',\n')};`;
    }
    await SaveFile(content, filename);
  }
</script>

<svelte:window onkeydown={handleKeydown} />

<div class="flex flex-col h-full relative" style="background-color: {colors.background.primary}">

  <!-- Toolbar -->
  <div class="flex items-center gap-2 px-3 py-2 shrink-0 flex-wrap" style="background-color: {colors.background.secondary}; border-bottom: 1px solid {colors.border.primary}">
    <span class="text-xs font-semibold uppercase tracking-wider" style="color: {colors.text.muted}">Results</span>

    {#if result}
      <span class="text-xs" style="color: {selectedRows.size > 0 ? colors.accent.primary : colors.text.muted}">
        {statusLabel}
      </span>
    {/if}

    {#if tab?.executionTime != null}
      <span class="text-xs" style="color: {colors.text.muted}">· {tab.executionTime}ms</span>
    {/if}

    {#if result}
      <div class="ml-auto flex items-center gap-1">
        <!-- Export dropdown -->
        <div class="relative">
          <button
            onclick={() => { showExportMenu = !showExportMenu; }}
            class="px-2 py-1 rounded text-xs transition-colors cursor-pointer"
            style="color: {colors.text.muted}; border: 1px solid {colors.border.primary}"
            onmouseenter={e => (e.currentTarget as HTMLElement).style.color = colors.text.primary}
            onmouseleave={e => (e.currentTarget as HTMLElement).style.color = colors.text.muted}
          >Export ▾</button>

          {#if showExportMenu}
            <div class="fixed inset-0 z-40" onclick={() => showExportMenu = false} role="presentation"></div>
            <div class="absolute right-0 top-full mt-1 z-50 rounded-md shadow-xl py-1 min-w-[150px]"
                 style="background-color: {colors.background.tertiary}; border: 1px solid {colors.border.primary}">
              {#each (['csv', 'json', 'sql', 'xlsx'] as ExportFormat[]) as fmt}
                <button
                  onclick={() => openExportModal(fmt)}
                  class="w-full text-left px-3 py-1.5 text-xs cursor-pointer transition-colors"
                  style="color: {colors.text.primary}"
                  onmouseenter={e => (e.currentTarget as HTMLElement).style.backgroundColor = colors.background.secondary}
                  onmouseleave={e => (e.currentTarget as HTMLElement).style.backgroundColor = 'transparent'}
                >{FORMAT_LABELS[fmt]}</button>
              {/each}
            </div>
          {/if}
        </div>
      </div>
    {/if}

    <button
      onclick={onToggleCollapse}
      class="w-5 h-5 flex items-center justify-center rounded text-xs transition-colors cursor-pointer {result ? '' : 'ml-auto'}"
      style="color: {colors.text.muted}"
      onmouseenter={e => (e.currentTarget as HTMLElement).style.color = colors.text.primary}
      onmouseleave={e => (e.currentTarget as HTMLElement).style.color = colors.text.muted}
      title={collapsed ? 'Expand results' : 'Collapse results'}
    >{collapsed ? '▴' : '▾'}</button>
  </div>

  <!-- Results body -->
  <div class="flex-1 overflow-auto select-none" style="display: {collapsed ? 'none' : 'block'}">
    {#if tab?.isExecuting}
      <div class="flex items-center justify-center h-full gap-2 text-sm" style="color: {colors.text.muted}">
        <span class="animate-spin">⟳</span> Executing…
      </div>

    {:else if tab?.error}
      <div class="p-4">
        <p class="text-sm font-mono whitespace-pre-wrap" style="color: #f87171">{tab.error}</p>
      </div>

    {:else if result}
      <table class="w-full text-xs border-collapse">
        <thead class="sticky top-0" style="background-color: {colors.background.secondary}">
          <tr>
            {#each result.Columns as col}
              <th
                class="text-left px-3 py-2 font-medium whitespace-nowrap"
                style="color: {colors.text.muted}; border-bottom: 1px solid {colors.border.primary}"
              >
                {col}
              </th>
            {/each}
          </tr>
        </thead>
        <tbody>
          {#each allRows as row, i}
            <tr
              onclick={e => handleRowClick(i, e)}
              class="cursor-pointer"
              style="background-color: {selectedRows.has(i) ? colors.accent.primary + '22' : i % 2 === 0 ? colors.background.primary : colors.background.secondary}"
              onmouseenter={e => { if (!selectedRows.has(i)) (e.currentTarget as HTMLElement).style.backgroundColor = colors.background.tertiary }}
              onmouseleave={e => { if (!selectedRows.has(i)) (e.currentTarget as HTMLElement).style.backgroundColor = i % 2 === 0 ? colors.background.primary : colors.background.secondary }}
            >
              {#each row as cell}
                <td
                  class="px-3 py-1.5 whitespace-nowrap max-w-xs truncate font-mono"
                  style="color: {colors.text.primary}; border-bottom: 1px solid {colors.border.primary}"
                >
                  {#if cell === null}
                    <span class="italic" style="color: {colors.text.muted}">NULL</span>
                  {:else}
                    {String(cell)}
                  {/if}
                </td>
              {/each}
            </tr>
          {/each}
          {#if allRows.length === 0 && result.Columns.length > 0}
            <tr>
              <td colspan={result.Columns.length} class="px-3 py-6 text-center text-sm" style="color: {colors.text.muted}">
                No rows
              </td>
            </tr>
          {/if}
        </tbody>
      </table>

    {:else}
      <div class="flex items-center justify-center h-full text-sm" style="color: {colors.text.muted}">
        Run a query to see results
      </div>
    {/if}
  </div>

  <!-- Export modal -->
  {#if exportModal && result}
    <div class="fixed inset-0 z-50 flex items-center justify-center" style="background: rgba(0,0,0,0.6)">
      <div class="rounded-lg shadow-2xl w-96" style="background-color: {colors.background.tertiary}; border: 1px solid {colors.border.primary}">
        <!-- Header -->
        <div class="px-5 py-4" style="border-bottom: 1px solid {colors.border.primary}">
          <p class="text-sm font-semibold" style="color: {colors.text.primary}">Export {FORMAT_LABELS[exportModal.format]}</p>
        </div>

        <div class="px-5 py-4 flex flex-col gap-4">
          <!-- Filename -->
          <div>
            <span class="block text-xs mb-1 font-medium" style="color: {colors.text.muted}">Filename</span>
            <input
              type="text"
              value={exportModal.filename}
              oninput={e => exportModal = { ...exportModal!, filename: (e.target as HTMLInputElement).value }}
              class="w-full px-2 py-1.5 rounded text-xs font-mono"
              style="background-color: {colors.background.secondary}; color: {colors.text.primary}; border: 1px solid {colors.border.primary}; outline: none"
            />
          </div>

          <!-- Row scope -->
          <div class="flex gap-2">
            <button
              onclick={() => exportModal = { ...exportModal!, selectedRowsOnly: false }}
              class="flex-1 py-1.5 rounded text-xs cursor-pointer transition-colors"
              style="border: 1px solid {!exportModal.selectedRowsOnly ? colors.accent.primary : colors.border.primary}; color: {!exportModal.selectedRowsOnly ? colors.accent.primary : colors.text.muted}; background: transparent"
            >All rows ({allRows.length})</button>
            <button
              onclick={() => { if (selectedRows.size > 0) exportModal = { ...exportModal!, selectedRowsOnly: true }; }}
              class="flex-1 py-1.5 rounded text-xs transition-colors"
              class:cursor-pointer={selectedRows.size > 0}
              class:cursor-not-allowed={selectedRows.size === 0}
              style="border: 1px solid {exportModal.selectedRowsOnly ? colors.accent.primary : colors.border.primary}; color: {exportModal.selectedRowsOnly ? colors.accent.primary : selectedRows.size === 0 ? colors.text.muted + '66' : colors.text.muted}; background: transparent; opacity: {selectedRows.size === 0 ? '0.4' : '1'}"
            >Selected rows ({selectedRows.size})</button>
          </div>

          <!-- Columns -->
          <div>
            <div class="flex items-center justify-between mb-2">
              <span class="text-xs font-medium" style="color: {colors.text.muted}">Columns</span>
              <div class="flex gap-2">
                <button onclick={() => toggleAllModalColumns(true)} class="text-xs cursor-pointer" style="color: {colors.accent.primary}">All</button>
                <button onclick={() => toggleAllModalColumns(false)} class="text-xs cursor-pointer" style="color: {colors.accent.primary}">None</button>
              </div>
            </div>
            <div class="max-h-44 overflow-y-auto flex flex-col gap-0.5">
              {#each result.Columns as col}
                <label class="flex items-center gap-2 py-0.5 text-xs cursor-pointer" style="color: {colors.text.primary}">
                  <input
                    type="checkbox"
                    checked={exportModal.includedColumns.has(col)}
                    onchange={() => toggleModalColumn(col)}
                    class="cursor-pointer"
                  />
                  {col}
                </label>
              {/each}
            </div>
          </div>
        </div>

        <!-- Footer -->
        <div class="px-5 py-3 flex justify-end gap-2" style="border-top: 1px solid {colors.border.primary}">
          <button
            onclick={() => exportModal = null}
            class="px-3 py-1.5 rounded text-xs cursor-pointer transition-colors"
            style="color: {colors.text.muted}; border: 1px solid {colors.border.primary}"
            onmouseenter={e => (e.currentTarget as HTMLElement).style.color = colors.text.primary}
            onmouseleave={e => (e.currentTarget as HTMLElement).style.color = colors.text.muted}
          >Cancel</button>
          <button
            onclick={confirmExport}
            class="px-3 py-1.5 rounded text-xs font-medium cursor-pointer"
            style="background-color: {colors.accent.primary}; color: #fff; border: none"
          >Export</button>
        </div>
      </div>
    </div>
  {/if}

</div>
