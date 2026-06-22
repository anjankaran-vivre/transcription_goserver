package database

import (
	"fmt"
	"log"
	"strings"

	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"transcription-goserver/internal/config"
)

var DB *gorm.DB

func InitDB() {
	cfg := config.Settings
	dsn := fmt.Sprintf("server=%s;port=%d;user id=%s;password=%s;database=%s;trustservercertificate=true",
		strings.TrimRight(cfg.DBHost, ","), cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)

	logLevel := logger.Silent
	if cfg.Debug {
		logLevel = logger.Info
	}

	var err error
	DB, err = gorm.Open(sqlserver.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatalf("Failed to get underlying DB: %v", err)
	}
	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetMaxIdleConns(10)

	log.Println("✅ Database connection initialized")
}

func AutoMigrate(models ...interface{}) {
	// Drop indexes on call_id so GORM can ALTER COLUMN (SQL Server limitation)
	DB.Exec(`DECLARE @sql NVARCHAR(MAX) = '';
SELECT @sql = @sql + 'DROP INDEX ' + QUOTENAME(i.name) + ' ON call_logs;'
FROM sys.indexes i
INNER JOIN sys.index_columns ic ON i.object_id = ic.object_id AND i.index_id = ic.index_id
INNER JOIN sys.columns c ON ic.object_id = c.object_id AND ic.column_id = c.column_id
WHERE OBJECT_NAME(i.object_id) = 'call_logs'
  AND c.name = 'call_id'
  AND i.is_primary_key = 0
  AND i.name IS NOT NULL;
EXEC sp_executesql @sql;`)

	
	log.Println("✅ Database tables initialized")
}
