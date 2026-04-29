<script lang="ts">
  import { onMount } from 'svelte'
  import Sidebar from "$lib/components/layout/Sidebar.svelte";
  import TabBar from "$lib/components/layout/TabBar.svelte";
  import StatusBar from "$lib/components/layout/StatusBar.svelte";
  import QueryEditor from "$lib/components/editor/QueryEditor.svelte";
  import ResultsPane from "$lib/components/editor/ResultsPane.svelte";
  import SchemaEditor from "$lib/components/editor/SchemaEditor.svelte";
  import { colors } from "$lib/colors";
  import { connectionStore } from "$lib/stores/connections.svelte";
  import { tabStore } from "$lib/stores/tabs.svelte";

  let activeTab = $derived(tabStore.active);

  onMount(async () => {
    await connectionStore.loadSaved()
  })

  const EDITOR_HEADER_H = 40
  const RESULTS_HEADER_H = 37

  let resultsHeight = $state(220)
  let editorCollapsed = $state(false)
  let resultsCollapsed = $state(false)

  let dragging = false
  let dragStartY = 0
  let dragStartHeight = 0

  function startDrag(e: MouseEvent) {
    dragging = true
    dragStartY = e.clientY
    dragStartHeight = resultsHeight
    document.body.style.cursor = 'ns-resize'
    document.body.style.userSelect = 'none'
    e.preventDefault()
  }

  function onMouseMove(e: MouseEvent) {
    if (!dragging) return
    const delta = dragStartY - e.clientY
    resultsHeight = Math.max(60, Math.min(700, dragStartHeight + delta))
  }

  function onMouseUp() {
    if (!dragging) return
    dragging = false
    document.body.style.cursor = ''
    document.body.style.userSelect = ''
  }

  function editorStyle() {
    if (editorCollapsed) return `flex: 0 0 ${EDITOR_HEADER_H}px; overflow: hidden`
    if (resultsCollapsed) return 'flex: 1 1 0; min-height: 0; overflow: hidden'
    return 'flex: 1 1 0; min-height: 0; overflow: hidden'
  }

  function resultsStyle() {
    if (resultsCollapsed) return `flex: 0 0 ${RESULTS_HEADER_H}px; overflow: hidden`
    if (editorCollapsed) return 'flex: 1 1 0; min-height: 0; overflow: hidden'
    return `flex: 0 0 ${resultsHeight}px; overflow: hidden`
  }
</script>

<svelte:window onmousemove={onMouseMove} onmouseup={onMouseUp} />

<div class="flex flex-col h-screen overflow-hidden" style="background-color: {colors.background.primary}; color: {colors.text.primary}">
  <TabBar />

  <div class="flex flex-1 overflow-hidden">
    <Sidebar />

    <div class="flex flex-col flex-1 overflow-hidden">
      {#if activeTab?.kind === 'schema'}
        <SchemaEditor />
      {:else}
        <div style={editorStyle()}>
          <QueryEditor
            collapsed={editorCollapsed}
            onToggleCollapse={() => editorCollapsed = !editorCollapsed}
          />
        </div>

        {#if !editorCollapsed && !resultsCollapsed}
          <button
            type="button"
            aria-label="Resize panels"
            class="h-1.5 w-full shrink-0 cursor-ns-resize transition-colors border-none p-0"
            style="background-color: {colors.border.primary}"
            onmousedown={startDrag}
            onmouseenter={e => (e.currentTarget as HTMLElement).style.backgroundColor = colors.accent.primary}
            onmouseleave={e => (e.currentTarget as HTMLElement).style.backgroundColor = colors.border.primary}
          ></button>
        {:else}
          <div class="h-px shrink-0" style="background-color: {colors.border.primary}"></div>
        {/if}

        <div style={resultsStyle()}>
          <ResultsPane
            collapsed={resultsCollapsed}
            onToggleCollapse={() => resultsCollapsed = !resultsCollapsed}
          />
        </div>
      {/if}
    </div>
  </div>

  <StatusBar />
</div>
