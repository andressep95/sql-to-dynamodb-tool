# Backend - Lambda Functions (Go)

Documentación detallada del backend serverless implementado con AWS Lambda y Go.

## Tabla de Contenidos

1. [Arquitectura General](#arquitectura-general)
2. [Funciones Lambda](#funciones-lambda)
3. [Process Handler](#process-handler)
4. [Conversion Worker](#conversion-worker)
5. [Access Pattern Worker](#access-pattern-worker)
6. [Query Handler](#query-handler)
7. [DLQ Handler](#dlq-handler)
8. [Modelos de Datos](#modelos-de-datos)
9. [Integración con Bedrock](#integración-con-bedrock)
10. [Testing](#testing)
11. [Build y Deploy](#build-y-deploy)

---

## Arquitectura General

### Stack Tecnológico

| Tecnología | Versión | Uso |
|------------|---------|-----|
| Go | 1.x | Lenguaje principal |
| AWS SDK v2 | Latest | Integración AWS |
| ARM64 | Graviton2 | Arquitectura optimizada |
| provided.al2 | Custom Runtime | Runtime Lambda |

### Estructura de Directorios

```
lambda/
├── diagrams/                 # Process Handler
│   ├── main.go               # Handler API Gateway
│   ├── model.go              # Tipos de datos
│   ├── service.go            # Lógica de validación SQL
│   ├── validator.go          # Regexes y validaciones
│   ├── dynamo.go             # Cliente DynamoDB
│   ├── sqs.go                # Cliente SQS
│   ├── validator_test.go     # Tests unitarios
│   └── main_test.go          # Tests de integración
│
├── conversion-worker/        # Conversion Worker
│   ├── main.go               # Handler SQS
│   ├── model.go              # Tipos de datos
│   ├── bedrock.go            # Cliente Bedrock
│   ├── dynamo.go             # Cliente DynamoDB
│   └── sqs.go                # Cliente SQS
│
├── access-pattern-worker/    # Access Pattern Worker
│   ├── main.go               # Handler SQS
│   ├── model.go              # Tipos de datos
│   ├── bedrock.go            # Cliente Bedrock
│   └── dynamo.go             # Cliente DynamoDB
│
├── query/                    # Query Handler
│   ├── main.go               # Handler API Gateway
│   ├── model.go              # Tipos de datos
│   └── dynamo.go             # Cliente DynamoDB
│
└── dlq-handler/              # DLQ Handler
    ├── main.go               # Handler SQS (DLQ)
    ├── model.go              # Tipos de datos
    └── dynamo.go             # Cliente DynamoDB
```

### Patrón Común

Cada Lambda sigue este patrón:

```go
package main

import (
    "context"
    "github.com/aws/aws-lambda-go/events"
    "github.com/aws/aws-lambda-go/lambda"
    "github.com/aws/aws-sdk-go-v2/config"
)

var (
    dynamoClient *dynamodb.Client
    sqsClient    *sqs.Client
)

func init() {
    cfg, _ := config.LoadDefaultConfig(context.Background())
    dynamoClient = dynamodb.NewFromConfig(cfg)
    sqsClient = sqs.NewFromConfig(cfg)
}

func handler(ctx context.Context, event EventType) (ResponseType, error) {
    // Lógica del handler
}

func main() {
    lambda.Start(handler)
}
```

---

## Funciones Lambda

### Resumen

| Lambda | Trigger | Input | Output | Responsabilidad |
|--------|---------|-------|--------|-----------------|
| process-handler | API Gateway | HTTP POST | HTTP Response | Validar SQL, crear record |
| conversion-worker | SQS | SQS Message | - | Generar diseño DynamoDB |
| access-pattern-worker | SQS | SQS Message | - | Generar access patterns |
| query-handler | API Gateway | HTTP GET | HTTP Response | Consultar conversiones |
| dlq-handler | SQS (DLQ) | SQS Message | - | Marcar como fallido |

### Flujo de Estados

```
                    ┌─────────────────┐
                    │     PENDING     │
                    │  (Creado en DB) │
                    └────────┬────────┘
                             │
                             ▼
                    ┌─────────────────┐
                    │   PROCESSING    │
                    │ (Generando      │
                    │  diseño)        │
                    └────────┬────────┘
                             │
              ┌──────────────┼──────────────┐
              │              │              │
              ▼              ▼              │
     ┌────────────────┐  ┌────────────────┐ │
     │    FAILED      │  │ DESIGN_        │ │
     │ (Error en      │  │ COMPLETED      │ │
     │  conversión)   │  │                │ │
     └────────────────┘  └────────┬───────┘ │
                                  │         │
                                  ▼         │
                         ┌────────────────┐ │
                         │ PROCESSING_    │ │
                         │ PATTERNS       │ │
                         └────────┬───────┘ │
                                  │         │
              ┌───────────────────┼─────────┘
              │                   │
              ▼                   ▼
     ┌────────────────┐  ┌────────────────┐
     │ PATTERNS_      │  │   COMPLETED    │
     │ FAILED         │  │                │
     └────────────────┘  └────────────────┘
```

---

## Process Handler

**Ubicación:** `lambda/diagrams/`
**Trigger:** API Gateway HTTP v2 (`POST /api/v1/schemas`)

### Flujo de Ejecución

```
1. Recibir request HTTP POST
2. Parsear JSON body → ConvertRequest
3. Validar campos requeridos
4. Ejecutar ValidateSQL()
   4.1. Parsear sentencias CREATE TABLE
   4.2. Extraer información de tablas
   4.3. Validar sintaxis SQL
   4.4. Detectar warnings (ej: sin PRIMARY KEY)
5. Si válido:
   5.1. Generar UUID para conversionId
   5.2. Crear record PENDING en DynamoDB
   5.3. Encolar mensaje en SQS
   5.4. Retornar 202 Accepted
6. Si inválido:
   6.1. Retornar 400 Bad Request con detalles
```

### Handler Principal

```go
// main.go
func handleValidateSQL(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
    // Parse request body
    var convertRequest ConvertRequest
    if err := json.Unmarshal([]byte(req.Body), &convertRequest); err != nil {
        return errorResponse(400, "Invalid JSON body")
    }

    // Validate required fields
    if convertRequest.SQLContent == "" {
        return errorResponse(400, "sqlContent is required")
    }

    // Validate optimization type
    validTypes := []string{"read_heavy", "write_heavy", "balanced", "cost_optimized"}
    if !contains(validTypes, convertRequest.OptimizationType) {
        return errorResponse(400, "Invalid optimizationType")
    }

    // Execute SQL validation
    result := ValidateSQL(convertRequest.SQLContent)

    if !result.Valid {
        return validationErrorResponse(result)
    }

    // Generate conversion ID
    conversionId := uuid.New().String()

    // Create DynamoDB record
    err := CreateConversionRecord(ctx, conversionId, convertRequest, result)
    if err != nil {
        return errorResponse(500, "Failed to create conversion record")
    }

    // Send to SQS queue
    err = SendToQueue(ctx, conversionId, convertRequest)
    if err != nil {
        return errorResponse(500, "Failed to queue conversion")
    }

    // Return success response
    return successResponse(202, ConversionResponse{
        ConversionID: conversionId,
        Status:       "PENDING",
        CreatedAt:    time.Now().Unix(),
        ExpiresAt:    time.Now().Add(24 * time.Hour).Unix(),
    })
}
```

### Validador SQL

```go
// validator.go
func ValidateSQL(sql string) ValidationResult {
    result := ValidationResult{Valid: true}

    // Regex para detectar CREATE TABLE
    createTableRegex := regexp.MustCompile(
        `(?i)CREATE\s+TABLE\s+(?:IF\s+NOT\s+EXISTS\s+)?` +
        `(?:(\w+)\.)?(\w+)\s*\(([\s\S]*?)\)`,
    )

    matches := createTableRegex.FindAllStringSubmatch(sql, -1)
    if len(matches) == 0 {
        result.Valid = false
        result.Errors = append(result.Errors, ValidationError{
            Type:    "SYNTAX_ERROR",
            Message: "No valid CREATE TABLE statements found",
        })
        return result
    }

    for _, match := range matches {
        tableName := match[2]
        columnsStr := match[3]

        // Validar nombre de tabla
        if !isValidIdentifier(tableName) {
            result.Valid = false
            result.Errors = append(result.Errors, ValidationError{
                Type:    "INVALID_IDENTIFIER",
                Message: fmt.Sprintf("Invalid table name: %s", tableName),
            })
            continue
        }

        // Parsear columnas
        columns := parseColumns(columnsStr)
        tableInfo := TableInfo{
            Name:    tableName,
            Columns: columns,
        }

        // Detectar PRIMARY KEY
        if !hasPrimaryKey(columnsStr) {
            result.Warnings = append(result.Warnings, ValidationWarning{
                Type:    "NO_PRIMARY_KEY",
                Message: fmt.Sprintf("Table %s has no PRIMARY KEY", tableName),
            })
        }

        result.Tables = append(result.Tables, tableInfo)
    }

    return result
}
```

### Tipos de Datos Soportados

```go
// validator.go
var postgresDataTypes = []string{
    // Numéricos
    "SMALLINT", "INTEGER", "INT", "BIGINT",
    "DECIMAL", "NUMERIC", "REAL", "DOUBLE PRECISION",
    "SMALLSERIAL", "SERIAL", "BIGSERIAL",

    // Monetarios
    "MONEY",

    // Caracteres
    "CHAR", "CHARACTER", "VARCHAR", "CHARACTER VARYING",
    "TEXT",

    // Binarios
    "BYTEA",

    // Fecha/Hora
    "TIMESTAMP", "TIMESTAMPTZ", "DATE", "TIME",
    "TIMETZ", "INTERVAL",

    // Booleanos
    "BOOLEAN", "BOOL",

    // UUID
    "UUID",

    // JSON
    "JSON", "JSONB",

    // Arrays
    "ARRAY",

    // Geométricos
    "POINT", "LINE", "LSEG", "BOX", "PATH", "POLYGON", "CIRCLE",

    // Red
    "CIDR", "INET", "MACADDR", "MACADDR8",
}
```

---

## Conversion Worker

**Ubicación:** `lambda/conversion-worker/`
**Trigger:** SQS (`sql-to-nosql-conversion-queue`)

### Flujo de Ejecución

```
1. Recibir mensaje SQS
2. Parsear body → SQSMessageBody
3. Actualizar status a PROCESSING
4. Construir prompt para Bedrock
5. Invocar Claude 3.5 Sonnet v2
6. Parsear respuesta JSON
7. Validar diseño generado
8. Actualizar DynamoDB con diseño
9. Encolar mensaje para access patterns
10. Retornar nil (éxito) o error (reintento)
```

### Handler Principal

```go
// main.go
func handler(ctx context.Context, sqsEvent events.SQSEvent) error {
    for _, record := range sqsEvent.Records {
        if err := processMessage(ctx, record); err != nil {
            log.Printf("ERROR processing message: %v", err)
            return err // SQS reintentará
        }
    }
    return nil
}

func processMessage(ctx context.Context, record events.SQSMessage) error {
    var msg SQSMessageBody
    if err := json.Unmarshal([]byte(record.Body), &msg); err != nil {
        return fmt.Errorf("failed to parse message: %w", err)
    }

    log.Printf("[%s] Processing conversion", msg.ConversionID)

    // Actualizar status
    if err := UpdateStatusToProcessing(ctx, msg.ConversionID); err != nil {
        return err
    }

    // Invocar Bedrock
    design, err := InvokeConversion(ctx, msg.SQLContent, msg.OptimizationType)
    if err != nil {
        UpdateStatusToFailed(ctx, msg.ConversionID, err.Error())
        return nil // No reintentar errores de Bedrock
    }

    // Validar diseño
    if err := ValidateDesignFull(design); err != nil {
        UpdateStatusToFailed(ctx, msg.ConversionID, "Invalid design: "+err.Error())
        return nil
    }

    // Guardar diseño
    if err := UpdateStatusToDesignCompleted(ctx, msg.ConversionID, design); err != nil {
        return err
    }

    // Encolar para access patterns
    return SendToAccessPatternQueue(ctx, msg.ConversionID, msg.SQLContent, design)
}
```

### Integración Bedrock

```go
// bedrock.go
func InvokeConversion(ctx context.Context, sql string, optimizationType string) (*DynamoDBDesign, error) {
    prompt := buildConversionPrompt(sql, optimizationType)

    input := &bedrockruntime.InvokeModelInput{
        ModelId:     aws.String("us.anthropic.claude-3-5-sonnet-20241022-v2:0"),
        ContentType: aws.String("application/json"),
        Accept:      aws.String("application/json"),
        Body:        []byte(prompt),
    }

    result, err := bedrockClient.InvokeModel(ctx, input)
    if err != nil {
        return nil, fmt.Errorf("bedrock invoke failed: %w", err)
    }

    var response BedrockResponse
    if err := json.Unmarshal(result.Body, &response); err != nil {
        return nil, fmt.Errorf("failed to parse response: %w", err)
    }

    var design DynamoDBDesign
    if err := json.Unmarshal([]byte(response.Content[0].Text), &design); err != nil {
        return nil, fmt.Errorf("failed to parse design: %w", err)
    }

    return &design, nil
}

func buildConversionPrompt(sql string, optimizationType string) string {
    return fmt.Sprintf(`{
        "anthropic_version": "bedrock-2023-05-31",
        "max_tokens": 4096,
        "messages": [{
            "role": "user",
            "content": "Convert the following SQL schema to DynamoDB design..."
        }]
    }`, sql, optimizationType)
}
```

---

## Access Pattern Worker

**Ubicación:** `lambda/access-pattern-worker/`
**Trigger:** SQS (`sql-to-nosql-access-pattern-queue`)

### Flujo de Ejecución

```
1. Recibir mensaje SQS
2. Parsear body → AccessPatternMessage
3. Actualizar status a PROCESSING_PATTERNS
4. Invocar Bedrock con SQL + diseño
5. Recibir access patterns generados
6. Mergear con diseño existente
7. Actualizar DynamoDB con resultado final
8. Status → COMPLETED
```

### Handler Principal

```go
// main.go
func processAccessPatternMessage(ctx context.Context, record events.SQSMessage) error {
    var msg AccessPatternMessage
    if err := json.Unmarshal([]byte(record.Body), &msg); err != nil {
        return err
    }

    log.Printf("[%s] Generating access patterns", msg.ConversionID)

    // Actualizar status
    if err := UpdateStatusToProcessingPatterns(ctx, msg.ConversionID); err != nil {
        return err
    }

    // Generar access patterns
    patterns, err := GenerateAccessPatterns(ctx, msg.SQLContent, msg.DynamoDBDesign)
    if err != nil {
        UpdateStatusToPatternsFailed(ctx, msg.ConversionID, err.Error())
        return nil
    }

    // Mergear con diseño
    finalDesign := mergeDesignWithPatterns(msg.DynamoDBDesign, patterns)

    // Actualizar como completado
    return UpdateStatusToCompleted(ctx, msg.ConversionID, finalDesign)
}
```

---

## Query Handler

**Ubicación:** `lambda/query/`
**Trigger:** API Gateway HTTP v2 (`GET /api/v1/schemas`, `GET /api/v1/schemas/{id}`)

### Handler Principal

```go
// main.go
func handler(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
    // Determinar ruta
    pathParams := req.PathParameters
    id, hasID := pathParams["id"]

    if hasID {
        return handleGetByID(ctx, id)
    }
    return handleListAll(ctx)
}

func handleGetByID(ctx context.Context, id string) (events.APIGatewayV2HTTPResponse, error) {
    record, err := GetConversionByID(ctx, id)
    if err != nil {
        if errors.Is(err, ErrNotFound) {
            return errorResponse(404, "Conversion not found")
        }
        return errorResponse(500, "Internal server error")
    }

    // Parsear noSqlSchema si es string
    if record.NoSqlSchema != "" {
        var schema DynamoDBDesign
        json.Unmarshal([]byte(record.NoSqlSchema), &schema)
        record.ParsedSchema = &schema
    }

    return jsonResponse(200, record)
}

func handleListAll(ctx context.Context) (events.APIGatewayV2HTTPResponse, error) {
    records, err := ListConversions(ctx)
    if err != nil {
        return errorResponse(500, "Failed to list conversions")
    }

    return jsonResponse(200, ListResponse{
        Conversions: records,
        Count:       len(records),
    })
}
```

### Cliente DynamoDB

```go
// dynamo.go
func GetConversionByID(ctx context.Context, id string) (*ConversionRecord, error) {
    input := &dynamodb.GetItemInput{
        TableName: aws.String(tableName),
        Key: map[string]types.AttributeValue{
            "conversionId": &types.AttributeValueMemberS{Value: id},
        },
    }

    result, err := dynamoClient.GetItem(ctx, input)
    if err != nil {
        return nil, err
    }

    if result.Item == nil {
        return nil, ErrNotFound
    }

    var record ConversionRecord
    if err := attributevalue.UnmarshalMap(result.Item, &record); err != nil {
        return nil, err
    }

    return &record, nil
}

func ListConversions(ctx context.Context) ([]ConversionRecord, error) {
    today := time.Now().Format("2006-01-02")

    input := &dynamodb.QueryInput{
        TableName:              aws.String(tableName),
        IndexName:              aws.String("byDate"),
        KeyConditionExpression: aws.String("conversionDate = :date"),
        ExpressionAttributeValues: map[string]types.AttributeValue{
            ":date": &types.AttributeValueMemberS{Value: today},
        },
        ScanIndexForward: aws.Bool(false), // Más recientes primero
    }

    result, err := dynamoClient.Query(ctx, input)
    if err != nil {
        return nil, err
    }

    var records []ConversionRecord
    if err := attributevalue.UnmarshalListOfMaps(result.Items, &records); err != nil {
        return nil, err
    }

    return records, nil
}
```

---

## DLQ Handler

**Ubicación:** `lambda/dlq-handler/`
**Trigger:** SQS Dead Letter Queues

### Propósito

Procesa mensajes que fallaron 3 veces, marcando la conversión como FAILED.

### Handler

```go
// main.go
func handler(ctx context.Context, sqsEvent events.SQSEvent) error {
    for _, record := range sqsEvent.Records {
        var msg SQSMessageBody
        if err := json.Unmarshal([]byte(record.Body), &msg); err != nil {
            log.Printf("Failed to parse DLQ message: %v", err)
            continue
        }

        log.Printf("[%s] Processing from DLQ - max retries exceeded", msg.ConversionID)

        if err := UpdateStatusToFailed(ctx, msg.ConversionID, "Max retries exceeded"); err != nil {
            log.Printf("Failed to update status: %v", err)
        }
    }
    return nil
}
```

---

## Modelos de Datos

### Request/Response

```go
// ConvertRequest - Input del usuario
type ConvertRequest struct {
    SQLContent       string `json:"sqlContent"`
    OptimizationType string `json:"optimizationType"`
}

// ConversionResponse - Respuesta inicial
type ConversionResponse struct {
    ConversionID string `json:"conversionId"`
    Status       string `json:"status"`
    CreatedAt    int64  `json:"createdAt"`
    ExpiresAt    int64  `json:"expiresAt"`
}

// ErrorResponse - Respuesta de error
type ErrorResponse struct {
    Error   string            `json:"error"`
    Details []ValidationError `json:"details,omitempty"`
}
```

### DynamoDB Record

```go
// ConversionRecord - Registro en DynamoDB
type ConversionRecord struct {
    ConversionID     string `dynamodbav:"conversionId"`
    ConversionDate   string `dynamodbav:"conversionDate"`
    CreatedAt        int64  `dynamodbav:"createdAt"`
    ExpiresAt        int64  `dynamodbav:"expiresAt"`
    Status           string `dynamodbav:"status"`
    SQLContent       string `dynamodbav:"sqlContent"`
    OptimizationType string `dynamodbav:"optimizationType"`
    NoSqlSchema      string `dynamodbav:"noSqlSchema"` // JSON string
    TablesExtracted  int    `dynamodbav:"tablesExtracted"`
}
```

### DynamoDB Design (Output de Bedrock)

```go
// DynamoDBDesign - Diseño generado por Claude
type DynamoDBDesign struct {
    Analysis struct {
        Entities       []Entity        `json:"entities"`
        AccessPatterns []AccessPattern `json:"accessPatterns"`
    } `json:"analysis"`

    Design struct {
        TableName   string `json:"tableName"`
        BillingMode string `json:"billingMode"`
        PrimaryKey  struct {
            PartitionKey KeyDefinition `json:"partitionKey"`
            SortKey      KeyDefinition `json:"sortKey,omitempty"`
        } `json:"primaryKey"`
        GlobalSecondaryIndexes []GSIDefinition  `json:"globalSecondaryIndexes"`
        EntitySchemas          []EntitySchema   `json:"entitySchemas"`
        EdgeItems              []EdgeItemSchema `json:"edgeItems,omitempty"`
    } `json:"design"`

    SampleData                  []SampleDataItem       `json:"sampleData"`
    AccessPatternImplementation []AccessPatternImpl    `json:"accessPatternImplementation"`
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
    EntityType string      `json:"entityType"`
    PKPattern  string      `json:"pkPattern"`
    SKPattern  string      `json:"skPattern"`
    Attributes []Attribute `json:"attributes"`
}

type AccessPatternImpl struct {
    PatternID    string `json:"patternId"`
    Description  string `json:"description"`
    Operation    string `json:"operation"` // GET, QUERY, SCAN
    Index        string `json:"index"`     // PRIMARY, GSI1, etc.
    KeyCondition string `json:"keyCondition"`
}
```

---

## Integración con Bedrock

### Modelos Utilizados

| Modelo | ID | Lambda | Propósito |
|--------|-----|--------|-----------|
| Claude 3.5 Sonnet v2 | `us.anthropic.claude-3-5-sonnet-20241022-v2:0` | conversion-worker | Diseño DynamoDB |
| Claude 3.5 Haiku | `us.anthropic.claude-3-5-haiku-20241022-v1:0` | access-pattern-worker | Access patterns |

### Prompt de Conversión

```go
const conversionPrompt = `You are a DynamoDB design expert. Convert the following SQL schema to an optimized DynamoDB single-table design.

SQL Schema:
%s

Optimization Type: %s

Requirements:
1. Use single-table design pattern
2. Define clear PK and SK patterns
3. Include GSIs for access patterns
4. Provide entity schemas with attribute mappings
5. Include sample data examples

Return your response as valid JSON following this structure:
{
    "analysis": {...},
    "design": {...},
    "sampleData": [...],
    "accessPatternImplementation": [...]
}`
```

### Cliente Bedrock

```go
// Inicialización
func init() {
    cfg, err := config.LoadDefaultConfig(context.Background(),
        config.WithRegion("us-east-1"),
    )
    if err != nil {
        log.Fatal(err)
    }
    bedrockClient = bedrockruntime.NewFromConfig(cfg)
}

// Invocación
func invokeModel(ctx context.Context, prompt string) (string, error) {
    body := map[string]interface{}{
        "anthropic_version": "bedrock-2023-05-31",
        "max_tokens":        4096,
        "messages": []map[string]string{
            {"role": "user", "content": prompt},
        },
    }

    bodyBytes, _ := json.Marshal(body)

    output, err := bedrockClient.InvokeModel(ctx, &bedrockruntime.InvokeModelInput{
        ModelId:     aws.String(modelID),
        ContentType: aws.String("application/json"),
        Accept:      aws.String("application/json"),
        Body:        bodyBytes,
    })

    if err != nil {
        return "", err
    }

    var response struct {
        Content []struct {
            Text string `json:"text"`
        } `json:"content"`
    }
    json.Unmarshal(output.Body, &response)

    return response.Content[0].Text, nil
}
```

---

## Testing

### Tests Unitarios

```bash
cd lambda/diagrams
go test -v ./...
```

### Ejemplo de Test

```go
// validator_test.go
func TestValidateSQL_ValidCreateTable(t *testing.T) {
    sql := `CREATE TABLE users (
        id SERIAL PRIMARY KEY,
        email VARCHAR(255) NOT NULL,
        created_at TIMESTAMP DEFAULT NOW()
    );`

    result := ValidateSQL(sql)

    if !result.Valid {
        t.Errorf("Expected valid, got errors: %v", result.Errors)
    }

    if len(result.Tables) != 1 {
        t.Errorf("Expected 1 table, got %d", len(result.Tables))
    }

    if result.Tables[0].Name != "users" {
        t.Errorf("Expected table name 'users', got '%s'", result.Tables[0].Name)
    }
}

func TestValidateSQL_NoPrimaryKey(t *testing.T) {
    sql := `CREATE TABLE logs (
        message TEXT,
        created_at TIMESTAMP
    );`

    result := ValidateSQL(sql)

    if !result.Valid {
        t.Error("Table without PK should be valid (warning only)")
    }

    if len(result.Warnings) == 0 {
        t.Error("Expected warning for missing PRIMARY KEY")
    }
}
```

---

## Build y Deploy

### Compilación

```bash
# Makefile target
lambda:
    @echo "Building Lambda functions..."
    cd lambda/diagrams && GOOS=linux GOARCH=arm64 go build -o bootstrap
    cd lambda/diagrams && zip -j ../../dist/diagrams.zip bootstrap

    cd lambda/conversion-worker && GOOS=linux GOARCH=arm64 go build -o bootstrap
    cd lambda/conversion-worker && zip -j ../../dist/conversion-worker.zip bootstrap

    # ... resto de lambdas
```

### Variables de Entorno por Lambda

| Lambda | Variable | Valor |
|--------|----------|-------|
| process-handler | `DYNAMODB_TABLE_NAME` | schemas |
| process-handler | `SQS_QUEUE_URL` | conversion-queue URL |
| conversion-worker | `DYNAMODB_TABLE_NAME` | schemas |
| conversion-worker | `SQS_ACCESS_PATTERN_QUEUE_URL` | access-pattern-queue URL |
| conversion-worker | `BEDROCK_MODEL_ID` | us.anthropic.claude-3-5-sonnet-20241022-v2:0 |
| access-pattern-worker | `DYNAMODB_TABLE_NAME` | schemas |
| access-pattern-worker | `BEDROCK_MODEL_ID` | us.anthropic.claude-3-5-haiku-20241022-v1:0 |
| query-handler | `DYNAMODB_TABLE_NAME` | schemas |
| dlq-handler | `DYNAMODB_TABLE_NAME` | schemas |

### Despliegue

```bash
# Build + Deploy
make lambda
make prod
```
