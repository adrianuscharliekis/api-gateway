package config

import (
	"api-gateway/model"
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var Config *model.Config
var DB *gorm.DB

func Startup() {
	config, err := model.LoadConfig()
	if err != nil {
		fmt.Println(err.Error())
	}
	Config = config
}

func ConnectDB() {
	configDB := Config.Database
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,   // Slow SQL threshold
			LogLevel:                  logger.Silent, // Set log level to silent
			IgnoreRecordNotFoundError: true,          // Ignore ErrRecordNotFound error for logger
			Colorful:                  false,         // Disable color
		},
	)
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		configDB["user"],
		configDB["password"],
		configDB["host"],
		configDB["port"],
		configDB["schema"],
	)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("Failed to get DB from GORM: ", err)
	}

	// Set connection pool
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	DB = db
}
