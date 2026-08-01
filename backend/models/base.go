package models

import (
	"time"
)

// Base Model for Audit Fields
type BaseModel struct {
	CreatedAt   *time.Time `gorm:"column:created_at;autoCreateTime"`
	CreatedUser *string    `gorm:"column:created_user"`
	UpdatedAt   *time.Time `gorm:"column:updated_at;autoUpdateTime"`
	UpdatedUser *string    `gorm:"column:updated_user"`
}

// BaseModel with Soft Delete for Master Tables
type MasterBaseModel struct {
	BaseModel
	DeletedAt   *time.Time `gorm:"column:deleted_at"`
	DeletedUser *string    `gorm:"column:deleted_user"`
}
