package services

import (
    "database/sql"
    "fmt"
    "chat2sr/config"
    _ "github.com/go-sql-driver/mysql"
)

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
        var field, fieldType, null, key, extra string
        var defaultValue sql.NullString
        err := rows.Scan(&field, &fieldType, &null, &key, &defaultValue, &extra)
        if err != nil {
            return nil, err
        }
        columns = append(columns, field)
    }

    return columns, nil
}

func GetAllTables() ([]string, error) {
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

    rows, err := db.Query("SHOW TABLES")
    if err != nil {
        return nil, fmt.Errorf("failed to get tables: %v", err)
    }
    defer rows.Close()

    var tables []string
    for rows.Next() {
        var table string
        if err := rows.Scan(&table); err != nil {
            return nil, err
        }
        tables = append(tables, table)
    }

    return tables, nil
}

func ExecuteSQL(sqlQuery string) ([]map[string]interface{}, error) {
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

    rows, err := db.Query(sqlQuery)
    if err != nil {
        return nil, fmt.Errorf("failed to execute query: %v", err)
    }
    defer rows.Close()

    columns, err := rows.Columns()
    if err != nil {
        return nil, fmt.Errorf("failed to get columns: %v", err)
    }

    var results []map[string]interface{}
    
    count := len(columns)
    values := make([]interface{}, count)
    valuePtrs := make([]interface{}, count)
    
    for rows.Next() {
        for i := range columns {
            valuePtrs[i] = &values[i]
        }
        
        if err := rows.Scan(valuePtrs...); err != nil {
            return nil, fmt.Errorf("failed to scan row: %v", err)
        }
        
        row := make(map[string]interface{})
        for i, col := range columns {
            val := values[i]
            
            b, ok := val.([]byte)
            if ok {
                row[col] = string(b)
            } else {
                row[col] = val
            }
        }
        
        results = append(results, row)
    }
    
    return results, nil
}