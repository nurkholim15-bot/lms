package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"lms-backend/utils"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

var DB *gorm.DB

// SelectOnlyLogger filters GORM SQL logs to output only SELECT statements
type SelectOnlyLogger struct {
	logger.Interface
}

func (l *SelectOnlyLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	sqlStr, _ := fc()
	trimmed := strings.TrimSpace(strings.ToUpper(sqlStr))
	if strings.HasPrefix(trimmed, "SELECT") {
		l.Interface.Trace(ctx, begin, fc, err)
	}
}

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
		gormLogger = logger.Default.LogMode(logger.Info) // All SQL Queries (SELECT, INSERT, UPDATE, DELETE)
		log.Println("Database TRACE_LEVEL=3: SQL Logging ENABLED (ALL Queries: SELECT, INSERT, UPDATE, DELETE)")
	} else if traceLevel == "2" {
		gormLogger = &SelectOnlyLogger{Interface: logger.Default.LogMode(logger.Info)} // SELECT SQL Queries Only
		log.Println("Database TRACE_LEVEL=2: SQL Logging ENABLED (SELECT Queries Only)")
	} else if traceLevel == "1" {
		gormLogger = logger.Default.LogMode(logger.Silent) // API Request Log Only (NO SQL Log)
		log.Println("Database TRACE_LEVEL=1: API Request Log Only (SQL Logging OFF)")
	} else {
		gormLogger = logger.Default.LogMode(logger.Silent) // All Logs Off
		log.Println("Database TRACE_LEVEL=0: SQL Logging OFF")
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

	sqlDB, err := db.DB()
	if err == nil {
		sqlDB.SetMaxOpenConns(200)                // Maximum 200 open connections for 500 VUs load test
		sqlDB.SetMaxIdleConns(100)               // 100 idle connections ready
		sqlDB.SetConnMaxLifetime(5 * time.Minute) // Conn max lifetime
	}
	log.Println("Database connection successfully established with optimized Connection Pool (Max 200 Conns).")
}

