package zeek

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// Service Zeek 分析服务
type Service struct {
	client *Client
	config ServiceConfig
}

// ServiceConfig 服务配置
type ServiceConfig struct {
	GRPCAddress string        `yaml:"grpc_address"`
	HTTPAddress string        `yaml:"http_address"`
	Timeout     time.Duration `yaml:"timeout"`
}

// NewService 创建服务
func NewService(cfg ServiceConfig) (*Service, error) {
	client, err := NewClient(Config{
		GRPCAddress: cfg.GRPCAddress,
		HTTPAddress: cfg.HTTPAddress,
		Timeout:     cfg.Timeout,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create zeek client: %w", err)
	}

	return &Service{
		client: client,
		config: cfg,
	}, nil
}

// AnalyzePCAP 分析 PCAP 文件
func (s *Service) AnalyzePCAP(ctx context.Context, req AnalyzePCAPRequest) (*AnalyzeResult, error) {
	grpcReq := &AnalyzeRequest{
		TaskID:               req.TaskID,
		UUID:                 req.UUID,
		OnlyNotice:           req.OnlyNotice,
		PcapID:               req.PcapID,
		PcapPath:             req.PcapPath,
		ScriptID:             req.ScriptID,
		ScriptPath:           req.ScriptPath,
		ExtractedFilePath:    req.ExtractedFilePath,
		ExtractedFileMinSize: req.ExtractedFileMinSize,
	}

	// 优先使用 gRPC 接口
	resp, err := s.client.Analyze(ctx, grpcReq)
	if err != nil {
		// gRPC 失败时降级到 HTTP
		fmt.Printf("Warning: gRPC failed, fallback to HTTP: %v\n", err)
		return s.analyzeHTTP(ctx, req)
	}

	return &AnalyzeResult{
		TaskID:    resp.TaskID,
		Status:    "running",
		StartTime: resp.StartTime,
	}, nil
}

// analyzeHTTP HTTP 降级方案
func (s *Service) analyzeHTTP(ctx context.Context, req AnalyzePCAPRequest) (*AnalyzeResult, error) {
	httpReq := &AnalyzeRequest{
		TaskID:               req.TaskID,
		UUID:                 req.UUID,
		OnlyNotice:           req.OnlyNotice,
		PcapID:               req.PcapID,
		PcapPath:             req.PcapPath,
		ScriptID:             req.ScriptID,
		ScriptPath:           req.ScriptPath,
		ExtractedFilePath:    req.ExtractedFilePath,
		ExtractedFileMinSize: req.ExtractedFileMinSize,
	}

	resp, err := s.client.HTTPClient().AnalyzeHTTP(ctx, httpReq)
	if err != nil {
		return nil, fmt.Errorf("HTTP analyze failed: %w", err)
	}

	return &AnalyzeResult{
		TaskID:    resp.TaskID,
		Status:    "running",
		StartTime: resp.StartTime,
	}, nil
}

// CheckHealth 检查服务健康状态
func (s *Service) CheckHealth(ctx context.Context) (*HealthStatus, error) {
	health, err := s.client.HTTPClient().Healthz(ctx)
	if err != nil {
		return nil, fmt.Errorf("health check failed: %w", err)
	}

	return &HealthStatus{
		Status:      health.Status,
		PoolRunning: health.PoolRunning,
	}, nil
}

// GetZeekVersion 获取 Zeek 版本
func (s *Service) GetZeekVersion(ctx context.Context) (string, error) {
	return s.client.HTTPClient().GetZeekVersion(ctx)
}

// GetZeekKafkaVersion 获取 Zeek-Kafka 版本
func (s *Service) GetZeekKafkaVersion(ctx context.Context) (string, error) {
	return s.client.HTTPClient().GetZeekKafkaVersion(ctx)
}

// GetVersions 获取所有版本信息
func (s *Service) GetVersions(ctx context.Context) (*VersionInfo, error) {
	zeekVersion, err := s.GetZeekVersion(ctx)
	if err != nil {
		zeekVersion = "未知"
	}

	zeekKafkaVersion, err := s.GetZeekKafkaVersion(ctx)
	if err != nil {
		zeekKafkaVersion = "未知"
	}

	return &VersionInfo{
		ZeekVersion:      zeekVersion,
		ZeekKafkaVersion: zeekKafkaVersion,
	}, nil
}

// Close 关闭服务
func (s *Service) Close() error {
	return s.client.Close()
}

// AnalyzePCAPRequest 分析请求
type AnalyzePCAPRequest struct {
	TaskID               string
	UUID                 string
	OnlyNotice           bool
	PcapID               string
	PcapPath             string
	ScriptID             string
	ScriptPath           string
	ExtractedFilePath    string
	ExtractedFileMinSize int32
}

// AnalyzeResult 分析结果
type AnalyzeResult struct {
	TaskID    string
	Status    string
	StartTime string
}

// HealthStatus 健康状态
type HealthStatus struct {
	Status      string
	PoolRunning int
}

// VersionInfo 版本信息
type VersionInfo struct {
	ZeekVersion      string `json:"zeek_version"`
	ZeekKafkaVersion string `json:"zeek_kafka_version"`
}

// IsHealthy 检查服务是否健康
func (h *HealthStatus) IsHealthy() bool {
	return h.Status == "ok"
}

// GetStatusMessage 获取状态消息
func (h *HealthStatus) GetStatusMessage() string {
	switch h.Status {
	case "ok":
		return "服务正常"
	case "kafka_down":
		return "Kafka 不可用"
	default:
		return fmt.Sprintf("未知状态：%s", h.Status)
	}
}

// String 实现 Stringer 接口
func (r *AnalyzeResult) String() string {
	return fmt.Sprintf("TaskID: %s, Status: %s, StartTime: %s", r.TaskID, r.Status, r.StartTime)
}

// Validate 验证请求
func (r *AnalyzePCAPRequest) Validate() error {
	if r.TaskID == "" {
		return errors.New("taskID 不能为空")
	}
	if r.PcapPath == "" {
		return errors.New("pcapPath 不能为空")
	}
	if r.UUID == "" {
		return errors.New("uuid 不能为空")
	}
	return nil
}
