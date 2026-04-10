<script lang="ts">
  import Sidebar from "$lib/components/layout/Sidebar.svelte";
  import TabBar from "$lib/components/layout/TabBar.svelte";
  import StatusBar from "$lib/components/layout/StatusBar.svelte";
  import QueryEditor from "$lib/components/editor/QueryEditor.svelte";
  import ResultsPane from "$lib/components/editor/ResultsPane.svelte";
</script>

<div class="flex flex-col h-screen bg-[#111113] text-[#f4f4f5] overflow-hidden">
  <TabBar />

  <div class="flex flex-1 overflow-hidden">
    <Sidebar />

    <div class="flex flex-col flex-1 overflow-hidden">
      <div class="flex-1 overflow-hidden">
        <QueryEditor />
      </div>
      <div class="h-px bg-[#3f3f46]"></div>
      <div class="h-64 overflow-hidden shrink-0">
        <ResultsPane />
      </div>
    </div>
  </div>

  <StatusBar />
</div>
