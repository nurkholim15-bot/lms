//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"os"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Penggunaan: go run generate_bcrypt_hash.go \"password_anda\"")
		os.Exit(1)
	}

	plainPassword := os.Args[1]
	hash, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
	if err != nil {
		fmt.Printf("Gagal membuat hash: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Hash Bcrypt Anda:")
	fmt.Println(string(hash))
}
