// Package database provides database connection management for the application.
package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DatabaseConfig contains the configuration for the database connection.
type DatabaseConfig struct {
	Host            string        `yaml:"host"`
	Port            int           `yaml:"port"`
	User            string        `yaml:"user"`
	Password        string        `yaml:"password"`
	DBName          string        `yaml:"dbname"`
	SSLMode         string        `yaml:"sslmode"`
	MaxOpenConns    int           `yaml:"max_open_conns"`
	MaxIdleConns    int           `yaml:"max_idle_conns"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`
}

// DefaultDatabaseConfig returns the default database configuration.
func DefaultDatabaseConfig() *DatabaseConfig {
	return &DatabaseConfig{
		Host:            "localhost",
		Port:            5432,
		User:            "postgres",
		Password:        "",
		DBName:          "jiufang",
		SSLMode:         "disable",
		MaxOpenConns:    100,
		MaxIdleConns:    10,
		ConnMaxLifetime: 30 * time.Minute,
	}
}

// DSN returns the data source name for PostgreSQL.
func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s client_encoding=utf8",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode)
}

// Connect creates a new database connection.
func Connect(config *DatabaseConfig) (*gorm.DB, error) {
	if config == nil {
		config = DefaultDatabaseConfig()
	}

	// Create GORM config
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	// Open database connection using pgx driver
	// The pgx driver has better UTF-8 support than pq
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  config.DSN(),
		PreferSimpleProtocol: true, // Disables implicit prepared statement usage
	}), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get underlying SQL DB
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(config.ConnMaxLifetime)

	// Set client encoding for all connections using a callback
	// This ensures every new connection from the pool has UTF-8 encoding
	if err := setClientEncodingForAllConnections(sqlDB); err != nil {
		return nil, fmt.Errorf("failed to set client encoding for connections: %w", err)
	}

	return db, nil
}

// setClientEncodingForAllConnections sets UTF-8 encoding for all database connections
func setClientEncodingForAllConnections(sqlDB *sql.DB) error {
	// Test a connection to ensure encoding is set
	conn, err := sqlDB.Conn(context.Background())
	if err != nil {
		return err
	}
	defer conn.Close()

	// Execute SET command on this connection
	_, err = conn.ExecContext(context.Background(), "SET client_encoding = 'UTF8'")
	if err != nil {
		return err
	}

	return nil
}

// Close closes the database connection.
func Close(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}

	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("failed to close database connection: %w", err)
	}

	return nil
}
