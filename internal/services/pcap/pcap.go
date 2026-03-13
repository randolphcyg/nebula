package pcap

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"gorm.io/gorm"

	"nebula/internal/database"
	"nebula/internal/models"
	"nebula/internal/utils"
)

// ================= 请求和响应数据结构 =================

type FileQueryReq struct {
	FileName  string `json:"fileName"`
	FileSize  string `json:"fileSize"`
	StartDate string `json:"startDate"`
	EndDate   string `json:"endDate"`
	Page      int    `json:"page"`
	PageSize  int    `json:"pageSize"`
}

type FileQueryResp struct {
	List  []models.PcapFile `json:"list"`
	Total int64             `json:"total"`
}

// ================= 进度读取器 =================

type ProgressReader struct {
	Reader     io.Reader
	Total      int64
	Loaded     int64
	Ctx        context.Context
	FileID     string
	LastEmitPt int
}

func (pr *ProgressReader) Read(p []byte) (int, error) {
	n, err := pr.Reader.Read(p)
	pr.Loaded += int64(n)

	if pr.Total > 0 {
		percent := int((float64(pr.Loaded) / float64(pr.Total)) * 100)
		if percent-pr.LastEmitPt >= 2 || percent == 100 {
			pr.LastEmitPt = percent
			runtime.EventsEmit(pr.Ctx, "pcap:progress", map[string]interface{}{
				"fileId": pr.FileID, "percent": percent,
			})
		}
	}
	return n, err
}

type Service struct {
	ctx       context.Context
	db        *gorm.DB
	workspace string
}

func NewService() *Service {
	return &Service{}
}

func (s *Service) Start(ctx context.Context, dbService interface{}) error {
	s.ctx = ctx

	switch svc := dbService.(type) {
	case *database.Database:
		s.db = svc.GetDB()
	case *gorm.DB:
		s.db = svc
	default:
		return fmt.Errorf("无效的数据库服务类型")
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("获取用户目录失败：%v", err)
	}
	s.workspace = filepath.Join(homeDir, ".nebula", "data", "pcaps")
	os.MkdirAll(s.workspace, os.ModePerm)

	if err := s.db.AutoMigrate(&models.PcapFile{}); err != nil {
		return fmt.Errorf("数据库迁移失败：%v", err)
	}

	return nil
}

func (s *Service) ImportPcapsDialog() ([]models.PcapFile, error) {
	selectedFiles, err := runtime.OpenMultipleFilesDialog(s.ctx, runtime.OpenDialogOptions{
		Title:   "选择 PCAP 流量包",
		Filters: []runtime.FileFilter{{DisplayName: "PCAP Files", Pattern: "*.pcap;*.pcapng"}},
	})
	if err != nil || len(selectedFiles) == 0 {
		return nil, nil
	}
	return s.ImportFromPaths(selectedFiles)
}

func (s *Service) ImportFromPaths(paths []string) ([]models.PcapFile, error) {
	var importedFiles []models.PcapFile

	for _, srcPath := range paths {
		fileName := filepath.Base(srcPath)
		fileID := utils.NewRandomID()

		monthDir := time.Now().Format("2006-01")
		saveDir := filepath.Join(s.workspace, monthDir)
		os.MkdirAll(saveDir, os.ModePerm)

		saveFileName := fmt.Sprintf("%s_%s", fileID, fileName)
		destPath := filepath.Join(saveDir, saveFileName)

		fileInfo, err := os.Stat(srcPath)
		if err != nil {
			continue
		}

		pcapFile := models.PcapFile{
			FileID:   fileID,
			FileName: fileName,
			FilePath: destPath,
			FileSize: formatFileSize(fileInfo.Size()),
			Status:   "导入中",
		}
		s.db.Create(&pcapFile)

		if fileInfo.Size() == 0 {
			pcapFile.Status = "导入失败"
			s.db.Model(&models.PcapFile{}).Where("file_id = ?", fileID).Update("status", "导入失败")

			runtime.EventsEmit(s.ctx, "pcap:progress", map[string]interface{}{
				"fileId": fileID, "percent": -1, "fileName": fileName, "error": "无效的流量包 (0 字节)",
			})

			importedFiles = append(importedFiles, pcapFile)
			continue
		}

		importedFiles = append(importedFiles, pcapFile)

		go func(src, dst string, size int64, fId string, fName string) {
			runtime.EventsEmit(s.ctx, "pcap:progress", map[string]interface{}{
				"fileId": fId, "percent": 0, "fileName": fName,
			})

			err := s.copyWithProgress(src, dst, size, fId)

			if err != nil {
				s.db.Model(&models.PcapFile{}).Where("file_id = ?", fId).Update("status", "导入失败")
				runtime.EventsEmit(s.ctx, "pcap:progress", map[string]interface{}{
					"fileId": fId, "percent": -1, "fileName": fName, "error": err.Error(),
				})
				return
			}

			s.db.Model(&models.PcapFile{}).Where("file_id = ?", fId).Update("status", "导入成功")
			runtime.EventsEmit(s.ctx, "pcap:progress", map[string]interface{}{
				"fileId": fId, "percent": 100, "fileName": fName,
			})
		}(srcPath, destPath, fileInfo.Size(), fileID, fileName)
	}

	return importedFiles, nil
}

func (s *Service) copyWithProgress(src, dst string, totalSize int64, fileID string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	progressReader := &ProgressReader{Reader: in, Total: totalSize, Ctx: s.ctx, FileID: fileID}
	_, err = io.Copy(out, progressReader)
	return err
}

func formatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func (s *Service) GetFileList(req FileQueryReq) (*FileQueryResp, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 10
	}

	var files []models.PcapFile
	var total int64

	query := s.db.Model(&models.PcapFile{})

	if req.FileName != "" {
		query = query.Where("file_name LIKE ?", "%"+req.FileName+"%")
	}

	if req.FileSize != "" {
		query = query.Where("file_size LIKE ?", "%"+req.FileSize+"%")
	}

	if req.StartDate != "" {
		query = query.Where("created_at >= ?", req.StartDate+" 00:00:00")
	}
	if req.EndDate != "" {
		query = query.Where("created_at <= ?", req.EndDate+" 23:59:59")
	}

	query.Count(&total)

	offset := (req.Page - 1) * req.PageSize
	if err := query.Order("created_at desc").Limit(req.PageSize).Offset(offset).Find(&files).Error; err != nil {
		return nil, err
	}

	return &FileQueryResp{List: files, Total: total}, nil
}

func (s *Service) DeleteFile(id uint) error {
	var file models.PcapFile
	if err := s.db.First(&file, id).Error; err != nil {
		return err
	}
	os.Remove(file.FilePath)
	return s.db.Delete(&file).Error
}

func (s *Service) BatchDeleteFiles(ids []uint) error {
	if len(ids) == 0 {
		return nil
	}

	var files []models.PcapFile
	if err := s.db.Where("id IN ?", ids).Find(&files).Error; err != nil {
		return err
	}

	for _, file := range files {
		_ = os.Remove(file.FilePath)
	}

	return s.db.Delete(&models.PcapFile{}, ids).Error
}
