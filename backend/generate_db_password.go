//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"log"
	"lms-backend/utils"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Cara Penggunaan:")
		fmt.Println("go run generate_db_password.go \"PasswordBaruAnda\"")
		os.Exit(1)
	}

	// Load file .env untuk mengambil JWT_SECRET yang sama
	err := godotenv.Load()
	if err != nil {
		log.Println("Peringatan: File .env tidak ditemukan, pastikan file ini ada di folder yang sama.")
	}

	newPassword := os.Args[1]
	secretKey := os.Getenv("JWT_SECRET")

	if secretKey == "" {
		log.Fatal("GAGAL: Parameter JWT_SECRET di file .env kosong, tidak bisa melakukan enkripsi.")
	}

	encrypted, err := utils.EncryptAES(newPassword, secretKey)
	if err != nil {
		log.Fatal("Gagal mengenkripsi password:", err)
	}

	fmt.Println("\n==========================================================================")
	fmt.Println("BERHASIL!")
	fmt.Println("Silahkan copy nilai di bawah ini dan jadikan isi dari variabel DB_PASSWORD_ENCRYPTED di file .env Anda:")
	fmt.Println("==========================================================================")
	fmt.Printf("%s\n", encrypted)
	fmt.Println("==========================================================================\n")
}
