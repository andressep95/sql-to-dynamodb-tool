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
You are a DynamoDB access pattern specialist. Your job is to analyze SQL schemas and existing DynamoDB designs to generate comprehensive access patterns, implementations and sample data.

## YOUR FOCUS
- Generate realistic access patterns based on SQL relationships
- Create detailed implementation guides for each pattern
- Provide sample data that demonstrates the patterns
- Suggest performance and cost optimizations

## ANALYSIS APPROACH
1. **SQL Analysis**: Understand the original relational queries
2. **DynamoDB Design Review**: Leverage the existing table structure
3. **Access Pattern Inference**: Generate patterns for common operations
4. **Implementation Details**: Provide exact DynamoDB operations
</system>

<task>
Given the original SQL schema and the DynamoDB design, generate comprehensive access patterns.

**Optimization Type**: %s

**Original SQL Schema**:
<sql_schema>
%s
</sql_schema>

**Existing DynamoDB Design**:
<dynamodb_design>
%s
</dynamodb_design>
</task>

<instructions>
## REQUIRED ACCESS PATTERNS

### Core Patterns (Minimum)
1. **Entity Retrieval**: Get single entity by ID (one per entity type)
2. **Relationship Queries**: Follow foreign key relationships
3. **List Operations**: Get all entities of a type with pagination
4. **Search Patterns**: Query by secondary attributes using GSIs

### Advanced Patterns
5. **Hierarchical Queries**: Parent-child relationships
6. **Many-to-Many**: Junction table queries via edge items
7. **Filtering**: Complex conditions using filter expressions
8. **Aggregation**: Count, sum operations where applicable

## IMPLEMENTATION REQUIREMENTS
- Use exact DynamoDB operation names (Query, GetItem, BatchGetItem)
- Specify correct index names from the design
- Provide realistic key conditions
- Include filter expressions when needed

## SAMPLE DATA REQUIREMENTS
- At least one sample per entity type
- Demonstrate PK/SK patterns from the design
- Include GSI attribute values
- Show edge items for relationships
</instructions>

<output_format>
Respond ONLY with valid JSON:
{
  "accessPatterns": [
    {
      "id": "AP1",
      "description": "Get user by ID",
      "operation": "Query",
      "index": "Table",
      "keyCondition": "PK = 'USER#123' AND SK begins_with 'METADATA#'"
    }
  ],
  "accessPatternImplementation": [
    {
      "patternId": "AP1",
      "description": "Get user by ID implementation",
      "implementation": {
        "operation": "Query",
        "table": "SingleTable",
        "key": {"PK": "USER#123", "SK": "begins_with(METADATA#)"},
        "keyCondition": "PK = :pk AND begins_with(SK, :sk)"
      }
    }
  ],
  "sampleData": [
    {
      "description": "User entity sample",
      "item": {
        "PK": "USER#123",
        "SK": "METADATA#",
        "GSI1PK": "METADATA#",
        "GSI1SK": "USER#123",
        "name": "John Doe",
        "email": "john@example.com"
      }
    }
  ]
}
</output_format>`, optimizationType, sqlContent, string(designJSON))
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
	raw = strings.TrimSpace(raw)

	// Remove markdown code blocks
	if strings.Contains(raw, "```") {
		start := strings.Index(raw, "{")
		end := strings.LastIndex(raw, "}")
		if start != -1 && end > start {
			return raw[start : end+1]
		}
	}

	if strings.HasPrefix(raw, "{") {
		return raw
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
