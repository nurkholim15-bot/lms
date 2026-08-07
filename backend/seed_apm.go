//go:build ignore
// +build ignore

package main

import (
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Role struct {
	RoleID   int64  `gorm:"primaryKey;column:role_id"`
	RoleName string `gorm:"column:role_name"`
}
func (Role) TableName() string { return "lms_sch.roles" }

type Menu struct {
	MenuID int64  `gorm:"primaryKey;column:menu_id"`
	Title  string `gorm:"column:title"`
	Icon   string `gorm:"column:icon"`
	Path   string `gorm:"column:path"`
	Order  int    `gorm:"column:order_seq"`
}
func (Menu) TableName() string { return "lms_sch.menus" }

type RoleMenu struct {
	RoleID int64 `gorm:"primaryKey;column:role_id"`
	MenuID int64 `gorm:"primaryKey;column:menu_id"`
}
func (RoleMenu) TableName() string { return "lms_sch.role_menus" }

type Employee struct {
	EmployeeID int64 `gorm:"primaryKey;column:employee_id"`
	RoleID     int64 `gorm:"column:role_id"`
}
func (Employee) TableName() string { return "lms_sch.employees" }

func main() {
	dsn := "host=localhost user=postgres password=password dbname=lms_db port=5432 sslmode=disable TimeZone=Asia/Jakarta"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	// AutoMigrate first to ensure tables exist
	db.AutoMigrate(&Role{}, &Menu{}, &RoleMenu{})

	db.Exec("INSERT INTO lms_sch.roles (role_id, role_name, description) VALUES (1, 'Admin', 'Administrator') ON CONFLICT DO NOTHING")
	db.Exec("INSERT INTO lms_sch.roles (role_id, role_name, description) VALUES (2, 'Anggota', 'Standard Member') ON CONFLICT DO NOTHING")

	menus := []Menu{
		{MenuID: 1, Title: "Dashboard", Icon: "📊", Path: "dashboard", Order: 1},
		{MenuID: 2, Title: "Pengajuan Pinjaman", Icon: "📝", Path: "pengajuan", Order: 2},
		{MenuID: 3, Title: "Daftar Pinjaman", Icon: "💰", Path: "pinjaman", Order: 3},
		{MenuID: 4, Title: "Master Departments", Icon: "🏢", Path: "master-departments", Order: 4},
		{MenuID: 5, Title: "Master Employees", Icon: "👥", Path: "master-employees", Order: 5},
		{MenuID: 6, Title: "Master Roles", Icon: "🔑", Path: "master-roles", Order: 6},
		{MenuID: 7, Title: "Master Menus", Icon: "📌", Path: "master-menus", Order: 7},
		{MenuID: 8, Title: "Master Role Menus", Icon: "🔗", Path: "master-role-menus", Order: 8},
	}
	for _, m := range menus {
		db.Exec("INSERT INTO lms_sch.menus (menu_id, title, icon, path, order_seq) VALUES (?, ?, ?, ?, ?) ON CONFLICT DO NOTHING", m.MenuID, m.Title, m.Icon, m.Path, m.Order)
	}

	roleMenus := []RoleMenu{
		{1, 1}, {1, 3}, {1, 4}, {1, 5}, {1, 6}, {1, 7}, {1, 8}, // Admin gets all but pengajuan
		{2, 1}, {2, 2}, {2, 3}, // Anggota gets dashboard, pengajuan, pinjaman
	}
	for _, rm := range roleMenus {
		db.Exec("INSERT INTO lms_sch.role_menus (role_id, menu_id) VALUES (?, ?) ON CONFLICT DO NOTHING", rm.RoleID, rm.MenuID)
	}

	// Update Employee 10101 to Role 2 (Anggota)
	db.Exec("UPDATE lms_sch.employees SET role_id = 2 WHERE employee_id = 10101")
	log.Println("Database seeded successfully!")
}
