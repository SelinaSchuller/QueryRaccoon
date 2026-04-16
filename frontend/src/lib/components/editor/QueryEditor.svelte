<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { tabStore } from "$lib/stores/tabs.svelte";
  import { connectionStore } from "$lib/stores/connections.svelte";
  import { colors } from "$lib/colors";
  import { EditorView, keymap, placeholder } from '@codemirror/view';
  import { EditorState } from '@codemirror/state';
  import { sql } from '@codemirror/lang-sql';
  import { defaultKeymap, history, historyKeymap } from '@codemirror/commands';
  import { syntaxHighlighting, HighlightStyle } from '@codemirror/language';
  import { tags } from '@lezer/highlight';

  type Props = { collapsed?: boolean; onToggleCollapse?: () => void }
  let { collapsed = false, onToggleCollapse }: Props = $props()

  let tab = $derived(tabStore.active);
  let activeConn = $derived(connectionStore.active);

  let editorEl: HTMLDivElement;
  let view: EditorView | undefined;

  const raccoonTheme = EditorView.theme({
    '&': {
      height: '100%',
      fontSize: '13px',
      backgroundColor: '#111113',
    },
    '.cm-scroller': {
      fontFamily: 'ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace',
      lineHeight: '1.7',
      padding: '12px 0',
      overflow: 'auto',
    },
    '.cm-content': { caretColor: '#10b981', padding: '0 16px' },
    '.cm-cursor': { borderLeftColor: '#10b981' },
    '.cm-activeLine': { backgroundColor: '#18181b' },
    '.cm-activeLineGutter': { backgroundColor: '#18181b' },
    '.cm-selectionBackground, ::selection': { backgroundColor: '#10b98133 !important' },
    '.cm-gutters': {
      backgroundColor: '#111113',
      borderRight: '1px solid #3f3f46',
      color: '#52525b',
    },
    '.cm-lineNumbers .cm-gutterElement': { padding: '0 12px 0 8px' },
    '.cm-foldGutter': { width: '0' },
    '.cm-placeholder': { color: '#52525b' },
    '&.cm-focused .cm-selectionBackground': { backgroundColor: '#10b98133' },
    '&.cm-focused': { outline: 'none' },
  }, { dark: true });

  const raccoonHighlight = HighlightStyle.define([
    { tag: tags.keyword,          color: '#10b981', fontWeight: '600' },
    { tag: tags.operatorKeyword,  color: '#10b981', fontWeight: '600' },
    { tag: tags.typeName,         color: '#67e8f9' },
    { tag: tags.string,           color: '#fbbf24' },
    { tag: tags.number,           color: '#c084fc' },
    { tag: tags.bool,             color: '#c084fc' },
    { tag: tags.null,             color: '#c084fc' },
    { tag: tags.comment,          color: '#52525b', fontStyle: 'italic' },
    { tag: tags.operator,         color: '#f4f4f5' },
    { tag: tags.punctuation,      color: '#a1a1aa' },
    { tag: tags.name,             color: '#f4f4f5' },
    { tag: tags.special(tags.string), color: '#fbbf24' },
  ]);

  const runKeymap = keymap.of([{
    key: 'Mod-Enter',
    run: () => {
      if (tab) tabStore.execute(tab.id);
      return true;
    }
  }]);

  function createView(initialDoc: string) {
    view = new EditorView({
      state: EditorState.create({
        doc: initialDoc,
        extensions: [
          history(),
          keymap.of([...defaultKeymap, ...historyKeymap]),
          runKeymap,
          sql(),
          syntaxHighlighting(raccoonHighlight),
          raccoonTheme,
          placeholder('SELECT * FROM ...'),
          EditorView.lineWrapping,
          EditorView.updateListener.of(update => {
            if (update.docChanged && tab) {
              tabStore.updateSQL(tab.id, update.state.doc.toString());
            }
          }),
        ],
      }),
      parent: editorEl,
    });
  }

  // Recreate editor when active tab changes
  let lastTabId = $state<string | undefined>(undefined);

  $effect(() => {
    const currentTab = tab;
    if (!editorEl) return;

    if (currentTab?.id !== lastTabId) {
      lastTabId = currentTab?.id;
      view?.destroy();
      createView(currentTab?.sql ?? '');
      return;
    }

    // Sync external SQL changes (e.g. openDataTab auto-fills SQL)
    if (view && currentTab) {
      const editorContent = view.state.doc.toString();
      if (editorContent !== currentTab.sql) {
        view.dispatch({
          changes: { from: 0, to: editorContent.length, insert: currentTab.sql ?? '' },
        });
      }
    }
  });

  onMount(() => {
    createView(tab?.sql ?? '');
    lastTabId = tab?.id;
  });

  onDestroy(() => view?.destroy());
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

  <div
    bind:this={editorEl}
    class="flex-1 min-h-0 overflow-hidden"
    style="display: {collapsed ? 'none' : 'block'}"
  ></div>
</div>
