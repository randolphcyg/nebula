package main

import (
	"context"
	"fmt"
	"nebula/internal/models"
	"nebula/internal/services/analyzer"
	"nebula/internal/services/pcap"
)

// App struct
type App struct {
	ctx context.Context

	// 预留模块接口
	AnalyzerService *analyzer.Service
	PcapService     *pcap.Service
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		AnalyzerService: analyzer.NewService(),
		PcapService:     pcap.NewService(),
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	// 启动时初始化数据库和工作区
	if err := a.PcapService.Start(ctx); err != nil {
		fmt.Printf("服务启动失败: %v\n", err)
	}
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
