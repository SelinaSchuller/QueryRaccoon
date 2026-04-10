<script lang="ts">
  import { connectionStore } from "$lib/stores/connections.svelte";
  import type { DriverType } from "$lib/api/connections";
  import { colors } from "$lib/colors";

  type Props = { onclose: () => void };
  let { onclose }: Props = $props();

  let name = $state("");
  let host = $state("localhost");
  let port = $state(5432);
  let user = $state("postgres");
  let password = $state("");
  let database = $state("postgres");
  let driverType = $state<DriverType>("postgresql");
  let error = $state("");
  let loading = $state(false);

  const defaultPorts: Record<DriverType, number> = {
    postgresql: 5432,
    mysql: 3306,
    sqlite: 0,
    mssql: 1433,
  };

  const defaultUsers: Record<DriverType, string> = {
    postgresql: 'postgres',
    mysql: 'root',
    sqlite: '',
    mssql: '',
  }
  const defaultDatabases: Record<DriverType, string> = {
    postgresql: 'postgres',
    mysql: '',
    sqlite: '',
    mssql: 'master',
  }

  function onDriverChange() {
    port = defaultPorts[driverType];
    user = defaultUsers[driverType];
    database = defaultDatabases[driverType];
  }

  async function submit() {
    if (!name.trim()) { error = "Name is required"; return }
    if (!database.trim()) { error = "Database is required"; return }
    error = "";
    loading = true;
    try {
      await connectionStore.add(name, {
        Host: host,
        Port: port,
        User: user,
        Password: password,
        Database: database,
        DriverType: driverType,
      });
      onclose();
    } catch (e: any) {
      error = e?.message ?? String(e);
    } finally {
      loading = false;
    }
  }
</script>

<div class="fixed inset-0 flex items-center justify-center z-50" style="background-color: rgba(0,0,0,0.6)" role="dialog" aria-modal="true">
  <div class="rounded-lg w-[480px] shadow-2xl" style="background-color: {colors.background.tertiary}; border: 1px solid {colors.border.primary}">

    <div class="flex items-center justify-between px-6 py-4" style="border-bottom: 1px solid {colors.border.primary}">
      <h2 class="font-semibold text-sm" style="color: {colors.text.primary}">New Connection</h2>
      <button onclick={onclose} class="transition-colors cursor-pointer" style="color: {colors.text.muted}" onmouseenter={e => (e.currentTarget as HTMLElement).style.color = colors.text.primary} onmouseleave={e => (e.currentTarget as HTMLElement).style.color = colors.text.muted}>✕</button>
    </div>

    <div class="px-6 py-5 space-y-4">
      <div>
        <label for="conn-name" class="block text-xs mb-1" style="color: {colors.text.muted}">Connection Name</label>
        <input id="conn-name" bind:value={name} placeholder="My Database" autocapitalize="off" autocomplete="off" spellcheck={false} class="w-full rounded-md px-2.5 py-1.5 text-xs outline-none box-border" style="background-color: {colors.background.secondary}; border: 1px solid {colors.border.primary}; color: {colors.text.primary}" onfocus={e => (e.currentTarget as HTMLElement).style.borderColor = colors.accent.primary} onblur={e => (e.currentTarget as HTMLElement).style.borderColor = colors.border.primary} />
      </div>

      <div>
        <label for="conn-driver" class="block text-xs mb-1" style="color: {colors.text.muted}">Database Type</label>
        <select id="conn-driver" bind:value={driverType} onchange={onDriverChange} class="w-full rounded-md px-2.5 py-1.5 text-xs outline-none box-border appearance-none" style="background-color: {colors.background.secondary}; border: 1px solid {colors.border.primary}; color: {colors.text.primary}" onfocus={e => (e.currentTarget as HTMLElement).style.borderColor = colors.accent.primary} onblur={e => (e.currentTarget as HTMLElement).style.borderColor = colors.border.primary}>
          <option value="postgresql">PostgreSQL</option>
          <option value="mysql">MySQL</option>
          <option value="sqlite">SQLite</option>
          <option value="mssql">SQL Server</option>
        </select>
      </div>

      {#if driverType === "sqlite"}
        <div>
          <label for="conn-db-path" class="block text-xs mb-1" style="color: {colors.text.muted}">Database File Path</label>
          <input id="conn-db-path" bind:value={database} placeholder="/path/to/database.db" autocapitalize="off" autocomplete="off" spellcheck={false} class="w-full rounded-md px-2.5 py-1.5 text-xs outline-none box-border" style="background-color: {colors.background.secondary}; border: 1px solid {colors.border.primary}; color: {colors.text.primary}" onfocus={e => (e.currentTarget as HTMLElement).style.borderColor = colors.accent.primary} onblur={e => (e.currentTarget as HTMLElement).style.borderColor = colors.border.primary} />
        </div>
      {:else}
        <div class="flex gap-3">
          <div class="flex-1">
            <label for="conn-host" class="block text-xs mb-1" style="color: {colors.text.muted}">Host</label>
            <input id="conn-host" bind:value={host} placeholder="localhost" autocapitalize="off" autocomplete="off" spellcheck={false} class="w-full rounded-md px-2.5 py-1.5 text-xs outline-none box-border" style="background-color: {colors.background.secondary}; border: 1px solid {colors.border.primary}; color: {colors.text.primary}" onfocus={e => (e.currentTarget as HTMLElement).style.borderColor = colors.accent.primary} onblur={e => (e.currentTarget as HTMLElement).style.borderColor = colors.border.primary} />
          </div>
          <div class="w-24">
            <label for="conn-port" class="block text-xs mb-1" style="color: {colors.text.muted}">Port</label>
            <input id="conn-port" bind:value={port} type="number" class="w-full rounded-md px-2.5 py-1.5 text-xs outline-none box-border [&::-webkit-inner-spin-button]:appearance-none [&::-webkit-outer-spin-button]:appearance-none" style="background-color: {colors.background.secondary}; border: 1px solid {colors.border.primary}; color: {colors.text.primary}" onfocus={e => (e.currentTarget as HTMLElement).style.borderColor = colors.accent.primary} onblur={e => (e.currentTarget as HTMLElement).style.borderColor = colors.border.primary} onkeydown={e => { if (!/[\d]/.test(e.key) && !['Backspace','Delete','ArrowLeft','ArrowRight','Tab'].includes(e.key)) e.preventDefault() }} />
          </div>
        </div>

        <div class="flex gap-3">
          <div class="flex-1">
            <label for="conn-user" class="block text-xs mb-1" style="color: {colors.text.muted}">Username</label>
            <input id="conn-user" bind:value={user} autocapitalize="off" autocomplete="off" spellcheck={false} class="w-full rounded-md px-2.5 py-1.5 text-xs outline-none box-border" style="background-color: {colors.background.secondary}; border: 1px solid {colors.border.primary}; color: {colors.text.primary}" onfocus={e => (e.currentTarget as HTMLElement).style.borderColor = colors.accent.primary} onblur={e => (e.currentTarget as HTMLElement).style.borderColor = colors.border.primary} />
          </div>
          <div class="flex-1">
            <label for="conn-pass" class="block text-xs mb-1" style="color: {colors.text.muted}">Password</label>
            <input id="conn-pass" bind:value={password} type="password" class="w-full rounded-md px-2.5 py-1.5 text-xs outline-none box-border" style="background-color: {colors.background.secondary}; border: 1px solid {colors.border.primary}; color: {colors.text.primary}" onfocus={e => (e.currentTarget as HTMLElement).style.borderColor = colors.accent.primary} onblur={e => (e.currentTarget as HTMLElement).style.borderColor = colors.border.primary} />
          </div>
        </div>

        <div>
          <label for="conn-db" class="block text-xs mb-1" style="color: {colors.text.muted}">Database</label>
          <input id="conn-db" bind:value={database} autocapitalize="off" autocomplete="off" spellcheck={false} class="w-full rounded-md px-2.5 py-1.5 text-xs outline-none box-border" style="background-color: {colors.background.secondary}; border: 1px solid {colors.border.primary}; color: {colors.text.primary}" onfocus={e => (e.currentTarget as HTMLElement).style.borderColor = colors.accent.primary} onblur={e => (e.currentTarget as HTMLElement).style.borderColor = colors.border.primary} />
        </div>
      {/if}

      {#if error}
        <p class="text-xs" style="color: #f87171">{error}</p>
      {/if}
    </div>

    <div class="flex justify-end gap-2 px-6 py-4" style="border-top: 1px solid {colors.border.primary}">
      <button onclick={onclose} class="px-3.5 py-1.5 rounded-md text-xs cursor-pointer transition-colors" style="color: {colors.text.muted}; border: 1px solid {colors.border.primary}; background: transparent" onmouseenter={e => (e.currentTarget as HTMLElement).style.color = colors.text.primary} onmouseleave={e => (e.currentTarget as HTMLElement).style.color = colors.text.muted}>
        Cancel
      </button>
      <button onclick={submit} disabled={loading} class="px-3.5 py-1.5 rounded-md text-xs font-medium cursor-pointer transition-colors disabled:opacity-50 disabled:cursor-not-allowed" style="background-color: {colors.accent.primary}; color: #fff; border: none" onmouseenter={e => { if (!loading) (e.currentTarget as HTMLElement).style.backgroundColor = colors.accent.hover }} onmouseleave={e => (e.currentTarget as HTMLElement).style.backgroundColor = colors.accent.primary}>
        {loading ? "Connecting…" : "Connect"}
      </button>
    </div>
  </div>
</div>
