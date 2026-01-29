package main

import (
	"fmt"
	"strings"
	"testing"
)

func TestIsValidSchema(t *testing.T) {
	tests := []struct {
		name   string
		schema string
		want   bool
	}{
		{
			name: "Tabla simple válida",
			schema: `CREATE TABLE users (
				id SERIAL PRIMARY KEY,
				name VARCHAR(100) NOT NULL,
				email TEXT UNIQUE
			);`,
			want: true,
		},
		{
			name: "Tabla con constraints",
			schema: `CREATE TABLE orders (
				id BIGSERIAL,
				user_id INTEGER NOT NULL,
				amount DECIMAL(10,2),
				created_at TIMESTAMPTZ DEFAULT NOW(),
				PRIMARY KEY (id),
				FOREIGN KEY (user_id) REFERENCES users(id)
			);`,
			want: true,
		},
		{
			name: "Múltiples tablas",
			schema: `
				CREATE TABLE categories (id SERIAL PRIMARY KEY, name TEXT);
				CREATE TABLE products (id SERIAL, category_id INT, price NUMERIC(10,2));
			`,
			want: true,
		},
		{
			name:   "Tipo de dato inválido",
			schema: `CREATE TABLE bad (id INVALID_TYPE);`,
			want:   false,
		},
		{
			name:   "Sin columnas",
			schema: `CREATE TABLE empty ();`,
			want:   false,
		},
		{
			name:   "Sin CREATE TABLE",
			schema: `SELECT * FROM users;`,
			want:   false,
		},
		{
			name: "Columna duplicada",
			schema: `CREATE TABLE dup (
				id SERIAL,
				name TEXT,
				name VARCHAR(100)
			);`,
			want: false,
		},
		{
			name:   "IF NOT EXISTS",
			schema: `CREATE TABLE IF NOT EXISTS users (id SERIAL PRIMARY KEY);`,
			want:   true,
		},
		{
			name:   "Con schema",
			schema: `CREATE TABLE public.users (id SERIAL PRIMARY KEY, name TEXT);`,
			want:   true,
		},
		{
			name: "Arrays",
			schema: `CREATE TABLE with_arrays (
				id SERIAL,
				tags TEXT[],
				numbers INTEGER[]
			);`,
			want: true,
		},
		{
			name: "JSON types",
			schema: `CREATE TABLE with_json (
				id SERIAL,
				data JSON,
				metadata JSONB
			);`,
			want: true,
		},
		{
			name: "Todos los tipos numéricos",
			schema: `CREATE TABLE numerics (
				a SMALLINT,
				b INTEGER,
				c BIGINT,
				d DECIMAL(10,2),
				e NUMERIC(15,4),
				f REAL,
				g DOUBLE PRECISION,
				h SERIAL,
				i BIGSERIAL,
				j MONEY
			);`,
			want: true,
		},
		{
			name: "Tipos fecha/hora",
			schema: `CREATE TABLE dates (
				a TIMESTAMP,
				b TIMESTAMPTZ,
				c DATE,
				d TIME,
				e INTERVAL
			);`,
			want: true,
		},
		{
			name:   "Nombre de tabla inválido (número al inicio)",
			schema: `CREATE TABLE 123table (id SERIAL);`,
			want:   false,
		},
		{
			name:   "Constraint CHECK",
			schema: `CREATE TABLE products (id SERIAL, price NUMERIC CHECK (price > 0));`,
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidSchema(tt.schema)
			if got != tt.want {
				t.Errorf("isValidSchema() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtractCreateTableStatements(t *testing.T) {
	tests := []struct {
		name      string
		schema    string
		wantCount int
	}{
		{
			name:      "Una tabla",
			schema:    `CREATE TABLE users (id SERIAL);`,
			wantCount: 1,
		},
		{
			name: "Tres tablas",
			schema: `
				CREATE TABLE a (id INT);
				CREATE TABLE b (id INT);
				CREATE TABLE c (id INT);
			`,
			wantCount: 3,
		},
		{
			name:      "Sin tablas",
			schema:    `SELECT * FROM users;`,
			wantCount: 0,
		},
		{
			name: "Mixto con otros statements",
			schema: `
				DROP TABLE IF EXISTS old;
				CREATE TABLE new (id SERIAL);
				INSERT INTO new VALUES (1);
				CREATE TABLE another (name TEXT);
			`,
			wantCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			statements := extractCreateTableStatements(tt.schema)
			fmt.Println("Table: ", len(statements))
			if len(statements) != tt.wantCount {
				t.Errorf("extractCreateTableStatements() count = %d, want %d", len(statements), tt.wantCount)
				t.Logf("Statements encontrados: %v", statements)
			}
		})
	}
}

func TestExtractTableName(t *testing.T) {
	tests := []struct {
		stmt string
		want string
	}{
		{`CREATE TABLE users (id INT);`, "users"},
		{`CREATE TABLE public.users (id INT);`, "users"},
		{`CREATE TABLE IF NOT EXISTS users (id INT);`, "users"},
		{`CREATE TABLE "MyTable" (id INT);`, "MyTable"},
		{`create table lowercase (id int);`, "lowercase"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := extractTableName(tt.stmt)
			if got != tt.want {
				t.Errorf("extractTableName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestExtractTableBody(t *testing.T) {
	tests := []struct {
		name string
		stmt string
		want string
	}{
		{
			name: "Simple",
			stmt: `CREATE TABLE t (id INT);`,
			want: "id INT",
		},
		{
			name: "Múltiples columnas",
			stmt: `CREATE TABLE t (id INT, name TEXT);`,
			want: "id INT, name TEXT",
		},
		{
			name: "Con constraint",
			stmt: `CREATE TABLE t (id INT, PRIMARY KEY (id));`,
			want: "id INT, PRIMARY KEY (id)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractTableBody(tt.stmt)
			if strings.TrimSpace(got) != tt.want {
				t.Errorf("extractTableBody() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSplitTableElements(t *testing.T) {
	tests := []struct {
		name      string
		body      string
		wantCount int
	}{
		{"Una columna", "id INT", 1},
		{"Dos columnas", "id INT, name TEXT", 2},
		{"Con constraint anidado", "id INT, CHECK (id > 0)", 2},
		{"FK compleja", "id INT, FOREIGN KEY (id) REFERENCES other(id)", 2},
		{"Múltiples", "a INT, b TEXT, c BOOL, PRIMARY KEY (a)", 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			elements := splitTableElements(tt.body)
			if len(elements) != tt.wantCount {
				t.Errorf("splitTableElements() count = %d, want %d", len(elements), tt.wantCount)
				t.Logf("Elements: %v", elements)
			}
		})
	}
}

func TestIsValidIdentifier(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"users", true},
		{"_private", true},
		{"table123", true},
		{"my_table", true},
		{"123invalid", false},
		{"-invalid", false},
		{"", false},
		{strings.Repeat("a", 64), false}, // > 63 chars
		{strings.Repeat("a", 63), true},  // exactamente 63
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidIdentifier(tt.name)
			if got != tt.want {
				t.Errorf("isValidIdentifier(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestIsValidDataType(t *testing.T) {
	validTypes := []string{
		"INT", "INTEGER", "BIGINT", "SMALLINT",
		"VARCHAR(100)", "VARCHAR(255)", "CHAR(10)",
		"TEXT", "BOOLEAN", "BOOL",
		"DECIMAL(10,2)", "NUMERIC(15,4)",
		"TIMESTAMP", "TIMESTAMPTZ", "DATE", "TIME",
		"JSON", "JSONB", "UUID",
		"SERIAL", "BIGSERIAL",
		"INTEGER[]", "TEXT[]",
		"double precision",
	}

	invalidTypes := []string{
		"INVALID", "NOTREAL", "MYTYPE",
		"", "123",
	}

	for _, dt := range validTypes {
		t.Run("valid_"+dt, func(t *testing.T) {
			if !isValidDataType(dt) {
				t.Errorf("isValidDataType(%q) = false, want true", dt)
			}
		})
	}

	for _, dt := range invalidTypes {
		t.Run("invalid_"+dt, func(t *testing.T) {
			if isValidDataType(dt) {
				t.Errorf("isValidDataType(%q) = true, want false", dt)
			}
		})
	}
}

func TestIsValidColumnDefinition(t *testing.T) {
	tests := []struct {
		elem    string
		wantCol string
		wantOk  bool
	}{
		{"id SERIAL", "id", true},
		{"name VARCHAR(100) NOT NULL", "name", true},
		{"price DECIMAL(10,2) DEFAULT 0", "price", true},
		{"created_at TIMESTAMPTZ", "created_at", true},
		{"data JSONB", "data", true},
		{"id", "", false},              // sin tipo
		{"id INVALID_TYPE", "", false}, // tipo inválido
		{"123col INT", "", false},      // nombre inválido
	}

	for _, tt := range tests {
		t.Run(tt.elem, func(t *testing.T) {
			col, ok := isValidColumnDefinition(tt.elem)
			if ok != tt.wantOk {
				t.Errorf("isValidColumnDefinition(%q) ok = %v, want %v", tt.elem, ok, tt.wantOk)
			}
			if col != tt.wantCol {
				t.Errorf("isValidColumnDefinition(%q) col = %q, want %q", tt.elem, col, tt.wantCol)
			}
		})
	}
}

func TestIsTableLevelConstraint(t *testing.T) {
	constraints := []string{
		"PRIMARY KEY (id)",
		"FOREIGN KEY (user_id) REFERENCES users(id)",
		"UNIQUE (email)",
		"CHECK (price > 0)",
		"CONSTRAINT pk_users PRIMARY KEY (id)",
		"CONSTRAINT fk_orders FOREIGN KEY (user_id) REFERENCES users(id)",
	}

	notConstraints := []string{
		"id SERIAL",
		"name VARCHAR(100)",
		"created_at TIMESTAMP",
	}

	for _, c := range constraints {
		t.Run("is_constraint", func(t *testing.T) {
			if !isTableLevelConstraint(c) {
				t.Errorf("isTableLevelConstraint(%q) = false, want true", c)
			}
		})
	}

	for _, c := range notConstraints {
		t.Run("not_constraint", func(t *testing.T) {
			if isTableLevelConstraint(c) {
				t.Errorf("isTableLevelConstraint(%q) = true, want false", c)
			}
		})
	}
}

func TestIsValidTableConstraint(t *testing.T) {
	valid := []string{
		"PRIMARY KEY (id)",
		"FOREIGN KEY (user_id) REFERENCES users(id)",
		"UNIQUE (email)",
		"CHECK (price > 0)",
		"CONSTRAINT pk PRIMARY KEY (id)",
		"CONSTRAINT chk CHECK (x > 0)",
	}

	invalid := []string{
		"CONSTRAINT",    // incompleto
		"CONSTRAINT pk", // sin tipo
		"INVALID (id)",
	}

	for _, c := range valid {
		t.Run("valid", func(t *testing.T) {
			if !isValidTableConstraint(c) {
				t.Errorf("isValidTableConstraint(%q) = false, want true", c)
			}
		})
	}

	for _, c := range invalid {
		t.Run("invalid", func(t *testing.T) {
			if isValidTableConstraint(c) {
				t.Errorf("isValidTableConstraint(%q) = true, want false", c)
			}
		})
	}
}

// ============================================================================
// TEST DETALLADO CON OUTPUT VERBOSE
// ============================================================================

func TestValidatorVerbose(t *testing.T) {
	schema := `
		CREATE TABLE users (
			id SERIAL PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			email TEXT UNIQUE,
			created_at TIMESTAMPTZ DEFAULT NOW()
		);

		CREATE TABLE orders (
			id BIGSERIAL,
			user_id INTEGER NOT NULL,
			total DECIMAL(10,2),
			FOREIGN KEY (user_id) REFERENCES users(id)
		);

		CREATE TABLE products (
			id SERIAL,
			name TEXT NOT NULL,
			price NUMERIC(10,2) CHECK (price >= 0),
			tags TEXT[]
		);
	`

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("ANÁLISIS DETALLADO DEL SCHEMA")
	fmt.Println(strings.Repeat("=", 60))

	// 1. Verificar si contiene CREATE TABLE
	hasCreateTable := containsCreateTableStatement(schema)
	fmt.Printf("\n✓ Contiene CREATE TABLE: %v\n", hasCreateTable)

	// 2. Extraer sentencias
	statements := extractCreateTableStatements(schema)
	fmt.Printf("✓ Sentencias encontradas: %d\n", len(statements))

	// 3. Analizar cada sentencia
	for i, stmt := range statements {
		fmt.Printf("\n%s\n", strings.Repeat("-", 50))
		fmt.Printf("TABLA #%d\n", i+1)
		fmt.Printf("%s\n", strings.Repeat("-", 50))

		// Nombre
		tableName := extractTableName(stmt)
		fmt.Printf("  Nombre: %s\n", tableName)
		fmt.Printf("  Identificador válido: %v\n", isValidIdentifier(tableName))

		// Cuerpo
		body := extractTableBody(stmt)
		elements := splitTableElements(body)
		fmt.Printf("  Elementos totales: %d\n", len(elements))

		// Columnas y constraints
		var columns []string
		var constraints []string

		for _, elem := range elements {
			elem = strings.TrimSpace(elem)
			if elem == "" {
				continue
			}

			if isTableLevelConstraint(elem) {
				constraints = append(constraints, elem)
			} else {
				columns = append(columns, elem)
			}
		}

		fmt.Printf("  Columnas: %d\n", len(columns))
		for _, col := range columns {
			tokens := tokenize(col)
			if len(tokens) >= 2 {
				colName := strings.Trim(tokens[0], `"`)
				dataType := extractDataType(tokens[1:])
				validType := isValidDataType(dataType)
				status := "✓"
				if !validType {
					status = "✗"
				}
				fmt.Printf("    %s %s (%s) - tipo válido: %v\n", status, colName, dataType, validType)
			}
		}

		fmt.Printf("  Constraints: %d\n", len(constraints))
		for _, c := range constraints {
			valid := isValidTableConstraint(c)
			status := "✓"
			if !valid {
				status = "✗"
			}
			// Truncar si es muy largo
			display := c
			if len(display) > 50 {
				display = display[:47] + "..."
			}
			fmt.Printf("    %s %s\n", status, display)
		}

		// Validación completa
		valid := isValidCreateTableStatement(stmt)
		fmt.Printf("  Sentencia válida: %v\n", valid)
	}

	// Resultado final
	fmt.Printf("\n%s\n", strings.Repeat("=", 60))
	result := isValidSchema(schema)
	if result {
		fmt.Println("✅ SCHEMA VÁLIDO")
	} else {
		fmt.Println("❌ SCHEMA INVÁLIDO")
	}
	fmt.Printf("%s\n\n", strings.Repeat("=", 60))

	if !result {
		t.Error("Expected valid schema")
	}
}

// Test con schema inválido para ver errores
func TestValidatorVerboseInvalid(t *testing.T) {
	schema := `
		CREATE TABLE bad_table (
			id INVALID_TYPE,
			name VARCHAR(100),
			name TEXT,
			price MONEY
		);
	`

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("ANÁLISIS DE SCHEMA INVÁLIDO")
	fmt.Println(strings.Repeat("=", 60))

	statements := extractCreateTableStatements(schema)
	fmt.Printf("\n✓ Sentencias encontradas: %d\n", len(statements))

	for i, stmt := range statements {
		fmt.Printf("\nTABLA #%d: %s\n", i+1, extractTableName(stmt))

		body := extractTableBody(stmt)
		elements := splitTableElements(body)

		columnNames := make(map[string]bool)

		for _, elem := range elements {
			elem = strings.TrimSpace(elem)
			if elem == "" || isTableLevelConstraint(elem) {
				continue
			}

			tokens := tokenize(elem)
			if len(tokens) >= 2 {
				colName := strings.Trim(tokens[0], `"`)
				dataType := extractDataType(tokens[1:])

				issues := []string{}

				if !isValidIdentifier(colName) {
					issues = append(issues, "nombre inválido")
				}
				if !isValidDataType(dataType) {
					issues = append(issues, "tipo inválido")
				}
				if columnNames[strings.ToLower(colName)] {
					issues = append(issues, "DUPLICADA")
				}
				columnNames[strings.ToLower(colName)] = true

				status := "✓"
				if len(issues) > 0 {
					status = "✗"
				}
				fmt.Printf("  %s %s %s", status, colName, dataType)
				if len(issues) > 0 {
					fmt.Printf(" [%s]", strings.Join(issues, ", "))
				}
				fmt.Println()
			}
		}
	}

	result := isValidSchema(schema)
	fmt.Printf("\n%s\n", strings.Repeat("=", 60))
	if result {
		fmt.Println("✅ SCHEMA VÁLIDO")
	} else {
		fmt.Println("❌ SCHEMA INVÁLIDO")
	}
	fmt.Printf("%s\n\n", strings.Repeat("=", 60))

	if result {
		t.Error("Expected invalid schema")
	}
}
