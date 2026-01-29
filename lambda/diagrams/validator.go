package main

import (
	"regexp"
	"strings"
)

// ============================================================================
// REGEX PATTERNS
// ============================================================================

var (
	createTableRegex = regexp.MustCompile(`(?is)CREATE\s+TABLE\s+.*?;`)
	tableNameRegex   = regexp.MustCompile(`(?i)CREATE\s+TABLE\s+(?:IF\s+NOT\s+EXISTS\s+)?(?:"?(\w+)"?\.)?"?(\w+)"?\s*\(`)
	identifierRegex  = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_$]*$`)
)

// Tipos de datos válidos de PostgreSQL
var postgresDataTypes = map[string]bool{
	// Numéricos
	"smallint": true, "int2": true, "integer": true, "int": true, "int4": true,
	"bigint": true, "int8": true, "decimal": true, "numeric": true,
	"real": true, "float4": true, "double precision": true, "float8": true,
	"smallserial": true, "serial2": true, "serial": true, "serial4": true,
	"bigserial": true, "serial8": true, "money": true, "float": true,
	// Caracteres
	"character varying": true, "varchar": true, "character": true, "char": true,
	"text": true, "citext": true,
	// Binarios
	"bytea": true,
	// Fecha/Hora
	"timestamp": true, "timestamp without time zone": true,
	"timestamp with time zone": true, "timestamptz": true,
	"date": true, "time": true, "time without time zone": true,
	"time with time zone": true, "timetz": true, "interval": true,
	// Booleano
	"boolean": true, "bool": true,
	// JSON
	"json": true, "jsonb": true,
	// UUID
	"uuid": true,
	// Red
	"cidr": true, "inet": true, "macaddr": true, "macaddr8": true,
	// Geométricos
	"point": true, "line": true, "lseg": true, "box": true,
	"path": true, "polygon": true, "circle": true,
	// Text search
	"tsvector": true, "tsquery": true,
	// XML
	"xml": true,
	// Bit
	"bit": true, "bit varying": true, "varbit": true,
	// Rangos
	"int4range": true, "int8range": true, "numrange": true,
	"tsrange": true, "tstzrange": true, "daterange": true,
}

// ============================================================================
// VALIDACIÓN PRINCIPAL
// ============================================================================

func isValidSchema(schema string) bool {
	// 1. Debe contener al menos un CREATE TABLE
	if !containsCreateTableStatement(schema) {
		return false
	}

	// 2. Extraer y validar cada sentencia
	statements := extractCreateTableStatements(schema)
	for _, stmt := range statements {
		if !isValidCreateTableStatement(stmt) {
			return false
		}
	}

	return true
}

func containsCreateTableStatement(schema string) bool {
	return createTableRegex.MatchString(schema)
}

func extractCreateTableStatements(schema string) []string {
	return createTableRegex.FindAllString(schema, -1)
}

// ============================================================================
// VALIDACIÓN DE SENTENCIA CREATE TABLE
// ============================================================================

func hasTrailingGarbage(stmt string) bool {
	end := strings.LastIndex(stmt, ")")
	if end == -1 {
		return true
	}
	trailing := strings.TrimSpace(stmt[end+1:])
	return trailing != ";" && trailing != ""
}

func isValidCreateTableStatement(stmt string) bool {
	// 0. Verificar que no haya caracteres invalidos despues del cierre )
	if hasTrailingGarbage(stmt) {
		return false
	}

	// 1. Validar nombre de tabla
	tableName := extractTableName(stmt)
	if tableName == "" || !isValidIdentifier(tableName) {
		return false
	}

	// 2. Extraer y validar el cuerpo (columnas y constraints)
	body := extractTableBody(stmt)
	if body == "" {
		return false
	}

	// 3. Parsear elementos (columnas y constraints)
	elements := splitTableElements(body)
	if len(elements) == 0 {
		return false
	}

	// 4. Validar cada elemento
	columnCount := 0
	columnNames := make(map[string]bool)

	for _, elem := range elements {
		elem = strings.TrimSpace(elem)
		if elem == "" {
			continue
		}

		if isTableLevelConstraint(elem) {
			// Validar constraint de tabla
			if !isValidTableConstraint(elem) {
				return false
			}
		} else {
			// Validar columna
			colName, valid := isValidColumnDefinition(elem)
			if !valid {
				return false
			}

			// Verificar duplicados
			colLower := strings.ToLower(colName)
			if columnNames[colLower] {
				return false // Columna duplicada
			}
			columnNames[colLower] = true
			columnCount++
		}
	}

	// Debe tener al menos una columna
	return columnCount > 0
}

// ============================================================================
// EXTRACCIÓN
// ============================================================================

func extractTableName(stmt string) string {
	matches := tableNameRegex.FindStringSubmatch(stmt)
	if len(matches) >= 3 {
		return matches[2] // Nombre de tabla (sin schema)
	}
	return ""
}

func extractTableBody(stmt string) string {
	// Encontrar primer ( y último )
	start := strings.Index(stmt, "(")
	end := strings.LastIndex(stmt, ")")

	if start == -1 || end == -1 || end <= start {
		return ""
	}

	return strings.TrimSpace(stmt[start+1 : end])
}

func splitTableElements(body string) []string {
	var elements []string
	var current strings.Builder
	depth := 0
	inQuotes := false

	for i := 0; i < len(body); i++ {
		ch := body[i]

		switch ch {
		case '\'':
			if i == 0 || body[i-1] != '\\' {
				inQuotes = !inQuotes
			}
			current.WriteByte(ch)
		case '(':
			if !inQuotes {
				depth++
			}
			current.WriteByte(ch)
		case ')':
			if !inQuotes {
				depth--
			}
			current.WriteByte(ch)
		case ',':
			if !inQuotes && depth == 0 {
				elements = append(elements, current.String())
				current.Reset()
			} else {
				current.WriteByte(ch)
			}
		default:
			current.WriteByte(ch)
		}
	}

	if current.Len() > 0 {
		elements = append(elements, current.String())
	}

	return elements
}

// ============================================================================
// VALIDACIÓN DE IDENTIFICADORES
// ============================================================================

func isValidIdentifier(name string) bool {
	if name == "" || len(name) > 63 {
		return false
	}
	return identifierRegex.MatchString(name)
}

// ============================================================================
// VALIDACIÓN DE COLUMNAS
// ============================================================================

// columnConstraintKeywords son palabras clave validas despues del tipo de dato.
// KEY no es valido solo; debe ir como "PRIMARY KEY".
var columnConstraintKeywords = map[string]bool{
	"NOT": true, "NULL": true, "DEFAULT": true,
	"PRIMARY": true, "UNIQUE": true,
	"CHECK": true, "REFERENCES": true, "CONSTRAINT": true,
	"COLLATE": true, "GENERATED": true, "ALWAYS": true,
	"AS": true, "STORED": true, "IDENTITY": true,
	"ON": true, "UPDATE": true, "DELETE": true,
	"CASCADE": true, "RESTRICT": true, "NO": true,
	"ACTION": true, "SET": true, "DEFERRABLE": true,
	"INITIALLY": true, "DEFERRED": true, "IMMEDIATE": true,
	"NOW()": true, "TRUE": true, "FALSE": true,
}

// Tokens que requieren otro token adyacente para ser validos
var mustBeFollowedBy = map[string][]string{
	"PRIMARY": {"KEY"},
	"NOT":     {"NULL", "DEFERRABLE"},
}

var mustBePrecededBy = map[string]string{
	"KEY": "PRIMARY", // KEY solo valido despues de PRIMARY
}

func isValidColumnDefinition(elem string) (string, bool) {
	tokens := tokenize(elem)
	if len(tokens) < 2 {
		return "", false
	}

	// Primer token: nombre de columna
	colName := strings.Trim(tokens[0], `"`)
	if !isValidIdentifier(colName) {
		return "", false
	}

	// Segundo token: tipo de dato
	dataType := extractDataType(tokens[1:])
	if !isValidDataType(dataType) {
		return "", false
	}

	// Determinar cuantos tokens consume el tipo de dato
	constraintStart := 2
	normalizedDT := strings.ToLower(dataType)
	if strings.Contains(normalizedDT, " ") {
		constraintStart = 3 // tipo de dos palabras (e.g. "double precision")
	}

	// Validar tokens restantes como constraint de columna
	for i := constraintStart; i < len(tokens); i++ {
		tok := tokens[i]
		upper := strings.ToUpper(tok)

		// Ignorar valores entre parentesis (CHECK(...), REFERENCES table(col))
		if strings.Contains(tok, "(") {
			continue
		}

		// Ignorar literales numericos y strings (DEFAULT 0, DEFAULT 'value')
		if len(tok) > 0 && (tok[0] == '\'' || (tok[0] >= '0' && tok[0] <= '9') || tok[0] == '-') {
			continue
		}

		// Verificar tokens que requieren predecesor (e.g. KEY solo valido despues de PRIMARY)
		if requiredPrev, isDep := mustBePrecededBy[upper]; isDep {
			if i == constraintStart || strings.ToUpper(tokens[i-1]) != requiredPrev {
				return "", false
			}
			continue
		}

		// Verificar que sea una keyword de constraint conocida
		if columnConstraintKeywords[upper] {
			// Verificar tokens que requieren sucesor (e.g. PRIMARY debe ir seguido de KEY)
			if allowedNext, needs := mustBeFollowedBy[upper]; needs {
				if i+1 >= len(tokens) {
					return "", false
				}
				next := strings.ToUpper(tokens[i+1])
				valid := false
				for _, a := range allowedNext {
					if next == a {
						valid = true
						break
					}
				}
				if !valid {
					return "", false
				}
			}
			continue
		}

		// Token desconocido: podria ser un nombre de tabla despues de REFERENCES,
		// o un valor DEFAULT. Verificar contexto.
		if i > constraintStart {
			prev := strings.ToUpper(tokens[i-1])
			if prev == "REFERENCES" || prev == "DEFAULT" || prev == "COLLATE" {
				continue
			}
		}

		// Si llegamos aqui, es un token no reconocido -> invalido
		return "", false
	}

	return colName, true
}

func tokenize(s string) []string {
	var tokens []string
	var current strings.Builder
	inParens := 0

	s = strings.TrimSpace(s)

	for i := 0; i < len(s); i++ {
		ch := s[i]

		switch ch {
		case '(':
			inParens++
			current.WriteByte(ch)
		case ')':
			inParens--
			current.WriteByte(ch)
		case ' ', '\t', '\n', '\r':
			if inParens > 0 {
				current.WriteByte(ch)
			} else if current.Len() > 0 {
				tokens = append(tokens, current.String())
				current.Reset()
			}
		default:
			current.WriteByte(ch)
		}
	}

	if current.Len() > 0 {
		tokens = append(tokens, current.String())
	}

	return tokens
}

func extractDataType(tokens []string) string {
	if len(tokens) == 0 {
		return ""
	}

	// Tipos de dos palabras
	if len(tokens) >= 2 {
		twoWord := strings.ToLower(tokens[0] + " " + tokens[1])
		twoWordTypes := []string{
			"double precision", "character varying", "bit varying",
			"timestamp without", "timestamp with", "time without", "time with",
		}
		for _, t := range twoWordTypes {
			if strings.HasPrefix(twoWord, t) {
				return twoWord
			}
		}
	}

	return tokens[0]
}

func isValidDataType(dataType string) bool {
	if dataType == "" {
		return false
	}

	// Normalizar
	normalized := strings.ToLower(dataType)

	// Quitar arrays []
	normalized = strings.TrimSuffix(normalized, "[]")

	// Quitar precisión (10,2)
	if idx := strings.Index(normalized, "("); idx != -1 {
		normalized = strings.TrimSpace(normalized[:idx])
	}

	return postgresDataTypes[normalized]
}

// ============================================================================
// VALIDACIÓN DE CONSTRAINTS
// ============================================================================

func isTableLevelConstraint(elem string) bool {
	upper := strings.ToUpper(strings.TrimSpace(elem))
	return strings.HasPrefix(upper, "CONSTRAINT ") ||
		strings.HasPrefix(upper, "PRIMARY KEY") ||
		strings.HasPrefix(upper, "FOREIGN KEY") ||
		strings.HasPrefix(upper, "UNIQUE") ||
		strings.HasPrefix(upper, "CHECK") ||
		strings.HasPrefix(upper, "EXCLUDE")
}

func isValidTableConstraint(elem string) bool {
	upper := strings.ToUpper(strings.TrimSpace(elem))

	// Quitar "CONSTRAINT nombre" si existe
	if strings.HasPrefix(upper, "CONSTRAINT ") {
		parts := strings.Fields(elem)
		if len(parts) < 3 {
			return false
		}
		// Validar nombre del constraint
		if !isValidIdentifier(parts[1]) {
			return false
		}
		// Continuar con el resto
		upper = strings.ToUpper(strings.Join(parts[2:], " "))
	}

	// Verificar tipo de constraint válido
	validPrefixes := []string{
		"PRIMARY KEY", "FOREIGN KEY", "UNIQUE", "CHECK", "EXCLUDE",
	}

	for _, prefix := range validPrefixes {
		if strings.HasPrefix(upper, prefix) {
			return true
		}
	}

	return false
}
