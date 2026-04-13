import { getDatabases, getSchemas, getTables, getViews, getColumns } from '$lib/api/schema'
import type { Column } from '$lib/api/schema'

export type TreeNode = {
  id: string
  label: string
  type: 'database' | 'schema' | 'folder' | 'table' | 'view' | 'column'
  folderKind?: 'tables' | 'views'
  schemaName?: string
  children: TreeNode[]
  loaded: boolean
  expanded: boolean
  meta?: Column
}

class SchemaStore {
  trees = $state<Record<string, TreeNode[]>>({})
  loading = $state<Record<string, boolean>>({})

  async loadDatabases(connectionId: string): Promise<void> {
    this.loading[connectionId] = true
    try {
      const dbs = await getDatabases(connectionId)
      this.trees[connectionId] = dbs.map(db => ({
        id: `${connectionId}/${db}`,
        label: db,
        type: 'database',
        children: [],
        loaded: false,
        expanded: false,
      }))
    } finally {
      this.loading[connectionId] = false
    }
  }

  async refresh(connectionId: string): Promise<void> {
    delete this.trees[connectionId]
    await this.loadDatabases(connectionId)
  }

  async refreshNode(connectionId: string, node: TreeNode): Promise<void> {
    node.loaded = false
    node.children = []
    if (node.expanded) {
      await this.loadChildren(connectionId, node)
    }
  }

  async expand(connectionId: string, node: TreeNode): Promise<void> {
    node.expanded = !node.expanded
    if (node.expanded && !node.loaded) {
      await this.loadChildren(connectionId, node)
    }
  }

  private async loadChildren(connectionId: string, node: TreeNode): Promise<void> {
    node.loaded = true
    if (node.type === 'database') {
      const schemas = await getSchemas(connectionId, node.label)
      node.children = schemas.map(s => ({
        id: `${node.id}/${s}`,
        label: s,
        type: 'schema',
        children: [],
        loaded: false,
        expanded: false,
      }))
    } else if (node.type === 'schema') {
      node.children = [
        {
          id: `${node.id}/tables`,
          label: 'Tables',
          type: 'folder',
          folderKind: 'tables',
          schemaName: node.label,
          children: [],
          loaded: false,
          expanded: false,
        },
        {
          id: `${node.id}/views`,
          label: 'Views',
          type: 'folder',
          folderKind: 'views',
          schemaName: node.label,
          children: [],
          loaded: false,
          expanded: false,
        },
      ]
    } else if (node.type === 'folder') {
      if (node.folderKind === 'tables') {
        const tables = await getTables(connectionId, node.schemaName!)
        node.children = tables.map(t => ({
          id: `${node.id}/${t}`,
          label: t,
          type: 'table',
          schemaName: node.schemaName,
          children: [],
          loaded: false,
          expanded: false,
        }))
      } else if (node.folderKind === 'views') {
        const views = await getViews(connectionId, node.schemaName!)
        node.children = views.map(v => ({
          id: `${node.id}/${v}`,
          label: v,
          type: 'view',
          schemaName: node.schemaName,
          children: [],
          loaded: true,
          expanded: false,
        }))
      }
    } else if (node.type === 'table') {
      const columns = await getColumns(connectionId, node.schemaName!, node.label)
      node.children = columns.map(c => ({
        id: `${node.id}/${c.Name}`,
        label: c.Name,
        type: 'column',
        children: [],
        loaded: true,
        expanded: false,
        meta: c,
      }))
    }
  }
}

export const schemaStore = new SchemaStore()
