import { addConnection, disconnectDB } from '$lib/api/connections'
import type { ConnectionConfig } from '$lib/api/connections'

export type SavedConnection = {
  id: string
  name: string
  config: ConnectionConfig
  connected: boolean
}

class ConnectionStore {
  list = $state<SavedConnection[]>([])
  activeId = $state<string | null>(null)

  get active(): SavedConnection | undefined {
    return this.list.find(c => c.id === this.activeId)
  }

  async add(name: string, config: ConnectionConfig): Promise<string> {
    if (this.list.some(c => c.name === name)) {
      throw new Error(`A connection named "${name}" already exists`)
    }
    const id = await addConnection(config)
    this.list.push({ id, name, config, connected: true })
    this.activeId = id
    return id
  }

  async disconnect(id: string): Promise<void> {
    await disconnectDB(id)
    const conn = this.list.find(c => c.id === id)
    if (conn) conn.connected = false
    if (this.activeId === id) this.activeId = null
  }

  remove(id: string): void {
    this.list = this.list.filter(c => c.id !== id)
    if (this.activeId === id) this.activeId = null
  }
}

export const connectionStore = new ConnectionStore()
