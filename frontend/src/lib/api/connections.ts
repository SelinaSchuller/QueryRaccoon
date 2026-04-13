import {
  AddConnection,
  Connect,
  Disconnect,
  GetConnections,
} from '$wailsjs/go/bindings/ConnectionService'
import type { connections } from '$wailsjs/go/models'

export type DriverType = 'postgresql' | 'mysql' | 'sqlite' | 'mssql'

export type ConnectionConfig = {
  Host: string
  Port: number
  User: string
  Password: string
  Database: string
  DriverType: DriverType
}

export type ConnectionInfo = connections.ConnectionInfo

export async function addConnection(name: string, config: ConnectionConfig): Promise<string> {
  return AddConnection(name, config)
}

export async function getConnections(): Promise<ConnectionInfo[]> {
  return GetConnections()
}

export async function reconnect(id: string): Promise<void> {
  return Connect(id)
}

export async function disconnectDB(id: string): Promise<void> {
  return Disconnect(id)
}
