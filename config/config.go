package config

import (
	"api-gateway/model"
	"database/sql"
	"fmt"
	"log"
	"time"
)

var Config *model.Config
var DB *sql.DB

func Startup() {
	config, err := model.LoadConfig()
	if err != nil {
		fmt.Println(err.Error())
	}
	Config = config
}

func ConnectDB() {
	configDB := Config.Database

	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4&loc=Local",
		configDB["user"],
		configDB["password"],
		configDB["host"],
		configDB["port"],
		configDB["schema"],
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to open DB: %v", err)
	}

	// Test DB connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping DB: %v", err)
	}

	// Set connection pool settings
	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(100)
	db.SetConnMaxLifetime(time.Hour)

	DB = db
}
