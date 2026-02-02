package main

// AccessPatternMessage represents the message sent from conversion-worker to access-pattern-worker
type AccessPatternMessage struct {
	ConversionID     string          `json:"conversionId"`
	SQLContent       string          `json:"sqlContent"`
	DynamoDBDesign   *DynamoDBDesign `json:"dynamodbDesign"`
	OptimizationType string          `json:"optimizationType"`
}

// AccessPatternResult represents the output from Bedrock for access patterns only
type AccessPatternResult struct {
	AccessPatterns              []AccessPattern     `json:"accessPatterns"`
	AccessPatternImplementation []AccessPatternImpl `json:"accessPatternImplementation"`
	SampleData                  []SampleDataItem    `json:"sampleData"`
}

// AccessPattern represents a single access pattern
type AccessPattern struct {
	ID           string `json:"id"`
	Description  string `json:"description"`
	Operation    string `json:"operation"`
	Index        string `json:"index"`
	KeyCondition string `json:"keyCondition"`
}

// AccessPatternImpl represents the implementation of an access pattern
type AccessPatternImpl struct {
	PatternID      string                 `json:"patternId"`
	Description    string                 `json:"description"`
	Implementation map[string]interface{} `json:"implementation"`
}

// SampleDataItem represents sample data for testing
type SampleDataItem struct {
	Description string                 `json:"description"`
	Item        map[string]interface{} `json:"item"`
}

// DynamoDBDesign represents the base design from conversion-worker
type DynamoDBDesign struct {
	Analysis struct {
		Entities []struct {
			Name          string `json:"name"`
			OriginalTable string `json:"originalTable"`
			Relationships []struct {
				Type        string `json:"type"`
				With        string `json:"with"`
				Description string `json:"description"`
			} `json:"relationships"`
		} `json:"entities"`
		AccessPatterns []AccessPattern `json:"accessPatterns"`
	} `json:"analysis"`
	Design struct {
		TableName   string `json:"tableName"`
		BillingMode string `json:"billingMode"`
		PrimaryKey  struct {
			PartitionKey KeyDefinition `json:"partitionKey"`
			SortKey      KeyDefinition `json:"sortKey"`
		} `json:"primaryKey"`
		GlobalSecondaryIndexes []GSIDefinition  `json:"globalSecondaryIndexes"`
		EntitySchemas          []EntitySchema   `json:"entitySchemas"`
		EdgeItems              []EdgeItemSchema `json:"edgeItems"`
	} `json:"design"`
	SampleData                  []SampleDataItem    `json:"sampleData"`
	AccessPatternImplementation []AccessPatternImpl `json:"accessPatternImplementation"`
}

type KeyDefinition struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type GSIDefinition struct {
	IndexName           string        `json:"indexName"`
	PartitionKey        KeyDefinition `json:"partitionKey"`
	SortKey             KeyDefinition `json:"sortKey,omitempty"`
	Projection          string        `json:"projection"`
	ProjectedAttributes []string      `json:"projectedAttributes,omitempty"`
	Purpose             string        `json:"purpose"`
}

type EntitySchema struct {
	EntityType  string            `json:"entityType"`
	PKPattern   string            `json:"pkPattern"`
	SKPattern   string            `json:"skPattern"`
	Attributes  []AttributeDef    `json:"attributes"`
	GSIMappings map[string]string `json:"gsiMappings,omitempty"`
}

type AttributeDef struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Required bool   `json:"required"`
	Pattern  string `json:"pattern,omitempty"`
}

type EdgeItemSchema struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	PKPattern   string         `json:"pkPattern"`
	SKPattern   string         `json:"skPattern"`
	Attributes  []AttributeDef `json:"attributes"`
}
