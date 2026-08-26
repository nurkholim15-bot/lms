package models

import (
	"strings"
	"time"

	"gorm.io/gorm"
)

// Base Model for Audit Fields
type BaseModel struct {
	CreatedAt   *time.Time `gorm:"column:created_at;autoCreateTime;<-:create" json:"created_at"`
	CreatedUser *string    `gorm:"column:created_user;<-:create" json:"created_user"`
	UpdatedAt   *time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	UpdatedUser *string    `gorm:"column:updated_user" json:"updated_user"`
}

func (b *BaseModel) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	if b.CreatedAt == nil {
		b.CreatedAt = &now
	}
	if b.UpdatedAt == nil {
		b.UpdatedAt = &now
	}
	if b.CreatedUser == nil || strings.TrimSpace(*b.CreatedUser) == "" {
		defUser := "SYSTEM_AUTO"
		b.CreatedUser = &defUser
	}
	if b.UpdatedUser == nil || strings.TrimSpace(*b.UpdatedUser) == "" {
		b.UpdatedUser = b.CreatedUser
	}
	return nil
}

func (b *BaseModel) BeforeUpdate(tx *gorm.DB) error {
	now := time.Now()
	b.UpdatedAt = &now
	if b.UpdatedUser == nil || strings.TrimSpace(*b.UpdatedUser) == "" {
		if b.CreatedUser != nil && strings.TrimSpace(*b.CreatedUser) != "" {
			b.UpdatedUser = b.CreatedUser
		} else {
			defUser := "SYSTEM_AUTO"
			b.UpdatedUser = &defUser
		}
	}
	return nil
}

// BaseModel with Soft Delete for Master Tables & Audit Protection
type MasterBaseModel struct {
	BaseModel
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at;index;<-:create" json:"deleted_at"`
	DeletedUser *string        `gorm:"column:deleted_user" json:"deleted_user"`
}

func (m *MasterBaseModel) BeforeDelete(tx *gorm.DB) error {
	if m.DeletedUser == nil || strings.TrimSpace(*m.DeletedUser) == "" {
		if m.UpdatedUser != nil && strings.TrimSpace(*m.UpdatedUser) != "" {
			m.DeletedUser = m.UpdatedUser
		} else if m.CreatedUser != nil && strings.TrimSpace(*m.CreatedUser) != "" {
			m.DeletedUser = m.CreatedUser
		} else {
			defUser := "SYSTEM_AUTO"
			m.DeletedUser = &defUser
		}
	}
	return nil
}
