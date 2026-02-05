package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
)

// DynamoDBDesign representa la estructura de respuesta completa
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
}

type KeyDefinition struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type GSIDefinition struct {
	IndexName    string        `json:"indexName"`
	PartitionKey KeyDefinition `json:"partitionKey"`
	SortKey      KeyDefinition `json:"sortKey,omitempty"`
	Projection   string        `json:"projection"`
	Purpose      string        `json:"purpose"`
}

type EntitySchema struct {
	EntityType  string         `json:"entityType"`
	PKPattern   string         `json:"pkPattern"`
	SKPattern   string         `json:"skPattern"`
	Attributes  []AttributeDef `json:"attributes"`
	GSIMappings FlexibleMap    `json:"gsiMappings,omitempty"`
}

type AttributeDef struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Required bool   `json:"required"`
}

type EdgeItemSchema struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	PKPattern   string         `json:"pkPattern"`
	SKPattern   string         `json:"skPattern"`
	Attributes  []AttributeDef `json:"attributes,omitempty"`
}

// countSQLTables cuenta las tablas CREATE TABLE en el SQL
func countSQLTables(sqlContent string) int {
	re := regexp.MustCompile(`(?i)CREATE\s+TABLE\s+`)
	return len(re.FindAllString(sqlContent, -1))
}

// ValidateDesign ahora es más inteligente y menos punitivo
func ValidateDesign(design *DynamoDBDesign, sqlContent string) ([]string, []string) {
	var errors []string
	var warnings []string
	expectedTables := countSQLTables(sqlContent)

	// 1. Análisis de Reducción (Warning, no Error)
	if len(design.Analysis.Entities) < expectedTables {
		warnings = append(warnings, fmt.Sprintf("Desnormalization detected: %d tables merged into %d entities", expectedTables, len(design.Analysis.Entities)))
	}

	// 2. Estructura base (Error)
	if len(design.Design.EntitySchemas) == 0 {
		errors = append(errors, "Critical: No entity schemas generated")
	}

	// 3. Validación de GSI1 (Requerido para Single Table)
	hasGSI1 := false
	for _, gsi := range design.Design.GlobalSecondaryIndexes {
		if gsi.IndexName == "GSI1" {
			hasGSI1 = true
			break
		}
	}
	if !hasGSI1 {
		errors = append(errors, "Missing GSI1: Required for inverted index patterns")
	}

	return errors, warnings
}

// ValidationResult estructura para resultado de validación
type ValidationResult struct {
	IsValid  bool     `json:"isValid"`
	Errors   []string `json:"errors,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
	Stats    struct {
		SQLTables       int `json:"sqlTables"`
		EntitiesFound   int `json:"entitiesFound"`
		EntitySchemas   int `json:"entitySchemas"`
		EdgeItems       int `json:"edgeItems"`
		AccessPatterns  int `json:"accessPatterns"`
		GSICount        int `json:"gsiCount"`
		SampleDataCount int `json:"sampleDataCount"`
	} `json:"stats"`
}

// ValidateDesignFull validación completa con estadísticas
func ValidateDesignFull(design *DynamoDBDesign, sqlContent string) ValidationResult {
	result := ValidationResult{IsValid: true}

	// Calcular estadísticas
	result.Stats.SQLTables = countSQLTables(sqlContent)
	result.Stats.EntitiesFound = len(design.Analysis.Entities)
	result.Stats.EntitySchemas = len(design.Design.EntitySchemas)
	result.Stats.EdgeItems = len(design.Design.EdgeItems)
	result.Stats.GSICount = len(design.Design.GlobalSecondaryIndexes)

	// Ejecutar validaciones
	errors, warnings := ValidateDesign(design, sqlContent)

	for _, warning := range warnings {
		result.Warnings = append(result.Warnings, warning)
	}
	for _, error := range errors {
		result.Errors = append(result.Errors, error)
		result.IsValid = false
	}

	return result
}

var bedrockClient *bedrockruntime.Client

func initBedrockClient() {
	if os.Getenv("USE_MOCK_BEDROCK") == "true" {
		log.Println("Mock Bedrock enabled — skipping client initialization")
		return
	}

	var opts []func(*config.LoadOptions) error

	// Use dedicated Bedrock credentials if available (needed in LocalStack
	// where AWS_ACCESS_KEY_ID is overwritten with dummy values).
	if ak := os.Getenv("BEDROCK_AWS_ACCESS_KEY_ID"); ak != "" {
		sk := os.Getenv("BEDROCK_AWS_SECRET_ACCESS_KEY")
		opts = append(opts, config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(ak, sk, ""),
		))
	}

	if region := os.Getenv("BEDROCK_AWS_REGION"); region != "" {
		opts = append(opts, config.WithRegion(region))
	}

	cfg, err := config.LoadDefaultConfig(context.Background(), opts...)
	if err != nil {
		log.Printf("WARN: Failed to load AWS config for Bedrock: %v", err)
		return
	}

	endpoint := os.Getenv("BEDROCK_ENDPOINT")
	if endpoint != "" {
		bedrockClient = bedrockruntime.NewFromConfig(cfg, func(o *bedrockruntime.Options) {
			o.BaseEndpoint = aws.String(endpoint)
		})
	} else {
		bedrockClient = bedrockruntime.NewFromConfig(cfg)
	}
}

func buildOptimizedPrompt(sqlContent, optimizationType string) string {
	return fmt.Sprintf(`<system>
You are a senior Data Architect specialized in Amazon DynamoDB NoSQL modeling, certified in applying **Single Table Design** patterns following AWS official best practices.

## FUNDAMENTAL DYNAMODB PRINCIPLES (AWS Official)

### 1. NoSQL Mindset Shift
- In RDBMS, you design schemas first and query later. In DynamoDB, **you MUST know all access patterns BEFORE designing the schema**.
- DynamoDB eliminates JOINs by design—your schema must answer any query with a SINGLE operation (Query or GetItem).
- Optimize for the most frequent and critical access patterns; secondary patterns can use GSIs.

### 2. Single Table Design Philosophy
- Maintain the **minimum number of tables possible** (recommended: one table per microservice).
- A single table with inverted indexes supports complex hierarchical data structures.
- Exception: High-volume time-series data or datasets with fundamentally different access patterns may justify separate tables.

### 3. Key Design Principles
**Partition Key (PK):**
- Must have **high cardinality** to distribute load evenly across partitions.
- Use entity prefixes for Single Table Design: "USER#", "ORDER#", "PRODUCT#".
- Avoid hot partitions—if necessary, implement **write sharding** with random suffixes.

**Sort Key (SK):**
- Enables hierarchical data organization using **composite sort keys**: "TYPE#SUBTYPE#TIMESTAMP".
- Supports range queries with operators: begins_with, between, >, <.
- Critical for creating **item collections** (all items sharing the same PK).

### 4. Relationship Modeling Patterns
**Adjacency List Pattern (Primary for 1:N and N:M):**
- Top-level entities use PK as their identifier.
- Relationships (edges) are items within the partition where SK points to the related entity.
- Enables minimal data duplication and simplified query patterns.

**Vertical Partitioning:**
- Split large items into multiple items with same PK but different SK prefixes.
- Root item: "ENTITY#123" + "METADATA#" contains core attributes.
- Related items: "ENTITY#123" + "DETAIL#type" contains supplementary data.
- Benefits: Granular write control, efficient RCU consumption for targeted reads.

**Edge Items for Many-to-Many:**
- Create bidirectional items for each relationship.
- Example: USER#A -> FOLLOWS#USER#B and USER#B -> FOLLOWER#USER#A.

### 5. Global Secondary Index (GSI) Patterns
**GSI Overloading:**
- A single GSI can index multiple attribute types using overloaded keys.
- GSI1PK/GSI1SK can contain different data depending on entity type.
- Maximizes the 20 GSI limit per table.

**Inverted Index (GSI1):**
- PK=SK and SK=PK from base table.
- Enables reverse lookups for bidirectional relationships.

**Sparse Index:**
- Only items containing the GSI key attributes appear in the index.
- Ideal for filtering rare statuses or flags (e.g., "ESCALATED", "FEATURED").
- Significantly reduces storage and RCU costs.

### 6. Data Optimization Techniques
**Controlled Denormalization:**
- Duplicate frequently read-together data to eliminate the need for multiple queries.
- Trade-off: Increased write complexity for faster reads.
- Evaluate read-to-write ratio before denormalizing.

**Write Sharding for Hot Partitions:**
- Add computed suffix (0-N) to partition keys for high-write items.
- Example: "COUNTER#daily#2024-01-15#0" through "COUNTER#daily#2024-01-15#9".
- Requires scatter-gather on reads.

**Item Size Considerations:**
- DynamoDB limit: 400 KB per item.
- Store large blobs in S3, keep reference in DynamoDB.
- Use vertical partitioning for items approaching size limits.
</system>

<task>
Analyze the following relational SQL schema and design an optimized DynamoDB Single Table schema.

**Optimization Priority**: %s
**SQL Schema**:
<sql_schema>
%s
</sql_schema>
</task>

<analysis_process>
Execute the following steps systematically:

### Step 1: Entity Extraction
- Identify all SQL tables as distinct entity types.
- Map column data types to DynamoDB types (S, N, B, L, M, SS, NS, BS, BOOL, NULL).
- Note primary keys, foreign keys, and unique constraints.

### Step 2: Relationship Classification
- **1:1 Relationships**: Consider embedding or vertical partitioning.
- **1:N Relationships**: Parent PK + Child SK pattern, or adjacency list.
- **N:M Relationships**: Require edge items with bidirectional references.
- Document cardinality estimates where inferable from schema.

### Step 3: Access Pattern Inference
Based on the schema structure, infer these common patterns:
- **Primary lookups**: Get entity by primary key.
- **Foreign key traversals**: Get related entities (1:N, N:M).
- **Secondary attribute searches**: Filter by non-key columns.
- **Listing operations**: Get all entities of a type.
- **Hierarchical queries**: Parent-child traversals.

### Step 4: Primary Key Design
- Define PK pattern with entity prefix: "ENTITYTYPE#<id>".
- Define SK pattern for hierarchy: "METADATA#" for root, "RELATION#<related_id>" for edges.
- Ensure uniqueness across all entity types in single table.

### Step 5: GSI Strategy
- **GSI1 (Inverted Index)**: GSI1PK=SK, GSI1SK=PK for reverse lookups.
- **Additional GSIs**: Based on inferred secondary access patterns.
- Apply GSI overloading where multiple entity types share similar query needs.
- Consider sparse indexes for status-based filtering.

### Step 6: Edge Item Definition
For each N:M or significant 1:N relationship:
- Define bidirectional edge items with clear PK/SK patterns.
- Include minimal denormalized attributes needed for common queries.
- Document the purpose and query patterns each edge item supports.
</analysis_process>

<output_format>
Return ONLY a valid JSON object. No explanations, no sample data, no implementation notes.

{
  "analysis": {
    "entities": [
      {
        "name": "EntityName",
        "originalTable": "sql_table_name",
        "relationships": [
          {
            "type": "1:N|N:M|1:1",
            "with": "OtherEntity",
            "description": "Brief description of the relationship"
          }
        ]
      }
    ]
  },
  "design": {
    "tableName": "SingleTable",
    "billingMode": "PAY_PER_REQUEST",
    "primaryKey": {
      "partitionKey": {"name": "PK", "type": "S"},
      "sortKey": {"name": "SK", "type": "S"}
    },
    "globalSecondaryIndexes": [
      {
        "indexName": "GSI1",
        "partitionKey": {"name": "GSI1PK", "type": "S"},
        "sortKey": {"name": "GSI1SK", "type": "S"},
        "projection": "ALL",
        "purpose": "Inverted index for reverse lookups"
      }
    ],
    "entitySchemas": [
      {
        "entityType": "ENTITY_TYPE",
        "pkPattern": "ENTITY#<id>",
        "skPattern": "METADATA#",
        "attributes": [
          {"name": "attributeName", "type": "S|N|B|BOOL|L|M|SS|NS", "required": true}
        ],
        "gsiMappings": {
          "GSI1PK": "pattern or static value",
          "GSI1SK": "pattern or static value"
        }
      }
    ],
    "edgeItems": [
      {
        "name": "EdgeItemName",
        "description": "Purpose of this edge item",
        "pkPattern": "ENTITY_A#<id>",
        "skPattern": "RELATION#ENTITY_B#<id>",
        "attributes": [
          {"name": "denormalizedAttr", "type": "S", "required": false}
        ]
      }
    ]
  }
}
</output_format>

<critical_rules>
1. Every entity MUST have a METADATA# item as its root record.
2. GSI1 MUST be an inverted index (GSI1PK=SK, GSI1SK=PK) unless the schema clearly requires otherwise.
3. Edge items are REQUIRED for all N:M relationships.
4. Attribute types must use DynamoDB notation: S, N, B, BOOL, L, M, SS, NS, BS, NULL.
5. All patterns must use the "#" delimiter for composite keys.
6. Do not include sample data or UUID examples—use "<id>" placeholders.
</critical_rules>`, optimizationType, sqlContent)
}

func InvokeConversion(ctx context.Context, sqlContent, optimizationType string) (*DynamoDBDesign, error) {
	if os.Getenv("USE_MOCK_BEDROCK") == "true" {
		return mockBedrockResponse()
	}

	if bedrockClient == nil {
		return nil, fmt.Errorf("Bedrock client not initialized")
	}

	modelID := os.Getenv("BEDROCK_MODEL_ID")
	if modelID == "" {
		modelID = "anthropic.claude-sonnet-4-20250514-v1:0"
	}

	log.Printf("[DEBUG] Using Bedrock Model ID: %s", modelID)
	log.Printf("[DEBUG] SQL contains %d tables", countSQLTables(sqlContent))

	prompt := buildOptimizedPrompt(sqlContent, optimizationType)

	request := map[string]interface{}{
		"anthropic_version": "bedrock-2023-05-31",
		"max_tokens":        16384,
		"temperature":       0.1,
		"messages": []map[string]interface{}{
			{
				"role":    "user",
				"content": prompt,
			},
		},
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()

	output, err := bedrockClient.InvokeModel(ctx, &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String(modelID),
		Body:        requestBody,
		ContentType: aws.String("application/json"),
	})
	if err != nil {
		return nil, fmt.Errorf("error invoking model: %w", err)
	}

	var response struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}

	if err := json.Unmarshal(output.Body, &response); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	if len(response.Content) == 0 {
		return nil, fmt.Errorf("empty response from model")
	}

	rawJSON := cleanJSONResponse(response.Content[0].Text)

	var design DynamoDBDesign
	if err := json.Unmarshal([]byte(rawJSON), &design); err != nil {
		// Log más detallado para debugging
		log.Printf("[ERROR] JSON parse failed. First 1000 chars: %s", rawJSON[:min(1000, len(rawJSON))])
		return nil, fmt.Errorf("error parsing design JSON: %w", err)
	}

	// Validar el diseño automáticamente
	validation := ValidateDesignFull(&design, sqlContent)
	if !validation.IsValid {
		log.Printf("[WARN] Design validation failed: %v", validation.Errors)
	}
	if len(validation.Warnings) > 0 {
		log.Printf("[INFO] Design warnings: %v", validation.Warnings)
	}
	log.Printf("[DEBUG] Design stats: %+v", validation.Stats)

	return &design, nil
}

// cleanJSONResponse limpia la respuesta para extraer JSON válido
func cleanJSONResponse(raw string) string {
	raw = strings.TrimSpace(raw)

	// Remover bloques de código markdown
	if strings.Contains(raw, "```") {
		// Buscar ```json o solo ```
		re := regexp.MustCompile("```(?:json)?\\s*")
		parts := re.Split(raw, -1)
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if strings.HasPrefix(part, "{") {
				// Remover ``` del final si existe
				if idx := strings.LastIndex(part, "```"); idx != -1 {
					part = part[:idx]
				}
				return strings.TrimSpace(part)
			}
		}
	}

	// Si empieza con {, asumir que es JSON directo
	if strings.HasPrefix(raw, "{") {
		return raw
	}

	// Buscar el primer { y último }
	start := strings.Index(raw, "{")
	end := strings.LastIndex(raw, "}")
	if start != -1 && end > start {
		return raw[start : end+1]
	}

	return raw
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func mockBedrockResponse() (*DynamoDBDesign, error) {
	// Mock para testing
	return &DynamoDBDesign{
		Design: struct {
			TableName   string `json:"tableName"`
			BillingMode string `json:"billingMode"`
			PrimaryKey  struct {
				PartitionKey KeyDefinition `json:"partitionKey"`
				SortKey      KeyDefinition `json:"sortKey"`
			} `json:"primaryKey"`
			GlobalSecondaryIndexes []GSIDefinition  `json:"globalSecondaryIndexes"`
			EntitySchemas          []EntitySchema   `json:"entitySchemas"`
			EdgeItems              []EdgeItemSchema `json:"edgeItems"`
		}{
			TableName:   "SingleTable",
			BillingMode: "PAY_PER_REQUEST",
		},
	}, nil
}

// FlexibleMap puede manejar tanto un objeto JSON como un string (que intenta convertir a objeto)
type FlexibleMap map[string]string

func (fm *FlexibleMap) UnmarshalJSON(b []byte) error {
	// Caso 1: Es un objeto JSON estándar { "key": "value" }
	var m map[string]string
	if err := json.Unmarshal(b, &m); err == nil {
		*fm = FlexibleMap(m)
		return nil
	}

	// Caso 2: Claude envió un string por error (ej: "GSI1PK: METADATA#")
	var s string
	if err := json.Unmarshal(b, &s); err == nil {
		// Creamos un mapa simple para no romper el flujo
		*fm = FlexibleMap{"raw": s}
		return nil
	}

	return nil // Si es nulo o vacío, no hacemos nada
}
