<script lang="ts">
  import { tabStore } from "$lib/stores/tabs.svelte";
  import { connectionStore } from "$lib/stores/connections.svelte";
  import { colors } from "$lib/colors";
  import { SaveFile, SaveXLSX } from '$wailsjs/go/bindings/ExportService';
  import { execute } from '$lib/api/query';
  import { tick } from 'svelte';

  type Props = { collapsed?: boolean; onToggleCollapse?: () => void }
  let { collapsed = false, onToggleCollapse }: Props = $props()

  let editInputEl = $state<HTMLInputElement | null>(null);

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
    if (editingCell) return; // let cell input handle its own keys
    if (!result) return;
    const mod = e.metaKey || e.ctrlKey;
    if (mod && e.key === 'z' && !e.shiftKey) { e.preventDefault(); undo(); return; }
    if (mod && (e.key === 'Z' || (e.key === 'z' && e.shiftKey))) { e.preventDefault(); redo(); return; }
    if (mod && e.key === 'a') { e.preventDefault(); selectedRows = new Set(allRows.map((_, i) => i)); return; }
    if (e.key === 'Escape') {
      if (deleteConfirm) { deleteConfirm = false; return; }
      selectedRows = new Set(); lastClickedRow = null;
    }
    if ((e.key === 'Delete' || e.key === 'Backspace') && selectedRows.size > 0) {
      e.preventDefault();
      if (deleteConfirm) { deleteSelectedRows(); } else { deleteConfirm = true; }
      return;
    }
  }

  // --- Cell editing ---
  type EditingCell = { row: number; col: number; value: string };
  type PendingChange = { row: number; col: number; newValue: string };

  let editingCell = $state<EditingCell | null>(null);
  let pendingChanges = $state(new Map<string, PendingChange>()); // key: "row-col"
  let failedCells = $state(new Set<string>()); // keys of cells that failed to save
  let saveError = $state<string | null>(null);
  let saveSuccess = $state(false);
  let isSaving = $state(false);
  let undoStack = $state<Map<string, PendingChange>[]>([]);
  let redoStack = $state<Map<string, PendingChange>[]>([]);

  // Reset on tab change or new result
  $effect(() => {
    tab?.id; // track tab id
    pendingChanges = new Map();
    failedCells = new Set();
    saveError = null;
    undoStack = [];
    redoStack = [];
  });

  function pushUndo() {
    undoStack = [...undoStack, new Map(pendingChanges)];
    redoStack = [];
    failedCells = new Set();
  }

  function undo() {
    if (undoStack.length === 0) return;
    redoStack = [...redoStack, new Map(pendingChanges)];
    pendingChanges = undoStack[undoStack.length - 1];
    undoStack = undoStack.slice(0, -1);
    failedCells = new Set();
    saveError = null;
  }

  function redo() {
    if (redoStack.length === 0) return;
    undoStack = [...undoStack, new Map(pendingChanges)];
    pendingChanges = redoStack[redoStack.length - 1];
    redoStack = redoStack.slice(0, -1);
    failedCells = new Set();
    saveError = null;
  }

  function cellKey(row: number, col: number) { return `${row}-${col}`; }

  function getCellDisplayValue(rowIndex: number, colIndex: number): any {
    const pending = pendingChanges.get(cellKey(rowIndex, colIndex));
    if (pending) return pending.newValue === '' ? null : pending.newValue;
    return allRows[rowIndex][colIndex];
  }

  async function startEdit(rowIndex: number, colIndex: number) {
    const current = getCellDisplayValue(rowIndex, colIndex);
    editingCell = { row: rowIndex, col: colIndex, value: current === null ? '' : String(current) };
    await tick();
    editInputEl?.focus();
    editInputEl?.select();
  }

  function commitEdit() {
    if (!editingCell || !result) { editingCell = null; return; }
    const original = allRows[editingCell.row][editingCell.col];
    const originalStr = original === null ? '' : String(original);
    const key = cellKey(editingCell.row, editingCell.col);
    if (editingCell.value === originalStr && !pendingChanges.has(key)) {
      editingCell = null; return; // nothing changed
    }
    pushUndo();
    const next = new Map(pendingChanges);
    if (editingCell.value === originalStr) {
      next.delete(key);
    } else {
      next.set(key, { row: editingCell.row, col: editingCell.col, newValue: editingCell.value });
    }
    pendingChanges = next;
    editingCell = null;
  }

  function cancelEdit() {
    editingCell = null;
  }

  function discardChanges() {
    pendingChanges = new Map();
    failedCells = new Set();
    saveError = null;
  }

  function extractTableName(): string | null {
    const sql = tab?.sql ?? '';
    // Capture schema.table or just table, stripping brackets/quotes
    const match = sql.match(/\bFROM\b\s+((?:[\w\[\]"'`]+\.)?[\w\[\]"'`]+)/i);
    if (!match) return null;
    return match[1].replace(/[\[\]"'`]/g, '');
  }

  function sqlVal(v: any): string {
    if (v === null) return 'NULL';
    if (typeof v === 'number' || typeof v === 'boolean') return String(v);
    return `'${String(v).replace(/'/g, "''")}'`;
  }

  async function saveAllChanges() {
    if (!result || !tab?.connectionId || pendingChanges.size === 0) return;

    const tableName = extractTableName() ?? (tab.name !== 'Query' ? tab.name : null);
    if (!tableName) {
      saveError = "Can't determine table name — add a FROM clause to your query.";
      return;
    }

    isSaving = true;
    saveError = null;
    failedCells = new Set();

    const newFailed = new Set<string>();
    let firstError: string | null = null;

    for (const [key, change] of pendingChanges.entries()) {
      const col = result.Columns[change.col];
      const originalRow = allRows[change.row];
      const newVal = change.newValue.trim() === '' || change.newValue.trim().toUpperCase() === 'NULL'
        ? null : change.newValue;

      const where = result.Columns.map((c, i) => {
        const v = originalRow[i];
        return v === null ? `"${c}" IS NULL` : `"${c}" = ${sqlVal(v)}`;
      }).join(' AND ');

      try {
        await execute(tab.connectionId!, `UPDATE ${tableName} SET "${col}" = ${sqlVal(newVal)} WHERE ${where}`);
      } catch (e: any) {
        newFailed.add(key);
        if (!firstError) firstError = e?.message ?? String(e);
      }
    }

    failedCells = newFailed;
    isSaving = false;

    if (newFailed.size === 0) {
      // Build revert map from current allRows (pre-refresh) so Cmd+Z after save
      // creates pending changes that undo the save
      const revertMap = new Map<string, PendingChange>();
      for (const [key, change] of pendingChanges.entries()) {
        const originalVal = allRows[change.row][change.col];
        revertMap.set(key, { row: change.row, col: change.col, newValue: originalVal === null ? '' : String(originalVal) });
      }

      pendingChanges = new Map();
      await tabStore.execute(tab.id);

      // Push revert as undo entry — Cmd+Z will surface it as a new pending change
      undoStack = [...undoStack, revertMap];
      redoStack = [];

      saveSuccess = true;
      setTimeout(() => saveSuccess = false, 3000);
    } else {
      saveError = firstError;
    }
  }

  // --- Row deletion ---
  let deleteConfirm = $state(false);
  let isDeleting = $state(false);
  let deleteError = $state<string | null>(null);

  async function deleteSelectedRows() {
    if (!result || !tab?.connectionId || selectedRows.size === 0) return;

    const tableName = extractTableName() ?? (tab.name !== 'Query' ? tab.name : null);
    if (!tableName) {
      deleteError = "Can't determine table name — add a FROM clause to your query.";
      deleteConfirm = false;
      return;
    }

    isDeleting = true;
    deleteError = null;

    try {
      for (const rowIndex of selectedRows) {
        const originalRow = allRows[rowIndex];
        const where = result.Columns.map((c, i) => {
          const v = originalRow[i];
          return v === null ? `"${c}" IS NULL` : `"${c}" = ${sqlVal(v)}`;
        }).join(' AND ');
        await execute(tab.connectionId!, `DELETE FROM ${tableName} WHERE ${where}`);
      }
      selectedRows = new Set();
      lastClickedRow = null;
      deleteConfirm = false;
      await tabStore.execute(tab.id);
    } catch (e: any) {
      deleteError = e?.message ?? String(e);
      deleteConfirm = false;
    } finally {
      isDeleting = false;
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

    {#if selectedRows.size > 0 && result}
      {#if deleteConfirm}
        <span class="text-xs" style="color: #f87171">Delete {selectedRows.size} row{selectedRows.size !== 1 ? 's' : ''}?</span>
        <button
          onclick={deleteSelectedRows}
          disabled={isDeleting}
          class="px-2 py-0.5 rounded text-xs font-medium cursor-pointer"
          style="background-color: #ef4444; color: #fff; border: none; opacity: {isDeleting ? '0.6' : '1'}"
        >{isDeleting ? 'Deleting…' : 'Confirm'}</button>
        <button
          onclick={() => deleteConfirm = false}
          class="px-2 py-0.5 rounded text-xs cursor-pointer"
          style="color: {colors.text.muted}; border: 1px solid {colors.border.primary}"
          onmouseenter={e => (e.currentTarget as HTMLElement).style.color = colors.text.primary}
          onmouseleave={e => (e.currentTarget as HTMLElement).style.color = colors.text.muted}
        >Cancel</button>
      {:else}
        <button
          onclick={() => deleteConfirm = true}
          class="px-2 py-0.5 rounded text-xs cursor-pointer transition-colors"
          style="color: #f87171; border: 1px solid #f8717144"
          onmouseenter={e => { (e.currentTarget as HTMLElement).style.backgroundColor = '#7f1d1d44' }}
          onmouseleave={e => { (e.currentTarget as HTMLElement).style.backgroundColor = 'transparent' }}
        >Delete {selectedRows.size}</button>
      {/if}
    {/if}

    {#if result}
      <div class="ml-auto flex items-center gap-1">
        <!-- Save / Discard pending changes -->
        {#if pendingChanges.size > 0}
          <button
            onclick={discardChanges}
            class="px-2 py-1 rounded text-xs cursor-pointer transition-colors"
            style="color: {colors.text.muted}; border: 1px solid {colors.border.primary}"
            onmouseenter={e => (e.currentTarget as HTMLElement).style.color = colors.text.primary}
            onmouseleave={e => (e.currentTarget as HTMLElement).style.color = colors.text.muted}
          >Discard</button>
          <button
            onclick={saveAllChanges}
            disabled={isSaving}
            class="px-2 py-1 rounded text-xs font-medium cursor-pointer"
            style="background-color: {colors.accent.primary}; color: #fff; border: none; opacity: {isSaving ? '0.6' : '1'}"
          >{isSaving ? 'Saving…' : `Save ${pendingChanges.size} change${pendingChanges.size !== 1 ? 's' : ''}`}</button>
        {/if}

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
        <thead class="sticky top-0 z-10" style="background-color: {colors.background.secondary}">
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
              {#each row as _cell, j}
                {@const isPending = pendingChanges.has(cellKey(i, j))}
                {@const displayVal = getCellDisplayValue(i, j)}
                <td
                  class="whitespace-nowrap max-w-xs font-mono relative"
                  style="border-bottom: 1px solid {colors.border.primary}; padding: 0; background-color: {isPending ? (failedCells.has(cellKey(i, j)) ? '#7f1d1d33' : '#78350f33') : 'transparent'}"
                  ondblclick={e => { e.stopPropagation(); startEdit(i, j); }}
                >
                  {#if editingCell?.row === i && editingCell?.col === j}
                    <input
                      type="text"
                      class="w-full px-3 py-1.5 font-mono text-xs outline-none"
                      style="background-color: {colors.background.tertiary}; color: {colors.text.primary}; border: 1px solid {colors.accent.primary}; box-sizing: border-box"
                      value={editingCell.value}
                      oninput={e => { if (editingCell) editingCell.value = (e.target as HTMLInputElement).value; }}
                      onkeydown={e => { if (e.key === 'Enter') { e.preventDefault(); commitEdit(); } if (e.key === 'Escape') cancelEdit(); e.stopPropagation(); }}
                      onblur={commitEdit}
                      bind:this={editInputEl}
                    />
                  {:else}
                    <span class="block px-3 py-1.5 truncate" style="color: {displayVal === null ? colors.text.muted : isPending ? (failedCells.has(cellKey(i, j)) ? '#fca5a5' : '#fbbf24') : colors.text.primary}; font-style: {displayVal === null ? 'italic' : 'normal'}">
                      {displayVal === null ? 'NULL' : String(displayVal)}
                    </span>
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

  <!-- Save success toast -->
  {#if saveSuccess}
    <div class="absolute bottom-4 right-4 z-50 px-3 py-2 rounded text-xs" style="background-color: #14532d; color: #86efac; border: 1px solid #22c55e">
      Changes saved successfully
    </div>
  {/if}

  <!-- Delete error toast -->
  {#if deleteError}
    <div class="absolute bottom-4 right-4 z-50 px-3 py-2 rounded text-xs max-w-sm" style="background-color: #7f1d1d; color: #fca5a5; border: 1px solid #ef4444">
      <span class="font-semibold">Delete failed: </span>{deleteError}
      <button onclick={() => deleteError = null} class="ml-2 underline cursor-pointer">dismiss</button>
    </div>
  {/if}

  <!-- Save error toast -->
  {#if saveError}
    <div class="absolute bottom-4 right-4 z-50 px-3 py-2 rounded text-xs max-w-sm" style="background-color: #7f1d1d; color: #fca5a5; border: 1px solid #ef4444">
      <span class="font-semibold">Save failed: </span>{saveError}
      <button onclick={() => saveError = null} class="ml-2 underline cursor-pointer">dismiss</button>
    </div>
  {/if}

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
