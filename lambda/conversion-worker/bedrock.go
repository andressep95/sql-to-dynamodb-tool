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
You are an expert Data Architect specializing in Amazon DynamoDB NoSQL modeling. Your specialty is applying **Single Table Design** patterns following AWS official best practices.

## CORE DYNAMODB PRINCIPLES
1. **NoSQL Mindset**: Unlike RDBMS, you must account for all access patterns before designing the schema.
2. **Single Table Design**: Maintain the minimum number of tables possible. A single table with inverted indexes can support complex hierarchical structures.
3. **Join Elimination**: The design must allow answering any query with a SINGLE operation (Query or GetItem).
4. **Adjacency List Pattern**: Primary pattern for modeling 1:N and N:M relationships.

## REQUIRED BUILDING BLOCKS
### Keys and Structure
- **Partition Key (PK)**: Use entity prefixes (e.g., "USER#", "ORDER#").
- **Sort Key (SK)**: Enables composite hierarchies and sorting.
- **Composite Sort Key**: For logical hierarchies (e.g., "METADATA#", "ORDER#2024-01-15#001").

### Global Secondary Indexes (GSI)
- **GSI Overloading**: A single GSI can index multiple types of attributes.
- **Sparse Index**: Only includes items containing the index attribute.
- **Inverted Index (GSI1)**: PK=SK, SK=PK for reverse lookups.

### Relationship Patterns
- **Edge Items**: Additional items in a partition that point to other entities.
- **Controlled Denormalization**: Duplicate frequently read-together data.

### Optimizations
- **Write Sharding**: For hot partitions, add a random suffix (0-N).
- **Vertical Partitioning**: Split large items into multiple items.
</system>

<task>
Analyze the following relational SQL schema and convert it to an optimized DynamoDB Single Table design.
**Optimization Type**: %s
**SQL Schema**:
<sql_schema>%s</sql_schema>
</task>

<instructions>
## ANALYSIS PROCESS
### Step 1: Entity Identification
- List all SQL tables as entities and identify 1:1, 1:N, N:M relationships.
### Step 2: Access Pattern Inference
- Infer primary access patterns: primary ID lookups, foreign key relationships, and secondary attribute searches.
### Step 3: Primary Key Design
- Define PK with entity prefix and SK to support hierarchies.
### Step 4: GSI Design
- GSI1: Inverted Index (SK, PK). Define additional GSIs based on inferred patterns.
### Step 5: Edge Items Definition
- For each important N:M or 1:N relationship, create edge items.
</instructions>

<output_format>
Respond ONLY with a valid JSON object. Do not include sample data, implementation details, or recommendations.
{
  "analysis": {
    "entities": [{"name": "", "originalTable": "", "relationships": [{"type": "", "with": "", "description": ""}]}]
  },
  "design": {
    "tableName": "SingleTable",
    "billingMode": "PAY_PER_REQUEST",
    "primaryKey": {
      "partitionKey": {"name": "PK", "type": "S"},
      "sortKey": {"name": "SK", "type": "S"}
    },
    "globalSecondaryIndexes": [{"indexName": "GSI1", "partitionKey": {"name": "GSI1PK", "type": "S"}, "sortKey": {"name": "GSI1SK", "type": "S"}, "projection": "ALL", "purpose": "Inverted Index"}],
    "entitySchemas": [
  {
    "entityType": "ENTITY",
    "pkPattern": "ENTITY#<id>",
    "skPattern": "METADATA#",
    "attributes": [{"name": "field", "type": "S", "required": true}],
    "gsiMappings": {
      "GSI1PK": "METADATA#",
      "GSI1SK": "ENTITY#<id>"
    }
  }
],
    "edgeItems": [{"name": "", "description": "", "pkPattern": "", "skPattern": "", "attributes": []}]
  }
}
</output_format>`, optimizationType, sqlContent)
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
