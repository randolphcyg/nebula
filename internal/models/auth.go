package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	Username  string         `gorm:"uniqueIndex;size:50;not null" json:"username"`
	Password  string         `gorm:"size:255;not null" json:"-"`
	Email     string         `gorm:"size:100" json:"email"`
	RoleID    uint           `gorm:"not null" json:"roleId"`
	Role      Role           `gorm:"foreignKey:RoleID" json:"role"`
	Status    int            `gorm:"default:1" json:"status"` // 1: 正常，0: 禁用
	LastLogin *time.Time     `json:"lastLogin"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type Role struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	Name        string         `gorm:"uniqueIndex;size:50;not null" json:"name"`
	Code        string         `gorm:"uniqueIndex;size:50;not null" json:"code"`
	Description string         `gorm:"size:255" json:"description"`
	Permissions []Permission   `gorm:"many2many:role_permissions" json:"permissions"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

type Permission struct {
	ID         uint           `gorm:"primarykey" json:"id"`
	Name       string         `gorm:"size:100;not null" json:"name"`
	Code       string         `gorm:"uniqueIndex;size:100;not null" json:"code"`
	Resource   string         `gorm:"size:100" json:"resource"`
	Action     string         `gorm:"size:50" json:"action"`
	CreatedAt  time.Time      `json:"createdAt"`
	UpdatedAt  time.Time      `json:"updatedAt"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}
