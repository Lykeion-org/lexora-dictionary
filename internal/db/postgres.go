package db

import (
    "fmt"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

func ConnectAndMigrate(dsn string) (*gorm.DB, error) {
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        return nil, err
    }

    err = db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`).Error
    if err != nil {
        return nil, fmt.Errorf("enable uuid extension: %w", err)
    }

    err = db.AutoMigrate(&Referent{}, &Symbol{}, &Word{}, &WordAttributes{})
    if err != nil {
        return nil, fmt.Errorf("auto-migrate failed: %w", err)
    }

    return db, nil
}

func CreateConnectionString(host string, port int, dbName string, dbUser string, dbPassword string) string {
    return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable",
		host,
		dbUser,
		dbPassword,
		dbName,
		port,
	)
}