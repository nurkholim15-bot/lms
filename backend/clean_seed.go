//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Script DML Cleanup Go untuk Menghapus Data Dummy Load Test
func main() {
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5433")
	dbUser := getEnv("DB_USER", "postgres")
	dbPass := getEnv("DB_PASSWORD", "password")
	dbName := getEnv("DB_NAME", "lms_db")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Jakarta",
		dbHost, dbUser, dbPass, dbName, dbPort)

	log.Printf("[CLEANUP] Connecting to Database: %s:%s/%s ...\n", dbHost, dbPort, dbName)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		log.Fatalf("[CLEANUP ERROR] Failed to connect database: %v", err)
	}

	log.Println("[CLEANUP] Database connected successfully!")
	log.Println("[CLEANUP] Starting DML Cleanup of Load Test Dummy Data...")

	// 1. Delete Loan Schedules
	res1 := db.Exec(`DELETE FROM lms_sch.loan_schedules WHERE loan_no BETWEEN 400001 AND 415000;`)
	log.Printf("[CLEANUP] Deleted %d Loan Schedules.\n", res1.RowsAffected)

	// 2. Delete Loans
	res2 := db.Exec(`DELETE FROM lms_sch.loans WHERE loan_no BETWEEN 400001 AND 415000;`)
	log.Printf("[CLEANUP] Deleted %d Loans.\n", res2.RowsAffected)

	// 3. Delete Loan Trackings
	res3 := db.Exec(`DELETE FROM lms_sch.loan_trackings WHERE application_no BETWEEN 300001 AND 315000;`)
	log.Printf("[CLEANUP] Deleted %d Loan Trackings.\n", res3.RowsAffected)

	// 4. Delete Loan Contracts
	res4 := db.Exec(`DELETE FROM lms_sch.loan_contracts WHERE application_no BETWEEN 300001 AND 315000;`)
	log.Printf("[CLEANUP] Deleted %d Loan Contracts.\n", res4.RowsAffected)

	// 5. Delete Loan Applications
	res5 := db.Exec(`DELETE FROM lms_sch.loan_applications WHERE application_no BETWEEN 300001 AND 315000;`)
	log.Printf("[CLEANUP] Deleted %d Loan Applications.\n", res5.RowsAffected)

	// 6. Delete Members
	res6 := db.Exec(`DELETE FROM lms_sch.members WHERE member_no BETWEEN 200001 AND 210000;`)
	log.Printf("[CLEANUP] Deleted %d Members.\n", res6.RowsAffected)

	// 7. Delete Employees
	res7 := db.Exec(`DELETE FROM lms_sch.employees WHERE employee_id BETWEEN 100001 AND 110000;`)
	log.Printf("[CLEANUP] Deleted %d Employees.\n", res7.RowsAffected)

	// 8. Delete Loan Products
	res8 := db.Exec(`DELETE FROM lms_sch.loan_products WHERE id IN (1, 2, 3);`)
	log.Printf("[CLEANUP] Deleted %d Loan Products.\n", res8.RowsAffected)

	// 9. Delete Master Parent Tables
	db.Exec(`DELETE FROM lms_sch.departments WHERE deptno IN ('DEPT01', 'DEPT02', 'DEPT03');`)
	db.Exec(`DELETE FROM lms_sch.roles WHERE role_id IN (1, 2);`)
	db.Exec(`DELETE FROM lms_sch.employee_categories WHERE category_code IN ('PERM', 'CONT');`)
	db.Exec(`DELETE FROM lms_sch.employee_statuses WHERE status_code IN ('ACTIVE', 'RESIGNED');`)
	db.Exec(`DELETE FROM lms_sch.kopkara_statuses WHERE status_code IN ('ACTIVE', 'INACTIVE');`)

	log.Println("[CLEANUP SUCCESS] All Load Test Dummy Data Cleaned Up Successfully!")
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}
