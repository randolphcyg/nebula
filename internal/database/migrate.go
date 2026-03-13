package database

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"nebula/internal/models"
)

func (d *Database) Migrate() error {
	return d.db.AutoMigrate(
		&models.PcapFile{},
		&models.User{},
		&models.Role{},
		&models.Permission{},
	)
}

func (d *Database) Seed() error {
	var count int64
	if err := d.db.Model(&models.Role{}).Count(&count).Error; err != nil {
		return err
	}

	if count == 0 {
		roles := []models.Role{
			{
				Name:        "超级管理员",
				Code:        "admin",
				Description: "拥有系统所有权限",
			},
			{
				Name:        "分析师",
				Code:        "analyst",
				Description: "可以查看和分析数据",
			},
			{
				Name:        "访客",
				Code:        "guest",
				Description: "只读权限",
			},
		}

		for _, role := range roles {
			if err := d.db.Create(&role).Error; err != nil {
				return fmt.Errorf("创建角色 %s 失败：%v", role.Name, err)
			}
		}
	}

	permissions := []models.Permission{
		{Resource: "pcap", Action: "read", Name: "查看 PCAP"},
		{Resource: "pcap", Action: "write", Name: "管理 PCAP"},
		{Resource: "pcap", Action: "delete", Name: "删除 PCAP"},
		{Resource: "analyzer", Action: "read", Name: "使用分析引擎"},
		{Resource: "analyzer", Action: "write", Name: "配置分析引擎"},
		{Resource: "user", Action: "read", Name: "查看用户"},
		{Resource: "user", Action: "write", Name: "管理用户"},
		{Resource: "role", Action: "read", Name: "查看角色"},
		{Resource: "role", Action: "write", Name: "管理角色"},
	}

	for _, perm := range permissions {
		perm.Code = perm.Resource + "_" + perm.Action
		if err := d.db.Where("code = ?", perm.Code).FirstOrCreate(&perm).Error; err != nil {
			continue
		}
	}

	var adminRole models.Role
	if err := d.db.Where("code = ?", "admin").First(&adminRole).Error; err != nil {
		return fmt.Errorf("获取管理员角色失败：%v", err)
	}

	var allPermissions []models.Permission
	if err := d.db.Find(&allPermissions).Error; err != nil {
		return err
	}

	if err := d.db.Model(&adminRole).Association("Permissions").Replace(allPermissions); err != nil {
		return fmt.Errorf("分配权限失败：%v", err)
	}

	var analystRole models.Role
	if err := d.db.Where("code = ?", "analyst").First(&analystRole).Error; err != nil {
		return err
	}

	var analystPermissions []models.Permission
	if err := d.db.Where("resource IN ? AND action IN ?", []string{"pcap", "analyzer"}, []string{"read", "write"}).Find(&analystPermissions).Error; err != nil {
		return err
	}
	if err := d.db.Model(&analystRole).Association("Permissions").Replace(analystPermissions); err != nil {
		return fmt.Errorf("分配分析师权限失败：%v", err)
	}

	var guestRole models.Role
	if err := d.db.Where("code = ?", "guest").First(&guestRole).Error; err != nil {
		return err
	}

	var guestPermissions []models.Permission
	if err := d.db.Where("action = ?", "read").Find(&guestPermissions).Error; err != nil {
		return err
	}
	if err := d.db.Model(&guestRole).Association("Permissions").Replace(guestPermissions); err != nil {
		return fmt.Errorf("分配访客权限失败：%v", err)
	}

	var adminUser models.User
	if err := d.db.Where("username = ?", "admin").First(&adminUser).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
			if err != nil {
				return fmt.Errorf("加密密码失败：%v", err)
			}

			adminUser = models.User{
				Username: "admin",
				Password: string(hashedPassword),
				Email:    "admin@nebula.local",
				RoleID:   adminRole.ID,
				Status:   1,
			}
			if err := d.db.Create(&adminUser).Error; err != nil {
				return fmt.Errorf("创建管理员用户失败：%v", err)
			}
		} else {
			return fmt.Errorf("查询管理员用户失败：%v", err)
		}
	}

	return nil
}
