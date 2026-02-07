export interface Attribute {
  name: string
  type: string
  required: boolean
  pattern?: string
}

export interface KeyDefinition {
  name: string
  type: string
}

export interface GlobalSecondaryIndex {
  indexName: string
  partitionKey: KeyDefinition
  sortKey?: KeyDefinition
  projection: string
  projectedAttributes?: string[]
  purpose: string
}

export interface EntitySchema {
  entityType: string
  pkPattern: string
  skPattern: string
  attributes: Attribute[]
  gsiMappings?: Record<string, string>
}

export interface EdgeItemSchema {
  name: string
  description: string
  pkPattern: string
  skPattern: string
  attributes: Attribute[]
}

export interface Entity {
  name: string
  originalTable: string
  relationships: {
    type: string
    with: string
    description: string
  }[]
}

export interface AccessPattern {
  id: string
  description: string
  operation: string
  index: string
  keyCondition: string
}

export interface AccessPatternImpl {
  patternId: string
  description: string
  implementation: Record<string, any>
}

export interface SampleDataItem {
  description: string
  item: Record<string, any>
}

export interface DynamoDBDesign {
  analysis: {
    entities: Entity[]
    accessPatterns: AccessPattern[]
  }
  design: {
    tableName: string
    billingMode: string
    primaryKey: {
      partitionKey: KeyDefinition
      sortKey: KeyDefinition
    }
    globalSecondaryIndexes: GlobalSecondaryIndex[]
    entitySchemas: EntitySchema[]
    edgeItems: EdgeItemSchema[]
  }
  sampleData: SampleDataItem[]
  accessPatternImplementation: AccessPatternImpl[]
}

// Legacy interfaces for backward compatibility
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

export interface ConversionSummary {
  conversionId: string
  conversionDate: string
  createdAt: string
  status: string
  optimizationType: string
  tablesExtracted: string
}

export interface Conversion {
  conversionId: string
  conversionDate: string
  createdAt: string
  sqlContent: string
  noSqlSchema: DynamoDBDesign | NoSqlSchema // Support both formats
  optimizationType: string
  status: string
  tablesExtracted: string
}

export interface GetSchemasResponse {
  conversions: ConversionSummary[]
  count: number
}

export interface ParseSqlRequest {
  sqlContent: string
  optimizationType?: 'balanced' | 'read-heavy' | 'write-heavy' | 'cost-optimized'
}
