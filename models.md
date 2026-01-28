package org.example.database.model;

public enum AlterType {
ADD_COLUMN,
DROP_COLUMN,
MODIFY_COLUMN,
ADD_CONSTRAINT,
RENAME_COLUMN,
OTHER
}

---

package org.example.database.model;

public class ColumnMetadata {
private String columnName;
private String columnType;
private boolean isNotNull;
private String defaultValue;
}

---

package org.example.database.model;

public class RelationMetadata {
private String sourceColumn;
private String targetTable;
private String targetColumn;
private boolean isManyToOne;

}

---

package org.example.database.model;

public class TableAlteration {
private String tableName;
private AlterType alterType;
private String fullStatement;
private String targetColumn;
}

---

package org.example.database.model;

import java.util.List;
import java.util.Objects;

public class TableConstraintData {

    private String tableName;
    private String constraintName;
    private List<String> targetColumnNames;

}

---

package org.example.database.model;

import java.util.List;

public class TableIndexData {

    private String tableName;
    private String indexName;
    private List<String> targetColumnName;

}

---

package org.example.database.model;

import java.util.ArrayList;
import java.util.List;

public class TableMetadata {
private String tableName;
private List<TableIndexData> indexes = new ArrayList<>();
private List<ColumnMetadata> columns = new ArrayList<>();
private List<String> primaryKeys = new ArrayList<>();
private List<RelationMetadata> relations = new ArrayList<>();
private List<TableConstraintData> uniqueConstraints = new ArrayList<>();
}
