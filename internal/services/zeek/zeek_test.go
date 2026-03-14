package zeek

import (
	"context"
	"testing"
	"time"
)

// TestServiceConfig 测试配置
func TestServiceConfig(t *testing.T) {
	cfg := ServiceConfig{
		GRPCAddress: "localhost:50051",
		HTTPAddress: "http://localhost:8080",
		Timeout:     30 * time.Second,
	}

	if cfg.GRPCAddress == "" {
		t.Error("GRPCAddress should not be empty")
	}
	if cfg.HTTPAddress == "" {
		t.Error("HTTPAddress should not be empty")
	}
	if cfg.Timeout == 0 {
		t.Error("Timeout should not be zero")
	}
}

// TestAnalyzePCAPRequestValidate 测试请求验证
func TestAnalyzePCAPRequestValidate(t *testing.T) {
	tests := []struct {
		name    string
		req     AnalyzePCAPRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: AnalyzePCAPRequest{
				TaskID:   "task-001",
				UUID:     "user-uuid",
				PcapPath: "/data/test.pcap",
			},
			wantErr: false,
		},
		{
			name: "missing taskID",
			req: AnalyzePCAPRequest{
				TaskID:   "",
				UUID:     "user-uuid",
				PcapPath: "/data/test.pcap",
			},
			wantErr: true,
		},
		{
			name: "missing pcapPath",
			req: AnalyzePCAPRequest{
				TaskID:   "task-001",
				UUID:     "user-uuid",
				PcapPath: "",
			},
			wantErr: true,
		},
		{
			name: "missing uuid",
			req: AnalyzePCAPRequest{
				TaskID:   "task-001",
				UUID:     "",
				PcapPath: "/data/test.pcap",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestHealthStatus 测试健康状态
func TestHealthStatus(t *testing.T) {
	tests := []struct {
		name     string
		status   HealthStatus
		expected bool
		message  string
	}{
		{
			name: "healthy",
			status: HealthStatus{
				Status:      "ok",
				PoolRunning: 10,
			},
			expected: true,
			message:  "服务正常",
		},
		{
			name: "kafka down",
			status: HealthStatus{
				Status:      "kafka_down",
				PoolRunning: 0,
			},
			expected: false,
			message:  "Kafka 不可用",
		},
		{
			name: "unknown status",
			status: HealthStatus{
				Status:      "unknown",
				PoolRunning: 0,
			},
			expected: false,
			message:  "未知状态：unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.status.IsHealthy() != tt.expected {
				t.Errorf("IsHealthy() = %v, want %v", tt.status.IsHealthy(), tt.expected)
			}

			msg := tt.status.GetStatusMessage()
			if msg != tt.message {
				t.Errorf("GetStatusMessage() = %v, want %v", msg, tt.message)
			}
		})
	}
}

// TestAnalyzeResultString 测试字符串表示
func TestAnalyzeResultString(t *testing.T) {
	result := AnalyzeResult{
		TaskID:    "task-001",
		Status:    "running",
		StartTime: "2024-01-01 12:00:00",
	}

	expected := "TaskID: task-001, Status: running, StartTime: 2024-01-01 12:00:00"
	if result.String() != expected {
		t.Errorf("String() = %v, want %v", result.String(), expected)
	}
}

// TestHTTPClientCreation 测试 HTTP 客户端创建
func TestHTTPClientCreation(t *testing.T) {
	client := NewHTTPClient("http://localhost:8080", 30*time.Second)

	if client == nil {
		t.Fatal("HTTPClient should not be nil")
	}
	if client.baseURL != "http://localhost:8080" {
		t.Errorf("baseURL = %v, want %v", client.baseURL, "http://localhost:8080")
	}
	if client.userAgent != "Nebula/1.0" {
		t.Errorf("userAgent = %v, want %v", client.userAgent, "Nebula/1.0")
	}
	if client.httpClient.Timeout != 30*time.Second {
		t.Errorf("timeout = %v, want %v", client.httpClient.Timeout, 30*time.Second)
	}
}

// TestContextTimeout 测试上下文超时
func TestContextTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	<-ctx.Done()

	if ctx.Err() != context.DeadlineExceeded {
		t.Errorf("Expected deadline exceeded, got %v", ctx.Err())
	}
}
