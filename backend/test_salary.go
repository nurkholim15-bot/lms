package main

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func main() {
	dsn := "host=localhost user=admin_lms password=Nkl@130200 dbname=lms_db port=5433 sslmode=disable TimeZone=Asia/Jakarta"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix: "lms_sch.",
		},
	})
	if err != nil {
		log.Fatal("failed to connect database")
	}

	type Result struct {
		MemberNo string
		EmployeeId string
		Name string
		Salary float64
	}
	var results []Result
	db.Raw(`
		SELECT m.member_no, e.employee_id, e.name, e.salary 
		FROM lms_sch.members m 
		JOIN lms_sch.employees e ON m.employee_id = e.employee_id 
	`).Scan(&results)

	for _, r := range results {
		fmt.Printf("Member: %s (%s) - %s | Salary: %.2f\n", r.MemberNo, r.EmployeeId, r.Name, r.Salary)
	}
}
