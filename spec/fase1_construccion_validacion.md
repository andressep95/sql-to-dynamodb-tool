# Fase 1: Construcción del Sistema de Recepción y Validación SQL

## Contexto

Este documento detalla la construcción de la primera fase del sistema SQL to DynamoDB Parser. El equipo **NO cuenta con Go ni Terraform instalado**, por lo que esta fase se enfoca en:

1. Diseñar la arquitectura de infraestructura (sin implementarla aún)
2. Definir las estructuras de datos y lógica de validación SQL
3. Establecer los contratos de API
4. Preparar la documentación técnica para futuras implementaciones

## Objetivo de la Fase

Crear la especificación completa y la lógica de validación de esquemas SQL (PostgreSQL) que serán:
- Recibidos vía API Gateway como String
- Validados en la Lambda Process Handler
- Estructurados en modelos de datos consistentes

---

## Arquitectura General

```
Usuario → API Gateway → Process Handler Lambda → Validación SQL
                              ↓
                    [PENDING] → DynamoDB
                              ↓
                          SQS Queue
```

---

# Especificaciones por Componente

## 1. API Gateway - Endpoint de Recepción

### Archivo: `spec/api_contracts.md`

**Crear este archivo nuevo** con el siguiente contenido:

```markdown
# API Contracts - SQL to DynamoDB Parser

## Endpoint: POST /convert

### Request

**URL**: `POST /api/convert`

**Headers**:
```json
{
  "Content-Type": "application/json",
  "X-Origin-Verify": "<secret-token>"
}
```

**Body**:
```json
{
  "sqlContent": "CREATE TABLE users (\n  id SERIAL PRIMARY KEY,\n  email VARCHAR(255) UNIQUE NOT NULL\n);",
  "optimizationType": "balanced"
}
```

**Campos**:
- `sqlContent` (string, requerido): Contenido SQL completo con sentencias CREATE TABLE e índices
- `optimizationType` (string, opcional): Tipo de optimización deseada
  - Valores válidos: `read_heavy`, `write_heavy`, `balanced`
  - Default: `balanced`

### Response

**Success (202 Accepted)**:
```json
{
  "conversionId": "550e8400-e29b-41d4-a716-446655440000",
  "status": "PENDING",
  "createdAt": "2026-01-28T10:30:00Z",
  "expiresAt": "2026-01-29T10:30:00Z"
}
```

**Error (400 Bad Request)**:
```json
{
  "error": "INVALID_SQL_SYNTAX",
  "message": "Syntax error at line 3: unexpected token 'CRATE'",
  "details": {
    "line": 3,
    "column": 1,
    "suggestion": "Did you mean 'CREATE'?"
  }
}
```

**Error Codes**:
- `INVALID_SQL_SYNTAX`: Error de sintaxis SQL
- `EMPTY_SQL_CONTENT`: Campo sqlContent vacío
- `INVALID_OPTIMIZATION_TYPE`: Tipo de optimización no válido
- `NO_CREATE_TABLES_FOUND`: No se encontraron sentencias CREATE TABLE
- `INTERNAL_SERVER_ERROR`: Error interno del servidor

---

## Endpoint: GET /conversions

### Request

**URL**: `GET /api/conversions?date=2026-01-28`

**Query Parameters**:
- `date` (string, opcional): Fecha en formato YYYY-MM-DD
  - Default: fecha actual

### Response

**Success (200 OK)**:
```json
{
  "date": "2026-01-28",
  "conversions": [
    {
      "conversionId": "550e8400-e29b-41d4-a716-446655440000",
      "status": "COMPLETED",
      "createdAt": "2026-01-28T10:30:00Z",
      "tablesCount": 5,
      "preview": "users, orders, products..."
    },
    {
      "conversionId": "660e8400-e29b-41d4-a716-446655440001",
      "status": "PROCESSING",
      "createdAt": "2026-01-28T11:15:00Z",
      "tablesCount": 3,
      "preview": "customers, addresses..."
    }
  ],
  "total": 2
}
```

---

## Endpoint: GET /conversions/{id}

### Request

**URL**: `GET /api/conversions/550e8400-e29b-41d4-a716-446655440000`

### Response

**Success (200 OK)**:
```json
{
  "conversionId": "550e8400-e29b-41d4-a716-446655440000",
  "status": "COMPLETED",
  "createdAt": "2026-01-28T10:30:00Z",
  "completedAt": "2026-01-28T10:32:15Z",
  "expiresAt": "2026-01-29T10:30:00Z",
  "input": {
    "sqlContent": "CREATE TABLE users...",
    "optimizationType": "balanced"
  },
  "result": {
    "tables": [...],
    "accessPatterns": [...],
    "terraformCode": "..."
  }
}
```

**Status Values**:
- `PENDING`: En cola, esperando procesamiento
- `PROCESSING`: Siendo procesado por Bedrock
- `COMPLETED`: Conversión exitosa
- `FAILED`: Error durante la conversión

---

## Endpoint: GET /health

### Request

**URL**: `GET /health`

### Response

**Success (200 OK)**:
```json
{
  "status": "healthy",
  "timestamp": "2026-01-28T10:30:00Z",
  "components": {
    "dynamodb": "healthy",
    "sqs": "healthy",
    "bedrock": "healthy"
  }
}
```
```

---

## 2. Modelos de Datos

### Archivo: `spec/data_models.md`

**Crear este archivo nuevo**:

```markdown
# Data Models - SQL to DynamoDB Parser

## 1. Conversion Record (DynamoDB)

**Tabla**: `schemas_table`

**Schema**:
```json
{
  "conversionId": "string (UUID, PK)",
  "status": "string (PENDING|PROCESSING|COMPLETED|FAILED)",
  "createdAt": "string (ISO 8601)",
  "completedAt": "string (ISO 8601, opcional)",
  "expiresAt": "number (Unix timestamp, TTL)",
  "conversionDate": "string (YYYY-MM-DD, GSI PK)",
  "sqlContent": "string (SQL original)",
  "optimizationType": "string",
  "tablesExtracted": "number",
  "result": {
    "tables": [...],
    "indexes": [...],
    "accessPatterns": [...],
    "terraformCode": "string"
  },
  "errorMessage": "string (opcional)"
}
```

**Indexes**:
- **Primary Key**: `conversionId`
- **GSI**: `conversionDate-createdAt-index`
  - PK: `conversionDate`
  - SK: `createdAt`

**TTL**: Campo `expiresAt` (24 horas desde creación)

---

## 2. Table Metadata

Estructura extraída del SQL para representar una tabla:

```json
{
  "tableName": "string",
  "columns": [
    {
      "columnName": "string",
      "columnType": "string",
      "isNotNull": "boolean",
      "isUnique": "boolean",
      "defaultValue": "string (opcional)"
    }
  ],
  "primaryKeys": ["string"],
  "indexes": [
    {
      "indexName": "string",
      "targetColumns": ["string"],
      "isUnique": "boolean"
    }
  ],
  "relations": [
    {
      "sourceColumn": "string",
      "targetTable": "string",
      "targetColumn": "string",
      "isManyToOne": "boolean"
    }
  ],
  "uniqueConstraints": [
    {
      "constraintName": "string",
      "targetColumns": ["string"]
    }
  ]
}
```

---

## 3. SQS Message Format

Mensaje enviado a la cola para procesamiento asíncrono:

```json
{
  "conversionId": "550e8400-e29b-41d4-a716-446655440000",
  "sqlContent": "CREATE TABLE...",
  "optimizationType": "balanced",
  "createdAt": "2026-01-28T10:30:00Z"
}
```

---

## 4. Validation Result

Estructura interna para resultados de validación:

```json
{
  "isValid": "boolean",
  "errors": [
    {
      "code": "string",
      "message": "string",
      "line": "number (opcional)",
      "column": "number (opcional)",
      "suggestion": "string (opcional)"
    }
  ],
  "warnings": [
    {
      "code": "string",
      "message": "string"
    }
  ],
  "extractedTables": "number",
  "extractedIndexes": "number"
}
```
```

---

## 3. Lógica de Validación SQL

### Archivo: `spec/sql_validation_logic.md`

**Crear este archivo nuevo**:

```markdown
# SQL Validation Logic

## Objetivo

Validar la sintaxis y estructura de sentencias SQL (PostgreSQL) antes de encolarlas para procesamiento. Esta validación se ejecuta en el **Process Handler Lambda**.

---

## Proceso de Validación

### 1. Validación de Entrada Básica

**Checks**:
- ✅ Campo `sqlContent` no está vacío
- ✅ Campo `sqlContent` es un string válido
- ✅ Contenido tiene al menos 10 caracteres
- ✅ `optimizationType` está en los valores permitidos

**Errores posibles**:
- `EMPTY_SQL_CONTENT`: El campo sqlContent está vacío
- `INVALID_OPTIMIZATION_TYPE`: Tipo de optimización no válido

---

### 2. Extracción de Sentencias CREATE TABLE

**Algoritmo**:

```plaintext
1. Limpiar comentarios SQL (-- y /* */)
2. Buscar patrón: CREATE\s+TABLE\s+.*?;
3. Flags: CASE_INSENSITIVE | DOTALL
4. Extraer todas las coincidencias
```

**Validación**:
- ✅ Al menos 1 sentencia CREATE TABLE encontrada
- ✅ Cada sentencia termina con `;`
- ✅ Sintaxis válida: `CREATE TABLE [IF NOT EXISTS] nombre (...)`

**Errores posibles**:
- `NO_CREATE_TABLES_FOUND`: No se encontraron CREATE TABLE
- `INCOMPLETE_STATEMENT`: Sentencia sin terminar (falta `;`)

---

### 3. Validación de Estructura de Tabla

Para cada tabla extraída:

#### 3.1 Validar Nombre de Tabla

**Checks**:
- ✅ Nombre extraído correctamente
- ✅ Formato válido: `[esquema.]nombre_tabla`
- ✅ Caracteres permitidos: letras, números, guiones bajos

**Pattern**: `CREATE\s+TABLE\s+(?:IF\s+NOT\s+EXISTS\s+)?(["\w.-]+)`

---

#### 3.2 Validar Definiciones de Columnas

**Extracción**:
```plaintext
1. Extraer contenido entre paréntesis: CREATE TABLE nombre (...)
2. Dividir por comas, ignorando comas dentro de paréntesis
3. Filtrar líneas de constraints (PRIMARY KEY, FOREIGN KEY, etc.)
```

**Checks por columna**:
- ✅ Nombre de columna válido
- ✅ Tipo de dato especificado
- ✅ Sintaxis de constraints correcta

**Validaciones**:
- Nombre columna: `^"?[\w.-]+"?\s+`
- Tipo dato: Lista de tipos PostgreSQL válidos
- NOT NULL, UNIQUE, DEFAULT, CHECK correctamente formateados

---

#### 3.3 Validar Primary Keys

**Patterns soportados**:

**Inline**:
```sql
id SERIAL PRIMARY KEY
```

**Table-level**:
```sql
PRIMARY KEY (id)
CONSTRAINT pk_nombre PRIMARY KEY (col1, col2)
```

**Checks**:
- ✅ Al menos 1 PK por tabla (warning si no existe)
- ✅ Columnas PK existen en definiciones
- ✅ Sintaxis correcta

---

#### 3.4 Validar Foreign Keys

**Patterns soportados**:

**Inline**:
```sql
user_id INTEGER REFERENCES users(id)
```

**Table-level**:
```sql
FOREIGN KEY (user_id) REFERENCES users(id)
CONSTRAINT fk_name FOREIGN KEY (col) REFERENCES table(col)
```

**Checks**:
- ✅ Sintaxis correcta de REFERENCES
- ✅ Tabla referenciada mencionada
- ✅ Columnas fuente y destino especificadas

**Warning**: No se valida que las tablas referenciadas existan (puede estar en otro schema)

---

#### 3.5 Validar Constraints UNIQUE

**Patterns**:

**Inline**:
```sql
email VARCHAR(255) UNIQUE
```

**Table-level**:
```sql
UNIQUE (email)
CONSTRAINT uk_name UNIQUE (col1, col2)
```

**Checks**:
- ✅ Sintaxis correcta
- ✅ Columnas existen

---

### 4. Extracción y Validación de Índices

**Pattern**:
```sql
CREATE [UNIQUE] INDEX [IF NOT EXISTS] nombre ON tabla (columnas);
```

**Checks**:
- ✅ Nombre de índice válido
- ✅ Tabla especificada
- ✅ Al menos 1 columna
- ✅ Sintaxis de columnas correcta (soporta funciones como LOWER(email))

---

## Algoritmos de Validación

### Algoritmo 1: Limpieza de SQL

```plaintext
FUNCTION cleanSQL(sql: string) -> string:
  1. Eliminar comentarios de línea: sql.replaceAll("--[^\n]*", "")
  2. Eliminar comentarios de bloque: sql.replaceAll("/\*.*?\*/", " ")
  3. Trim espacios
  4. RETURN sql limpio
END FUNCTION
```

---

### Algoritmo 2: Extracción de Nombre de Tabla

```plaintext
FUNCTION extractTableName(sql: string) -> string:
  1. Pattern = "CREATE\s+TABLE\s+(?:IF\s+NOT\s+EXISTS\s+)?(["\w.-]+)"
  2. Matcher = Pattern.match(sql)
  3. IF matcher.found():
       tableName = matcher.group(1)
       tableName = tableName.replaceAll("\"", "")  // Quitar comillas
       tableName = tableName.replaceAll(".*\.", "") // Quitar esquema
       RETURN tableName
  4. RETURN null
END FUNCTION
```

---

### Algoritmo 3: Extracción de Columnas

```plaintext
FUNCTION extractColumns(sql: string) -> List<string>:
  1. Pattern = "CREATE\s+TABLE\s+["\w.-]+\s*\((.*?)\);"
  2. Matcher = Pattern.match(sql)
  3. IF NOT matcher.found():
       RETURN []

  4. content = matcher.group(1)
  5. content = cleanSQL(content)

  6. lines = content.split(",(?![^\(]*\))")  // Split por coma, ignorando dentro de ()

  7. columns = []
  8. FOR EACH line IN lines:
       line = line.trim()
       IF NOT line.startsWith("PRIMARY") AND
          NOT line.startsWith("FOREIGN") AND
          NOT line.startsWith("CONSTRAINT") AND
          NOT line.startsWith("CHECK") AND
          NOT line.startsWith("UNIQUE"):
            columns.add(line)

  9. RETURN columns
END FUNCTION
```

---

### Algoritmo 4: Validación de Tipo de Dato

```plaintext
FUNCTION extractColumnType(columnDef: string) -> string:
  1. Pattern = "\s+([A-Za-z]+(?:\s*\([^)]*\))?)\s*"
  2. Matcher = Pattern.match(columnDef)
  3. IF matcher.found():
       dataType = matcher.group(1).trim().toUpperCase()

       // Detectar arrays
       IF columnDef.contains("ARRAY") OR columnDef.contains("[]"):
          dataType = dataType + "[]"

       RETURN dataType

  4. RETURN null
END FUNCTION

FUNCTION isValidPostgresType(type: string) -> boolean:
  validTypes = [
    "INTEGER", "SERIAL", "BIGSERIAL", "BIGINT", "SMALLINT",
    "NUMERIC", "DECIMAL", "REAL", "DOUBLE PRECISION",
    "VARCHAR", "CHAR", "TEXT", "UUID",
    "BOOLEAN", "DATE", "TIMESTAMP", "TIMESTAMPTZ",
    "JSON", "JSONB", "BYTEA", "ARRAY"
  ]

  baseType = type.replaceAll("\(.*?\)", "").replaceAll("\[\]", "")
  RETURN validTypes.contains(baseType)
END FUNCTION
```

---

### Algoritmo 5: Extracción de Primary Keys

```plaintext
FUNCTION extractPrimaryKeys(sql: string) -> List<string>:
  1. cleanSql = cleanSQL(sql)
  2. primaryKeys = []

  3. // Pattern para PK inline
     inlinePattern = "(["\w.-]+)\s+\w+(?:\([^)]*\))?\s*(?:NOT\s+NULL\s+)?PRIMARY\s+KEY"

  4. // Pattern para PK a nivel tabla
     tablePattern = "PRIMARY\s+KEY\s*\(([^)]+)\)"

  5. Matcher inlineMatcher = inlinePattern.match(cleanSql)
  6. WHILE inlineMatcher.found():
       column = inlineMatcher.group(1).trim()
       column = cleanColumnName(column)
       primaryKeys.add(column)

  7. Matcher tableMatcher = tablePattern.match(cleanSql)
  8. WHILE tableMatcher.found():
       columnsGroup = tableMatcher.group(1)
       columns = columnsGroup.split(",")
       FOR EACH col IN columns:
          cleanCol = cleanColumnName(col.trim())
          IF NOT primaryKeys.contains(cleanCol):
             primaryKeys.add(cleanCol)

  9. RETURN primaryKeys
END FUNCTION
```

---

### Algoritmo 6: Extracción de Foreign Keys

```plaintext
FUNCTION extractForeignKeys(sql: string) -> List<RelationMetadata>:
  1. cleanSql = cleanSQL(sql)
  2. relations = []

  3. // Pattern para FK inline
     inlinePattern = "(["\w.-]+)\s+\w+(?:\([^)]*\))?\s+REFERENCES\s+(["\w.-]+)\s*\(([^)]+)\)"

  4. // Pattern para FK a nivel tabla
     tablePattern = "FOREIGN\s+KEY\s*\(([^)]+)\)\s+REFERENCES\s+(["\w.-]+)\s*\(([^)]+)\)"

  5. Matcher inlineMatcher = inlinePattern.match(cleanSql)
  6. WHILE inlineMatcher.found():
       sourceCol = cleanColumnName(inlineMatcher.group(1))
       targetTable = cleanTableName(inlineMatcher.group(2))
       targetCol = cleanColumnName(inlineMatcher.group(3))

       relation = {
         sourceColumn: sourceCol,
         targetTable: targetTable,
         targetColumn: targetCol,
         isManyToOne: !isUniqueColumn(sourceCol, cleanSql)
       }
       relations.add(relation)

  7. Matcher tableMatcher = tablePattern.match(cleanSql)
  8. WHILE tableMatcher.found():
       // Similar al inline

  9. RETURN relations
END FUNCTION
```

---

## Códigos de Error de Validación

| Código | Descripción | Severidad |
|--------|-------------|-----------|
| `EMPTY_SQL_CONTENT` | Campo sqlContent vacío | ERROR |
| `INVALID_SQL_SYNTAX` | Sintaxis SQL inválida | ERROR |
| `NO_CREATE_TABLES_FOUND` | No se encontraron CREATE TABLE | ERROR |
| `INCOMPLETE_STATEMENT` | Sentencia sin terminar (falta `;`) | ERROR |
| `INVALID_TABLE_NAME` | Nombre de tabla inválido | ERROR |
| `INVALID_COLUMN_NAME` | Nombre de columna inválido | ERROR |
| `INVALID_DATA_TYPE` | Tipo de dato no reconocido | ERROR |
| `INVALID_CONSTRAINT_SYNTAX` | Sintaxis de constraint incorrecta | ERROR |
| `NO_PRIMARY_KEY` | Tabla sin primary key | WARNING |
| `DUPLICATE_COLUMN` | Columna duplicada | ERROR |
| `FK_INVALID_REFERENCE` | Foreign key con sintaxis inválida | ERROR |

---

## Formato de Respuesta de Validación

```json
{
  "isValid": true,
  "errors": [],
  "warnings": [
    {
      "code": "NO_PRIMARY_KEY",
      "message": "Table 'logs' has no primary key defined",
      "table": "logs"
    }
  ],
  "summary": {
    "tablesExtracted": 5,
    "indexesExtracted": 8,
    "foreignKeysExtracted": 12,
    "uniqueConstraintsExtracted": 3
  },
  "tables": [
    {
      "tableName": "users",
      "columnsCount": 8,
      "primaryKey": ["id"],
      "foreignKeys": 0,
      "indexes": 2
    }
  ]
}
```
```

---

## 4. Flujo del Process Handler Lambda

### Archivo: `spec/process_handler_flow.md`

**Crear este archivo nuevo**:

```markdown
# Process Handler Lambda - Flujo de Ejecución

## Responsabilidades

1. Recibir request del API Gateway (POST /convert)
2. Validar el payload básico
3. Ejecutar validación SQL completa
4. Crear registro en DynamoDB con estado PENDING
5. Encolar mensaje en SQS
6. Retornar respuesta inmediata al cliente

---

## Diagrama de Flujo

```plaintext
START
  ↓
[1. Recibir Request]
  ↓
[2. Validar Payload Básico]
  ├─ NO → [Retornar 400 Bad Request]
  ↓ SÍ
[3. Ejecutar Validación SQL]
  ├─ ERRORES → [Retornar 400 con detalles]
  ↓ VÁLIDO
[4. Generar Conversion ID (UUID)]
  ↓
[5. Crear Registro DynamoDB]
  └─ conversionId: UUID
  └─ status: PENDING
  └─ createdAt: timestamp
  └─ expiresAt: timestamp + 24h
  └─ conversionDate: YYYY-MM-DD
  └─ sqlContent: input
  └─ optimizationType: input
  ↓
[6. Encolar Mensaje SQS]
  └─ conversionId
  └─ sqlContent
  └─ optimizationType
  ↓
[7. Retornar 202 Accepted]
  └─ conversionId
  └─ status: PENDING
  └─ createdAt
  └─ expiresAt
  ↓
END
```

---

## Pseudocódigo

```plaintext
FUNCTION handleConvertRequest(event: APIGatewayEvent) -> Response:

  // 1. Parse request body
  TRY:
    body = JSON.parse(event.body)
  CATCH:
    RETURN Response(400, {error: "INVALID_JSON", message: "Request body is not valid JSON"})

  // 2. Validación básica
  IF body.sqlContent IS empty:
    RETURN Response(400, {error: "EMPTY_SQL_CONTENT", message: "Field sqlContent is required"})

  IF body.optimizationType IS set AND NOT IN ["read_heavy", "write_heavy", "balanced"]:
    RETURN Response(400, {error: "INVALID_OPTIMIZATION_TYPE", message: "Invalid optimization type"})

  optimizationType = body.optimizationType OR "balanced"

  // 3. Validación SQL
  validationResult = validateSQL(body.sqlContent)

  IF NOT validationResult.isValid:
    RETURN Response(400, {
      error: "INVALID_SQL_SYNTAX",
      message: validationResult.errors[0].message,
      details: validationResult.errors
    })

  // 4. Generar ID de conversión
  conversionId = generateUUID()
  now = getCurrentTimestamp()
  expiresAt = now + 86400  // +24 horas en segundos
  conversionDate = formatDate(now, "YYYY-MM-DD")

  // 5. Crear registro en DynamoDB
  conversionRecord = {
    conversionId: conversionId,
    status: "PENDING",
    createdAt: now,
    expiresAt: expiresAt,
    conversionDate: conversionDate,
    sqlContent: body.sqlContent,
    optimizationType: optimizationType,
    tablesExtracted: validationResult.summary.tablesExtracted
  }

  TRY:
    dynamoDB.putItem("schemas_table", conversionRecord)
  CATCH error:
    LOG.error("Failed to create DynamoDB record", error)
    RETURN Response(500, {error: "INTERNAL_SERVER_ERROR", message: "Failed to create conversion"})

  // 6. Encolar en SQS
  sqsMessage = {
    conversionId: conversionId,
    sqlContent: body.sqlContent,
    optimizationType: optimizationType,
    createdAt: now
  }

  TRY:
    sqs.sendMessage("conversion_queue", JSON.stringify(sqsMessage))
  CATCH error:
    LOG.error("Failed to enqueue message", error)
    // No fallar la request, el mensaje puede reintentar via DLQ

  // 7. Retornar respuesta
  RETURN Response(202, {
    conversionId: conversionId,
    status: "PENDING",
    createdAt: now,
    expiresAt: expiresAt
  })

END FUNCTION
```

---

## Funciones Auxiliares

### generateUUID()

```plaintext
FUNCTION generateUUID() -> string:
  RETURN randomUUIDv4()
END FUNCTION
```

### getCurrentTimestamp()

```plaintext
FUNCTION getCurrentTimestamp() -> string:
  RETURN new Date().toISOString()
END FUNCTION
```

### formatDate()

```plaintext
FUNCTION formatDate(timestamp: string, format: string) -> string:
  date = parseDate(timestamp)
  IF format == "YYYY-MM-DD":
    RETURN date.getFullYear() + "-" +
           padZero(date.getMonth() + 1) + "-" +
           padZero(date.getDate())
  END IF
END FUNCTION
```

---

## Manejo de Errores

### Errores Recuperables
- Validación SQL fallida → Retornar 400 con detalles
- Payload inválido → Retornar 400

### Errores No Recuperables
- Fallo DynamoDB → Retornar 500
- Fallo SQS → Log error pero NO fallar request (mensaje se reintenta)

---

## Logs y Métricas

### Logs Requeridos

```plaintext
INFO: [conversionId] Conversion request received
INFO: [conversionId] SQL validation passed (tables: N, indexes: M)
INFO: [conversionId] DynamoDB record created
INFO: [conversionId] SQS message enqueued
INFO: [conversionId] Response sent (202)

ERROR: [conversionId] SQL validation failed: <error>
ERROR: [conversionId] DynamoDB put failed: <error>
ERROR: [conversionId] SQS send failed: <error>
```

### Métricas CloudWatch

- `ConversionRequestsReceived` (Count)
- `ValidationPassed` (Count)
- `ValidationFailed` (Count)
- `DynamoDBWriteSuccess` (Count)
- `DynamoDBWriteFailure` (Count)
- `SQSEnqueueSuccess` (Count)
- `SQSEnqueueFailure` (Count)
- `ProcessHandlerDuration` (Milliseconds)

---

## Configuración Lambda

```yaml
Function: ProcessHandlerLambda
Runtime: Go 1.x (cuando se implemente)
Architecture: ARM64 (Graviton2)
Memory: 512 MB
Timeout: 10 segundos
Environment Variables:
  - DYNAMODB_TABLE_NAME: schemas_table
  - SQS_QUEUE_URL: <queue-url>
  - AWS_REGION: us-east-1
IAM Permissions:
  - dynamodb:PutItem
  - sqs:SendMessage
  - logs:CreateLogGroup
  - logs:CreateLogStream
  - logs:PutLogEvents
```
```

---

## 5. Query Handler Lambda

### Archivo: `spec/query_handler_flow.md`

**Crear este archivo nuevo**:

```markdown
# Query Handler Lambda - Flujo de Ejecución

## Responsabilidades

1. Manejar GET /conversions (listar conversiones del día)
2. Manejar GET /conversions/{id} (obtener detalle de conversión)
3. Consultar DynamoDB eficientemente usando GSI
4. Retornar datos formateados al cliente

---

## Endpoint 1: GET /conversions

### Flujo

```plaintext
START
  ↓
[1. Recibir Request]
  ↓
[2. Extraer Query Parameter 'date']
  └─ Si no existe → usar fecha actual
  ↓
[3. Validar formato fecha (YYYY-MM-DD)]
  ├─ NO → [Retornar 400]
  ↓ SÍ
[4. Query DynamoDB usando GSI]
  └─ GSI: conversionDate-createdAt-index
  └─ PK: conversionDate = <fecha>
  └─ SortKey: createdAt DESC
  └─ Limit: 100
  ↓
[5. Formatear Resultados]
  └─ Mapear cada item a formato de listado
  └─ Preview: primeras 3 tablas
  ↓
[6. Retornar 200 OK]
  └─ date
  └─ conversions[]
  └─ total
  ↓
END
```

### Pseudocódigo

```plaintext
FUNCTION handleListConversions(event: APIGatewayEvent) -> Response:

  // 1. Extraer parámetro date
  date = event.queryStringParameters?.date

  IF date IS null:
    date = formatDate(getCurrentTimestamp(), "YYYY-MM-DD")

  // 2. Validar formato
  IF NOT isValidDateFormat(date):
    RETURN Response(400, {error: "INVALID_DATE_FORMAT", message: "Date must be YYYY-MM-DD"})

  // 3. Query DynamoDB usando GSI
  TRY:
    result = dynamoDB.query({
      TableName: "schemas_table",
      IndexName: "conversionDate-createdAt-index",
      KeyConditionExpression: "conversionDate = :date",
      ExpressionAttributeValues: {
        ":date": date
      },
      ScanIndexForward: false,  // DESC order
      Limit: 100
    })
  CATCH error:
    LOG.error("DynamoDB query failed", error)
    RETURN Response(500, {error: "INTERNAL_SERVER_ERROR"})

  // 4. Formatear resultados
  conversions = []
  FOR EACH item IN result.Items:
    preview = generateTablesPreview(item.sqlContent)

    conversions.add({
      conversionId: item.conversionId,
      status: item.status,
      createdAt: item.createdAt,
      tablesCount: item.tablesExtracted OR 0,
      preview: preview
    })

  // 5. Retornar respuesta
  RETURN Response(200, {
    date: date,
    conversions: conversions,
    total: conversions.length
  })

END FUNCTION
```

---

## Endpoint 2: GET /conversions/{id}

### Flujo

```plaintext
START
  ↓
[1. Recibir Request]
  ↓
[2. Extraer Path Parameter 'id']
  ↓
[3. Validar formato UUID]
  ├─ NO → [Retornar 400]
  ↓ SÍ
[4. GetItem DynamoDB]
  └─ PK: conversionId = <id>
  ↓
[5. Verificar si existe]
  ├─ NO → [Retornar 404]
  ↓ SÍ
[6. Retornar 200 OK con datos completos]
  ↓
END
```

### Pseudocódigo

```plaintext
FUNCTION handleGetConversion(event: APIGatewayEvent) -> Response:

  // 1. Extraer ID del path
  conversionId = event.pathParameters.id

  // 2. Validar UUID
  IF NOT isValidUUID(conversionId):
    RETURN Response(400, {error: "INVALID_CONVERSION_ID", message: "ID must be a valid UUID"})

  // 3. GetItem de DynamoDB
  TRY:
    result = dynamoDB.getItem({
      TableName: "schemas_table",
      Key: {
        conversionId: conversionId
      }
    })
  CATCH error:
    LOG.error("DynamoDB getItem failed", error)
    RETURN Response(500, {error: "INTERNAL_SERVER_ERROR"})

  // 4. Verificar existencia
  IF result.Item IS null:
    RETURN Response(404, {error: "CONVERSION_NOT_FOUND", message: "Conversion not found"})

  item = result.Item

  // 5. Formatear respuesta
  response = {
    conversionId: item.conversionId,
    status: item.status,
    createdAt: item.createdAt,
    completedAt: item.completedAt,
    expiresAt: item.expiresAt,
    input: {
      sqlContent: item.sqlContent,
      optimizationType: item.optimizationType
    }
  }

  // Incluir resultado si está completado
  IF item.status == "COMPLETED" AND item.result IS NOT null:
    response.result = item.result

  // Incluir error si falló
  IF item.status == "FAILED" AND item.errorMessage IS NOT null:
    response.errorMessage = item.errorMessage

  // 6. Retornar
  RETURN Response(200, response)

END FUNCTION
```

---

## Funciones Auxiliares

### generateTablesPreview()

```plaintext
FUNCTION generateTablesPreview(sqlContent: string) -> string:
  tables = extractTableNames(sqlContent)

  IF tables.length == 0:
    RETURN ""

  IF tables.length <= 3:
    RETURN tables.join(", ")

  RETURN tables[0] + ", " + tables[1] + ", " + tables[2] + "..."
END FUNCTION
```

### isValidUUID()

```plaintext
FUNCTION isValidUUID(id: string) -> boolean:
  pattern = "^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$"
  RETURN id.matches(pattern)
END FUNCTION
```

### isValidDateFormat()

```plaintext
FUNCTION isValidDateFormat(date: string) -> boolean:
  pattern = "^\d{4}-\d{2}-\d{2}$"
  IF NOT date.matches(pattern):
    RETURN false

  // Validar fecha real
  TRY:
    parsedDate = parseDate(date, "YYYY-MM-DD")
    RETURN true
  CATCH:
    RETURN false
END FUNCTION
```

---

## Configuración Lambda

```yaml
Function: QueryHandlerLambda
Runtime: Go 1.x (cuando se implemente)
Architecture: ARM64 (Graviton2)
Memory: 256 MB
Timeout: 5 segundos
Environment Variables:
  - DYNAMODB_TABLE_NAME: schemas_table
  - AWS_REGION: us-east-1
IAM Permissions:
  - dynamodb:GetItem
  - dynamodb:Query
  - logs:CreateLogGroup
  - logs:CreateLogStream
  - logs:PutLogEvents
```

---

## Logs y Métricas

### Logs Requeridos

```plaintext
INFO: List conversions request (date: <date>)
INFO: List conversions returned N results
INFO: Get conversion request (id: <id>)
INFO: Get conversion found (id: <id>, status: <status>)

WARN: Conversion not found (id: <id>)
ERROR: DynamoDB query failed: <error>
ERROR: DynamoDB getItem failed: <error>
```

### Métricas CloudWatch

- `ListConversionsRequests` (Count)
- `GetConversionRequests` (Count)
- `ConversionNotFound` (Count)
- `DynamoDBQuerySuccess` (Count)
- `DynamoDBQueryFailure` (Count)
- `QueryHandlerDuration` (Milliseconds)
```

---

## 6. Infraestructura Terraform (Especificación)

### Archivo: `spec/terraform_infrastructure.md`

**Crear este archivo nuevo**:

```markdown
# Terraform Infrastructure Specification

## Nota Importante

Este documento es una **especificación** de la infraestructura que se construirá en el futuro. El equipo NO tiene Terraform instalado actualmente.

---

## Estructura de Módulos

```
infra/terraform/
├── modules/
│   ├── dynamodb/          # Tabla schemas_table con GSI y TTL
│   ├── sqs/               # Cola conversion_queue + DLQ
│   ├── lambda/            # Process Handler + Query Handler
│   ├── gateway/           # API Gateway (REST v1 / HTTP v2)
│   └── iam/               # Roles y políticas
└── environments/
    ├── dev/               # LocalStack
    └── prod/              # AWS
```

---

## Módulo: DynamoDB

### Recursos

**1. Tabla Principal: schemas_table**

```hcl
resource "aws_dynamodb_table" "schemas_table" {
  name           = "schemas_table"
  billing_mode   = "PAY_PER_REQUEST"  # On-demand
  hash_key       = "conversionId"

  attribute {
    name = "conversionId"
    type = "S"
  }

  attribute {
    name = "conversionDate"
    type = "S"
  }

  attribute {
    name = "createdAt"
    type = "S"
  }

  # GSI para consultas por fecha
  global_secondary_index {
    name            = "conversionDate-createdAt-index"
    hash_key        = "conversionDate"
    range_key       = "createdAt"
    projection_type = "ALL"
  }

  # TTL de 24 horas
  ttl {
    attribute_name = "expiresAt"
    enabled        = true
  }

  tags = {
    Environment = var.environment
    Project     = "sql-to-dynamodb-parser"
  }
}
```

**Outputs**:
- `table_name`
- `table_arn`
- `gsi_name`

---

## Módulo: SQS

### Recursos

**1. Cola Principal: conversion_queue**

```hcl
resource "aws_sqs_queue" "conversion_queue" {
  name                       = "conversion-queue"
  visibility_timeout_seconds = 180  # 3 min (más que timeout de Lambda)
  message_retention_seconds  = 345600  # 4 días
  receive_wait_time_seconds  = 20  # Long polling

  # Configurar DLQ
  redrive_policy = jsonencode({
    deadLetterTargetArn = aws_sqs_queue.conversion_dlq.arn
    maxReceiveCount     = 3
  })

  tags = {
    Environment = var.environment
    Project     = "sql-to-dynamodb-parser"
  }
}
```

**2. Dead Letter Queue**

```hcl
resource "aws_sqs_queue" "conversion_dlq" {
  name                      = "conversion-dlq"
  message_retention_seconds = 1209600  # 14 días

  tags = {
    Environment = var.environment
    Project     = "sql-to-dynamodb-parser"
  }
}
```

**Outputs**:
- `queue_url`
- `queue_arn`
- `dlq_url`

---

## Módulo: Lambda

### Process Handler Lambda

```hcl
resource "aws_lambda_function" "process_handler" {
  function_name = "sql-parser-process-handler"
  role          = aws_iam_role.process_handler_role.arn

  # Cuando se compile Go
  filename      = "../../lambda/process-handler/bootstrap.zip"
  handler       = "bootstrap"
  runtime       = "provided.al2"  # Custom runtime para Go
  architectures = ["arm64"]       # Graviton2

  memory_size = 512
  timeout     = 10

  environment {
    variables = {
      DYNAMODB_TABLE_NAME = var.dynamodb_table_name
      SQS_QUEUE_URL       = var.sqs_queue_url
      AWS_REGION          = var.aws_region
      LOG_LEVEL           = "INFO"
    }
  }

  tags = {
    Environment = var.environment
    Project     = "sql-to-dynamodb-parser"
  }
}
```

### Query Handler Lambda

```hcl
resource "aws_lambda_function" "query_handler" {
  function_name = "sql-parser-query-handler"
  role          = aws_iam_role.query_handler_role.arn

  filename      = "../../lambda/query-handler/bootstrap.zip"
  handler       = "bootstrap"
  runtime       = "provided.al2"
  architectures = ["arm64"]

  memory_size = 256
  timeout     = 5

  environment {
    variables = {
      DYNAMODB_TABLE_NAME = var.dynamodb_table_name
      AWS_REGION          = var.aws_region
      LOG_LEVEL           = "INFO"
    }
  }

  tags = {
    Environment = var.environment
    Project     = "sql-to-dynamodb-parser"
  }
}
```

**Outputs**:
- `process_handler_arn`
- `process_handler_invoke_arn`
- `query_handler_arn`
- `query_handler_invoke_arn`

---

## Módulo: API Gateway

### Variables

```hcl
variable "gateway_type" {
  description = "rest-v1 (LocalStack) or http-v2 (AWS)"
  type        = string

  validation {
    condition     = contains(["rest-v1", "http-v2"], var.gateway_type)
    error_message = "Must be rest-v1 or http-v2"
  }
}
```

### Rutas

```hcl
routes = {
  "POST /api/convert"        = { lambda: process_handler }
  "GET /api/conversions"     = { lambda: query_handler }
  "GET /api/conversions/{id}" = { lambda: query_handler }
  "GET /health"              = { lambda: query_handler }
}
```

---

## Módulo: IAM

### Process Handler Role

**Permisos**:
- `dynamodb:PutItem` en schemas_table
- `sqs:SendMessage` en conversion_queue
- `logs:CreateLogGroup`, `logs:CreateLogStream`, `logs:PutLogEvents`

### Query Handler Role

**Permisos**:
- `dynamodb:GetItem` en schemas_table
- `dynamodb:Query` en schemas_table (GSI)
- `logs:CreateLogGroup`, `logs:CreateLogStream`, `logs:PutLogEvents`

---

## Variables de Environment

### dev (LocalStack)

```hcl
environment       = "dev"
aws_region        = "us-east-1"
gateway_type      = "rest-v1"
dynamodb_endpoint = "http://localhost:4566"
sqs_endpoint      = "http://localhost:4566"
```

### prod (AWS)

```hcl
environment  = "prod"
aws_region   = "us-east-1"
gateway_type = "http-v2"
```

---

## Comandos Terraform (Futuros)

```bash
# Inicializar
cd infra/terraform/environments/dev
terraform init

# Plan
terraform plan

# Apply
terraform apply

# Destroy
terraform destroy
```
```

---

# Resumen de Archivos a Crear

| Archivo | Ubicación | Descripción |
|---------|-----------|-------------|
| `api_contracts.md` | `spec/` | Contratos de API completos |
| `data_models.md` | `spec/` | Modelos de datos DynamoDB, SQS, etc. |
| `sql_validation_logic.md` | `spec/` | Algoritmos de validación SQL |
| `process_handler_flow.md` | `spec/` | Flujo del Process Handler Lambda |
| `query_handler_flow.md` | `spec/` | Flujo del Query Handler Lambda |
| `terraform_infrastructure.md` | `spec/` | Especificación de infraestructura |

---

# Próximos Pasos

## Fase 2: Implementación de Lambdas (Requiere Go)

1. Instalar Go 1.21+
2. Implementar lógica de validación SQL basada en `sql_validation_logic.md`
3. Implementar Process Handler Lambda
4. Implementar Query Handler Lambda
5. Escribir tests unitarios

## Fase 3: Despliegue de Infraestructura (Requiere Terraform)

1. Instalar Terraform
2. Implementar módulos según `terraform_infrastructure.md`
3. Desplegar en LocalStack (dev)
4. Probar integración completa
5. Desplegar en AWS (prod)

## Fase 4: Conversión Worker y Bedrock

1. Implementar Conversion Worker Lambda
2. Integrar con Amazon Bedrock
3. Implementar generación de diseños DynamoDB
4. Implementar generación de código Terraform

---

# Validación de la Fase 1

✅ Contratos de API documentados
✅ Modelos de datos especificados
✅ Lógica de validación SQL definida
✅ Flujos de lambdas diagramados
✅ Infraestructura especificada
✅ Referencias a `models.md`, `postgres_extraction.md`, `postgres_index.md` incorporadas

**Estado**: Especificación completa lista para implementación
