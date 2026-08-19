package models

import (
	"time"

	"gorm.io/gorm"
)

// Base Model for Audit Fields
type BaseModel struct {
	CreatedAt   *time.Time `gorm:"column:created_at;autoCreateTime"`
	CreatedUser *string    `gorm:"column:created_user"`
	UpdatedAt   *time.Time `gorm:"column:updated_at;autoUpdateTime"`
	UpdatedUser *string    `gorm:"column:updated_user"`
}

// BaseModel with Soft Delete for Master Tables & Audit Protection
type MasterBaseModel struct {
	BaseModel
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at;index"`
	DeletedUser *string        `gorm:"column:deleted_user"`
}
