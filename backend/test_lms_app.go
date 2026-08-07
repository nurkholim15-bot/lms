//go:build ignore
// +build ignore

package main

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dsn := "host=localhost user=lms_app password=Nkl@130200 dbname=lms_db port=5433 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Println("GAGAL koneksi:", err)
		return
	}

	// Test SELECT members
	type Member struct {
		MemberNo   string
		EmployeeId int64
	}
	var members []Member
	if err := db.Raw("SELECT member_no, employee_id FROM lms_sch.members LIMIT 5").Scan(&members).Error; err != nil {
		fmt.Println("GAGAL SELECT members:", err)
	} else {
		fmt.Println("OK SELECT members:", len(members), "rows")
	}

	// Cek privileges
	type Priv struct {
		TableName string
		Privilege string
	}
	var privs []Priv
	db.Raw(
		"SELECT table_name, privilege_type as privilege FROM information_schema.role_table_grants " +
			"WHERE grantee = 'lms_app' AND table_schema = 'lms_sch' ORDER BY table_name, privilege_type",
	).Scan(&privs)

	if len(privs) == 0 {
		fmt.Println("TIDAK ADA privileges untuk lms_app di schema lms_sch!")
		fmt.Println("Solusi: Jalankan GRANT dari psql dengan user admin_lms atau postgres")
	} else {
		fmt.Println("Privileges lms_app di schema lms_sch:")
		for _, p := range privs {
			fmt.Printf("  %-30s -> %s\n", p.TableName, p.Privilege)
		}
	}
}
