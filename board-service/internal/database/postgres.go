package database

import (
	"context"
	"fmt"
	"time"

	"board-service/internal/domain"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// Connect establishes a connection to PostgreSQL database
// If useAutoMigrate is true, it will automatically migrate all domain models
func Connect(databaseURL string, logger *zap.Logger, useAutoMigrate bool) (*gorm.DB, error) {
	// Configure GORM with custom logger
	gormConfig := &gorm.Config{
		Logger: newGormLogger(logger),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	}

	// Open database connection
	db, err := gorm.Open(postgres.Open(databaseURL), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get underlying SQL database
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Ping database to verify connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("Database connected successfully",
		zap.Int("max_idle_conns", 10),
		zap.Int("max_open_conns", 100),
		zap.Duration("conn_max_lifetime", time.Hour),
	)

	// Run AutoMigrate if enabled (development mode)
	if useAutoMigrate {
		if err := autoMigrateAll(db, logger); err != nil {
			return nil, fmt.Errorf("failed to auto-migrate: %w", err)
		}
		logger.Info("GORM AutoMigrate completed successfully")
	} else {
		logger.Info("AutoMigrate disabled - using manual migrations")
	}

	return db, nil
}

// autoMigrateAll runs GORM AutoMigrate for all domain models
func autoMigrateAll(db *gorm.DB, logger *zap.Logger) error {
	logger.Info("Running GORM AutoMigrate...")

	models := []interface{}{
		&domain.SchemaVersion{},
		&domain.Role{},
		// Workspace models removed - managed by User Service
		&domain.Project{},
		&domain.ProjectMember{},
		&domain.ProjectJoinRequest{},
		&domain.Board{},
		&domain.Comment{},
		// Custom fields system (new ProjectField system replaces CustomRole/CustomStage/CustomImportance)
		&domain.ProjectField{},
		&domain.FieldOption{},
		&domain.BoardFieldValue{},
		&domain.SavedView{},
		&domain.UserBoardOrder{}, // Fractional indexing for board ordering in views
	}

	return db.AutoMigrate(models...)
}

// gormLogger is a custom logger adapter for GORM
type gormLogger struct {
	zapLogger *zap.Logger
}

func newGormLogger(logger *zap.Logger) gormlogger.Interface {
	return &gormLogger{zapLogger: logger}
}

func (l *gormLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	return l
}

func (l *gormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	l.zapLogger.Sugar().Infof(msg, data...)
}

func (l *gormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	l.zapLogger.Sugar().Warnf(msg, data...)
}

func (l *gormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	l.zapLogger.Sugar().Errorf(msg, data...)
}

func (l *gormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()

	if err != nil {
		l.zapLogger.Error("Database query error",
			zap.Error(err),
			zap.Duration("elapsed", elapsed),
			zap.String("sql", sql),
			zap.Int64("rows", rows),
		)
	} else {
		l.zapLogger.Debug("Database query",
			zap.Duration("elapsed", elapsed),
			zap.String("sql", sql),
			zap.Int64("rows", rows),
		)
	}
}
