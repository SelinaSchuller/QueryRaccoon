import { Execute } from '$wailsjs/go/bindings/QueryService'

export type QueryResult = {
  Columns: string[]
  Rows: any[][]
}

export async function execute(connectionId: string, sql: string): Promise<QueryResult> {
  return Execute(connectionId, sql)
}
