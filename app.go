package main

import (
	"context"
	"fmt"

	"nebula/internal/auth"
	"nebula/internal/database"
	"nebula/internal/models"
	"nebula/internal/services/analyzer"
	"nebula/internal/services/pcap"
)

// App struct
type App struct {
	ctx context.Context

	// 核心服务
	Database        *database.Database
	AuthService     *auth.Service
	PermissionCheck *auth.PermissionChecker
	AnalyzerService *analyzer.Service
	PcapService     *pcap.Service
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// 1. 初始化数据库
	db, err := database.NewDatabase(database.Config{})
	if err != nil {
		fmt.Printf("数据库初始化失败：%v\n", err)
		return
	}
	a.Database = db

	// 2. 执行数据库迁移和种子数据
	if err := db.Migrate(); err != nil {
		fmt.Printf("数据库迁移失败：%v\n", err)
		return
	}

	if err := db.Seed(); err != nil {
		fmt.Printf("种子数据初始化失败：%v\n", err)
		return
	}

	// 3. 初始化认证服务
	a.AuthService = auth.NewService(db.GetDB(), auth.Config{
		JWTSecret: "nebula-jwt-secret-key-change-in-production-2024",
	})

	// 4. 初始化权限检查器
	a.PermissionCheck = auth.NewPermissionChecker(db.GetDB())

	// 5. 初始化 PCAP 服务（使用已有的数据库连接）
	a.PcapService = pcap.NewService()
	if err := a.PcapService.Start(ctx, db); err != nil {
		fmt.Printf("PCAP 服务启动失败：%v\n", err)
	}

	// 6. 初始化分析器服务
	a.AnalyzerService = analyzer.NewService()

	fmt.Println("NEBULA 系统初始化完成")
}

// ============================
// wireshark
// ============================

// 获取底层分析引擎版本
func (a *App) GetWiresharkVersion() (string, error) {
	return a.AnalyzerService.GetWiresharkVersion()
}

// 分页获取数据包列表
func (a *App) GetPacketsByPage(filepath string, page int, size int, bpfFilter string) (string, error) {
	return a.AnalyzerService.GetPacketsByPage(filepath, page, size, bpfFilter)
}

// 获取单帧的详细解析数据 (Protocol Tree & Hex)
func (a *App) GetPacketDetail(filepath string, index int) (string, error) {
	return a.AnalyzerService.GetPacketDetail(filepath, index)
}

func (a *App) GetPacketHex(filepath string, index int) (string, error) {
	return a.AnalyzerService.GetPacketHex(filepath, index)
}

// 一次性获取所有数据包
func (a *App) GetAllFrames(filepath string, bpfFilter string) (string, error) {
	return a.AnalyzerService.GetAllFrames(filepath, bpfFilter)
}

// 安全且极速地追踪并重组数据流
func (a *App) FollowStream(filepath string, bpfFilter string, protocol string) (string, error) {
	return a.AnalyzerService.FollowStream(filepath, bpfFilter, protocol)
}

// 获取网卡列表
func (a *App) GetInterfaces() (string, error) {
	return a.AnalyzerService.GetInterfaces()
}

// ============================
// pcap
// ============================

// 唤起多选文件对话框并导入
func (a *App) ImportPcapsDialog() ([]models.PcapFile, error) {
	return a.PcapService.ImportPcapsDialog()
}

// 接收前端拖拽传递过来的绝对路径数组并导入
func (a *App) ImportFromPaths(paths []string) ([]models.PcapFile, error) {
	return a.PcapService.ImportFromPaths(paths)
}

// 获取文件列表
func (a *App) GetFileList(req pcap.FileQueryReq) (*pcap.FileQueryResp, error) {
	return a.PcapService.GetFileList(req)
}

// 删除指定记录及底层文件
func (a *App) DeleteFile(id uint) error {
	return a.PcapService.DeleteFile(id)
}

// 批量删除文件
func (a *App) BatchDeleteFiles(ids []uint) error {
	return a.PcapService.BatchDeleteFiles(ids)
}

// ============================
// 认证与权限
// ============================

// 用户登录
func (a *App) Login(username, password string) (map[string]interface{}, error) {
	resp, err := a.AuthService.Login(auth.LoginRequest{
		Username: username,
		Password: password,
	})
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"token":     resp.Token,
		"expiresAt": resp.ExpiresAt,
		"user":      resp.User,
	}, nil
}

// 验证 Token
func (a *App) ValidateToken(token string) (map[string]interface{}, error) {
	claims, err := a.AuthService.ValidateToken(token)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"userID":   claims.UserID,
		"username": claims.Username,
		"role":     claims.Role,
		"roleCode": claims.RoleCode,
	}, nil
}

// 获取当前用户信息
func (a *App) GetCurrentUser(token string) (map[string]interface{}, error) {
	claims, err := a.AuthService.ValidateToken(token)
	if err != nil {
		return nil, err
	}

	user, err := a.AuthService.GetUserByID(claims.UserID)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"id":        user.ID,
		"username":  user.Username,
		"email":     user.Email,
		"role":      user.Role.Name,
		"roleCode":  user.Role.Code,
		"status":    user.Status,
		"lastLogin": user.LastLogin,
	}, nil
}

// 检查权限
func (a *App) CheckPermission(token, resource, action string) (bool, error) {
	claims, err := a.AuthService.ValidateToken(token)
	if err != nil {
		return false, err
	}

	return a.PermissionCheck.HasPermission(context.Background(), claims.RoleCode, resource, action)
}

// 修改密码
func (a *App) ChangePassword(token, oldPassword, newPassword string) error {
	claims, err := a.AuthService.ValidateToken(token)
	if err != nil {
		return err
	}

	return a.AuthService.ChangePassword(claims.UserID, oldPassword, newPassword)
}
