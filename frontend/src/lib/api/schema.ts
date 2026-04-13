import { GetDatabases, GetSchemas, GetTables, GetViews, GetColumns } from '$wailsjs/go/bindings/SchemaService'

export type Column = {
  Name: string
  Type: string
  Nullable: boolean
  Default: string
}

export async function getDatabases(connectionId: string): Promise<string[]> {
  return GetDatabases(connectionId)
}

export async function getSchemas(connectionId: string, database: string): Promise<string[]> {
  return GetSchemas(connectionId, database)
}

export async function getTables(connectionId: string, schemaName: string): Promise<string[]> {
  return GetTables(connectionId, schemaName)
}

export async function getViews(connectionId: string, schemaName: string): Promise<string[]> {
  return GetViews(connectionId, schemaName)
}

export async function getColumns(connectionId: string, schemaName: string, table: string): Promise<Column[]> {
  return GetColumns(connectionId, schemaName, table)
}
