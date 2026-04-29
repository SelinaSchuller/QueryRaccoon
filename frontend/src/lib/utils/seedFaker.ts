import { faker } from '@faker-js/faker'

export type FakerStrategy = {
  label: string
  kind: 'skip' | 'uuid' | 'fk' | 'data'
  generate: () => any
}

type FKConfig = { min: number; max: number }

// Normalise column name for matching
function norm(s: string): string {
  return s.toLowerCase().replace(/[\s_\-]/g, '')
}

function isIntegerType(type: string): boolean {
  const t = type.toLowerCase()
  return /\b(int|integer|bigint|smallint|tinyint|serial|bigserial|smallserial)\b/.test(t)
}

function isUUIDType(type: string): boolean {
  const t = type.toLowerCase()
  return t.includes('uuid') || t.includes('uniqueidentifier')
}

function isDateType(type: string): boolean {
  const t = type.toLowerCase()
  return t.includes('date') || t.includes('time') || t.includes('timestamp')
}

function isBoolType(type: string): boolean {
  const t = type.toLowerCase()
  return t.includes('bool') || t === 'bit'
}

function isNumberType(type: string): boolean {
  const t = type.toLowerCase()
  return /\b(numeric|decimal|float|double|real|money|number)\b/.test(t)
}

/**
 * Detect if a column is a FK by naming convention.
 * e.g. user_id, userId, company_id → true
 * but plain "id" → false (that's a PK)
 */
export function isFKColumn(colName: string): boolean {
  const n = norm(colName)
  return n !== 'id' && (n.endsWith('id') || n.endsWith('ids'))
}

/**
 * Detect if a column is an auto-increment primary key to skip.
 */
export function isAutoKey(colName: string, colType: string): boolean {
  const n = norm(colName)
  return n === 'id' && isIntegerType(colType)
}

/**
 * Return the referenced table name guess from a FK column name.
 * e.g. "company_id" → "companies", "user_id" → "users"
 */
export function guessRefTable(colName: string): string {
  const base = colName.toLowerCase().replace(/_?id$/i, '').replace(/id$/i, '')
  // naive pluralisation
  if (base.endsWith('y')) return base.slice(0, -1) + 'ies'
  if (base.endsWith('s')) return base
  return base + 's'
}

export function getStrategy(
  colName: string,
  colType: string,
  tableName: string,
  useLorem: boolean,
  fkConfig?: FKConfig
): FakerStrategy {
  const n = norm(colName)
  const t = norm(tableName)

  // Skip auto-increment PK
  if (isAutoKey(colName, colType)) {
    return { label: 'Skip (auto-increment)', kind: 'skip', generate: () => undefined }
  }

  // UUID PK / FK
  if (isUUIDType(colType)) {
    return { label: 'UUID', kind: 'uuid', generate: () => faker.string.uuid() }
  }

  // FK columns
  if (isFKColumn(colName)) {
    const { min = 1, max = 100 } = fkConfig ?? {}
    return {
      label: `FK → ${guessRefTable(colName)} (${min}–${max})`,
      kind: 'fk',
      generate: () => faker.number.int({ min, max }),
    }
  }

  // --- Name fields ---
  if (n === 'firstname' || n === 'fname' || n === 'givenname') {
    return { label: 'First name', kind: 'data', generate: () => faker.person.firstName() }
  }
  if (n === 'lastname' || n === 'lname' || n === 'surname' || n === 'familyname') {
    return { label: 'Last name', kind: 'data', generate: () => faker.person.lastName() }
  }
  if (n === 'fullname' || n === 'displayname' || n === 'name') {
    if (t.includes('compan') || t.includes('org') || t.includes('brand') || t.includes('vendor')) {
      return { label: 'Company name', kind: 'data', generate: () => faker.company.name() }
    }
    if (t.includes('product') || t.includes('item') || t.includes('good')) {
      return { label: 'Product name', kind: 'data', generate: () => faker.commerce.productName() }
    }
    return { label: 'Full name', kind: 'data', generate: () => faker.person.fullName() }
  }
  if (n === 'companyname' || n === 'company' || n === 'organization' || n === 'organisation' || n === 'employer') {
    return { label: 'Company name', kind: 'data', generate: () => faker.company.name() }
  }
  if (n === 'productname' || n === 'product' || n === 'productlabel') {
    return { label: 'Product name', kind: 'data', generate: () => faker.commerce.productName() }
  }

  // --- Contact ---
  if (n === 'email' || n === 'emailaddress' || n === 'mail') {
    return { label: 'Email address', kind: 'data', generate: () => faker.internet.email() }
  }
  if (n === 'phone' || n === 'phonenumber' || n === 'mobile' || n === 'cellphone' || n === 'telephone' || n === 'tel') {
    return { label: 'Phone number', kind: 'data', generate: () => faker.phone.number() }
  }

  // --- Address ---
  if (n === 'address' || n === 'streetaddress' || n === 'street') {
    return { label: 'Street address', kind: 'data', generate: () => faker.location.streetAddress() }
  }
  if (n === 'city' || n === 'town') {
    return { label: 'City', kind: 'data', generate: () => faker.location.city() }
  }
  if (n === 'state' || n === 'province' || n === 'region') {
    return { label: 'State / province', kind: 'data', generate: () => faker.location.state() }
  }
  if (n === 'country' || n === 'countryname') {
    return { label: 'Country', kind: 'data', generate: () => faker.location.country() }
  }
  if (n === 'countrycode') {
    return { label: 'Country code', kind: 'data', generate: () => faker.location.countryCode() }
  }
  if (n === 'zipcode' || n === 'postalcode' || n === 'zip' || n === 'postcode') {
    return { label: 'Zip / postal code', kind: 'data', generate: () => faker.location.zipCode() }
  }
  if (n === 'latitude' || n === 'lat') {
    return { label: 'Latitude', kind: 'data', generate: () => faker.location.latitude() }
  }
  if (n === 'longitude' || n === 'lng' || n === 'lon') {
    return { label: 'Longitude', kind: 'data', generate: () => faker.location.longitude() }
  }

  // --- Internet ---
  if (n === 'username' || n === 'login' || n === 'handle' || n === 'nickname') {
    return { label: 'Username', kind: 'data', generate: () => faker.internet.username() }
  }
  if (n === 'password' || n === 'passwordhash' || n === 'pwd') {
    return { label: 'Password hash', kind: 'data', generate: () => faker.internet.password() }
  }
  if (n === 'website' || n === 'url' || n === 'homepage' || n === 'link') {
    return { label: 'URL', kind: 'data', generate: () => faker.internet.url() }
  }
  if (n === 'avatar' || n === 'avatarurl' || n === 'profilepicture' || n === 'photo' || n === 'image' || n === 'imageurl') {
    return { label: 'Avatar URL', kind: 'data', generate: () => faker.image.avatar() }
  }
  if (n === 'ipaddress' || n === 'ip') {
    return { label: 'IP address', kind: 'data', generate: () => faker.internet.ip() }
  }

  // --- Job / company ---
  if (n === 'jobtitle' || n === 'position' || n === 'role' || n === 'occupation') {
    return { label: 'Job title', kind: 'data', generate: () => faker.person.jobTitle() }
  }
  if (n === 'department') {
    return { label: 'Department', kind: 'data', generate: () => faker.commerce.department() }
  }
  if (n === 'industry') {
    return { label: 'Industry', kind: 'data', generate: () => faker.company.buzzNoun() }
  }

  // --- Commerce ---
  if (n === 'price' || n === 'amount' || n === 'cost' || n === 'salary' || n === 'wage' || n === 'fee') {
    return { label: 'Price', kind: 'data', generate: () => parseFloat(faker.commerce.price()) }
  }
  if (n === 'currency' || n === 'currencycode') {
    return { label: 'Currency code', kind: 'data', generate: () => faker.finance.currencyCode() }
  }
  if (n === 'category' || n === 'productcategory') {
    return { label: 'Category', kind: 'data', generate: () => faker.commerce.department() }
  }
  if (n === 'sku' || n === 'barcode' || n === 'productcode') {
    return { label: 'SKU', kind: 'data', generate: () => faker.string.alphanumeric(8).toUpperCase() }
  }

  // --- Dates ---
  if (n === 'createdat' || n === 'createddate' || n === 'created') {
    return { label: 'Past date', kind: 'data', generate: () => faker.date.past().toISOString() }
  }
  if (n === 'updatedat' || n === 'updateddate' || n === 'modified' || n === 'modifiedat') {
    return { label: 'Recent date', kind: 'data', generate: () => faker.date.recent().toISOString() }
  }
  if (n === 'deletedat' || n === 'archivedat') {
    return { label: 'null (not deleted)', kind: 'data', generate: () => null }
  }
  if (n === 'birthday' || n === 'dateofbirth' || n === 'dob' || n === 'birthdate') {
    return { label: 'Birthdate', kind: 'data', generate: () => faker.date.birthdate().toISOString().split('T')[0] }
  }
  if (n === 'expirydate' || n === 'expiresat' || n === 'expirationdate') {
    return { label: 'Future date', kind: 'data', generate: () => faker.date.future().toISOString() }
  }
  if (isDateType(colType)) {
    return { label: 'Recent date', kind: 'data', generate: () => faker.date.recent().toISOString().split('T')[0] }
  }

  // --- Numbers ---
  if (n === 'age') {
    return { label: 'Age (18–80)', kind: 'data', generate: () => faker.number.int({ min: 18, max: 80 }) }
  }
  if (n === 'quantity' || n === 'qty' || n === 'stock' || n === 'count' || n === 'total') {
    return { label: 'Quantity', kind: 'data', generate: () => faker.number.int({ min: 1, max: 999 }) }
  }
  if (n === 'rating' || n === 'score' || n === 'rank') {
    return { label: 'Rating (1–5)', kind: 'data', generate: () => faker.number.int({ min: 1, max: 5 }) }
  }
  if (n === 'percentage' || n === 'percent' || n === 'discount') {
    return { label: 'Percentage', kind: 'data', generate: () => faker.number.int({ min: 0, max: 100 }) }
  }
  if (isNumberType(colType)) {
    return { label: 'Decimal number', kind: 'data', generate: () => faker.number.float({ min: 0, max: 1000, fractionDigits: 2 }) }
  }
  if (isIntegerType(colType)) {
    return { label: 'Integer', kind: 'data', generate: () => faker.number.int({ min: 1, max: 10000 }) }
  }

  // --- Boolean ---
  if (n === 'isactive' || n === 'active' || n === 'enabled' || n === 'verified' || n === 'confirmed') {
    return { label: 'Boolean (mostly true)', kind: 'data', generate: () => faker.datatype.boolean({ probability: 0.8 }) }
  }
  if (isBoolType(colType)) {
    return { label: 'Boolean', kind: 'data', generate: () => faker.datatype.boolean() }
  }

  // --- Status ---
  if (n === 'status') {
    return { label: 'Status', kind: 'data', generate: () => faker.helpers.arrayElement(['active', 'inactive', 'pending', 'archived']) }
  }
  if (n === 'gender' || n === 'sex') {
    return { label: 'Gender', kind: 'data', generate: () => faker.person.sex() }
  }
  if (n === 'locale' || n === 'language' || n === 'lang') {
    return { label: 'Locale', kind: 'data', generate: () => faker.helpers.arrayElement(['en', 'de', 'fr', 'es', 'nl', 'pt']) }
  }
  if (n === 'timezone' || n === 'tz') {
    return { label: 'Timezone', kind: 'data', generate: () => faker.location.timeZone() }
  }
  if (n === 'color' || n === 'colour') {
    return { label: 'Color', kind: 'data', generate: () => faker.color.human() }
  }
  if (n === 'tags' || n === 'keywords') {
    return { label: 'Tags', kind: 'data', generate: () => faker.lorem.words(3) }
  }

  // --- Text fallback ---
  if (n === 'title' || n === 'headline' || n === 'subject') {
    return { label: useLorem ? 'Lorem sentence' : 'Sentence', kind: 'data', generate: () => useLorem ? faker.lorem.sentence() : faker.lorem.sentence() }
  }
  if (n === 'description' || n === 'bio' || n === 'about' || n === 'summary' || n === 'excerpt') {
    return {
      label: useLorem ? 'Lorem paragraph' : 'Bio paragraph',
      kind: 'data',
      generate: () => useLorem ? faker.lorem.paragraph() : faker.person.bio(),
    }
  }
  if (n === 'notes' || n === 'comments' || n === 'remark' || n === 'message' || n === 'body' || n === 'content') {
    return { label: 'Sentences', kind: 'data', generate: () => faker.lorem.sentences(2) }
  }

  // Generic text fallback based on type
  return {
    label: useLorem ? 'Lorem word' : 'Random word',
    kind: 'data',
    generate: () => faker.lorem.word(),
  }
}

export function generateRows(
  columns: { name: string; type: string; nullable: boolean }[],
  tableName: string,
  count: number,
  useLorem: boolean,
  fkConfigs: Record<string, FKConfig>
): Record<string, any>[] {
  const strategies = columns.map(c => ({
    col: c,
    strategy: getStrategy(c.name, c.type, tableName, useLorem, fkConfigs[c.name]),
  }))

  return Array.from({ length: count }, () => {
    const row: Record<string, any> = {}
    for (const { col, strategy } of strategies) {
      if (strategy.kind === 'skip') continue
      const val = strategy.generate()
      if (val === undefined) continue
      if (val === null && !col.nullable) continue // skip null for non-nullable
      row[col.name] = val
    }
    return row
  })
}

export function buildInsertSQL(
  driver: string,
  schema: string,
  tableName: string,
  rows: Record<string, any>[]
): string {
  if (rows.length === 0) return ''

  const cols = Object.keys(rows[0])

  function q(name: string): string {
    if (driver === 'mssql') return `[${name}]`
    if (driver === 'mysql') return `\`${name}\``
    return `"${name}"`
  }

  function tableRef(): string {
    if (driver === 'sqlite' || !schema) return q(tableName)
    return `${q(schema)}.${q(tableName)}`
  }

  function val(v: any): string {
    if (v === null || v === undefined) return 'NULL'
    if (typeof v === 'boolean') return v ? '1' : '0'
    if (typeof v === 'number') return String(v)
    return `'${String(v).replace(/'/g, "''")}'`
  }

  const colList = cols.map(q).join(', ')
  const valueParts = rows.map(row => `(${cols.map(c => val(row[c])).join(', ')})`).join(',\n')
  return `INSERT INTO ${tableRef()} (${colList})\nVALUES\n${valueParts};`
}
