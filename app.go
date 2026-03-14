package main

import (
	"context"
	"fmt"

	"nebula/internal/auth"
	"nebula/internal/config"
	"nebula/internal/database"
	"nebula/internal/models"
	"nebula/internal/services/analyzer"
	"nebula/internal/services/pcap"
	"nebula/internal/services/zeek"
	"nebula/internal/utils"
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
	ZeekService     *zeek.Service
	Crypto          *utils.Crypto
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

	// 7. 初始化加密工具
	a.Crypto, err = utils.NewCrypto()
	if err != nil {
		fmt.Printf("加密工具初始化失败：%v\n", err)
		return
	}

	// 8. 初始化 Zeek 服务
	cfg := config.Get()
	if cfg.ZeekRunner.IsEnabled() {
		a.ZeekService, err = zeek.NewService(zeek.ServiceConfig{
			GRPCAddress: cfg.ZeekRunner.GetGRPCAddress(),
			HTTPAddress: cfg.ZeekRunner.GetHTTPAddress(),
			Timeout:     cfg.ZeekRunner.GetTimeout(),
		})
		if err != nil {
			fmt.Printf("⚠️  Zeek 服务初始化失败：%v (功能将被禁用)\n", err)
			a.ZeekService = nil
		} else {
			fmt.Println("✅ Zeek 服务初始化成功")
		}
	} else {
		fmt.Println("ℹ️  Zeek 服务未启用")
		a.ZeekService = nil
	}

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

// 获取公钥（用于前端加密密码）
func (a *App) GetPublicKey() (string, error) {
	return a.Crypto.GetPublicKey()
}

// 用户登录（加密密码）
func (a *App) Login(username, encryptedPassword string) (map[string]interface{}, error) {
	// 解密密码
	password, err := a.Crypto.DecryptPassword(encryptedPassword)
	if err != nil {
		return nil, fmt.Errorf("密码解密失败：%v", err)
	}

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

// 用户注册（加密密码）
func (a *App) Register(username, email, encryptedPassword string) error {
	// 解密密码
	password, err := a.Crypto.DecryptPassword(encryptedPassword)
	if err != nil {
		return fmt.Errorf("密码解密失败：%v", err)
	}

	return a.AuthService.Register(auth.RegisterRequest{
		Username: username,
		Email:    email,
		Password: password,
	})
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

// 获取所有用户（管理员功能）
func (a *App) GetUsers() ([]map[string]interface{}, error) {
	users, err := a.AuthService.GetAllUsers()
	if err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, 0, len(users))
	for _, user := range users {
		result = append(result, map[string]interface{}{
			"id":        user.ID,
			"username":  user.Username,
			"email":     user.Email,
			"role":      user.Role.Name,
			"roleCode":  user.Role.Code,
			"roleID":    user.RoleID,
			"status":    user.Status,
			"lastLogin": user.LastLogin,
			"createdAt": user.CreatedAt,
		})
	}
	return result, nil
}

// 搜索用户
func (a *App) SearchUsers(keyword string) ([]map[string]interface{}, error) {
	users, err := a.AuthService.SearchUsers(keyword)
	if err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, 0, len(users))
	for _, user := range users {
		result = append(result, map[string]interface{}{
			"id":        user.ID,
			"username":  user.Username,
			"email":     user.Email,
			"role":      user.Role.Name,
			"roleCode":  user.Role.Code,
			"roleID":    user.RoleID,
			"status":    user.Status,
			"lastLogin": user.LastLogin,
			"createdAt": user.CreatedAt,
		})
	}
	return result, nil
}

// 更新用户状态（管理员功能）
func (a *App) UpdateUserStatus(userID int, status int, token string) error {
	// 获取当前操作者信息
	claims, err := a.AuthService.ValidateToken(token)
	if err != nil {
		return fmt.Errorf("验证 token 失败：%v", err)
	}

	return a.AuthService.UpdateUserStatus(userID, status, claims.UserID, claims.Username)
}

// 批量更新用户状态
func (a *App) BatchUpdateUserStatus(userIDs []int, status int, token string) error {
	claims, err := a.AuthService.ValidateToken(token)
	if err != nil {
		return fmt.Errorf("验证 token 失败：%v", err)
	}

	return a.AuthService.BatchUpdateUserStatus(userIDs, status, claims.UserID, claims.Username)
}

// 获取审核日志
func (a *App) GetAuditLogs(limit int) ([]map[string]interface{}, error) {
	logs, err := a.AuthService.GetAuditLogs(limit)
	if err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, 0, len(logs))
	for _, log := range logs {
		result = append(result, map[string]interface{}{
			"id":        log.ID,
			"userID":    log.UserID,
			"username":  log.Username,
			"operator":  log.Operator,
			"action":    log.Action,
			"oldStatus": log.OldStatus,
			"newStatus": log.NewStatus,
			"remark":    log.Remark,
			"createdAt": log.CreatedAt,
		})
	}
	return result, nil
}

// 获取所有角色
func (a *App) GetAllRoles() ([]map[string]interface{}, error) {
	roles, err := a.AuthService.GetAllRoles()
	if err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, 0, len(roles))
	for _, role := range roles {
		perms := make([]string, 0, len(role.Permissions))
		for _, p := range role.Permissions {
			perms = append(perms, p.Code)
		}

		result = append(result, map[string]interface{}{
			"id":          role.ID,
			"name":        role.Name,
			"code":        role.Code,
			"description": role.Description,
			"permissions": perms,
		})
	}
	return result, nil
}

// 更新用户角色
func (a *App) UpdateUserRole(userID int, roleID int) error {
	return a.AuthService.UpdateUserRole(userID, roleID)
}

// 更新用户资料
func (a *App) UpdateUserProfile(userID int, email string, token string) error {
	claims, err := a.AuthService.ValidateToken(token)
	if err != nil {
		return fmt.Errorf("验证 token 失败：%v", err)
	}
	return a.AuthService.UpdateUserProfile(userID, email, claims.UserID, claims.Username)
}

// 删除用户
func (a *App) DeleteUser(userID int, token string) error {
	claims, err := a.AuthService.ValidateToken(token)
	if err != nil {
		return fmt.Errorf("验证 token 失败：%v", err)
	}
	return a.AuthService.DeleteUser(userID, claims.UserID, claims.Username)
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

// ============================
// Zeek Runner
// ============================

// IsZeekEnabled 检查 Zeek 服务是否已启用
func (a *App) IsZeekEnabled() bool {
	return a.ZeekService != nil
}

// GetZeekVersion 获取 Zeek 版本（HTTP 接口）
func (a *App) GetZeekVersion() (string, error) {
	if a.ZeekService == nil {
		return "", fmt.Errorf("Zeek 服务未启用")
	}

	ctx := context.Background()
	version, err := a.ZeekService.GetZeekVersion(ctx)
	if err != nil {
		return "", fmt.Errorf("获取 Zeek 版本失败：%w", err)
	}

	return version, nil
}

// CheckZeekHealth 检查 Zeek 服务健康状态（HTTP 接口）
func (a *App) CheckZeekHealth() (map[string]interface{}, error) {
	if a.ZeekService == nil {
		return nil, fmt.Errorf("Zeek 服务未启用")
	}

	ctx := context.Background()
	health, err := a.ZeekService.CheckHealth(ctx)
	if err != nil {
		return nil, fmt.Errorf("健康检查失败：%w", err)
	}

	return map[string]interface{}{
		"status":       health.Status,
		"pool_running": health.PoolRunning,
		"is_healthy":   health.IsHealthy(),
		"message":      health.GetStatusMessage(),
	}, nil
}

// AnalyzePCAP 分析 PCAP 文件（gRPC 优先，失败降级到 HTTP）
func (a *App) AnalyzePCAP(req zeek.AnalyzePCAPRequest) (*zeek.AnalyzeResult, error) {
	if a.ZeekService == nil {
		return nil, fmt.Errorf("Zeek 服务未启用")
	}

	// 验证请求
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("请求验证失败：%w", err)
	}

	ctx := context.Background()
	result, err := a.ZeekService.AnalyzePCAP(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("分析失败：%w", err)
	}

	return result, nil
}

// GetZeekVersions 获取 Zeek 和 Zeek-Kafka 版本（HTTP 接口）
func (a *App) GetZeekVersions() (map[string]interface{}, error) {
	if a.ZeekService == nil {
		return map[string]interface{}{
			"zeek_version":       "N/A",
			"zeek_kafka_version": "N/A",
			"component_status":   "disabled",
			"message":            "Zeek 服务未启用",
		}, nil
	}

	ctx := context.Background()

	// 获取版本信息
	versions, err := a.ZeekService.GetVersions(ctx)
	if err != nil {
		return map[string]interface{}{
			"zeek_version":       "未知",
			"zeek_kafka_version": "未知",
			"component_status":   "error",
			"message":            fmt.Sprintf("获取版本失败：%v", err),
		}, nil
	}

	// 获取健康状态
	health, err := a.ZeekService.CheckHealth(ctx)
	if err != nil {
		return map[string]interface{}{
			"zeek_version":       versions.ZeekVersion,
			"zeek_kafka_version": versions.ZeekKafkaVersion,
			"component_status":   "error",
			"message":            fmt.Sprintf("健康检查失败：%v", err),
		}, nil
	}

	return map[string]interface{}{
		"zeek_version":       versions.ZeekVersion,
		"zeek_kafka_version": versions.ZeekKafkaVersion,
		"component_status":   health.Status,
		"pool_running":       health.PoolRunning,
		"is_healthy":         health.IsHealthy(),
		"message":            health.GetStatusMessage(),
	}, nil
}
