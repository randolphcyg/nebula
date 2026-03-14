package zeek

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// HTTPClient HTTP 客户端
type HTTPClient struct {
	baseURL    string
	httpClient *http.Client
	userAgent  string
}

// NewHTTPClient 创建 HTTP 客户端
func NewHTTPClient(baseURL string, timeout time.Duration) *HTTPClient {
	return &HTTPClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		userAgent: "Nebula/1.0",
	}
}

// Healthz 健康检查
func (c *HTTPClient) Healthz(ctx context.Context) (*HealthResponse, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/api/v1/healthz", nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var health HealthResponse
	if err := json.Unmarshal(body, &health); err != nil {
		return nil, err
	}

	return &health, nil
}

// AnalyzeHTTP 使用 HTTP 接口进行分析
func (c *HTTPClient) AnalyzeHTTP(ctx context.Context, req *AnalyzeRequest) (*AnalyzeResponse, error) {
	reqBody := map[string]interface{}{
		"task_id":                 req.TaskID,
		"uuid":                    req.UUID,
		"only_notice":             req.OnlyNotice,
		"pcap_id":                 req.PcapID,
		"pcap_path":               req.PcapPath,
		"script_id":               req.ScriptID,
		"script_path":             req.ScriptPath,
		"extracted_file_path":     req.ExtractedFilePath,
		"extracted_file_min_size": req.ExtractedFileMinSize,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/v1/analyze", bytes.NewReader(jsonData))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", c.userAgent)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("HTTP analyze failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP analyze failed: status=%d, body=%s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var analyzeResp AnalyzeResponse
	if err := json.Unmarshal(body, &analyzeResp); err != nil {
		return nil, err
	}

	return &analyzeResp, nil
}

// VersionResponse 版本响应结构
type VersionResponse struct {
	Code int `json:"code"`
	Data struct {
		Output string `json:"output"`
	} `json:"data"`
	Msg string `json:"msg"`
}

// extractVersion 从输出字符串中提取纯版本号
func extractVersion(output string) string {
	output = strings.TrimSpace(output)

	// 对于 "zeek version 8.1.1" 格式，提取 "8.1.1"
	zeekVersionRe := regexp.MustCompile(`zeek\s+version\s+([\d.]+)`)
	if matches := zeekVersionRe.FindStringSubmatch(strings.ToLower(output)); len(matches) > 1 {
		return matches[1]
	}

	// 对于 "Seiso::Kafka - Writes logs to Kafka (dynamic, version 0.3.0)" 格式，提取 "0.3.0"
	kafkaVersionRe := regexp.MustCompile(`version\s+([\d.]+)`)
	if matches := kafkaVersionRe.FindStringSubmatch(strings.ToLower(output)); len(matches) > 1 {
		return matches[1]
	}

	// 如果已经是纯版本号，直接返回
	if regexp.MustCompile(`^[\d.]+$`).MatchString(output) {
		return output
	}

	// 默认返回原字符串
	return output
}

// GetZeekVersion 获取 Zeek 版本
func (c *HTTPClient) GetZeekVersion(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/api/v1/version/zeek", nil)
	if err != nil {
		return "", err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求失败：%w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP 状态码错误：%d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// 解析 JSON 响应
	var versionResp VersionResponse
	if err := json.Unmarshal(body, &versionResp); err != nil {
		return "", fmt.Errorf("解析 JSON 失败：%w", err)
	}

	if versionResp.Code != 0 {
		return "", fmt.Errorf("接口返回错误：%s", versionResp.Msg)
	}

	// 提取纯版本号
	return extractVersion(versionResp.Data.Output), nil
}

// GetZeekKafkaVersion 获取 Zeek-Kafka 版本
func (c *HTTPClient) GetZeekKafkaVersion(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/api/v1/version/zeek-kafka", nil)
	if err != nil {
		return "", err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求失败：%w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP 状态码错误：%d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// 解析 JSON 响应
	var versionResp VersionResponse
	if err := json.Unmarshal(body, &versionResp); err != nil {
		return "", fmt.Errorf("解析 JSON 失败：%w", err)
	}

	if versionResp.Code != 0 {
		return "", fmt.Errorf("接口返回错误：%s", versionResp.Msg)
	}

	// 提取纯版本号
	return extractVersion(versionResp.Data.Output), nil
}

// HealthResponse 健康检查响应
type HealthResponse struct {
	Status      string `json:"status"`
	PoolRunning int    `json:"pool_running"`
}
