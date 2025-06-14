package database

import (
	"fmt"
	"log"

	"owlistic-notes/owlistic/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Database struct {
	DB *gorm.DB
}

func Setup(cfg config.Config) (*Database, error) {

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
	)

	// Configure GORM with performance settings for large datasets
	gormConfig := &gorm.Config{
		Logger:                 logger.Default.LogMode(logger.Warn),
		PrepareStmt:            true,  // Cache prepared statements for better performance
		AllowGlobalUpdate:      false, // Prevent global updates without conditions
		SkipDefaultTransaction: true,  // Skip default transaction for better performance
	}

	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Run migrations to properly set up tables and constraints
	log.Println("Running database migrations...")
	if err := RunMigrations(db); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}
	log.Println("Database migrations completed successfully")

	// Create single user from environment variables
	log.Println("Setting up single user from environment variables...")
	if err := SetupSingleUser(db, cfg); err != nil {
		return nil, fmt.Errorf("failed to setup single user: %w", err)
	}
	log.Println("Single user setup completed successfully")

	return &Database{DB: db}, nil
}

func (d *Database) Close() {
	if d.DB == nil {
		log.Println("Database connection is nil, nothing to close.")
		return
	}
	sqlDB, err := d.DB.DB()
	if err != nil {
		log.Printf("Failed to get database connection: %v", err)
		return
	}
	if err := sqlDB.Close(); err != nil {
		log.Printf("Failed to close database connection: %v", err)
	}
}

func (d *Database) Query(query string, args ...interface{}) (*gorm.DB, error) {
	result := d.DB.Raw(query, args...)
	return result, result.Error
}

func (d *Database) Execute(query string, args ...interface{}) error {
	result := d.DB.Exec(query, args...)
	return result.Error
}
