package services

import (
    "database/sql"
    "fmt"
    "chat2sr/config"
    _ "github.com/go-sql-driver/mysql"
)

// TableInfo 结构体用于存储表信息
type TableInfo struct {
    Name    string
    Comment string
}

// GetAllTablesWithComments 获取所有表及其注释
func GetAllTablesWithComments() ([]TableInfo, error) {
    dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
        config.AppConfig.DBUser,
        config.AppConfig.DBPassword,
        config.AppConfig.DBHost,
        config.AppConfig.DBPort,
        config.AppConfig.DBName)

    db, err := sql.Open("mysql", dsn)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to database: %v", err)
    }
    defer db.Close()

    query := `
        SELECT 
            TABLE_NAME, 
            TABLE_COMMENT 
        FROM 
            INFORMATION_SCHEMA.TABLES 
        WHERE 
            TABLE_SCHEMA = ?
    `
    
    rows, err := db.Query(query, config.AppConfig.DBName)
    if err != nil {
        return nil, fmt.Errorf("failed to get tables: %v", err)
    }
    defer rows.Close()

    var tables []TableInfo
    for rows.Next() {
        var name, comment string
        if err := rows.Scan(&name, &comment); err != nil {
            return nil, err
        }
        tables = append(tables, TableInfo{Name: name, Comment: comment})
    }

    return tables, nil
}

// GetTableSchema 获取表结构
func GetTableSchema(tableName string) ([]string, error) {
    dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
        config.AppConfig.DBUser,
        config.AppConfig.DBPassword,
        config.AppConfig.DBHost,
        config.AppConfig.DBPort,
        config.AppConfig.DBName)

    db, err := sql.Open("mysql", dsn)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to database: %v", err)
    }
    defer db.Close()

    query := fmt.Sprintf("SHOW COLUMNS FROM %s", tableName)
    rows, err := db.Query(query)
    if err != nil {
        return nil, fmt.Errorf("failed to get schema for table %s: %v", tableName, err)
    }
    defer rows.Close()

    var columns []string
    for rows.Next() {
        var field, fieldType, null, key, defaultValue, extra sql.NullString
        if err := rows.Scan(&field, &fieldType, &null, &key, &defaultValue, &extra); err != nil {
            return nil, err
        }
        columnDesc := fmt.Sprintf("%s %s", field.String, fieldType.String)
        columns = append(columns, columnDesc)
    }

    return columns, nil
}

// GetTableSchemaWithComments 获取表结构包括字段注释
func GetTableSchemaWithComments(tableName string) ([]map[string]string, error) {
    dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
        config.AppConfig.DBUser,
        config.AppConfig.DBPassword,
        config.AppConfig.DBHost,
        config.AppConfig.DBPort,
        config.AppConfig.DBName)

    db, err := sql.Open("mysql", dsn)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to database: %v", err)
    }
    defer db.Close()

    query := `
        SELECT 
            COLUMN_NAME,
            COLUMN_TYPE,
            COLUMN_COMMENT
        FROM 
            INFORMATION_SCHEMA.COLUMNS 
        WHERE 
            TABLE_SCHEMA = ? AND TABLE_NAME = ?
    `
    
    rows, err := db.Query(query, config.AppConfig.DBName, tableName)
    if err != nil {
        return nil, fmt.Errorf("failed to get schema for table %s: %v", tableName, err)
    }
    defer rows.Close()

    var columns []map[string]string
    for rows.Next() {
        var name, dataType, comment string
        if err := rows.Scan(&name, &dataType, &comment); err != nil {
            return nil, err
        }
        
        column := map[string]string{
            "name": name,
            "type": dataType,
            "comment": comment,
        }
        
        columns = append(columns, column)
    }

    return columns, nil
}

// GetAllTables 获取所有表
func GetAllTables() ([]string, error) {
    tables, err := GetAllTablesWithComments()
    if err != nil {
        return nil, err
    }
    
    var tableNames []string
    for _, table := range tables {
        tableNames = append(tableNames, table.Name)
    }
    
    return tableNames, nil
}

// ExecuteSQL 执行SQL查询
func ExecuteSQL(query string) ([]map[string]interface{}, error) {
    dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
        config.AppConfig.DBUser,
        config.AppConfig.DBPassword,
        config.AppConfig.DBHost,
        config.AppConfig.DBPort,
        config.AppConfig.DBName)

    db, err := sql.Open("mysql", dsn)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to database: %v", err)
    }
    defer db.Close()

    rows, err := db.Query(query)
    if err != nil {
        return nil, fmt.Errorf("failed to execute query: %v", err)
    }
    defer rows.Close()

    columns, err := rows.Columns()
    if err != nil {
        return nil, fmt.Errorf("failed to get column names: %v", err)
    }

    var results []map[string]interface{}
    for rows.Next() {
        values := make([]interface{}, len(columns))
        pointers := make([]interface{}, len(columns))
        for i := range values {
            pointers[i] = &values[i]
        }

        if err := rows.Scan(pointers...); err != nil {
            return nil, fmt.Errorf("failed to scan row: %v", err)
        }

        row := make(map[string]interface{})
        for i, column := range columns {
            val := values[i]
            if b, ok := val.([]byte); ok {
                row[column] = string(b)
            } else {
                row[column] = val
            }
        }
        results = append(results, row)
    }

    return results, nil
}