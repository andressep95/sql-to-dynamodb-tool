package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
)

var bedrockClient *bedrockruntime.Client

func initBedrockClient() {
	if os.Getenv("USE_MOCK_BEDROCK") == "true" {
		log.Println("Mock Bedrock enabled — skipping client initialization")
		return
	}

	var opts []func(*config.LoadOptions) error

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

func buildAccessPatternPrompt(sqlContent string, design *DynamoDBDesign, optimizationType string) string {
	designJSON, _ := json.MarshalIndent(design, "", "  ")

	return fmt.Sprintf(`<system>
You are a DynamoDB access pattern specialist and implementation architect. Your expertise is translating relational access patterns into optimized DynamoDB operations following AWS best practices.

## ACCESS PATTERN FUNDAMENTALS (AWS Official)

### 1. Query vs Scan Operations
- **Query**: Always preferred. Uses partition key (required) + optional sort key conditions.
- **Scan**: Reads entire table/index. Avoid in production—use only for analytics or one-time operations.
- **GetItem**: Single item retrieval by exact primary key. Most efficient operation.
- **BatchGetItem**: Up to 100 items across multiple partition keys. Parallel retrieval.

### 2. Key Condition Expressions
Supported operators for Sort Key in Query:
- **=**: Exact match
- **<, <=, >, >=**: Range comparisons
- **BETWEEN**: Inclusive range
- **begins_with(SK, :prefix)**: Prefix matching (critical for hierarchical keys)

### 3. Filter Expressions
- Applied AFTER data is read—does NOT reduce RCU consumption.
- Use for post-query refinement, not as primary access mechanism.
- Complex filters are acceptable when the key condition already narrows results significantly.

### 4. Consistency Levels
- **Eventually Consistent**: Default, 0.5 RCU per 4KB. Use for most read patterns.
- **Strongly Consistent**: 1 RCU per 4KB. Use when immediate consistency is required.
- Specify in implementation when strongly consistent reads are necessary.

### 5. Pagination Strategy
- DynamoDB returns up to 1MB per Query/Scan.
- Use **LastEvaluatedKey** for cursor-based pagination.
- **Limit** parameter controls maximum items evaluated (not returned after filters).
- Implement application-level pagination with ExclusiveStartKey.

### 6. Projection Expressions
- Retrieve only needed attributes to reduce data transfer and RCU.
- Critical for items with many attributes or large values.
- Always specify when only a subset of attributes is needed.

### 7. Batch Operations
- **BatchGetItem**: Retrieve multiple items by PK. Max 100 items, 16MB response.
- **BatchWriteItem**: Write/delete multiple items. Max 25 items, 16MB request.
- Handle UnprocessedKeys/UnprocessedItems with exponential backoff.

### 8. Transaction Operations
- **TransactWriteItems**: ACID writes across up to 100 items.
- **TransactGetItems**: Consistent reads across up to 100 items.
- Use for operations requiring atomicity across multiple items.
</system>

<task>
Given the original SQL schema and the DynamoDB design, generate comprehensive access patterns with implementation details and sample data.

**Optimization Priority**: %s

**Original SQL Schema**:
<sql_schema>
%s
</sql_schema>

**DynamoDB Design**:
<dynamodb_design>
%s
</dynamodb_design>
</task>

<pattern_generation_process>

### Phase 1: Core Access Patterns
Generate patterns for these fundamental operations (one per entity type):

**1.1 Entity Retrieval Patterns**
- Get single entity by ID (GetItem or Query with exact keys)
- Pattern naming: "Get[Entity]ById"

**1.2 Relationship Traversal Patterns**
- Navigate 1:N relationships (Query with begins_with on SK)
- Navigate N:M relationships (Query edge items, then BatchGetItem for full entities)
- Pattern naming: "Get[Child]By[Parent]", "Get[Related]For[Entity]"

**1.3 List Operations**
- List all entities of a type with pagination
- Use GSI with entity type as partition key
- Pattern naming: "List[Entities]"

**1.4 Search Patterns**
- Query by secondary attributes using appropriate GSI
- Pattern naming: "Find[Entity]By[Attribute]"

### Phase 2: Advanced Patterns
Based on schema relationships, generate:

**2.1 Hierarchical Queries**
- Multi-level parent-child traversals
- Use composite sort keys with begins_with

**2.2 Aggregation Patterns**
- Count items in a collection (Query with Select: COUNT)
- Note: DynamoDB doesn't support SUM/AVG natively

**2.3 Time-Based Patterns**
- If timestamps exist, patterns for time-range queries
- Use sort key with ISO timestamp format

**2.4 Conditional Operations**
- Updates with condition expressions
- Optimistic locking patterns if version attributes exist

### Phase 3: Implementation Details
For each pattern, provide:

1. **Operation Type**: Query | GetItem | BatchGetItem | Scan
2. **Target**: Table name or GSI name
3. **Key Condition**: Exact expression with placeholder syntax
4. **Filter Expression**: If needed (with note about RCU impact)
5. **Projection**: Attributes to retrieve (if not all)
6. **Consistency**: Eventually consistent unless noted
7. **Pagination**: How to handle large result sets

### Phase 4: Sample Data Generation
Create realistic sample data that:
- Demonstrates all PK/SK patterns from the design
- Shows GSI attribute mappings
- Includes edge items for relationships
- Uses realistic but fictional values
- Covers at least 2 instances per entity type
</pattern_generation_process>

<output_format>
Return ONLY a valid JSON object following this EXACT schema. 
Values must be strings, not objects.

{
  "accessPatterns": [
    {
      "id": "AP001",
      "name": "GetEntityById",
      "description": "Short explanation",
      "operation": "GetItem", 
      "index": "Table",
      "keyCondition": "PK = :pk AND SK = :sk"
    }
  ],
  "accessPatternImplementation": [
    {
      "patternId": "AP001",
      "name": "GetEntityById",
      "description": "Implementation details",
      "implementation": {
        "operation": "GetItem",
        "table": "TableName",
        "index": null,
        "keyConditionExpression": "PK = :pk",
        "expressionAttributeValues": { ":pk": {"S": "ENTITY#ID"} }
      }
    }
  ],
  "sampleData": [
    {
      "description": "Example entity",
      "item": {
        "PK": {"S": "ENTITY#001"},
        "SK": {"S": "METADATA#"},
        "Attribute": {"S": "Value"}
      }
    }
  ]
}
</output_format>

<critical_rules>
1. **Every entity type** must have at least one retrieval pattern (GetById).
2. **Every relationship** in the design must have corresponding traversal patterns.
3. **Key conditions** must use valid DynamoDB expression syntax with placeholders (:value).
4. **Sample data** must use DynamoDB JSON format with type descriptors ({"S": "value"}, {"N": "123"}).
5. **Pattern IDs** must be sequential: AP001, AP002, AP003...
6. **GSI queries** must reference the correct index name from the design.
7. **Pagination strategy** is REQUIRED for any List operation.
8. **Filter expressions** must include a note about RCU impact.
9. **Do not generate patterns** for operations not supported by the schema design.
10. **Sample data must be consistent**—if User usr_001 follows usr_002, both users must exist in sample data.
</critical_rules>

<validation_checklist>
Before outputting, verify:
- [ ] All entities have GetById patterns
- [ ] All relationships have traversal patterns
- [ ] All GSIs have at least one query pattern
- [ ] Sample data covers all entity types
- [ ] Sample data demonstrates all edge items
- [ ] Key conditions match the design's PK/SK patterns
- [ ] GSI queries use correct GSI attribute names
- [ ] Pagination is specified for list operations
</validation_checklist>`, optimizationType, sqlContent, string(designJSON))
}

func GenerateAccessPatterns(ctx context.Context, sqlContent string, design *DynamoDBDesign, optimizationType string) (*AccessPatternResult, error) {
	if os.Getenv("USE_MOCK_BEDROCK") == "true" {
		return mockAccessPatternResponse(), nil
	}

	if bedrockClient == nil {
		return nil, fmt.Errorf("Bedrock client not initialized")
	}

	modelID := os.Getenv("BEDROCK_MODEL_ID")
	if modelID == "" {
		modelID = "anthropic.claude-sonnet-4-20250514-v1:0"
	}

	log.Printf("[DEBUG] Generating access patterns with model: %s", modelID)

	prompt := buildAccessPatternPrompt(sqlContent, design, optimizationType)

	request := map[string]interface{}{
		"anthropic_version": "bedrock-2023-05-31",
		"max_tokens":        8192,
		"temperature":       0.0,
		"system":            "Return ONLY a valid JSON object. No preamble, no explanations, no markdown blocks.",
		"messages": []map[string]interface{}{
			{
				"role":    "user",
				"content": prompt,
			},
			{
				"role":    "assistant",
				"content": "{",
			},
		},
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, 90*time.Second)
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

	var result AccessPatternResult
	if err := json.Unmarshal([]byte(rawJSON), &result); err != nil {
		log.Printf("[ERROR] Access pattern JSON parse failed. Response: %s", rawJSON[:min(500, len(rawJSON))])
		return nil, fmt.Errorf("error parsing access pattern JSON: %w", err)
	}

	log.Printf("[DEBUG] Generated %d access patterns, %d implementations",
		len(result.AccessPatterns), len(result.AccessPatternImplementation))

	return &result, nil
}

func cleanJSONResponse(raw string) string {
	if !strings.HasPrefix(strings.TrimSpace(raw), "{") {
		raw = "{" + raw
	}

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

func mockAccessPatternResponse() *AccessPatternResult {
	return &AccessPatternResult{
		AccessPatterns: []AccessPattern{
			{
				ID:           "AP1",
				Description:  "Get user by ID",
				Operation:    "Query",
				Index:        "Table",
				KeyCondition: "PK = 'USER#123' AND SK begins_with 'METADATA#'",
			},
		},
		AccessPatternImplementation: []AccessPatternImpl{
			{
				PatternID:   "AP1",
				Description: "Get user by ID implementation",
				Implementation: map[string]interface{}{
					"operation": "Query",
					"table":     "SingleTable",
					"key":       map[string]interface{}{"PK": "USER#123"},
				},
			},
		},
		SampleData: []SampleDataItem{
			{
				Description: "User sample",
				Item: map[string]interface{}{
					"PK":   "USER#123",
					"SK":   "METADATA#",
					"name": "John Doe",
				},
			},
		},
	}
}

// mergeAccessPatterns combines the base design with generated access patterns
func mergeAccessPatterns(baseDesign *DynamoDBDesign, patterns *AccessPatternResult) *DynamoDBDesign {
	// Create a copy of the base design
	result := *baseDesign

	// Merge access patterns
	result.Analysis.AccessPatterns = patterns.AccessPatterns
	result.AccessPatternImplementation = patterns.AccessPatternImplementation
	result.SampleData = patterns.SampleData

	return &result
}
