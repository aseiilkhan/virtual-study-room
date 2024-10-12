package config

import (
	"fmt"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func DB_init() {
	var err error
	DB, err = ConnectDatabase()
	if err != nil {
		fmt.Printf("Error connecting to the database: %v\n", err)
		os.Exit(1) // Terminate the application if the database connection fails
	}
}
func ConnectDatabase() (*gorm.DB, error) {
	databaseUrl := os.Getenv("DATABASE_URL")
	if databaseUrl == "" {
		return nil, fmt.Errorf("DATABASE_URL environment variable not set")
	}

	var db *gorm.DB
	var err error

	// Retry connecting to the database 3 times
	for i := 0; i < 3; i++ {
		db, err = gorm.Open(postgres.Open(databaseUrl), &gorm.Config{})
		if err == nil {
			break
		}
		fmt.Println("Database connection failed. Retrying in 2 seconds...")
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	return db, nil
}
