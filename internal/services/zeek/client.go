package zeek

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"nebula/internal/services/zeek/pb"
)

// Client Zeek 服务客户端
type Client struct {
	grpcClient pb.ZeekAnalysisServiceClient
	httpClient *HTTPClient
	conn       *grpc.ClientConn
}

// Config 客户端配置
type Config struct {
	GRPCAddress string
	HTTPAddress string
	Timeout     time.Duration
}

// NewClient 创建新客户端
func NewClient(cfg Config) (*Client, error) {
	// 1. 创建 gRPC 连接
	conn, err := grpc.NewClient(
		cfg.GRPCAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.WaitForReady(true)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC connection: %w", err)
	}

	// 2. 创建 HTTP 客户端
	httpClient := NewHTTPClient(cfg.HTTPAddress, cfg.Timeout)

	return &Client{
		grpcClient: pb.NewZeekAnalysisServiceClient(conn),
		httpClient: httpClient,
		conn:       conn,
	}, nil
}

// Close 关闭连接
func (c *Client) Close() error {
	return c.conn.Close()
}

// Analyze 调用 gRPC 接口进行分析
func (c *Client) Analyze(ctx context.Context, req *AnalyzeRequest) (*AnalyzeResponse, error) {
	// 转换为 proto 请求
	protoReq := &pb.AnalyzeRequest{
		TaskID:               req.TaskID,
		Uuid:                 req.UUID,
		OnlyNotice:           req.OnlyNotice,
		PcapID:               req.PcapID,
		PcapPath:             req.PcapPath,
		ScriptID:             req.ScriptID,
		ScriptPath:           req.ScriptPath,
		ExtractedFilePath:    req.ExtractedFilePath,
		ExtractedFileMinSize: req.ExtractedFileMinSize,
	}

	// 设置超时
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	// 调用 gRPC
	resp, err := c.grpcClient.Analyze(ctx, protoReq)
	if err != nil {
		return nil, fmt.Errorf("gRPC analyze failed: %w", err)
	}

	// 转换为内部响应
	return &AnalyzeResponse{
		TaskID:     resp.TaskID,
		UUID:       resp.Uuid,
		PcapPath:   resp.PcapPath,
		ScriptPath: resp.ScriptPath,
		StartTime:  resp.StartTime,
	}, nil
}

// HTTPClient 返回 HTTP 客户端
func (c *Client) HTTPClient() *HTTPClient {
	return c.httpClient
}

// AnalyzeRequest 分析请求
type AnalyzeRequest struct {
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

// AnalyzeResponse 分析响应
type AnalyzeResponse struct {
	TaskID     string
	UUID       string
	PcapPath   string
	ScriptPath string
	StartTime  string
}
