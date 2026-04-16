import { addConnection, getConnections, reconnect, disconnectDB, removeConnection, updateConnection } from '$lib/api/connections'
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

  async loadSaved(): Promise<void> {
    const saved = await getConnections()
    this.list = saved.map(s => ({
      id: s.ID,
      name: s.Name,
      config: s.Config as ConnectionConfig,
      connected: s.Connected,
    }))
  }

  async add(name: string, config: ConnectionConfig): Promise<string> {
    if (this.list.some(c => c.name === name)) {
      throw new Error(`A connection named "${name}" already exists`)
    }
    const id = await addConnection(name, config)
    this.list.push({ id, name, config, connected: true })
    this.activeId = id
    return id
  }

  async connect(id: string): Promise<void> {
    await reconnect(id)
    const conn = this.list.find(c => c.id === id)
    if (conn) conn.connected = true
  }

  async disconnect(id: string): Promise<void> {
    await disconnectDB(id)
    const conn = this.list.find(c => c.id === id)
    if (conn) conn.connected = false
    if (this.activeId === id) this.activeId = null
  }

  async remove(id: string): Promise<void> {
    await removeConnection(id)
    this.list = this.list.filter(c => c.id !== id)
    if (this.activeId === id) this.activeId = null
  }

  async update(id: string, name: string, config: ConnectionConfig): Promise<void> {
    await updateConnection(id, name, config)
    const conn = this.list.find(c => c.id === id)
    if (conn) {
      conn.name = name
      conn.config = config
      conn.connected = true
    }
  }
}

export const connectionStore = new ConnectionStore()
