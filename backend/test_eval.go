package main

import (
	"fmt"
	"log"

	"github.com/Knetic/govaluate"
)

func main() {
	formula := "(DAY/30) * SALARY * 0.5"

	// Integer division!
	// Wait, govaluate parses numbers as float64, so 4/30 is float division.
	
	expr, err := govaluate.NewEvaluableExpression(formula)
	if err != nil {
		log.Fatal(err)
	}

	params := map[string]interface{}{
		"DAY":    float64(4),
		"SALARY": float64(10000000),
		"REQUESTED_AMOUNT": float64(10000000),
	}
	
	res, err := expr.Evaluate(params)
	fmt.Printf("Result: %v\n", res)
}
