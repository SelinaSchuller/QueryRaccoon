import { execute } from '$lib/api/query'
import type { QueryResult } from '$lib/api/query'

export type Tab = {
  id: string
  name: string
  connectionId: string | null
  sql: string
  result: QueryResult | null
  isExecuting: boolean
  error: string | null
  executionTime: number | null
}

function makeTab(overrides: Partial<Tab> = {}): Tab {
  return {
    id: crypto.randomUUID(),
    name: 'Query',
    connectionId: null,
    sql: '',
    result: null,
    isExecuting: false,
    error: null,
    executionTime: null,
    ...overrides,
  }
}

class TabStore {
  list = $state<Tab[]>([makeTab()])
  activeId = $state<string>(this.list[0].id)

  get active(): Tab | undefined {
    return this.list.find(t => t.id === this.activeId)
  }

  add(): void {
    const tab = makeTab()
    this.list.push(tab)
    this.activeId = tab.id
  }

  close(id: string): void {
    if (this.list.length === 1) return
    const idx = this.list.findIndex(t => t.id === id)
    this.list = this.list.filter(t => t.id !== id)
    if (this.activeId === id) {
      this.activeId = this.list[Math.max(0, idx - 1)].id
    }
  }

  setActive(id: string): void {
    this.activeId = id
  }

  updateSQL(id: string, sql: string): void {
    const tab = this.list.find(t => t.id === id)
    if (tab) tab.sql = sql
  }

  setConnection(id: string, connectionId: string): void {
    const tab = this.list.find(t => t.id === id)
    if (tab) tab.connectionId = connectionId
  }

  async execute(id: string): Promise<void> {
    const tab = this.list.find(t => t.id === id)
    if (!tab || !tab.connectionId || !tab.sql.trim()) return

    tab.isExecuting = true
    tab.error = null
    tab.result = null

    const start = Date.now()
    try {
      tab.result = await execute(tab.connectionId, tab.sql)
      tab.executionTime = Date.now() - start
    } catch (e: any) {
      tab.error = e?.message ?? String(e)
    } finally {
      tab.isExecuting = false
    }
  }
}

export const tabStore = new TabStore()
