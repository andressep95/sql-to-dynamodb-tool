package main

// ============================================================================
// Error Codes (spec/fase1_construccion_validacion.md)
// ============================================================================

const (
	ErrEmptySQLContent        = "EMPTY_SQL_CONTENT"
	ErrInvalidJSON            = "INVALID_JSON"
	ErrInvalidSQLSyntax       = "INVALID_SQL_SYNTAX"
	ErrInvalidOptimizationType = "INVALID_OPTIMIZATION_TYPE"
	ErrNoCreateTablesFound    = "NO_CREATE_TABLES_FOUND"
	ErrInvalidTableName       = "INVALID_TABLE_NAME"
	ErrInvalidColumnName      = "INVALID_COLUMN_NAME"
	ErrInvalidDataType        = "INVALID_DATA_TYPE"
	ErrInvalidConstraintSyntax = "INVALID_CONSTRAINT_SYNTAX"
	ErrDuplicateColumn        = "DUPLICATE_COLUMN"
	ErrFKInvalidReference     = "FK_INVALID_REFERENCE"
	ErrIncompleteStatement    = "INCOMPLETE_STATEMENT"
	ErrInternalServerError    = "INTERNAL_SERVER_ERROR"

	WarnNoPrimaryKey = "NO_PRIMARY_KEY"
)

// Severidad de errores de validacion
const (
	SeverityError   = "ERROR"
	SeverityWarning = "WARNING"
)

// Tipos de optimizacion validos
var validOptimizationTypes = map[string]bool{
	"read_heavy":  true,
	"write_heavy": true,
	"balanced":    true,
}

// ============================================================================
// API Gateway V2 Response (without null fields that break LocalStack)
// ============================================================================

// V2Response is a clean API Gateway V2 response without MultiValueHeaders/Cookies null fields.
type V2Response struct {
	StatusCode      int               `json:"statusCode"`
	Headers         map[string]string `json:"headers,omitempty"`
	Body            string            `json:"body"`
	IsBase64Encoded bool              `json:"isBase64Encoded,omitempty"`
}

// ============================================================================
// Request / Response
// ============================================================================

// ConvertRequest es el body esperado en POST /api/convert
type ConvertRequest struct {
	SQLContent       string `json:"sqlContent"`
	OptimizationType string `json:"optimizationType,omitempty"`
}

// ErrorResponse representa una respuesta de error de la API
type ErrorResponse struct {
	Error   string             `json:"error"`
	Message string             `json:"message"`
	Details []ValidationDetail `json:"details,omitempty"`
}

// ============================================================================
// Validation Result
// ============================================================================

// ValidationResult contiene el resultado completo de la validacion SQL
type ValidationResult struct {
	IsValid  bool              `json:"isValid"`
	Tables   []TableInfo       `json:"tables,omitempty"`
	Errors   []ValidationDetail `json:"errors,omitempty"`
	Warnings []ValidationDetail `json:"warnings,omitempty"`
}

// ValidationDetail describe un error o warning especifico
type ValidationDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Severity string `json:"severity"`
	Table   string `json:"table,omitempty"`
	Column  string `json:"column,omitempty"`
}

// TableInfo contiene metadata extraida de un CREATE TABLE
type TableInfo struct {
	Name        string       `json:"name"`
	Columns     []ColumnInfo `json:"columns"`
	Constraints []string     `json:"constraints,omitempty"`
	HasPrimaryKey bool       `json:"hasPrimaryKey"`
}

// ColumnInfo contiene metadata de una columna
type ColumnInfo struct {
	Name     string `json:"name"`
	DataType string `json:"dataType"`
	Raw      string `json:"raw"`
}
