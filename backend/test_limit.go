package main

import (
	"fmt"
	"log"

	"github.com/Knetic/govaluate"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type EmployeeCategory struct {
	CategoryCode string
	CategoryName string
	MaxLimit     float64
}

type GlobalParameter struct {
	KeyName  string
	KeyValue string
}

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

	var param GlobalParameter
	db.Where("key_name = ?", "LOAN_LIMIT_FORMULA").First(&param)
	fmt.Println("LOAN_LIMIT_FORMULA:", param.KeyValue)

	var cats []EmployeeCategory
	db.Find(&cats)
	for _, c := range cats {
		fmt.Printf("Category: %s, Max Limit: %.2f\n", c.CategoryCode, c.MaxLimit)
	}

	// Test govaluate
	formula := "(DAY/30) * SALARY * 0.5"
	if param.KeyValue != "" {
		formula = param.KeyValue
	}
	expr, err := govaluate.NewEvaluableExpression(formula)
	if err != nil {
		fmt.Println("Error parsing expression:", err)
		return
	}
	params := map[string]interface{}{
		"DAY":    float64(4),
		"SALARY": float64(10000000),
	}
	result, err := expr.Evaluate(params)
	if err != nil {
		fmt.Println("Error evaluating expression:", err)
	} else {
		fmt.Println("Result of", formula, "is:", result)
	}
}
