import type { Column } from '$lib/api/schema'

export type DriverType = 'postgresql' | 'mssql' | 'mysql' | 'sqlite'

export type ColumnEdit = {
  id: string
  original: Column | null // null = new column
  name: string
  type: string
  nullable: boolean
  default: string
  deleted: boolean
}

function q(driver: DriverType, name: string): string {
  if (driver === 'mssql') return `[${name}]`
  if (driver === 'mysql') return `\`${name}\``
  return `"${name}"`
}

function tableRef(driver: DriverType, schema: string, table: string): string {
  if (driver === 'sqlite' || !schema) return q(driver, table)
  return `${q(driver, schema)}.${q(driver, table)}`
}

export function generateDDL(
  driver: DriverType,
  schema: string,
  originalTableName: string,
  newTableName: string,
  edits: ColumnEdit[]
): { statements: string[]; warnings: string[] } {
  const statements: string[] = []
  const warnings: string[] = []
  const origRef = tableRef(driver, schema, originalTableName)

  // Rename table
  if (newTableName !== originalTableName) {
    if (driver === 'mssql') {
      statements.push(`EXEC sp_rename '${schema ? schema + '.' : ''}${originalTableName}', '${newTableName}'`)
    } else if (driver === 'mysql') {
      statements.push(`RENAME TABLE ${origRef} TO ${tableRef(driver, schema, newTableName)}`)
    } else {
      statements.push(`ALTER TABLE ${origRef} RENAME TO ${q(driver, newTableName)}`)
    }
  }

  const curRef = tableRef(driver, schema, newTableName || originalTableName)

  for (const edit of edits) {
    // Drop column
    if (edit.deleted && edit.original) {
      if (driver === 'sqlite') {
        warnings.push(`SQLite: DROP COLUMN requires SQLite ≥ 3.35`)
      }
      statements.push(`ALTER TABLE ${curRef} DROP COLUMN ${q(driver, edit.original.Name)}`)
      continue
    }

    // Add new column
    if (!edit.original && !edit.deleted) {
      const nullPart = edit.nullable ? '' : ' NOT NULL'
      const defPart = edit.default ? ` DEFAULT ${edit.default}` : ''
      const kw = driver === 'mssql' ? 'ADD' : 'ADD COLUMN'
      statements.push(`ALTER TABLE ${curRef} ${kw} ${q(driver, edit.name)} ${edit.type}${nullPart}${defPart}`)
      continue
    }

    if (!edit.original) continue

    const orig = edit.original
    const nameChanged = edit.name !== orig.Name
    const typeChanged = edit.type !== orig.Type
    const nullableChanged = edit.nullable !== orig.Nullable
    const defaultChanged = edit.default !== (orig.Default ?? '')

    if (!nameChanged && !typeChanged && !nullableChanged && !defaultChanged) continue

    // Rename column
    if (nameChanged) {
      if (driver === 'mssql') {
        statements.push(`EXEC sp_rename '${schema ? schema + '.' : ''}${originalTableName}.${orig.Name}', '${edit.name}', 'COLUMN'`)
      } else if (driver === 'sqlite') {
        statements.push(`ALTER TABLE ${curRef} RENAME COLUMN ${q(driver, orig.Name)} TO ${q(driver, edit.name)}`)
      } else {
        statements.push(`ALTER TABLE ${curRef} RENAME COLUMN ${q(driver, orig.Name)} TO ${q(driver, edit.name)}`)
      }
    }

    const col = q(driver, edit.name)

    // Type / nullable
    if (typeChanged || nullableChanged) {
      if (driver === 'sqlite') {
        warnings.push(`SQLite: Cannot change type or nullability of "${orig.Name}" — recreate table manually`)
      } else if (driver === 'postgresql') {
        if (typeChanged) {
          statements.push(`ALTER TABLE ${curRef} ALTER COLUMN ${col} TYPE ${edit.type}`)
        }
        if (nullableChanged) {
          statements.push(`ALTER TABLE ${curRef} ALTER COLUMN ${col} ${edit.nullable ? 'DROP NOT NULL' : 'SET NOT NULL'}`)
        }
      } else if (driver === 'mssql') {
        statements.push(`ALTER TABLE ${curRef} ALTER COLUMN ${col} ${edit.type}${edit.nullable ? ' NULL' : ' NOT NULL'}`)
      } else if (driver === 'mysql') {
        const defPart = edit.default ? ` DEFAULT ${edit.default}` : ''
        statements.push(`ALTER TABLE ${curRef} MODIFY COLUMN ${col} ${edit.type}${edit.nullable ? ' NULL' : ' NOT NULL'}${defPart}`)
      }
    }

    // Default (PostgreSQL; MySQL already handled above in MODIFY)
    if (defaultChanged && driver === 'postgresql') {
      statements.push(edit.default
        ? `ALTER TABLE ${curRef} ALTER COLUMN ${col} SET DEFAULT ${edit.default}`
        : `ALTER TABLE ${curRef} ALTER COLUMN ${col} DROP DEFAULT`)
    }
    if (defaultChanged && driver === 'mysql' && !typeChanged && !nullableChanged) {
      statements.push(edit.default
        ? `ALTER TABLE ${curRef} ALTER COLUMN ${col} SET DEFAULT ${edit.default}`
        : `ALTER TABLE ${curRef} ALTER COLUMN ${col} DROP DEFAULT`)
    }
    if (defaultChanged && driver === 'sqlite') {
      warnings.push(`SQLite: Cannot change default of "${orig.Name}" — recreate table manually`)
    }
  }

  return { statements, warnings }
}
