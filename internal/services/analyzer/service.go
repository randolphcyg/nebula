package analyzer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

const (
	// wireshark 容器映射到宿主机的 API 地址
	baseURL = "http://127.0.0.1:18090/api/v1"
	// 容器内部的挂载路径
	containerMountPath = "/app/pcaps/"
)

type Service struct {
	client *http.Client
}

func NewService() *Service {
	return &Service{
		// 解析大文件可能较慢，设置 60 秒超时
		client: &http.Client{Timeout: 60 * time.Second},
	}
}

// ======================= 数据结构 (DTO) =======================

type baseRequest struct {
	Filepath  string `json:"filepath"`
	IsDebug   bool   `json:"isDebug"`
	IgnoreErr bool   `json:"ignoreErr"`
}

type getByPageRequest struct {
	baseRequest
	Page int `json:"page"`
	Size int `json:"size"`
}

type getByIdxsRequest struct {
	baseRequest
	FrameIdxs []int `json:"frameIdxs"`
}

// 统一响应结构
type apiResponse struct {
	Code  int             `json:"code"`
	Msg   string          `json:"msg"`
	Error string          `json:"error"`
	Data  json.RawMessage `json:"data"` // 保持 Data 为原生 JSON，方便直接透传给前端
}

type getHexReq struct {
	baseRequest
	FrameIdx int `json:"frameIdx"`
}

// ======================= 核心调用方法 =======================

// 获取 Wireshark 版本
func (s *Service) GetWiresharkVersion() (string, error) {
	respData, err := s.doGet(baseURL + "/version/wireshark")
	if err != nil {
		return "", err
	}
	return string(respData), nil
}

// 分页获取数据包
func (s *Service) GetPacketsByPage(filepath string, page int, size int) (string, error) {
	reqBody := getByPageRequest{
		baseRequest: baseRequest{
			Filepath:  toContainerPath(filepath), // 转换为容器内路径
			IsDebug:   false,
			IgnoreErr: true, // 容错处理
		},
		Page: page,
		Size: size,
	}

	respData, err := s.doPost(baseURL+"/frames/page", reqBody)
	if err != nil {
		return "", err
	}
	return string(respData), nil
}

// 获取所有数据包 (慎用，仅限小文件)
func (s *Service) GetAllFrames(filepath string) (string, error) {
	reqBody := baseRequest{
		Filepath:  toContainerPath(filepath),
		IsDebug:   false,
		IgnoreErr: true,
	}

	respData, err := s.doPost(baseURL+"/frames/all", reqBody)
	if err != nil {
		return "", err
	}
	return string(respData), nil
}

// 获取指定帧详情 (用于点击某一行时查看协议树和 Hex)
func (s *Service) GetPacketDetail(filepath string, index int) (string, error) {
	reqBody := getByIdxsRequest{
		baseRequest: baseRequest{
			Filepath:  toContainerPath(filepath),
			IsDebug:   false,
			IgnoreErr: true,
		},
		FrameIdxs: []int{index}, // 取单帧
	}

	respData, err := s.doPost(baseURL+"/frames/idxs", reqBody)
	if err != nil {
		return "", err
	}
	return string(respData), nil
}

func (s *Service) GetPacketHex(filepath string, index int) (string, error) {
	reqBody := getHexReq{
		baseRequest: baseRequest{
			Filepath:  toContainerPath(filepath),
			IsDebug:   false,
			IgnoreErr: true,
		},
		FrameIdx: index,
	}

	respData, err := s.doPost(baseURL+"/frames/hex", reqBody)
	if err != nil {
		return "", err
	}
	return string(respData), nil
}

// ======================= 内部辅助工具 =======================

// 路径转换工具：无论前端传绝对路径还是文件名，都转换为容器内绝对路径
func toContainerPath(fullPath string) string {
	// 1. 跨平台兼容：将 Windows 的反斜杠 \ 统一转换为 Linux 的 /
	slashPath := filepath.ToSlash(fullPath)

	// 2. 动态定位挂载锚点：查找 "/pcaps/" 在路径中的位置
	// 因为宿主机是 ~/.nebula/data/pcaps/，Docker 是 /app/pcaps/
	idx := strings.LastIndex(slashPath, "/pcaps/")
	if idx != -1 {
		// 提取出 "/pcaps/" 之后的所有相对路径 (例如 "2026-03/xxx.pcap" 或 "xxx.pcap")
		relativePath := slashPath[idx+len("/pcaps/"):]
		return containerMountPath + relativePath
	}

	// 3. 兜底方案：如果没匹配到 /pcaps/，则取最后两级（目录名/文件名）
	dir := filepath.Base(filepath.Dir(slashPath))
	base := filepath.Base(slashPath)
	if dir == "pcaps" || dir == "." || dir == "/" {
		return containerMountPath + base
	}

	return containerMountPath + dir + "/" + base
}

// 发送 GET 请求
func (s *Service) doGet(url string) ([]byte, error) {
	resp, err := s.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("请求 wireshark 失败: %v", err)
	}
	defer resp.Body.Close()
	return handleResponse(resp)
}

// 发送 POST 请求
func (s *Service) doPost(url string, body interface{}) ([]byte, error) {
	jsonBytes, _ := json.Marshal(body)
	resp, err := s.client.Post(url, "application/json", bytes.NewBuffer(jsonBytes))
	if err != nil {
		return nil, fmt.Errorf("请求 wireshark 失败: %v", err)
	}
	defer resp.Body.Close()
	return handleResponse(resp)
}

// 统一处理 HTTP 响应
func handleResponse(resp *http.Response) ([]byte, error) {
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	var apiRes apiResponse
	if err := json.Unmarshal(bodyBytes, &apiRes); err != nil {
		return nil, fmt.Errorf("解析 wireshark 响应失败: %v", err)
	}

	if apiRes.Code != 0 {
		return nil, fmt.Errorf("wireshark 错误 [%d]: %s (err: %s)", apiRes.Code, apiRes.Msg, apiRes.Error)
	}

	// 成功则直接返回 Data 部分的 JSON 字符串，Svelte 前端可以直接 JSON.parse()
	return apiRes.Data, nil
}
