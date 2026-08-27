//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Data Seeder khusus untuk Load & Stress Testing LMS Kopkara
func main() {
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5433")
	dbUser := getEnv("DB_USER", "postgres")
	dbPass := getEnv("DB_PASSWORD", "password")
	dbName := getEnv("DB_NAME", "lms_db")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Jakarta",
		dbHost, dbUser, dbPass, dbName, dbPort)

	log.Printf("[SEEDER] Connecting to Database: %s:%s/%s ...\n", dbHost, dbPort, dbName)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		log.Fatalf("[SEEDER ERROR] Failed to connect database: %v", err)
	}

	log.Println("[SEEDER] Database connected successfully!")

	// Total target seeding
	totalEmployees := 10000
	totalLoans := 15000

	log.Printf("[SEEDER] Starting Seeding Process: %d Employees/Members & %d Loans...\n", totalEmployees, totalLoans)
	startTime := time.Now()

	// 1. Seed Master Employee Categories & Statuses
	db.Exec(`INSERT INTO lms_sch.employee_categories (category_code, description, max_limit, is_eligible, created_at, updated_at)
		VALUES 
		('PERM', 'Permanent Employee', 50000000, true, NOW(), NOW()),
		('CONT', 'Contract Employee', 15000000, true, NOW(), NOW())
		ON CONFLICT (category_code) DO NOTHING;`)

	db.Exec(`INSERT INTO lms_sch.employee_statuses (status_code, description, created_at, updated_at)
		VALUES 
		('ACTIVE', 'Active Employee', NOW(), NOW()),
		('RESIGNED', 'Resigned Employee', NOW(), NOW())
		ON CONFLICT (status_code) DO NOTHING;`)

	db.Exec(`INSERT INTO lms_sch.kopkara_statuses (status_code, description, created_at, updated_at)
		VALUES 
		('ACTIVE', 'Active Member', NOW(), NOW()),
		('INACTIVE', 'Inactive Member', NOW(), NOW())
		ON CONFLICT (status_code) DO NOTHING;`)

	// 2. Seed Master Loan Products
	db.Exec(`INSERT INTO lms_sch.loan_products (id, name, loan_type, max_tenor_months, submission_period_start, submission_period_end, min_service_months, created_at, updated_at)
		VALUES 
		(1, 'Pinjaman Multiguna', 'MULTIGUNA', 36, 1, 25, 6, NOW(), NOW()),
		(2, 'Pinjaman Pendidikan', 'EDUCATION', 24, 1, 25, 3, NOW(), NOW()),
		(3, 'Pinjaman Darurat', 'EMERGENCY', 12, 1, 30, 1, NOW(), NOW())
		ON CONFLICT (id) DO NOTHING;`)

	// 3. Bulk Insert Employees & Members using raw SQL batching for maximum speed
	log.Println("[SEEDER] Seeding Employees & Members in batches...")
	batchSize := 1000
	for i := 0; i < totalEmployees; i += batchSize {
		empSql := "INSERT INTO lms_sch.employees (employee_id, name, employee_status, deptno, category_code, role_id, salary, bank_name, bank_account_no, bank_account_name, created_at, updated_at) VALUES "
		memSql := "INSERT INTO lms_sch.members (member_no, employee_id, kopkara_status, join_date, created_at, updated_at) VALUES "

		for j := 0; j < batchSize && (i+j) < totalEmployees; j++ {
			empID := 100001 + i + j
			memNo := 200001 + i + j
			name := fmt.Sprintf("Member LoadTest %d", empID)
			salary := float64(6000000 + (rand.Intn(20) * 500000))
			joinDate := "2022-01-01"

			if j > 0 {
				empSql += ", "
				memSql += ", "
			}
			empSql += fmt.Sprintf("(%d, '%s', 'ACTIVE', 'DEPT01', 'PERM', 2, %.2f, 'BCA', '123456%d', '%s', NOW(), NOW())", empID, name, salary, empID, name)
			memSql += fmt.Sprintf("(%d, %d, 'ACTIVE', '%s', NOW(), NOW())", memNo, empID, joinDate)
		}

		empSql += " ON CONFLICT (employee_id) DO NOTHING;"
		memSql += " ON CONFLICT (member_no) DO NOTHING;"

		if err := db.Exec(empSql).Error; err != nil {
			log.Fatalf("[SEEDER ERROR] Insert employees batch failed: %v", err)
		}
		if err := db.Exec(memSql).Error; err != nil {
			log.Fatalf("[SEEDER ERROR] Insert members batch failed: %v", err)
		}
	}
	log.Printf("[SEEDER] Successfully seeded %d Employees & Members.\n", totalEmployees)

	// 4. Bulk Insert Loan Applications, Loans, and Loan Schedules
	log.Println("[SEEDER] Seeding Loan Applications & Schedules in batches...")
	for i := 0; i < totalLoans; i += batchSize {
		appSql := `INSERT INTO lms_sch.loan_applications 
			(application_no, member_no, product_id, submission_date, requested_amount, approved_amount, tenor, eligibility_result, status, principal_per_month, interest_per_month, admin_fee, total_installment, total_loan_cost, interest_rate, credit_limit, created_at, updated_at) VALUES `

		loanSql := `INSERT INTO lms_sch.loans
			(loan_no, application_no, member_no, principal_amount, admin_fee, disbursement_amount, outstanding_amount, status, created_at, updated_at) VALUES `

		for j := 0; j < batchSize && (i+j) < totalLoans; j++ {
			appNo := int64(300001 + i + j)
			loanNo := int64(400001 + i + j)
			memNo := int64(200001 + (rand.Intn(totalEmployees)))
			prodID := int64(1 + rand.Intn(3))
			amount := float64((rand.Intn(10) + 1) * 1000000)
			tenor := 12
			principalPM := amount / float64(tenor)
			interestPM := amount * 0.01
			adminFee := 50000.0
			totalInst := principalPM + interestPM

			if j > 0 {
				appSql += ", "
				loanSql += ", "
			}
			appSql += fmt.Sprintf("(%d, %d, %d, NOW(), %.2f, %.2f, %d, 'PASSED', 'DISBURSED', %.2f, %.2f, %.2f, %.2f, %.2f, 0.12, 50000000, NOW(), NOW())",
				appNo, memNo, prodID, amount, amount, tenor, principalPM, interestPM, adminFee, totalInst, amount+(interestPM*12))

			loanSql += fmt.Sprintf("(%d, %d, %d, %.2f, %.2f, %.2f, %.2f, 'ACTIVE', NOW(), NOW())",
				loanNo, appNo, memNo, amount, adminFee, amount-adminFee, amount)
		}

		appSql += " ON CONFLICT (application_no) DO NOTHING;"
		loanSql += " ON CONFLICT (loan_no) DO NOTHING;"

		if err := db.Exec(appSql).Error; err != nil {
			log.Fatalf("[SEEDER ERROR] Insert loan applications failed: %v", err)
		}
		if err := db.Exec(loanSql).Error; err != nil {
			log.Fatalf("[SEEDER ERROR] Insert loans failed: %v", err)
		}
	}

	log.Printf("[SEEDER SUCCESS] Completed All Seeding in %v!\n", time.Since(startTime))
	log.Println("[SEEDER] Test Environment is Ready for Load Testing.")
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}
