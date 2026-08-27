package models

import (
	"time"
)

// User represents lms_sch.users table
type User struct {
	ID                  int64      `gorm:"primaryKey;column:id;autoIncrement:true" json:"id"`
	Username            string     `gorm:"column:username;uniqueIndex" json:"username"`
	Password            string     `gorm:"column:password" json:"-"`
	Name                string     `gorm:"column:name" json:"name"`
	RoleID              int64      `gorm:"column:role_id" json:"role_id"`
	MemberNo            *int64     `gorm:"column:member_no;uniqueIndex" json:"member_no"`
	NoKTP               *string    `gorm:"column:no_ktp" json:"no_ktp,omitempty"`
	PhoneNumber         *string    `gorm:"column:phone_number" json:"phone_number,omitempty"`
	PIN                 *string    `gorm:"column:pin" json:"-"`
	FailedPINAttempts   int        `gorm:"column:failed_pin_attempts;default:0" json:"failed_pin_attempts"`
	PINLockedUntil      *time.Time `gorm:"column:pin_locked_until" json:"pin_locked_until,omitempty"`
	FailedLoginAttempts int        `gorm:"column:failed_login_attempts;default:0" json:"failed_login_attempts"`
	LockedUntil         *time.Time `gorm:"column:locked_until" json:"locked_until"`
	PasswordChangedAt   time.Time  `gorm:"column:password_changed_at;autoCreateTime" json:"password_changed_at"`
	CreatedAt           time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	CreatedUser         *string    `gorm:"column:created_user" json:"created_user,omitempty"`
	UpdatedAt           time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	UpdatedUser         *string    `gorm:"column:updated_user" json:"updated_user,omitempty"`
	DeletedAt           *time.Time `gorm:"column:deleted_at;index" json:"deleted_at,omitempty"`
	DeletedUser         *string    `gorm:"column:deleted_user" json:"deleted_user,omitempty"`
	ForcePwdChange      bool       `gorm:"column:force_pwd_change;default:false" json:"force_pwd_change"`
}

func (User) TableName() string {
	return "lms_sch.users"
}
