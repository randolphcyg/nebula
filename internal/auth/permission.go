package auth

import (
	"context"

	"gorm.io/gorm"

	"nebula/internal/models"
)

type PermissionChecker struct {
	db *gorm.DB
}

func NewPermissionChecker(db *gorm.DB) *PermissionChecker {
	return &PermissionChecker{
		db: db,
	}
}

func (p *PermissionChecker) HasPermission(ctx context.Context, roleCode, resource, action string) (bool, error) {
	var role models.Role
	if err := p.db.Preload("Permissions").Where("code = ?", roleCode).First(&role).Error; err != nil {
		return false, err
	}

	for _, perm := range role.Permissions {
		if perm.Resource == resource && perm.Action == action {
			return true, nil
		}
	}

	return false, nil
}

func (p *PermissionChecker) GetUserPermissions(ctx context.Context, userID uint) ([]models.Permission, error) {
	var user models.User
	if err := p.db.Preload("Role.Permissions").First(&user, userID).Error; err != nil {
		return nil, err
	}

	return user.Role.Permissions, nil
}

func (p *PermissionChecker) GetRoleByCode(ctx context.Context, code string) (*models.Role, error) {
	var role models.Role
	if err := p.db.Preload("Permissions").Where("code = ?", code).First(&role).Error; err != nil {
		return nil, err
	}
	return &role, nil
}
