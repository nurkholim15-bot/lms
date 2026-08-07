package config

import (
	"fmt"
	"log"
	"os"
	"strings"

	"lms-backend/utils"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

var DB *gorm.DB

func ConnectDatabase() {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	dbname := os.Getenv("DB_NAME")

	// Gunakan DB_PASSWORD_ENCRYPTED jika tersedia, fallback ke DB_PASSWORD biasa
	password := os.Getenv("DB_PASSWORD")
	encryptedPassword := strings.TrimSpace(os.Getenv("DB_PASSWORD_ENCRYPTED"))
	if encryptedPassword != "" {
		secretKey := os.Getenv("JWT_SECRET")
		if secretKey == "" {
			log.Fatal("GAGAL: DB_PASSWORD_ENCRYPTED ditemukan tapi JWT_SECRET kosong di file .env")
		}
		decrypted, err := utils.DecryptAES(encryptedPassword, secretKey)
		if err != nil {
			log.Fatalf("GAGAL mendekripsi DB_PASSWORD_ENCRYPTED: %v", err)
		}
		password = decrypted
		log.Println("Database: Menggunakan password terenkripsi (DB_PASSWORD_ENCRYPTED).")
	}

	if host == "" {
		log.Println("Database configuration not fully provided, skipping DB connection for now.")
		return
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Jakarta",
		host, user, password, dbname, port)

	traceLevel := strings.TrimSpace(os.Getenv("TRACE_LEVEL"))
	var gormLogger logger.Interface
	if traceLevel == "3" {
		gormLogger = logger.Default.LogMode(logger.Info) // Detailed SQL Log
		log.Println("Database TRACE_LEVEL=3: SQL Logging ENABLED (Detailed)")
	} else if traceLevel == "1" {
		gormLogger = logger.Default.LogMode(logger.Warn) // High Level Warning
		log.Println("Database TRACE_LEVEL=1: SQL Logging set to Warn (High Level)")
	} else {
		gormLogger = logger.Default.LogMode(logger.Silent) // Off
	}

	// Ambil schema dari environment, default ke "lms_sch"
	schemaName := os.Getenv("DB_SCHEMA")
	if schemaName == "" {
		schemaName = "lms_sch"
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger,
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   schemaName + ".",
			SingularTable: false,
		},
	})

	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	DB = db
	log.Println("Database connection successfully established.")
}

