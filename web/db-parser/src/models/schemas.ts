export interface Attribute {
  name: string
  type: string
}

export interface KeyDefinition {
  name: string
  type: string
}

export interface GlobalSecondaryIndex {
  indexName: string
  partitionKey: KeyDefinition
  projectionType: string
}

export interface TableSchema {
  tableName: string
  attributes: Attribute[]
  partitionKey: KeyDefinition
  sortKey: KeyDefinition | null
  billingMode: string
  globalSecondaryIndexes: GlobalSecondaryIndex[]
}

export interface NoSqlSchema {
  tables: TableSchema[]
}

export interface Conversion {
  conversionId: string
  conversionDate: string
  createdAt: string
  expiresAt: string
  sqlContent: string
  noSqlSchema: NoSqlSchema
  optimizationType: string
  status: string
  tablesExtracted: string
}

export interface GetSchemasResponse {
  conversions: Conversion[]
  count: number
}

export interface ParseSqlRequest {
  sqlContent: string
  optimizationType?: 'balanced' | 'read-heavy' | 'write-heavy' | 'cost-optimized'
}
