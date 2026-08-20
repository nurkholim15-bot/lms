package models

import (
	"time"
)

// Session represents lms_sch.sessions table
type Session struct {
	ID             int64     `gorm:"primaryKey;column:id;autoIncrement:true" json:"id"`
	Token          string    `gorm:"column:token;uniqueIndex" json:"token"`
	UserID         int64     `gorm:"column:user_id;index" json:"user_id"`
	Username       string    `gorm:"column:username" json:"username"`
	IPAddress      string    `gorm:"column:ip_address" json:"ip_address"`
	UserAgent      string    `gorm:"column:user_agent" json:"user_agent"`
	IsActive       bool      `gorm:"column:is_active;default:true" json:"is_active"`
	LoginAt        time.Time `gorm:"column:login_at;autoCreateTime" json:"login_at"`
	ExpiresAt      time.Time `gorm:"column:expires_at" json:"expires_at"`
	LastActivityAt time.Time `gorm:"column:last_activity_at" json:"last_activity_at"`
	CreatedAt      time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (Session) TableName() string {
	return "lms_sch.sessions"
}
