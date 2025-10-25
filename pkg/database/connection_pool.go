package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
)

func OptimizeConnectionPool(db *gorm.DB, maxOpenConns, maxIdleConns int, maxLifetime time.Duration) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get SQL DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(maxOpenConns)

	sqlDB.SetMaxIdleConns(maxIdleConns)

	sqlDB.SetConnMaxLifetime(maxLifetime)

	sqlDB.SetConnMaxIdleTime(time.Minute * 5)

	log.Printf("Database connection pool optimized: %d max open, %d max idle", maxOpenConns, maxIdleConns)
	return nil
}

func HealthCheck(db *gorm.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	return sqlDB.PingContext(ctx)
}

func GetConnectionStats(db *gorm.DB) (*sql.DBStats, error) {
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	stats := sqlDB.Stats()
	return &stats, nil
}
