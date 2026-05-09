package database

import (
	"context"
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Database 配置
type Database struct {
	Driver         string
	Host           string
	Port           int
	Name           string
	User           string
	Password       string
	Charset        string
	MaxOpenConns   int
	MaxIdleConns   int
	ConnMaxLifetime int
	ConnMaxIdleTime int
	LogLevel       string
}

// DefaultDatabaseConfig returns default database configuration
func DefaultDatabaseConfig() Database {
	return Database{
		MaxOpenConns:    50,
		MaxIdleConns:    10,
		ConnMaxLifetime: 3600,
		ConnMaxIdleTime: 600,
		LogLevel:        "warn",
	}
}

// DBPoolStats holds database connection pool statistics
type DBPoolStats struct {
	MaxOpenConnections int
	OpenConnections   int
	InUse             int
	Idle              int
	WaitCount         int64
	WaitDuration      time.Duration
	MaxIdleClosed     int64
	MaxLifetimeClosed int64
}

// NewMySQL 初始化 MySQL 连接
func NewMySQL(cfg Database) (*gorm.DB, error) {
	// Apply defaults if not set
	if cfg.MaxOpenConns == 0 {
		cfg.MaxOpenConns = 50
	}
	if cfg.MaxIdleConns == 0 {
		cfg.MaxIdleConns = 10
	}
	if cfg.ConnMaxLifetime == 0 {
		cfg.ConnMaxLifetime = 3600
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name, cfg.Charset)

	gormLogLevel := mapStrToLogLevel(cfg.LogLevel)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(gormLogLevel),
	})
	if err != nil {
		return nil, fmt.Errorf("connect mysql: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get underlying sql.DB: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)

	// Set max idle time for connection reuse
	if cfg.ConnMaxIdleTime > 0 {
		sqlDB.SetConnMaxIdleTime(time.Duration(cfg.ConnMaxIdleTime) * time.Second)
	}

	return db, nil
}

// GetPoolStats returns database connection pool statistics
func GetPoolStats(db *gorm.DB) (DBPoolStats, error) {
	sqlDB, err := db.DB()
	if err != nil {
		return DBPoolStats{}, err
	}

	stats := sqlDB.Stats()
	return DBPoolStats{
		MaxOpenConnections: stats.MaxOpenConnections,
		OpenConnections:    stats.OpenConnections,
		InUse:              stats.InUse,
		Idle:               stats.Idle,
		WaitCount:          stats.WaitCount,
		WaitDuration:       stats.WaitDuration,
		MaxIdleClosed:      stats.MaxIdleClosed,
		MaxLifetimeClosed:  stats.MaxLifetimeClosed,
	}, nil
}

// Ping checks database connectivity
func Ping(ctx context.Context, db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}

func mapStrToLogLevel(level string) logger.LogLevel {
	switch level {
	case "silent":
		return logger.Silent
	case "error":
		return logger.Error
	case "warn":
		return logger.Warn
	case "info":
		return logger.Info
	default:
		return logger.Warn
	}
}
