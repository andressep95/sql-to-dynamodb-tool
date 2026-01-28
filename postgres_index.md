package org.example.database.extractor.index.postgres;

import org.example.database.extractor.index.SqlCreateIndexStatementExtractor;

import java.util.ArrayList;
import java.util.List;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

public class PostgresSqlCreateIndexStatementExtractor implements SqlCreateIndexStatementExtractor {

    // Patrón más simple para encontrar el inicio de las sentencias CREATE INDEX
    private static final Pattern CREATE_INDEX_START_PATTERN = Pattern.compile(
        "CREATE\\s+(?:UNIQUE\\s+)?INDEX\\s+(?:IF\\s+NOT\\s+EXISTS\\s+)?\\w+\\s+ON\\s+\\w+",
        Pattern.CASE_INSENSITIVE
                                                                             );

    // Patrón para extraer el nombre del índice
    private static final Pattern INDEX_NAME_PATTERN = Pattern.compile(
        "CREATE\\s+(?:UNIQUE\\s+)?INDEX\\s+(?:IF\\s+NOT\\s+EXISTS\\s+)?(\\w+)",
        Pattern.CASE_INSENSITIVE
                                                                     );

    // Patrón para extraer el nombre de la tabla
    private static final Pattern TABLE_NAME_PATTERN = Pattern.compile(
        "ON\\s+(\\w+)\\s*\\(",
        Pattern.CASE_INSENSITIVE
                                                                     );

    // Patrón para extraer las columnas del índice (maneja paréntesis anidados)
    private static final Pattern COLUMNS_PATTERN = Pattern.compile(
        "\\(([^)]*(?:\\([^)]*\\)[^)]*)*)\\)(?:\\s*WHERE|\\s*;|\\s*$)",
        Pattern.CASE_INSENSITIVE
                                                                  );

    @Override
    public List<String> extractCreateIndexStatements(String sql) {
        List<String> statements = new ArrayList<>();

        if (sql == null || sql.trim().isEmpty()) {
            return statements;
        }

        // Buscar todas las coincidencias del patrón de inicio
        Matcher matcher = CREATE_INDEX_START_PATTERN.matcher(sql);

        while (matcher.find()) {
            int startPos = matcher.start();
            String statement = extractCompleteStatement(sql, startPos);

            if (statement != null && !statement.trim().isEmpty()) {
                // Asegurar que termine con punto y coma
                if (!statement.endsWith(";")) {
                    statement += ";";
                }
                statements.add(statement);
            }
        }

        return statements;
    }

    /**
     * Extrae una sentencia CREATE INDEX completa desde una posición de inicio,
     * manejando correctamente paréntesis anidados y cláusulas WHERE.
     */
    private String extractCompleteStatement(String sql, int startPos) {
        int pos = startPos;
        int parenthesesCount = 0;
        boolean foundMainParentheses = false;
        boolean inWhereClause = false;
        StringBuilder statement = new StringBuilder();

        while (pos < sql.length()) {
            char currentChar = sql.charAt(pos);
            statement.append(currentChar);

            // Detectar inicio de paréntesis principales (después de ON tabla)
            if (currentChar == '(' && !foundMainParentheses) {
                // Verificar si estamos después de "ON tablename" o "USING method"
                String beforeParen = statement.toString().toUpperCase();
                if (beforeParen.matches(".*ON\\s+\\w+\\s*$") ||
                    beforeParen.matches(".*ON\\s+\\w+\\s+USING\\s+\\w+\\s*$")) {
                    foundMainParentheses = true;
                    parenthesesCount = 1;
                } else if (foundMainParentheses) {
                    parenthesesCount++;
                }
            } else if (currentChar == '(' && foundMainParentheses) {
                parenthesesCount++;
            } else if (currentChar == ')' && foundMainParentheses) {
                parenthesesCount--;

                // Si cerramos los paréntesis principales
                if (parenthesesCount == 0) {
                    // Buscar si hay cláusula WHERE o terminar
                    int nextPos = pos + 1;
                    while (nextPos < sql.length() && Character.isWhitespace(sql.charAt(nextPos))) {
                        statement.append(sql.charAt(nextPos));
                        nextPos++;
                    }

                    // Verificar si viene WHERE
                    if (nextPos < sql.length() - 4) {
                        String nextWord = sql.substring(nextPos, Math.min(nextPos + 5, sql.length())).toUpperCase();
                        if (nextWord.startsWith("WHERE")) {
                            inWhereClause = true;
                            pos = nextPos - 1; // Retroceder para incluir WHERE en siguiente iteración
                        } else {
                            // No hay WHERE, terminamos aquí
                            return statement.toString().trim();
                        }
                    } else {
                        return statement.toString().trim();
                    }
                }
            } else if (currentChar == ';') {
                // Encontramos el final de la sentencia
                return statement.toString().trim();
            } else if (inWhereClause && currentChar == '\n') {
                // En WHERE, verificar si la siguiente línea empieza con CREATE, ALTER, etc.
                int nextLineStart = pos + 1;
                while (nextLineStart < sql.length() && Character.isWhitespace(sql.charAt(nextLineStart))) {
                    if (sql.charAt(nextLineStart) == '\n') break;
                    nextLineStart++;
                }

                if (nextLineStart < sql.length()) {
                    String nextLine = sql.substring(nextLineStart, Math.min(nextLineStart + 10, sql.length())).trim().toUpperCase();
                    if (nextLine.startsWith("CREATE") || nextLine.startsWith("ALTER") ||
                        nextLine.startsWith("--") || nextLine.startsWith("/*")) {
                        // Nueva sentencia o comentario, terminamos aquí
                        return statement.toString().trim();
                    }
                }
            }

            pos++;
        }

        return statement.toString().trim();
    }

    @Override
    public String extractTableName(String indexStatement) {
        if (indexStatement == null || indexStatement.trim().isEmpty()) {
            return null;
        }

        Matcher matcher = TABLE_NAME_PATTERN.matcher(indexStatement);
        if (matcher.find()) {
            return matcher.group(1).trim();
        }

        return null;
    }

    @Override
    public List<String> extractTargetColumnNames(String indexStatement) {
        List<String> columnNames = new ArrayList<>();

        if (indexStatement == null || indexStatement.trim().isEmpty()) {
            return columnNames;
        }

        Matcher matcher = COLUMNS_PATTERN.matcher(indexStatement);
        if (matcher.find()) {
            String columnsText = matcher.group(1).trim();
            String[] columns = columnsText.split(",");

            Pattern functionPattern = Pattern.compile("\\w+\\s*\\(\\s*(\\w+)\\s*\\)");

            for (String col : columns) {
                String clean = col.trim();

                // Eliminar ASC/DESC al final (por si existe)
                clean = clean.replaceAll("(?i)\\s+(ASC|DESC)\\s*$", "");

                // Detectar si es una función tipo LOWER(email)
                Matcher funcMatcher = functionPattern.matcher(clean);
                if (funcMatcher.find()) {
                    columnNames.add(funcMatcher.group(1));
                } else {
                    columnNames.add(clean);
                }
            }
        }

        return columnNames;
    }

    @Override
    public String extractIndexName(String indexStatement) {
        if (indexStatement == null || indexStatement.trim().isEmpty()) {
            return null;
        }

        Matcher matcher = INDEX_NAME_PATTERN.matcher(indexStatement);
        if (matcher.find()) {
            return matcher.group(1).trim();
        }

        return null;
    }

}
