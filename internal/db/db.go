package db

import (
	"fmt"
	"giftredeem/internal/models"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// Initialize sets up the database connection and performs migrations
func Initialize() error {
	// Database connection parameters
	username := getEnv("DB_USERNAME", "root")
	password := getEnv("DB_PASSWORD", "")
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "3306")
	dbName := getEnv("DB_NAME", "giftredeem")
	charset := "utf8mb4"
	loc := "Local"

	// Build DSN (Data Source Name)
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=True&loc=%s",
		username, password, host, port, dbName, charset, loc)

	// Configure GORM
	config := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	// Connect to the database
	var err error
	DB, err = gorm.Open(mysql.Open(dsn), config)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get DB instance: %w", err)
	}

	// Set connection pool parameters
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Run migrations
	err = runMigrations()
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// runMigrations performs database schema migrations
func runMigrations() error {
	// AutoMigrate will create tables, missing foreign keys, constraints, columns and indexes
	err := DB.AutoMigrate(
		&models.User{},
		&models.OAuthAccount{},
		&models.OAuthProvider{},
		&models.Benefit{},
		&models.RedemptionCode{},
		&models.Claim{},
	)
	if err != nil {
		return err
	}

	// 执行自定义迁移
	// 修改可能为空的字段
	migrations := []struct {
		name string
		sql  string
	}{
		{
			name: "修改 redemption_codes.claimed_by 为可为空",
			sql:  "ALTER TABLE redemption_codes MODIFY COLUMN claimed_by BIGINT UNSIGNED NULL;",
		},
		{
			name: "修改 oauth_accounts.refresh_token 为可为空",
			sql:  "ALTER TABLE oauth_accounts MODIFY COLUMN refresh_token TEXT NULL;",
		},
		{
			name: "修改 oauth_providers.client_secret 为可为空",
			sql:  "ALTER TABLE oauth_providers MODIFY COLUMN client_secret TEXT NULL;",
		},
	}

	// 执行所有迁移
	for _, migration := range migrations {
		err = DB.Exec(migration.sql).Error
		if err != nil {
			fmt.Printf("注意: %s 时出现错误（可能字段已经是 NULL）: %v\n", migration.name, err)
		} else {
			fmt.Printf("成功: %s\n", migration.name)
		}
	}

	return nil
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
