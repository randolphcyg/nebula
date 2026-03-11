package models

import (
	"time"

	"gorm.io/gorm"
)

// PcapFile 对应本地存储的流量包实体记录
type PcapFile struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	FileID    string         `gorm:"uniqueIndex" json:"fileId"`
	FileName  string         `json:"fileName"`
	FilePath  string         `json:"filePath"` // 真实存在系统里的路径
	FileSize  string         `json:"fileSize"`
	Status    string         `json:"status"` // 状态: 导入中, 导入失败, 导入成功
	CreatedAt time.Time      `json:"createdAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
