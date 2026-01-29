package main

import (
	"fmt"
	"strings"
)

// ValidateSQL orquesta la validacion completa de un schema SQL.
// Retorna un ValidationResult con tablas extraidas, errores y warnings.
func ValidateSQL(sqlContent string) ValidationResult {
	result := ValidationResult{IsValid: true}

	// 1. Contenido vacio
	if strings.TrimSpace(sqlContent) == "" {
		return validationFailed(ValidationDetail{
			Code:     ErrEmptySQLContent,
			Message:  "Field sqlContent is required",
			Severity: SeverityError,
		})
	}

	// 2. Debe contener al menos un CREATE TABLE
	if !containsCreateTableStatement(sqlContent) {
		return validationFailed(ValidationDetail{
			Code:     ErrNoCreateTablesFound,
			Message:  "No CREATE TABLE statements found",
			Severity: SeverityError,
		})
	}

	// 3. Extraer y validar cada sentencia
	statements := extractCreateTableStatements(sqlContent)

	for _, stmt := range statements {
		// Validar caracteres invalidos despues del cierre
		if hasTrailingGarbage(stmt) {
			result.IsValid = false
			result.Errors = append(result.Errors, ValidationDetail{
				Code:     ErrInvalidSQLSyntax,
				Message:  fmt.Sprintf("Unexpected characters after closing parenthesis in statement: %s", strings.TrimSpace(stmt)),
				Severity: SeverityError,
			})
			continue
		}

		tableName := extractTableName(stmt)

		// Validar nombre de tabla
		if tableName == "" || !isValidIdentifier(tableName) {
			result.IsValid = false
			result.Errors = append(result.Errors, ValidationDetail{
				Code:     ErrInvalidTableName,
				Message:  fmt.Sprintf("Invalid table name: %q", tableName),
				Severity: SeverityError,
				Table:    tableName,
			})
			continue
		}

		body := extractTableBody(stmt)
		if body == "" {
			result.IsValid = false
			result.Errors = append(result.Errors, ValidationDetail{
				Code:     ErrInvalidSQLSyntax,
				Message:  fmt.Sprintf("Table %q has empty body", tableName),
				Severity: SeverityError,
				Table:    tableName,
			})
			continue
		}

		elements := splitTableElements(body)
		tableInfo := TableInfo{Name: tableName}
		columnNames := make(map[string]bool)
		hasPK := false

		for _, elem := range elements {
			elem = strings.TrimSpace(elem)
			if elem == "" {
				continue
			}

			if isTableLevelConstraint(elem) {
				if !isValidTableConstraint(elem) {
					result.IsValid = false
					result.Errors = append(result.Errors, ValidationDetail{
						Code:     ErrInvalidConstraintSyntax,
						Message:  fmt.Sprintf("Invalid constraint in table %q: %s", tableName, elem),
						Severity: SeverityError,
						Table:    tableName,
					})
				}
				upper := strings.ToUpper(elem)
				if strings.Contains(upper, "PRIMARY KEY") {
					hasPK = true
				}
				tableInfo.Constraints = append(tableInfo.Constraints, strings.TrimSpace(elem))
			} else {
				colName, valid := isValidColumnDefinition(elem)
				if !valid {
					result.IsValid = false
					tokens := tokenize(elem)
					if len(tokens) >= 2 {
						dt := extractDataType(tokens[1:])
						if !isValidDataType(dt) {
							result.Errors = append(result.Errors, ValidationDetail{
								Code:     ErrInvalidDataType,
								Message:  fmt.Sprintf("Invalid data type %q for column %q in table %q", dt, tokens[0], tableName),
								Severity: SeverityError,
								Table:    tableName,
								Column:   tokens[0],
							})
						} else {
							result.Errors = append(result.Errors, ValidationDetail{
								Code:     ErrInvalidColumnName,
								Message:  fmt.Sprintf("Invalid column definition in table %q: %s", tableName, elem),
								Severity: SeverityError,
								Table:    tableName,
							})
						}
					} else {
						result.Errors = append(result.Errors, ValidationDetail{
							Code:     ErrInvalidSQLSyntax,
							Message:  fmt.Sprintf("Incomplete column definition in table %q: %s", tableName, elem),
							Severity: SeverityError,
							Table:    tableName,
						})
					}
					continue
				}

				colLower := strings.ToLower(colName)
				if columnNames[colLower] {
					result.IsValid = false
					result.Errors = append(result.Errors, ValidationDetail{
						Code:     ErrDuplicateColumn,
						Message:  fmt.Sprintf("Duplicate column %q in table %q", colName, tableName),
						Severity: SeverityError,
						Table:    tableName,
						Column:   colName,
					})
					continue
				}
				columnNames[colLower] = true

				// Extraer info de columna
				tokens := tokenize(elem)
				dt := ""
				if len(tokens) >= 2 {
					dt = extractDataType(tokens[1:])
				}

				// Detectar PK inline
				if strings.Contains(strings.ToUpper(elem), "PRIMARY KEY") {
					hasPK = true
				}

				tableInfo.Columns = append(tableInfo.Columns, ColumnInfo{
					Name:     colName,
					DataType: dt,
					Raw:      strings.TrimSpace(elem),
				})
			}
		}

		if len(tableInfo.Columns) == 0 {
			result.IsValid = false
			result.Errors = append(result.Errors, ValidationDetail{
				Code:     ErrInvalidSQLSyntax,
				Message:  fmt.Sprintf("Table %q has no valid columns", tableName),
				Severity: SeverityError,
				Table:    tableName,
			})
			continue
		}

		tableInfo.HasPrimaryKey = hasPK
		if !hasPK {
			result.Warnings = append(result.Warnings, ValidationDetail{
				Code:     WarnNoPrimaryKey,
				Message:  fmt.Sprintf("Table %q has no PRIMARY KEY defined", tableName),
				Severity: SeverityWarning,
				Table:    tableName,
			})
		}

		result.Tables = append(result.Tables, tableInfo)
	}

	return result
}

func validationFailed(detail ValidationDetail) ValidationResult {
	return ValidationResult{
		IsValid: false,
		Errors:  []ValidationDetail{detail},
	}
}
