package org.example.database.extractor.postgres;

import org.example.database.extractor.SqlCreateTableStatementExtractor;
import org.example.database.model.RelationMetadata;
import org.example.database.model.TableConstraintData;

import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;
import java.util.regex.Matcher;
import java.util.regex.Pattern;
import java.util.stream.Collectors;

public class PostgresSqlCreateTableStatementExtractor implements SqlCreateTableStatementExtractor {

    @Override
    public List<String> extractCreateTableStatements(String sql) {
        List<String> statements = new ArrayList<>();
        Pattern pattern = Pattern.compile(
            "CREATE\\s+TABLE\\s+.*?;",
            Pattern.CASE_INSENSITIVE | Pattern.DOTALL);
        Matcher matcher = pattern.matcher(sql);

        while (matcher.find()) {
            statements.add(matcher.group());
        }

        return statements;
    }

    @Override
    public String extractTableName(String sql) {
        Pattern pattern = Pattern.compile(
            "(?i)CREATE\\s+TABLE\\s+" +
                "(?:IF\\s+NOT\\s+EXISTS\\s+)?" +
                "((?:\"[^\"]+\"|\\w+)(?:\\.(?:\"[^\"]+\"|\\w+))?)"
                                         );

        Matcher matcher = pattern.matcher(sql);
        if (matcher.find()) {
            // Devuelve todo el nombre encontrado (puede incluir esquema o no)
            String tableName = matcher.group(1);

            // Remover comillas si las hubiera
            return tableName.replaceAll("\"", "").replaceAll(".*\\.", "");
        }
        return null;
    }

    @Override
    public List<String> extractColumnDefinitions(String sql) {
        List<String> columnDefinitions = new ArrayList<>();

        Pattern pattern = Pattern.compile(
            "CREATE\\s+TABLE\\s+[\"\\w.-]+\\s*\\((.*?)\\);",
            Pattern.CASE_INSENSITIVE | Pattern.DOTALL);
        Matcher matcher = pattern.matcher(sql);

        if (matcher.find()) {
            String columnsDefinition = matcher.group(1);

            // Limpiar comentarios
            columnsDefinition = columnsDefinition
                .replaceAll("/\\*.*?\\*/", "") // Eliminar comentarios multilínea
                .replaceAll("--.*?(\\n|$)", "") // Eliminar comentarios en línea
                .trim();

            // Dividir líneas por comas, excluyendo las que están dentro de paréntesis
            String[] lines = columnsDefinition.split(",(?![^\\(]*\\))");

            for (String line : lines) {
                line = line.trim();
                // Filtrar definiciones de restricciones y líneas vacías
                if (!line.isEmpty() &&
                    !line.toUpperCase().startsWith("PRIMARY") &&
                    !line.toUpperCase().startsWith("FOREIGN") &&
                    !line.toUpperCase().startsWith("CONSTRAINT") &&
                    !line.toUpperCase().startsWith("CHECK") &&
                    !line.toUpperCase().startsWith("UNIQUE")) {
                    columnDefinitions.add(line);
                }
            }
        }

        return columnDefinitions;
    }

    @Override
    public String extractColumnName(String columnDefinition) {
        // Expresión regular para capturar el nombre de la columna antes del tipo de dato o constraint
        Pattern pattern = Pattern.compile("^\\s*\"?([\\w.-]+)\"?\\s+");
        Matcher matcher = pattern.matcher(columnDefinition.trim());

        if (matcher.find()) {
            String columnName = matcher.group(1); // Capturar el nombre original de la columna

            // Limpiar el nombre eliminando caracteres no válidos
            columnName = columnName.replaceAll("[^a-zA-Z0-9_]", "_"); // Reemplazar caracteres no válidos con guiones bajos
            columnName = columnName.replaceAll("_+", "_"); // Consolidar guiones bajos consecutivos
            columnName = columnName.replaceAll("^_+|_+$", ""); // Eliminar guiones bajos iniciales o finales

            return columnName; // Retornar el nombre limpio y validado
        }

        return null; // Retornar null si no se encuentra un nombre válido
    }

    @Override
    public String extractColumnType(String sql) {
        // Eliminar comentarios SQL si existen
        sql = sql.replaceAll("--.*$", "");

        // Patrón para capturar el tipo de dato, incluyendo cualquier precisión/escala
        Pattern pattern = Pattern.compile("\\s+([A-Za-z]+(?:\\s*\\([^)]*\\))?)\\s*(?:ARRAY|\\[\\])?");
        Matcher matcher = pattern.matcher(sql);

        if (matcher.find()) {
            String dataType = matcher.group(1).trim();

            // Manejar tipos de arrays
            if (sql.contains("ARRAY") || sql.contains("[]")) {
                dataType += "[]";
            }

            return dataType.toUpperCase();
        }

        return null;
    }

    @Override
    public boolean isNotNullColumn(String columnDefinition) {
        // Verifica NOT NULL explícito
        Pattern notNullPattern = Pattern.compile(".*\\bNOT\\s+NULL\\b.*",
            Pattern.CASE_INSENSITIVE);
        return notNullPattern.matcher(columnDefinition).matches();
    }

    @Override
    public boolean isUniqueColumn(String columnDefinition, String fullTableDDL) {
        // 1. Verificar si la columna tiene UNIQUE explícito en su definición
        Pattern inlineUniquePattern = Pattern.compile("\\bUNIQUE\\b", Pattern.CASE_INSENSITIVE);
        if (inlineUniquePattern.matcher(columnDefinition).find()) {
            return true;
        }

        // 2. Extraer el nombre de la columna
        Pattern columnNamePattern = Pattern.compile("^\\s*\"?([\\w_]+)\"?\\s+", Pattern.CASE_INSENSITIVE);
        Matcher nameMatcher = columnNamePattern.matcher(columnDefinition);
        if (!nameMatcher.find()) {
            return false; // No se pudo extraer el nombre de la columna
        }
        String columnName = nameMatcher.group(1);

        // 3. Buscar UNIQUE constraints al final del DDL que incluyan esta columna
        Pattern constraintUniquePattern = Pattern.compile(
            "UNIQUE\\s*\\(([^)]+)\\)", Pattern.CASE_INSENSITIVE);
        Matcher constraintMatcher = constraintUniquePattern.matcher(fullTableDDL);

        while (constraintMatcher.find()) {
            String columnsGroup = constraintMatcher.group(1); // contenido entre paréntesis
            String[] columns = columnsGroup.split(",");
            for (String col : columns) {
                String cleanCol = col.trim().replaceAll("\"", "");
                if (cleanCol.equalsIgnoreCase(columnName)) {
                    return true;
                }
            }
        }

        return false;
    }

    @Override
    public String extractDefaultValue(String columnDefinition) {
        Pattern pattern = Pattern.compile(
            "DEFAULT\\s+(" +
                "true|false|" +                 // Booleanos
                "'[^']*'|" +                   // Cadenas delimitadas por comillas simples
                "\\d+(\\.\\d+)?|" +            // Números enteros y decimales
                "[a-zA-Z_][a-zA-Z0-9_]*\\(.*?\\)|" + // Funciones con o sin argumentos, como NOW(), UUID(), etc.
                "[a-zA-Z_][a-zA-Z0-9_]*" +     // Identificadores como CURRENT_DATE, CURRENT_TIMESTAMP
                ")",
            Pattern.CASE_INSENSITIVE);

        Matcher matcher = pattern.matcher(columnDefinition);

        if (matcher.find()) {
            return matcher.group(1).trim(); // Retorna el valor encontrado, eliminando espacios adicionales
        }
        return null;
    }

    @Override
    public List<String> extractPrimaryKeyColumns(String sql) {
        List<String> primaryKeys = new ArrayList<>();

        // Limpiar comentarios y procesar todo en una sola pasada
        String cleanSql = sql.replaceAll("--[^\\n]*", "") // Eliminar comentarios en línea
            .replaceAll("/\\*[^*]*\\*+(?:[^/*][^*]*\\*+)*/", " "); // Eliminar comentarios en bloque

        Pattern pkPattern = Pattern.compile(
            // Captura PKs simples
            "(?:(?:\\(|,)\\s*([\"\\w.-]+)\\s+(?:INTEGER|SERIAL|UUID|NUMERIC|BIGINT|VARCHAR|TEXT|TIMESTAMP|DATE|DECIMAL)(?:\\([^)]*\\))?\\s*(?:NOT\\s+NULL\\s+)?(?:UNIQUE\\s+)?(?:CONSTRAINT\\s+\\w+\\s+)?PRIMARY\\s+KEY\\b)" +
                "|" +
                // Captura PKs compuestas
                "(?:CONSTRAINT\\s+[\"\\w.-]+\\s+)?PRIMARY\\s+KEY\\s*\\(([^)]+)\\)",
            Pattern.CASE_INSENSITIVE | Pattern.MULTILINE
                                           );

        Matcher matcher = pkPattern.matcher(cleanSql);
        while (matcher.find()) {
            if (matcher.group(1) != null) {
                // PK simple
                String column = matcher.group(1).trim()
                    .replaceAll("^\"|\"$", "") // Eliminar comillas
                    .replaceAll("\\[.*?\\]", "") // Eliminar índices como [1]
                    .replaceAll("[^a-zA-Z0-9_]", "_") // Reemplazar caracteres especiales por guiones bajos
                    .replaceAll("_+", "_") // Consolidar guiones bajos consecutivos
                    .replaceAll("^_+|_+$", "") // Eliminar guiones bajos iniciales o finales
                    .toLowerCase(); // Convertir a minúsculas
                if (!primaryKeys.contains(column)) {
                    primaryKeys.add(column);
                }
            } else if (matcher.group(2) != null) {
                // PK compuesta
                for (String column : matcher.group(2).split(",")) {
                    String cleanColumn = column.trim()
                        .replaceAll("^\"|\"$", "") // Eliminar comillas
                        .replaceAll("\\[.*?\\]", "") // Eliminar índices como [1]
                        .replaceAll("[^a-zA-Z0-9_]", "_") // Reemplazar caracteres especiales por guiones bajos
                        .replaceAll("_+", "_") // Consolidar guiones bajos consecutivos
                        .replaceAll("^_+|_+$", "") // Eliminar guiones bajos iniciales o finales
                        .toLowerCase(); // Convertir a minúsculas
                    if (!primaryKeys.contains(cleanColumn)) {
                        primaryKeys.add(cleanColumn);
                    }
                }
            }
        }

        return primaryKeys;
    }

    @Override
    public List<RelationMetadata> extractTableRelations(String sql) {
        List<RelationMetadata> relations = new ArrayList<>();

        // Limpiar comentarios
        String cleanSql = sql.replaceAll("--[^\\n]*", "")
            .replaceAll("/\\*[^*]*\\*+(?:[^/*][^*]*\\*+)*/", " ");

        // Debug: Imprimir el SQL limpio para verificar lo que estamos procesando
        System.out.println("SQL limpio para procesar: " + cleanSql);

        // Primero extraemos el nombre de la tabla actual
        String currentTable = extractTableName(cleanSql);
        if (currentTable == null || currentTable.isEmpty()) {
            return relations;
        }

        // Patrón para FK inline en definición de columna
        Pattern inlineFkPattern = Pattern.compile(
            "(?:\\(|,)\\s*([\"\\w.-]+)\\s+(?:\\w+)(?:\\([^)]*\\))?\\s+(?:NOT\\s+NULL\\s+)?(?:UNIQUE\\s+)?(?:DEFAULT\\s+[^\\s,]+\\s+)?REFERENCES\\s+([\"\\w.-]+)\\s*\\(([^)]+)\\)",
            Pattern.CASE_INSENSITIVE | Pattern.MULTILINE
                                                 );

        // Patrón muy estricto para FK con CONSTRAINT nombrado
        Pattern constraintFkPattern = Pattern.compile(
            "CONSTRAINT\\s+(?:[\"\\w.-]+)\\s+FOREIGN\\s+KEY\\s*\\(([^)]+)\\)\\s+REFERENCES\\s+([\"\\w.-]+)\\s*\\(([^)]+)\\)",
            Pattern.CASE_INSENSITIVE | Pattern.MULTILINE
                                                     );

        // Patrón alternativo sin nombre de CONSTRAINT
        Pattern altFkPattern = Pattern.compile(
            "FOREIGN\\s+KEY\\s*\\(([^)]+)\\)\\s+REFERENCES\\s+([\"\\w.-]+)\\s*\\(([^)]+)\\)",
            Pattern.CASE_INSENSITIVE | Pattern.MULTILINE
                                              );

        // Debug: Buscar y mostrar cualquier coincidencia que parezca una clave foránea
        Pattern debugFkPattern = Pattern.compile(
            "(?:FOREIGN|REFERENCES)",
            Pattern.CASE_INSENSITIVE
                                                );
        Matcher debugMatcher = debugFkPattern.matcher(cleanSql);
        while (debugMatcher.find()) {
            int start = Math.max(0, debugMatcher.start() - 20);
            int end = Math.min(cleanSql.length(), debugMatcher.end() + 60);
            System.out.println("Posible FK encontrada: ..." + cleanSql.substring(start, end) + "...");
        }

        // Buscar FK inline
        Matcher inlineMatcher = inlineFkPattern.matcher(cleanSql);
        while (inlineMatcher.find()) {
            String sourceColumn = inlineMatcher.group(1).trim().replaceAll("^\"|\"$", "");
            String targetTable = inlineMatcher.group(2).trim().replaceAll("^\"|\"$", "");
            String targetColumn = inlineMatcher.group(3).trim().replaceAll("^\"|\"$", "");

            // Debug: Mostrar lo que se encontró
            System.out.println("FK inline encontrada: " + sourceColumn + " -> " + targetTable + "." + targetColumn);

            relations.add(new RelationMetadata(
                sourceColumn,
                targetTable,
                targetColumn,
                isColumnUnique(sourceColumn, cleanSql) ? false : true
            ));
        }

        // Buscar FK con CONSTRAINT nombrado
        Matcher constraintMatcher = constraintFkPattern.matcher(cleanSql);
        while (constraintMatcher.find()) {
            // Debug: Mostrar lo que se encontró
            System.out.println("FK con CONSTRAINT encontrada: " + constraintMatcher.group(0));

            processMatchForRelations(constraintMatcher, relations, cleanSql);
        }

        // Buscar FK sin nombre de CONSTRAINT
        Matcher altMatcher = altFkPattern.matcher(cleanSql);
        while (altMatcher.find()) {
            // Debug: Mostrar lo que se encontró
            System.out.println("FK sin CONSTRAINT encontrada: " + altMatcher.group(0));

            processMatchForRelations(altMatcher, relations, cleanSql);
        }

        // Verificación final muy estricta
        List<RelationMetadata> validatedRelations = new ArrayList<>();
        for (RelationMetadata relation : relations) {
            // Comprobar si realmente hay una referencia explícita a la tabla objetivo
            String referencePattern = "REFERENCES\\s+" + Pattern.quote(relation.getTargetTable()) + "\\s*\\(\\s*" +
                Pattern.quote(relation.getTargetColumn()) + "\\s*\\)";

            Matcher referenceMatcher = Pattern.compile(referencePattern, Pattern.CASE_INSENSITIVE).matcher(cleanSql);
            if (referenceMatcher.find()) {
                validatedRelations.add(relation);
                System.out.println("Relación validada: " + relation);
            } else {
                System.out.println("Relación rechazada (no se encontró REFERENCES explícito): " + relation);
            }
        }

        return validatedRelations;
    }

    @Override
    public List<TableConstraintData> extractUniqueConstraints(String fullTableDDL) {
        List<TableConstraintData> uniqueConstraints = new ArrayList<>();
        String cleanSql = fullTableDDL.replaceAll("--[^\\n]*", "")
            .replaceAll("/\\*[^*]*\\*+(?:[^/*][^*]*\\*+)*/", " ");

        String tableName = extractTableName(cleanSql);
        if (tableName == null || tableName.isEmpty()) {
            return uniqueConstraints;
        }

        Pattern tableContentPattern = Pattern.compile(
            "CREATE\\s+TABLE\\s+[\"\\w.-]+\\s*\\((?<content>.*?)\\);",
            Pattern.CASE_INSENSITIVE | Pattern.DOTALL
                                                     );
        Matcher tableContentMatcher = tableContentPattern.matcher(cleanSql);
        String tableDefinitionBlock = "";
        if (tableContentMatcher.find()) {
            tableDefinitionBlock = tableContentMatcher.group("content");
        } else {
            return uniqueConstraints;
        }

        // 1. Extraer restricciones UNIQUE en línea de las definiciones de columna
        List<String> columnDefinitions = extractColumnDefinitions(fullTableDDL); // Este método DEBE haber sido ajustado previamente para NO filtrar la línea 'email UNIQUE'

        for (String columnDef : columnDefinitions) {
            // Usamos el nuevo método auxiliar para construir la restricción si aplica
            TableConstraintData inlineConstraint = buildInlineUniqueConstraint(columnDef, tableName);
            if (inlineConstraint != null) {
                if (!uniqueConstraints.contains(inlineConstraint)) {
                    uniqueConstraints.add(inlineConstraint);
                }
            }
        }

        // 2. Extraer restricciones UNIQUE definidas a nivel de tabla
        Pattern tableUniquePattern = Pattern.compile(
            "(?:CONSTRAINT\\s+\"?(?<constraintName>[\\w.-]+)\"?\\s+)?UNIQUE\\s*\\((?<columns>[^)]+)\\)",
            Pattern.CASE_INSENSITIVE | Pattern.DOTALL
                                                    );
        Matcher tableMatcher = tableUniquePattern.matcher(tableDefinitionBlock);

        while (tableMatcher.find()) {
            String constraintName = tableMatcher.group("constraintName");
            String columnsGroup = tableMatcher.group("columns");

            List<String> columns = Arrays.stream(columnsGroup.split(","))
                .map(s -> cleanName(s.trim()))
                .collect(Collectors.toList());

            if (constraintName == null || constraintName.isEmpty()) {
                String generatedName = "uk_" + tableName.toLowerCase();
                for (String col : columns) {
                    generatedName += "_" + col.toLowerCase();
                }
                constraintName = generatedName;
            } else {
                constraintName = cleanName(constraintName);
            }

            TableConstraintData newConstraint = new TableConstraintData(tableName, constraintName, columns);
            if (!uniqueConstraints.contains(newConstraint)) { // Esto sigue dependiendo de equals/hashCode en TableConstraintData
                uniqueConstraints.add(newConstraint);
            }
        }

        return uniqueConstraints;
    }

    private TableConstraintData buildInlineUniqueConstraint(String columnDefinition, String tableName) {
        // Reutilizamos la lógica del paso 1 de isUniqueColumn
        Pattern inlineUniquePattern = Pattern.compile("\\bUNIQUE\\b", Pattern.CASE_INSENSITIVE);
        if (inlineUniquePattern.matcher(columnDefinition).find()) {
            String columnName = extractColumnName(columnDefinition); // Reutiliza tu método existente
            if (columnName != null) {
                List<String> columns = Arrays.asList(cleanName(columnName)); // Asegura que el nombre esté limpio
                String constraintName = "uk_" + tableName.toLowerCase() + "_" + cleanName(columnName).toLowerCase();
                return new TableConstraintData(tableName, constraintName, columns);
            }
        }
        return null;
    }

    // Método auxiliar para procesar matches de relaciones
    private void processMatchForRelations(Matcher matcher, List<RelationMetadata> relations, String cleanSql) {
        String[] sourceColumns = matcher.group(1).split(",");
        String targetTable = matcher.group(2).trim().replaceAll("^\"|\"$", "");
        String[] targetColumns = matcher.group(3).split(",");

        for (int i = 0; i < sourceColumns.length && i < targetColumns.length; i++) {
            String sourceColumn = sourceColumns[i].trim().replaceAll("^\"|\"$", "");
            String targetColumn = targetColumns[i].trim().replaceAll("^\"|\"$", "");

            relations.add(new RelationMetadata(
                sourceColumn,
                targetTable,
                targetColumn,
                isColumnUnique(sourceColumn, cleanSql) ? false : true
            ));
        }
    }

    // Método auxiliar para determinar si una columna es UNIQUE
    private boolean isColumnUnique(String columnName, String sql) {
        // Comprobar si hay una definición UNIQUE directamente en la columna
        Pattern columnUniquePattern = Pattern.compile(
            columnName + "\\s+\\w+(?:\\([^)]*\\))?\\s+(?:NOT\\s+NULL\\s+)?UNIQUE",
            Pattern.CASE_INSENSITIVE
                                                     );

        // Comprobar si hay una restricción UNIQUE separada
        Pattern constraintUniquePattern = Pattern.compile(
            "CONSTRAINT\\s+[\"\\w.-]+\\s+UNIQUE\\s*\\([^)]*" + columnName + "[^)]*\\)",
            Pattern.CASE_INSENSITIVE
                                                         );

        return columnUniquePattern.matcher(sql).find() || constraintUniquePattern.matcher(sql).find();
    }

    /**
     * Método auxiliar para limpiar los nombres de columnas o constraints.
     * Consolida lógicas de limpieza para evitar repetición.
     */
    private String cleanName(String name) {
        if (name == null) return null;
        String cleaned = name.trim().replaceAll("^\"|\"$", ""); // Eliminar comillas dobles
        cleaned = cleaned.replaceAll("[^a-zA-Z0-9_]", "_"); // Reemplazar caracteres no alfanuméricos por guiones bajos
        cleaned = cleaned.replaceAll("_+", "_"); // Consolidar guiones bajos consecutivos
        cleaned = cleaned.replaceAll("^_+|_+$", ""); // Eliminar guiones bajos al inicio o final
        return cleaned;
    }

}
